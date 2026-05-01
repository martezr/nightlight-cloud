package main

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"

	"github.com/hashicorp/go-hclog"
	"github.com/martezr/nightlight-cloud/utils"
)

// Subnet represents a subnet within a VPC
type Subnet struct {
	ID          string                   `json:"id" storm:"id,index"`
	Name        string                   `json:"name" storm:"index"`
	Description string                   `json:"description"`
	CIDRBlock   string                   `json:"cidrBlock"`
	Tags        []map[string]interface{} `json:"tags"`
	VPCId       string                   `json:"vpcId" storm:"index"`
	BridgeName  string                   `json:"bridgeName"`
}

// SubnetGetResponse is the response structure for getting a subnet by ID
type SubnetGetResponse struct {
	Subnet Subnet `json:"subnet"`
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
	//subnet.BridgeName = "sub" + subNumber

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
	err = db.DeleteStruct(&subnet)
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}
}

// GetSubnet fetches a subnet by ID and returns it
func GetSubnet(id string) (subnet Subnet) {
	err := db.One("ID", id, &subnet)
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}
	return subnet
}
