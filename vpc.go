package main

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/hashicorp/go-hclog"
	"github.com/martezr/nightlight-cloud/utils"
)

type VPC struct {
	ID          string                   `json:"id" storm:"id,index"`
	Name        string                   `json:"name" storm:"index"`
	Description string                   `json:"description"`
	CIDRBlock   string                   `json:"cidrBlock"`
	Tags        []map[string]interface{} `json:"tags"`
	DNSServers  []string                 `json:"dnsServers"`
	DomainName  string                   `json:"domainName"`
}

func ListVpcs(w http.ResponseWriter, r *http.Request) {
	var vpcs []VPC
	err := db.All(&vpcs)
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}
	json.NewEncoder(w).Encode(utils.NilSliceToEmptySlice(vpcs))
}

func CreateVPC(w http.ResponseWriter, r *http.Request) {
	var vpc VPC
	_ = json.NewDecoder(r.Body).Decode(&vpc)
	vpc.ID = "vpc-" + utils.IDGenerator(10)
	db.Save(&vpc)
	json.NewEncoder(w).Encode(utils.NilSliceToEmptySlice(vpc))
}

func GetVPC(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var vpc VPC
	err := db.One("ID", id, &vpc)
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}
	json.NewEncoder(w).Encode(utils.NilSliceToEmptySlice(vpc))
}

func UpdateVPC(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var vpc VPC
	err := db.One("ID", id, &vpc)
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}

	var data VPC
	_ = json.NewDecoder(r.Body).Decode(&data)
	data.ID = vpc.ID
	err = db.Update(&data)
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}
}

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
