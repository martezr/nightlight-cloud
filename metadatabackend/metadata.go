package metadatabackend

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/asdine/storm/v3"
	"github.com/go-chi/chi/v5"
	"github.com/martezr/nightlight-cloud/utils"
)

// listenAddr is the TCP address the metadata server binds to.  It is
// intentionally loopback-only so VMs cannot reach it directly — traffic is
// intercepted and proxied by the SDN controller.
const listenAddr = "127.0.0.1:8169"

var localdb *storm.DB

// subnet is a minimal mirror of the main-package Subnet struct.  Storm uses
// the struct type name as the bucket key, so naming it "Subnet" here ensures
// reads hit the same BoltDB bucket as the main package.
type Subnet struct {
	ID         string `storm:"id"`
	CIDRBlock  string
	Gateway    string
	DNSServers []string
	DomainName string
	NTPServers []string
}

// SwitchPortMapping caches the resolved instance ID for a given OVS switch
// port, keyed by "<hex-dpid>:<port>" as set in X-Forwarded-For by the SDN
// controller.  The mapping is self-populating on first request via the
// X-Instance-Mac header.
type SwitchPortMapping struct {
	DpidPort   string `storm:"id"`
	InstanceID string `storm:"index"`
}

// ClearInstanceMappings removes all SwitchPortMapping cache entries for the
// given instance ID.  Call this when an instance is deleted so that the next
// VM assigned to the same switch port is not mis-identified.
func ClearInstanceMappings(db *storm.DB, instanceID string) {
	var mappings []SwitchPortMapping
	if err := db.Find("InstanceID", instanceID, &mappings); err != nil {
		return
	}
	for i := range mappings {
		_ = db.DeleteStruct(&mappings[i])
	}
}

// StartMetadataServer starts an AWS-compatible IMDSv1 HTTP server on
// 127.0.0.1:8169 and blocks until ctx is cancelled.
func StartMetadataServer(ctx context.Context, db *storm.DB) error {
	localdb = db

	r := chi.NewRouter()

	// Root: list available API versions.
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("latest\n"))
	})

	// IMDSv2 token endpoint — returns a dummy token; auth is not enforced.
	r.Put("/{version}/api/token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("nightlight-imds-token"))
	})

	// Meta-data directory listing.
	r.Get("/{version}/meta-data/", metaDataIndexHandler)

	// Scalar metadata fields.
	for _, field := range []string{
		"ami-id", "ami-launch-index", "ami-manifest-path",
		"hostname", "instance-id", "instance-life-cycle", "instance-type",
		"local-hostname", "local-ipv4", "mac",
		"profile", "public-hostname", "public-ipv4",
		"reservation-id", "security-groups",
	} {
		f := field // capture
		r.Get("/{version}/meta-data/"+f, metaDataFieldHandler(f))
	}

	// Placement sub-tree.
	r.Get("/{version}/meta-data/placement/", func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("availability-zone\nregion\n"))
	})
	r.Get("/{version}/meta-data/placement/region", metaDataFieldHandler("placement/region"))
	r.Get("/{version}/meta-data/placement/availability-zone", metaDataFieldHandler("placement/availability-zone"))

	// Network interfaces sub-tree.
	r.Get("/{version}/meta-data/network/interfaces/macs/", networkMACsHandler)
	r.Get("/{version}/meta-data/network/interfaces/macs/{mac}/", networkMACIndexHandler)
	for _, field := range []string{
		"device-number", "gateway-ipv4", "local-ipv4s",
		"subnet-id", "subnet-ipv4-cidr-block",
		"vpc-id", "vpc-ipv4-cidr-block",
	} {
		f := field
		r.Get("/{version}/meta-data/network/interfaces/macs/{mac}/"+f, networkMACFieldHandler(f))
	}

	// Instance tags sub-tree.
	r.Get("/{version}/meta-data/tags/instance/", tagsIndexHandler)
	r.Get("/{version}/meta-data/tags/instance/{key}", tagValueHandler)

	// User-data (404 when not set).
	r.Get("/{version}/user-data", userDataHandler)

	// DHCP payload lookup by MAC — used by the SDN controller.
	r.Get("/dhcp/{mac}", dhcpByMACHandler)

	// iPXE script endpoint — served to PXE-booting VMs via the SDN controller.
	r.Get("/ipxe/{mac}", ipxeScriptHandler)

	srv := &http.Server{
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return err
	}
	log.Printf("metadata server listening on %s", listenAddr)

	errCh := make(chan error, 1)
	go func() {
		if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("metadata server shutdown error: %v", err)
		}
		return nil
	case err := <-errCh:
		return err
	}
}

