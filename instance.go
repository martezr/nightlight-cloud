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
	"golang.org/x/mobile/event/key"
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
	outputInstance := instance
	outputInstance.ID = "i-" + utils.IDGenerator(10)

	if instance.DatastoreId == "" {
		http.Error(w, "datastoreId is required", http.StatusBadRequest)
		return
	}

	// Find instance datastore
	datastore := FindDatastoreByID(outputInstance.DatastoreId)
	instancePath := fmt.Sprintf("%s/%s", datastore.LocalPath, outputInstance.ID)
	err := os.MkdirAll(instancePath, os.ModePerm)
	if err != nil {
		fmt.Println(err)
	}
	// iterate over storage disks and create disk images
	for i, disk := range outputInstance.Devices.StorageDisks {
		var diskPath string
		diskDatastore := FindDatastoreByID(disk.DatastoreId)
		if diskDatastore.ID == datastore.ID {
			diskPath = fmt.Sprintf("%s/%s-disk-%d.qcow2", instancePath, outputInstance.ID, i+1)
		} else {
			diskPath = fmt.Sprintf("%s/%s-disk-%d.qcow2", diskDatastore.LocalPath, outputInstance.ID, i+1)
		}
		err := compute.CreateDiskImage(diskPath, disk.SizeGB)
		if err != nil {
			hclog.Default().Named("core").Error(err.Error())
		}
		outputInstance.Devices.StorageDisks[i].Path = diskPath
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

func SendInstanceConsoleKeys(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var instance utils.Instance
	err := db.One("ID", id, &instance)
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}
	// create command struct to decode json body
	type Command struct {
		KeyCode    string `json:"keyCode"`
		RawMapping bool   `json:"rawMapping"`
		RawKeyCode uint32 `json:"rawKeyCode"`
	}

	var cmd Command

	err = json.NewDecoder(r.Body).Decode(&cmd)
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}

	if !cmd.RawMapping {
		fmt.Println("Sending keycode:", cmd.KeyCode)
		scancodeIndex := make(map[string]key.Code)
		scancodeIndex["abcdefghijklmnopqrstuvwxyz"] = key.CodeA
		scancodeIndex["ABCDEFGHIJKLMNOPQRSTUVWXYZ"] = key.CodeA
		scancodeIndex["1234567890"] = key.Code1
		scancodeIndex["!@#$%^&*()"] = key.Code1
		scancodeIndex[" "] = key.CodeSpacebar
		scancodeIndex["-=[]\\"] = key.CodeHyphenMinus
		scancodeIndex["_+{}|"] = key.CodeHyphenMinus
		scancodeIndex[";'`,./"] = key.CodeSemicolon
		scancodeIndex[":\"~<>?"] = key.CodeSemicolon

		var scancodeMap = make(map[rune]key.Code)
		for chars, start := range scancodeIndex {
			for i, r := range chars {
				scancodeMap[r] = start + key.Code(i)
			}
		}

		var keycodes []uint32

		// split cmd string into invidual characters
		for _, char := range cmd.KeyCode {
			fmt.Println(string(char))
			keycode1, ok := scancodeMap[char]
			if !ok {
				fmt.Printf("Unsupported character: %s\n", string(char))
				continue
			}
			fmt.Printf("Sending keycode for character %s: %d\n", string(char), keycode1)
			keycodes = append(keycodes, uint32(keycode1))
		}

		compute.SendConsoleKeyEvent(instance.ID, keycodes)
	} else {
		fmt.Println("Sending keycode:", cmd.RawKeyCode)
		// raw mapping, split by comma (comma-separated integer keycodes expected)
		var keycodes []uint32
		keycodes = append(keycodes, cmd.RawKeyCode)
		compute.SendConsoleKeyEvent(instance.ID, keycodes)
	}

	json.NewEncoder(w).Encode(utils.NilSliceToEmptySlice(instance))
}
