package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/hashicorp/go-hclog"
	"github.com/martezr/nightlight-cloud/compute"
	"github.com/martezr/nightlight-cloud/utils"
)

func ListInstances(w http.ResponseWriter, r *http.Request) {
	var instances []utils.Instance
	err := db.All(&instances)
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}
	json.NewEncoder(w).Encode(utils.NilSliceToEmptySlice(instances))
}

func CreateInstance(w http.ResponseWriter, r *http.Request) {
	var instance utils.Instance
	_ = json.NewDecoder(r.Body).Decode(&instance)
	var outputInstance utils.Instance
	outputInstance = instance
	outputInstance.ID = "i-" + utils.IDGenerator(10)

	// Find instance datastore
	datastore := FindDatastoreByID(outputInstance.DatastoreId)
	instancePath := fmt.Sprintf("%s/%s", datastore.LocalPath, outputInstance.ID)
	err := os.MkdirAll(instancePath, os.ModePerm)
	if err != nil {
		fmt.Println(err)
	}

	compute.CreateVM(outputInstance, instancePath)
	db.Save(&outputInstance)
	json.NewEncoder(w).Encode(utils.NilSliceToEmptySlice(outputInstance))
}

func GetInstance(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var instance utils.Instance
	err := db.One("ID", id, &instance)
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}
	json.NewEncoder(w).Encode(utils.NilSliceToEmptySlice(instance))
}

func DeleteInstance(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var instance utils.Instance
	err := db.One("ID", id, &instance)
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}
	datastore := FindDatastoreByID(instance.DatastoreId)
	compute.DeleteVM(id, datastore.Path)
	err = db.DeleteStruct(&instance)
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}
}

func RestartInstance(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var instance utils.Instance
	err := db.One("ID", id, &instance)
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}
	compute.RestartVM(instance.ID)
	json.NewEncoder(w).Encode(utils.NilSliceToEmptySlice(instance))
}
