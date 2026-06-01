package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"

	"github.com/go-chi/chi"
	"github.com/hashicorp/go-hclog"
	"github.com/martezr/go-openvswitch/ovs"
	"github.com/martezr/nightlight-cloud/utils"
)

// Switch represents a network switch backed by an OVS bridge.
type Switch struct {
	ID          string   `json:"id"          storm:"id,index"`
	Name        string   `json:"name"        storm:"index"`
	Description string   `json:"description"`
	SiteId      string   `json:"siteId"      storm:"index"`
	BridgeName  string   `json:"bridgeName"  storm:"index"`
	Type        string   `json:"type"`
	Tags        []string `json:"tags"`
}

// ListSwitches returns all switches.
func ListSwitches(w http.ResponseWriter, r *http.Request) {
	var switches []Switch
	if err := db.All(&switches); err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}
	json.NewEncoder(w).Encode(utils.NilSliceToEmptySlice(switches))
}

// provisionSwitch creates a Switch record and its backing OVS bridge.
// bridgeName defaults to the generated ID when empty.
// It is the shared implementation used by both the HTTP handler and
// internal callers such as site topology setup.
func provisionSwitch(name, description, bridgeName, siteId, switchType string, tags []string) (Switch, error) {
	sw := Switch{
		ID:          "switch-" + utils.IDGenerator(3),
		Name:        name,
		Description: description,
		SiteId:      siteId,
		Type:        switchType,
		Tags:        tags,
	}
	if bridgeName == "" {
		sw.BridgeName = sw.ID
	} else {
		sw.BridgeName = bridgeName
	}
	ovsClient := ovs.New()
	if err := ovsClient.VSwitch.AddBridge(sw.BridgeName); err != nil {
		return Switch{}, err
	}

	if err := ovsClient.VSwitch.Set.Bridge(sw.BridgeName, ovs.BridgeOptions{Protocols: []string{"OpenFlow13"}}); err != nil {
		return Switch{}, err
	}
	// Configure openflow controller for the bridge (used by the metadata agent to push DHCP flows)
	if err := ovsClient.VSwitch.SetController(sw.BridgeName, "tcp:127.0.0.1:6653"); err != nil {
		return Switch{}, err
	}
	// Save the switch record to the database
	if err := db.Save(&sw); err != nil {
		return Switch{}, err
	}

	// Create the iPXE agent namespace and OpenRC service for this bridge.
	if err := setupIPXENamespace(sw.BridgeName); err != nil {
		hclog.Default().Named("core").Warn("iPXE namespace setup failed", "bridge", sw.BridgeName, "error", err)
	}

	return sw, nil
}

// ipxeVethName returns the host-side veth name for a bridge's iPXE namespace,
// truncated to the 15-character Linux interface name limit.
func ipxeVethName(bridgeName string) string {
	name := "ipxe-" + bridgeName
	if len(name) > 15 {
		name = name[:15]
	}
	return name
}

// ipxePeerName returns a temporary host-side name for the namespace end of the
// veth pair. It must be distinct from ipxeVethName and ≤15 characters.
// The peer is renamed to eth0 after it is moved into the namespace.
func ipxePeerName(bridgeName string) string {
	name := "ipxeP-" + bridgeName
	if len(name) > 15 {
		name = name[:15]
	}
	return name
}

// setupIPXENamespace creates a network namespace for the iPXE agent, wires it
// to the OVS bridge via a veth pair, and installs an OpenRC service that keeps
// the agent running across reboots.
//
// Namespace layout:
//
//	host: <hostVeth> ──── OVS bridge <bridgeName>
//	namespace ipxe-<bridgeName>: eth0 @ 169.254.169.253/16
func setupIPXENamespace(bridgeName string) error {
	nsName := "ipxe-" + bridgeName
	hostVeth := ipxeVethName(bridgeName)
	nsPeer := ipxePeerName(bridgeName)
	svcName := "ipxeagent-" + bridgeName

	if err := exec.Command("ip", "netns", "add", nsName).Run(); err != nil {
		return fmt.Errorf("create namespace %s: %w", nsName, err)
	}

	// Create veth pair with a unique host-side peer name to avoid collisions
	// with existing interfaces (e.g. eth0). The peer is renamed to eth0 inside
	// the namespace after it is moved there.
	if err := exec.Command("ip", "link", "add", hostVeth, "type", "veth", "peer", "name", nsPeer).Run(); err != nil {
		return fmt.Errorf("create veth pair: %w", err)
	}

	if err := exec.Command("ip", "link", "set", nsPeer, "netns", nsName).Run(); err != nil {
		return fmt.Errorf("move %s into namespace %s: %w", nsPeer, nsName, err)
	}

	if err := exec.Command("ip", "link", "set", hostVeth, "up").Run(); err != nil {
		return fmt.Errorf("bring up %s: %w", hostVeth, err)
	}

	ovsClient := ovs.New()
	if err := ovsClient.VSwitch.AddPort(bridgeName, hostVeth); err != nil {
		return fmt.Errorf("add %s to bridge %s: %w", hostVeth, bridgeName, err)
	}

	nsExec := func(args ...string) error {
		return exec.Command("ip", append([]string{"netns", "exec", nsName}, args...)...).Run()
	}
	// Rename the peer to eth0 now that it is safely inside the namespace.
	if err := nsExec("ip", "link", "set", nsPeer, "name", "eth0"); err != nil {
		return fmt.Errorf("rename %s to eth0 in namespace: %w", nsPeer, err)
	}
	if err := nsExec("ip", "link", "set", "lo", "up"); err != nil {
		return fmt.Errorf("bring up namespace loopback: %w", err)
	}
	if err := nsExec("ip", "addr", "add", "169.254.169.253/16", "dev", "eth0"); err != nil {
		return fmt.Errorf("assign 169.254.169.253 in namespace: %w", err)
	}
	if err := nsExec("ip", "link", "set", "eth0", "up"); err != nil {
		return fmt.Errorf("bring up namespace eth0: %w", err)
	}

	if err := writeIPXEAgentService(bridgeName, nsName, svcName); err != nil {
		return fmt.Errorf("write OpenRC service: %w", err)
	}
	if err := exec.Command("rc-update", "add", svcName, "default").Run(); err != nil {
		return fmt.Errorf("enable service %s: %w", svcName, err)
	}
	if err := exec.Command("rc-service", svcName, "start").Run(); err != nil {
		return fmt.Errorf("start service %s: %w", svcName, err)
	}

	hclog.Default().Named("core").Info("iPXE namespace ready", "namespace", nsName, "veth", hostVeth, "service", svcName)
	return nil
}

