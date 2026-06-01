package main

import (
	"encoding/json"
	"net/http"
	"os/exec"
	"strings"

	"github.com/hashicorp/go-hclog"
	"github.com/martezr/go-openvswitch/ovs"
	"github.com/martezr/nightlight-cloud/utils"
)

// BridgeInterface describes a single port/interface on an OVS bridge.
type BridgeInterface struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	MacAddress string `json:"macAddress,omitempty"`
}

// OVSBridge is the API representation of an OVS bridge and its interfaces.
type OVSBridge struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Interfaces []BridgeInterface `json:"interfaces"`
}

// interfaceType returns the OVS type of a named interface ("system",
// "internal", "patch", "vxlan", …).  An empty string means system/physical.
func interfaceType(name string) string {
	out, err := exec.Command("ovs-vsctl", "get", "Interface", name, "type").Output()
	if err != nil {
		return "system"
	}
	t := strings.TrimSpace(string(out))
	// OVS returns an empty string for system/physical interfaces.
	if t == "" || t == "[]" {
		return "system"
	}
	return t
}

// ListBridges queries OVS live and returns every bridge with its ports.
func ListBridges(w http.ResponseWriter, r *http.Request) {
	client := ovs.New()

	bridgeNames, err := client.VSwitch.ListBridges()
	if err != nil {
		hclog.Default().Named("core").Error("list bridges: " + err.Error())
		json.NewEncoder(w).Encode(utils.NilSliceToEmptySlice([]OVSBridge{}))
		return
	}

	bridges := make([]OVSBridge, 0, len(bridgeNames))
	for _, name := range bridgeNames {
		ports, err := client.VSwitch.ListPorts(name)
		if err != nil {
			hclog.Default().Named("core").Error("list ports for " + name + ": " + err.Error())
		}

		ifaces := make([]BridgeInterface, 0, len(ports))
		for _, port := range ports {
			ifaces = append(ifaces, BridgeInterface{
				Name: port,
				Type: interfaceType(port),
			})
		}

		bridges = append(bridges, OVSBridge{
			ID:         name,
			Name:       name,
			Interfaces: ifaces,
		})
	}

	json.NewEncoder(w).Encode(utils.NilSliceToEmptySlice(bridges))
}