// clientInstance resolves the Instance record for the requesting VM.
//
// Lookup order:
//  1. X-Forwarded-For contains a "<hex-dpid>:<port>" identifier set by the SDN
//     controller → look up cached SwitchPortMapping in the DB (O(1)).
//  2. Cache miss: X-Instance-Mac contains the VM's Ethernet address → scan
//     instances by MAC, persist the mapping for future requests.
//  3. Fallback: treat X-Forwarded-For as a plain IP and look up by
//     MetadataIPAddress (legacy / direct-curl path).
func clientInstance(r *http.Request) (*utils.Instance, error) {
	xff := strings.TrimSpace(strings.SplitN(r.Header.Get("X-Forwarded-For"), ",", 2)[0])

	// Step 1 — dpid:port cache hit.
	if xff != "" && !isPlainIP(xff) {
		var mapping SwitchPortMapping
		if err := localdb.One("DpidPort", xff, &mapping); err == nil {
			var inst utils.Instance
			if err := localdb.One("ID", mapping.InstanceID, &inst); err == nil {
				return &inst, nil
			}
		}

		// Step 2 — resolve via MAC and register the mapping.
		if mac := r.Header.Get("X-Instance-Mac"); mac != "" {
			if inst := instanceByMAC(mac); inst != nil {
				_ = localdb.Save(&SwitchPortMapping{DpidPort: xff, InstanceID: inst.ID})
				log.Printf("metadata: registered switch-port mapping %s → %s", xff, inst.ID)
				return inst, nil
			}
		}

		return nil, fmt.Errorf("no instance for switch port %s", xff)
	}

	// Step 3 — legacy IP-based lookup (direct curl, metadataagent, etc.).
	ip := xff
	if ip == "" {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		} else {
			ip = host
		}
	}
	var inst utils.Instance
	if err := localdb.One("MetadataIPAddress", ip, &inst); err != nil {
		return nil, fmt.Errorf("no instance for IP %s: %v", ip, err)
	}
	return &inst, nil
}

// isPlainIP reports whether s is a bare IP address (not a dpid:port string).
func isPlainIP(s string) bool {
	return net.ParseIP(s) != nil
}

// instanceByMAC scans the instance bucket for a record whose PrimaryMacAddress
// or any NIC MacAddress matches mac.
func instanceByMAC(mac string) *utils.Instance {
	var instances []utils.Instance
	if err := localdb.All(&instances); err != nil {
		return nil
	}
	for i := range instances {
		if strings.EqualFold(instances[i].PrimaryMacAddress, mac) {
			return &instances[i]
		}
		for _, nic := range instances[i].Devices.NetworkInterfaces {
			if strings.EqualFold(nic.MacAddress, mac) {
				return &instances[i]
			}
		}
	}
	return nil
}

// instanceField maps an IMDS field name to the corresponding Instance value.
func instanceField(inst *utils.Instance, field string) string {
	switch field {
	case "instance-id":
		return inst.ID
	case "hostname", "local-hostname", "public-hostname":
		return inst.Name
	case "local-ipv4", "public-ipv4":
		return inst.PrimaryIPAddress
	case "mac":
		return inst.PrimaryMacAddress
	case "instance-type":
		if inst.InstanceType != "" {
			return inst.InstanceType
		}
		return "nightlight.custom"
	case "ami-id":
		if inst.ImageId != "" {
			return inst.ImageId
		}
		return "ami-nightlight"
	case "ami-launch-index":
		return "0"
	case "ami-manifest-path":
		return "(unknown)"
	case "reservation-id":
		return "r-" + inst.ID
	case "security-groups":
		return inst.Name
	case "profile":
		return "default-hvm"
	case "instance-life-cycle":
		return "on-demand"
	case "placement/region", "placement/availability-zone":
		if inst.SiteId != "" {
			return inst.SiteId
		}
		return "nightlight"
	}
	return ""
}

