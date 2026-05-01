package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi"
	"github.com/hashicorp/go-hclog"
	"github.com/martezr/nightlight-cloud/utils"
)

// VPC represents a virtual private cloud
type VPC struct {
	ID          string                   `json:"id" storm:"id,index"`
	Name        string                   `json:"name" storm:"index"`
	Description string                   `json:"description"`
	CIDRBlock   string                   `json:"cidrBlock"`
	Tags        []map[string]interface{} `json:"tags"`
	DNSServers  []string                 `json:"dnsServers"`
	NTPServers  []string                 `json:"ntpServers"`
	DomainName  string                   `json:"domainName"`
	BridgeName  string                   `json:"bridgeName"`
	DHCPServer  bool                     `json:"dhcpServer"`
	IPPoolRange string                   `json:"ipPoolRange"`
	Gateway     string                   `json:"gateway"`
}

type IPRecord struct {
	ID         int    `json:"id" storm:"id,increment"`
	IPAddress  string `json:"ipAddress" storm:"index"`
	SubnetId   string `json:"subnetId"`
	VPCId      string `json:"vpcId"`
	Status     string `json:"status"`
	InstanceId string `json:"instanceId"`
	MacAddress string `json:"macAddress"`
}

// ListVpcs lists all VPCs
func ListVpcs(w http.ResponseWriter, r *http.Request) {
	var vpcs []VPC
	err := db.All(&vpcs)
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}
	json.NewEncoder(w).Encode(utils.NilSliceToEmptySlice(vpcs))
}

// CreateVPC creates a new VPC
func CreateVPC(w http.ResponseWriter, r *http.Request) {
	var vpc VPC
	if err := json.NewDecoder(r.Body).Decode(&vpc); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	vpc.ID = "vpc-" + utils.IDGenerator(10)
	//vpc.BridgeName = "vpc" + utils.IDGenerator(10)
	vpc.BridgeName = "nightlight"
	if vpc.IPPoolRange != "" {
		// generate IP records for the VPC
		IPs := strings.Split(vpc.IPPoolRange, "-")
		startIP := IPs[0]
		endIP := IPs[1]

		ips, err := ipsInRange(startIP, endIP)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		for _, ip := range ips {
			fmt.Println(ip)
			ipRecord := IPRecord{
				IPAddress: ip,
				VPCId:     vpc.ID,
				Status:    "available",
			}
			db.Save(&ipRecord)
		}
	}
	//ovs.VSwitch.AddBridge(vpc.BridgeName)
	db.Save(&vpc)
	json.NewEncoder(w).Encode(utils.NilSliceToEmptySlice(vpc))
}

// GetVPC fetches a VPC by ID and returns it
func GetVPC(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var vpc VPC
	err := db.One("ID", id, &vpc)
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}
	json.NewEncoder(w).Encode(utils.NilSliceToEmptySlice(vpc))
}

func GetVPCGraph(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	fmt.Println("Fetching graph for VPC ID:", id)
	// read graph data from file, open graph.json and return its contents using os.ReadFile
	data, err := os.ReadFile("/graph.json")
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
		json.NewEncoder(w).Encode(utils.NilSliceToEmptySlice(nil))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func GetVPCFlowLogs(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	fmt.Println("Fetching flow logs for VPC ID:", id)
	// read flow log data from file, open flowlogs.json and return its contents using os.ReadFile
	data, err := os.ReadFile("/var/log/flowagent.json")
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
		json.NewEncoder(w).Encode(utils.NilSliceToEmptySlice(nil))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

// UpdateVPC updates an existing VPC by ID
func UpdateVPC(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var vpc VPC
	err := db.One("ID", id, &vpc)
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}

	var data VPC
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	data.ID = vpc.ID
	err = db.Update(&data)
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}
}

// DeleteVPC deletes a existing VPC by ID
func DeleteVPC(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var vpc VPC
	err := db.One("ID", id, &vpc)
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}

	err = db.DeleteStruct(&vpc)
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}
}

// FindVPCByName finds a VPC by name and returns it
func FindVPCByName(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	var vpc VPC
	err := db.One("Name", name, &vpc)
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}
	json.NewEncoder(w).Encode(utils.NilSliceToEmptySlice(vpc))
}

// List available IPs in a VPC
func ListAvailableIPs(w http.ResponseWriter, r *http.Request) {
	vpcId := chi.URLParam(r, "id")
	var ipRecords []IPRecord
	err := db.Find("VPCId", vpcId, &ipRecords)
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}
	var availableIPs []string
	for _, record := range ipRecords {
		if record.Status == "available" {
			availableIPs = append(availableIPs, record.IPAddress)
		}
	}
	json.NewEncoder(w).Encode(utils.NilSliceToEmptySlice(availableIPs))
}

func AllocateIPAddress(vpcId string, instanceId string, macAddress string) (string, error) {
	var ipRecords []IPRecord
	fmt.Printf("Allocating IP address for instance %s in VPC %s\n", instanceId, vpcId)
	err := db.Find("VPCId", vpcId, &ipRecords)
	if err != nil {
		return "", err
	}
	for _, record := range ipRecords {
		if record.Status == "available" {
			record.Status = "allocated"
			record.InstanceId = instanceId
			record.MacAddress = macAddress
			db.Update(&record)
			fmt.Println("Allocated IP address:", record.IPAddress)
			return record.IPAddress, nil
		}
	}
	return "", fmt.Errorf("no available IP addresses in VPC %s", vpcId)
}

func ReleaseVPCIPAddress(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var request struct {
		IPAddress string `json:"ipAddress"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	err := ReleaseIPAddress(request.IPAddress, id)
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
		json.NewEncoder(w).Encode(utils.NilSliceToEmptySlice(fmt.Sprintf("Error releasing IP address: %v", err)))
		return
	}
	json.NewEncoder(w).Encode(utils.NilSliceToEmptySlice(fmt.Sprintf("Released IP address: %s", request.IPAddress)))
}

func ReleaseIPAddress(ipAddress string, vpcId string) error {
	var ipRecords []IPRecord
	err := db.Find("VPCId", vpcId, &ipRecords)
	if err != nil {
		return err
	}
	for _, record := range ipRecords {
		if record.IPAddress == ipAddress {
			record.Status = "available"
			record.InstanceId = ""
			record.MacAddress = ""
			db.Update(&record)
			fmt.Println("Released IP address:", record.IPAddress)
			return nil
		}
	}
	return fmt.Errorf("IP address %s not found in VPC %s", ipAddress, vpcId)
}

// Convert IP to uint32
func ipToInt(ip net.IP) uint32 {
	ip = ip.To4()
	return uint32(ip[0])<<24 |
		uint32(ip[1])<<16 |
		uint32(ip[2])<<8 |
		uint32(ip[3])
}

// Convert uint32 back to IP
func intToIP(n uint32) net.IP {
	return net.IPv4(
		byte(n>>24),
		byte(n>>16),
		byte(n>>8),
		byte(n),
	)
}

func ipsInRange(startIP, endIP string) ([]string, error) {
	start := net.ParseIP(startIP)
	end := net.ParseIP(endIP)

	if start == nil || end == nil {
		return nil, fmt.Errorf("invalid IP address")
	}

	startInt := ipToInt(start)
	endInt := ipToInt(end)

	if startInt > endInt {
		return nil, fmt.Errorf("start IP is greater than end IP")
	}

	var ips []string
	for i := startInt; i <= endInt; i++ {
		ips = append(ips, intToIP(i).String())
	}

	return ips, nil
}

func FindVPCByID(id string) (vpc VPC) {
	err := db.One("ID", id, &vpc)
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}
	return vpc
}
