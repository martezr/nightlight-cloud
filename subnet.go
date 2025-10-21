package main

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/martezr/go-openvswitch/ovs"

	"github.com/hashicorp/go-hclog"
	"github.com/martezr/nightlight-cloud/utils"
)

type Subnet struct {
	ID          string                   `json:"id" storm:"id,index"`
	Name        string                   `json:"name" storm:"index"`
	Description string                   `json:"description"`
	CIDRBlock   string                   `json:"cidrBlock"`
	Tags        []map[string]interface{} `json:"tags"`
	VPCId       string                   `json:"vpcId" storm:"index"`
	BridgeName  string                   `json:"bridgeName"`
}

type SubnetGetResponse struct {
	Subnet Subnet `json:"subnet"`
}

func ListSubnets(w http.ResponseWriter, r *http.Request) {
	var subnets []Subnet
	err := db.All(&subnets)
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}
	json.NewEncoder(w).Encode(utils.NilSliceToEmptySlice(subnets))
}

func CreateSubnet(w http.ResponseWriter, r *http.Request) {
	var subnet Subnet
	_ = json.NewDecoder(r.Body).Decode(&subnet)
	subNumber := utils.IDGenerator(10)
	subnet.ID = "subnet-" + subNumber
	subnet.BridgeName = "sub" + subNumber
	c := ovs.New()

	c.VSwitch.AddBridge(subnet.BridgeName)
	db.Save(&subnet)
	json.NewEncoder(w).Encode(utils.NilSliceToEmptySlice(subnet))
}

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

func DeleteSubnet(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var subnet Subnet
	err := db.One("ID", id, &subnet)
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}

	c := ovs.New()

	c.VSwitch.DeleteBridge(subnet.BridgeName)

	err = db.DeleteStruct(&subnet)
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}
}

func FindSubnetByID(id string) (subnet Subnet) {
	err := db.One("ID", id, &subnet)
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}
	return subnet
}