func metaDataIndexHandler(w http.ResponseWriter, _ *http.Request) {
	keys := strings.Join([]string{
		"ami-id",
		"ami-launch-index",
		"ami-manifest-path",
		"hostname",
		"instance-id",
		"instance-life-cycle",
		"instance-type",
		"local-hostname",
		"local-ipv4",
		"mac",
		"network/",
		"placement/",
		"profile",
		"public-hostname",
		"public-ipv4",
		"reservation-id",
		"security-groups",
		"tags/",
	}, "\n")
	w.Write([]byte(keys))
}

func metaDataFieldHandler(field string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		inst, err := clientInstance(r)
		if err != nil {
			log.Printf("metadata: %v", err)
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Write([]byte(instanceField(inst, field)))
	}
}

// nicForMAC finds the NetworkInterface matching mac (case-insensitive).
func nicForMAC(inst *utils.Instance, mac string) (*utils.NetworkInterface, int) {
	for i := range inst.Devices.NetworkInterfaces {
		if strings.EqualFold(inst.Devices.NetworkInterfaces[i].MacAddress, mac) {
			return &inst.Devices.NetworkInterfaces[i], i
		}
	}
	return nil, -1
}

func networkMACsHandler(w http.ResponseWriter, r *http.Request) {
	inst, err := clientInstance(r)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	var macs []string
	for _, nic := range inst.Devices.NetworkInterfaces {
		if nic.MacAddress != "" {
			macs = append(macs, nic.MacAddress+"/")
		}
	}
	w.Write([]byte(strings.Join(macs, "\n")))
}

func networkMACIndexHandler(w http.ResponseWriter, r *http.Request) {
	inst, err := clientInstance(r)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if _, idx := nicForMAC(inst, chi.URLParam(r, "mac")); idx < 0 {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	keys := strings.Join([]string{
		"device-number",
		"gateway-ipv4",
		"local-ipv4s",
		"subnet-id",
		"subnet-ipv4-cidr-block",
		"vpc-id",
		"vpc-ipv4-cidr-block",
	}, "\n")
	w.Write([]byte(keys))
}

func networkMACFieldHandler(field string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		inst, err := clientInstance(r)
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		nic, idx := nicForMAC(inst, chi.URLParam(r, "mac"))
		if nic == nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		switch field {
		case "device-number":
			w.Write([]byte(fmt.Sprintf("%d", idx)))
		case "local-ipv4s":
			var mapping utils.InstanceIPMapping
			if err := localdb.One("MacAddress", nic.MacAddress, &mapping); err == nil {
				w.Write([]byte(mapping.IPAddress))
			}
		case "subnet-id":
			w.Write([]byte(nic.SubnetId))
		case "vpc-id":
			w.Write([]byte(nic.SubnetId))
		case "subnet-ipv4-cidr-block", "vpc-ipv4-cidr-block":
			if nic.SubnetId == "" {
				return
			}
			var sn Subnet
			if err := localdb.One("ID", nic.SubnetId, &sn); err == nil {
				w.Write([]byte(sn.CIDRBlock))
			}
		case "gateway-ipv4":
			if nic.SubnetId == "" {
				return
			}
			var sn Subnet
			if err := localdb.One("ID", nic.SubnetId, &sn); err == nil {
				w.Write([]byte(sn.Gateway))
			}
		}
	}
}

func tagsIndexHandler(w http.ResponseWriter, r *http.Request) {
	inst, err := clientInstance(r)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	var keys []string
	for _, tag := range inst.Tags {
		for k := range tag {
			keys = append(keys, k)
		}
	}
	w.Write([]byte(strings.Join(keys, "\n")))
}

func tagValueHandler(w http.ResponseWriter, r *http.Request) {
	inst, err := clientInstance(r)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	key := chi.URLParam(r, "key")
	for _, tag := range inst.Tags {
		if val, ok := tag[key]; ok {
			w.Write([]byte(fmt.Sprintf("%v", val)))
			return
		}
	}
	http.Error(w, "not found", http.StatusNotFound)
}