// teardownIPXENamespace stops and removes the iPXE agent service and the
// network namespace created for bridgeName. Errors are logged but not returned
// because this runs during bridge deletion where partial cleanup is acceptable.
func teardownIPXENamespace(bridgeName string) {
	nsName := "ipxe-" + bridgeName
	hostVeth := ipxeVethName(bridgeName)
	svcName := "ipxeagent-" + bridgeName

	exec.Command("rc-service", svcName, "stop").Run()
	exec.Command("rc-update", "del", svcName, "default").Run()
	os.Remove("/etc/init.d/" + svcName)
	// Deleting one veth end removes the peer inside the namespace automatically.
	exec.Command("ip", "link", "delete", hostVeth).Run()
	exec.Command("ip", "netns", "delete", nsName).Run()

	hclog.Default().Named("core").Info("iPXE namespace removed", "namespace", nsName)
}

// writeIPXEAgentService writes an OpenRC init script for the iPXE agent to
// /etc/init.d/<svcName>. The script runs ipxeagent inside nsName and manages
// it with start-stop-daemon so OpenRC can track the PID correctly.
func writeIPXEAgentService(bridgeName, nsName, svcName string) error {
	script := fmt.Sprintf(`#!/sbin/openrc-run

name="%s"
description="iPXE agent for OVS bridge %s"

pidfile="/run/${RC_SVCNAME}.pid"

depend() {
    need net
    after nightlight
}

start() {
    ebegin "Starting ${RC_SVCNAME}"
    ip netns exec %s \
        start-stop-daemon --start \
            --background \
            --make-pidfile \
            --pidfile "${pidfile}" \
            --exec /usr/local/bin/ipxeagent \
            -- \
            -listen 169.254.169.253:8170 \
            -socket /opt/nightlight/ipxe.sock \
            -log /var/log/${RC_SVCNAME}.log
    eend $?
}

stop() {
    ebegin "Stopping ${RC_SVCNAME}"
    start-stop-daemon --stop --pidfile "${pidfile}"
    eend $?
}
`, svcName, bridgeName, nsName)

	path := "/etc/init.d/" + svcName
	return os.WriteFile(path, []byte(script), 0755)
}

// CreateSwitch is the HTTP handler for POST /api/v1/switches.
func CreateSwitch(w http.ResponseWriter, r *http.Request) {
	var input Switch
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	sw, err := provisionSwitch(input.Name, input.Description, input.BridgeName, input.SiteId, input.Type, input.Tags)
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
		http.Error(w, "failed to create switch", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(utils.NilSliceToEmptySlice(sw))
}

// GetSwitch returns a switch by ID.
func GetSwitch(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var sw Switch
	if err := db.One("ID", id, &sw); err != nil {
		hclog.Default().Named("core").Error(err.Error())
		http.Error(w, "switch not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(utils.NilSliceToEmptySlice(sw))
}

// UpdateSwitch updates the metadata of an existing switch.
func UpdateSwitch(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var sw Switch
	if err := db.One("ID", id, &sw); err != nil {
		hclog.Default().Named("core").Error(err.Error())
		http.Error(w, "switch not found", http.StatusNotFound)
		return
	}
	var data Switch
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	data.ID = sw.ID
	data.BridgeName = sw.BridgeName
	if err := db.Update(&data); err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}
	json.NewEncoder(w).Encode(utils.NilSliceToEmptySlice(data))
}

// DeleteSwitch removes a switch and deletes its backing OVS bridge.
func DeleteSwitch(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var sw Switch
	if err := db.One("ID", id, &sw); err != nil {
		hclog.Default().Named("core").Error(err.Error())
		http.Error(w, "switch not found", http.StatusNotFound)
		return
	}
	teardownIPXENamespace(sw.BridgeName)

	ovsClient := ovs.New()
	if err := ovsClient.VSwitch.DeleteBridge(sw.BridgeName); err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}
	if err := db.DeleteStruct(&sw); err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}
}

// FindSwitchByID returns a switch by ID without writing an HTTP response.
func FindSwitchByID(id string) (sw Switch) {
	if err := db.One("ID", id, &sw); err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}
	return sw
}

// ListSwitchesBySite returns all switches belonging to a given site.
func ListSwitchesBySite(w http.ResponseWriter, r *http.Request) {
	siteId := chi.URLParam(r, "id")
	var switches []Switch
	if err := db.Find("SiteId", siteId, &switches); err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}
	json.NewEncoder(w).Encode(utils.NilSliceToEmptySlice(switches))
}
