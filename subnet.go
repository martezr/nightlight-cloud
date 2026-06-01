package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/go-chi/chi"

	"github.com/hashicorp/go-hclog"
	"github.com/martezr/nightlight-cloud/utils"
)

// Subnet represents a subnet within a VNet
type Subnet struct {
	ID          string                   `json:"id"`
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	CIDRBlock   string                   `json:"cidrBlock"`
	Tags        []map[string]interface{} `json:"tags"`
	SiteId      string                   `json:"siteId"`
	VLANId      int64                    `json:"vlanId"`
	DNSServers  []string                 `json:"dnsServers"`
	NTPServers  []string                 `json:"ntpServers"`
	DomainName  string                   `json:"domainName"`
	BridgeName  string                   `json:"bridgeName"`
	DHCPServer  bool                     `json:"dhcpServer"`
	IPPoolRange string                   `json:"ipPoolRange"`
	Gateway     string                   `json:"gateway"`
}

// ListSubnets lists all subnets
func ListSubnets(w http.ResponseWriter, r *http.Request) {
	var subnets []Subnet
	err := db.All(&subnets)
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}
	json.NewEncoder(w).Encode(utils.NilSliceToEmptySlice(subnets))
}

// CreateSubnet creates a new subnet
func CreateSubnet(w http.ResponseWriter, r *http.Request) {
	var subnet Subnet
	_ = json.NewDecoder(r.Body).Decode(&subnet)
	subNumber := utils.IDGenerator(10)
	subnet.ID = "subnet-" + subNumber

	if subnet.IPPoolRange != "" && subnet.DHCPServer {
		// generate IP records for the VNet
		IPs := strings.Split(subnet.IPPoolRange, "-")
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
				Status:    "available",
				SubnetId:  subnet.ID,
				SiteId:    subnet.SiteId,
			}
			db.Save(&ipRecord)
		}
	}

	db.Save(&subnet)
	json.NewEncoder(w).Encode(utils.NilSliceToEmptySlice(subnet))
}

// UpdateSubnet updates an existing subnet by ID
func UpdateSubnet(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var subnet Subnet
	err := db.One("ID", id, &subnet)
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}

	var data Subnet
	_ = json.NewDecoder(r.Body).Decode(&data)
	data.ID = subnet.ID
	err = db.Update(&data)
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}
}

// DeleteSubnet deletes a existing subnet by ID
func DeleteSubnet(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var subnet Subnet
	err := db.One("ID", id, &subnet)
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}

	// Check if there are any instances with network interfaces in this subnet
	var instances []utils.Instance
	err = db.All(&instances)
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}
	for _, instance := range instances {
		for _, nic := range instance.Devices.NetworkInterfaces {
			if nic.SubnetId == subnet.ID {
				http.Error(w, "Cannot delete subnet with active instances", http.StatusBadRequest)
				return
			}
		}
	}

	// Delete associated IP records
	var ipRecords []IPRecord
	err = db.Find("SubnetId", subnet.ID, &ipRecords)
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}
	for _, record := range ipRecords {
		db.DeleteStruct(&record)
	}

	err = db.DeleteStruct(&subnet)
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}
}

// GetSubnet fetches a subnet by ID and returns it
func GetSubnet(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var subnet Subnet
	err := db.One("ID", id, &subnet)
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}
	json.NewEncoder(w).Encode(utils.NilSliceToEmptySlice(subnet))
}

// GetSubnet fetches a subnet by ID and returns it
func GetSubnetByID(id string) (subnet Subnet) {
	err := db.One("ID", id, &subnet)
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}
	return subnet
}

func FindSubnetByID(id string) (subnet Subnet) {
	err := db.One("ID", id, &subnet)
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}
	return subnet
}

// List IP records for a subnet
func ListIPRecords(w http.ResponseWriter, r *http.Request) {
	subnetId := chi.URLParam(r, "id")
	var ipRecords []IPRecord
	err := db.Find("SubnetId", subnetId, &ipRecords)
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}
	json.NewEncoder(w).Encode(utils.NilSliceToEmptySlice(ipRecords))
}

// List available IPs in a Subnet
func ListAvailableIPs(w http.ResponseWriter, r *http.Request) {
	subnetId := chi.URLParam(r, "id")
	var ipRecords []IPRecord
	err := db.Find("SubnetId", subnetId, &ipRecords)
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

func AllocateIPAddress(subnetId string, instanceId string, macAddress string) (string, error) {
	var ipRecords []IPRecord
	fmt.Printf("Allocating IP address for instance %s in Subnet %s\n", instanceId, subnetId)
	err := db.Find("SubnetId", subnetId, &ipRecords)
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
	return "", fmt.Errorf("no available IP addresses in Subnet %s", subnetId)
}

func ReleaseSubnetIPAddress(w http.ResponseWriter, r *http.Request) {
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

func ReleaseIPAddress(ipAddress string, subnetId string) error {
	var ipRecords []IPRecord
	err := db.Find("SubnetId", subnetId, &ipRecords)
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
	return fmt.Errorf("IP address %s not found in Subnet %s", ipAddress, subnetId)
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

type IPRecord struct {
	ID         int    `json:"id" storm:"id,increment"`
	IPAddress  string `json:"ipAddress" storm:"index"`
	SubnetId   string `json:"subnetId" storm:"index"`
	SiteId     string `json:"siteId" storm:"index"`
	Status     string `json:"status"`
	InstanceId string `json:"instanceId"`
	MacAddress string `json:"macAddress"`
}