func userDataHandler(w http.ResponseWriter, r *http.Request) {
	inst, err := clientInstance(r)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if inst.UserData == "" {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Write([]byte(inst.UserData))
}

// dhcpPayload is the JSON response for the DHCP lookup endpoint.
type dhcpPayload struct {
	IPAddress  string   `json:"ipAddress"`
	Netmask    string   `json:"netmask"`
	Gateway    string   `json:"gateway"`
	DNSServers []string `json:"dnsServers"`
	DomainName string   `json:"domainName"`
	NTPServers []string `json:"ntpServers"`
}

// dhcpByMACHandler handles GET /dhcp/{mac} and returns the DHCP assignment
// for the instance whose NIC matches the given MAC address.
func dhcpByMACHandler(w http.ResponseWriter, r *http.Request) {
	mac := chi.URLParam(r, "mac")
	// Restore colons if the caller sent a 12-hex-char string without them.
	if len(mac) == 12 && !strings.Contains(mac, ":") {
		mac = strings.Join([]string{mac[0:2], mac[2:4], mac[4:6], mac[6:8], mac[8:10], mac[10:12]}, ":")
	}

	inst := instanceByMAC(mac)
	if inst == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(dhcpPayload{})
		return
	}

	// Find the NIC that matched and pull its subnet.
	var subnetID string
	for _, nic := range inst.Devices.NetworkInterfaces {
		if strings.EqualFold(nic.MacAddress, mac) {
			subnetID = nic.SubnetId
			break
		}
	}

	var payload dhcpPayload
	payload.IPAddress = inst.PrimaryIPAddress

	if subnetID != "" {
		var sn Subnet
		if err := localdb.One("ID", subnetID, &sn); err == nil {
			payload.Gateway = sn.Gateway
			payload.DNSServers = sn.DNSServers
			payload.DomainName = sn.DomainName
			payload.NTPServers = sn.NTPServers
			if _, ipnet, err := net.ParseCIDR(sn.CIDRBlock); err == nil {
				m := ipnet.Mask
				payload.Netmask = fmt.Sprintf("%d.%d.%d.%d", m[0], m[1], m[2], m[3])
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payload)
}

// ipxeScriptHandler handles GET /ipxe/{mac} and returns the iPXE boot script
// for the instance whose NIC matches the given MAC address.  The instance's
// Kickstart field is served verbatim when it starts with "#!ipxe"; otherwise
// a minimal script is generated that sets the IP and chainloads from that URL.
func ipxeScriptHandler(w http.ResponseWriter, r *http.Request) {
	mac := chi.URLParam(r, "mac")
	if len(mac) == 12 && !strings.Contains(mac, ":") {
		mac = strings.Join([]string{mac[0:2], mac[2:4], mac[4:6], mac[6:8], mac[8:10], mac[10:12]}, ":")
	}

	inst := instanceByMAC(mac)
	if inst == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/plain")

	// If the iPXE script field already contains an iPXE script, serve it directly.
	if strings.HasPrefix(strings.TrimSpace(inst.IPXEScript), "#!ipxe") {
		w.Write([]byte(inst.IPXEScript))
		return
	}

	// Generate a minimal iPXE script.  If Kickstart holds a URL, chain to it;
	// otherwise just print instance info and drop to the iPXE shell.
	script := "#!ipxe\n"
	script += fmt.Sprintf("set instance-id %s\n", inst.ID)
	script += fmt.Sprintf("set hostname %s\n", inst.Name)
	if inst.PrimaryIPAddress != "" {
		script += fmt.Sprintf("set ip %s\n", inst.PrimaryIPAddress)
	}
	if inst.Kickstart != "" {
		// Treat the Kickstart field as a chain URL.
		script += fmt.Sprintf("chain %s\n", inst.Kickstart)
	} else {
		script += "echo No boot script configured for ${hostname} (${instance-id})\n"
		script += "shell\n"
	}
	w.Write([]byte(script))
}
