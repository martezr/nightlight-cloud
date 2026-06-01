package main

import (
	"fmt"
	"os"

	"github.com/hashicorp/go-hclog"
	"github.com/martezr/nightlight-cloud/compute"
	"github.com/martezr/nightlight-cloud/utils"
)

func createWanRouter() {
	// Create the instance
	var routerInstance = utils.Instance{
		ID:           "wanrouter",
		Name:         "wanrouter-instance",
		Description:  "WAN Router Instance",
		BootType:     "bios",
		CPUSockets:   1,
		CPUCores:     1,
		MemoryMB:     1024,
		DatastoreId:  "defaultdatastore",
		InstanceType: "router",
		Devices: utils.Devices{
			NetworkInterfaces: []utils.NetworkInterface{
				{
					BootOrder:  3,
					Model:      "e1000",
					BridgeName: "nl-external",
					Connected:  true,
				},
				{
					BootOrder:  4,
					Model:      "e1000",
					BridgeName: "nl-transit",
					Connected:  true,
				},
			},
			CDROMs: []utils.CDROM{
				{
					BootOrder: 1,
					Connected: true,
					Path:      "/etc/alpine-nlrouter-v3.22-x86_64.iso",
				},
			},
		},
	}
	instancePath := fmt.Sprintf("%s/%s", "/opt/nightlight/volumes/defaultdatastore", "wanrouter")
	err := os.MkdirAll(instancePath, os.ModePerm)
	if err != nil {
		fmt.Println(err)
	}
	out, err := compute.CreateVM(routerInstance, instancePath)
	if err != nil {
		// Handle error
		fmt.Printf("Error creating WAN Router: %s\n", err)
		return
	}
	fmt.Printf("Create WAN Router Output: %v\n", out)
	vncPort, err := compute.GetVNCPort(routerInstance.ID)
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}
	routerInstance.VNCPort = vncPort
	db.Save(&routerInstance)
}

func createTier1Router(siteName string, siteCoreSwitch string) {
	// Create the instance
	var routerInstance = utils.Instance{
		ID:           fmt.Sprintf("tier1router-%s", siteName),
		Name:         fmt.Sprintf("tier1router-instance-%s", siteName),
		Description:  fmt.Sprintf("Tier 1 Router Instance for site %s", siteName),
		BootType:     "bios",
		CPUSockets:   1,
		CPUCores:     1,
		MemoryMB:     1024,
		DatastoreId:  "defaultdatastore",
		InstanceType: "router",
		Devices: utils.Devices{
			NetworkInterfaces: []utils.NetworkInterface{
				{
					BootOrder:  3,
					Model:      "e1000",
					BridgeName: "nl-transit",
					Connected:  true,
				},
				{
					BootOrder:  4,
					Model:      "e1000",
					BridgeName: siteCoreSwitch,
					Connected:  true,
				},
			},
			CDROMs: []utils.CDROM{
				{
					BootOrder: 1,
					Connected: true,
					Path:      "/etc/alpine-nlrouter-v3.22-x86_64.iso",
				},
			},
		},
	}
	instancePath := fmt.Sprintf("%s/%s", "/opt/nightlight/volumes/defaultdatastore", routerInstance.ID)
	err := os.MkdirAll(instancePath, os.ModePerm)
	if err != nil {
		fmt.Println(err)
	}
	out, err := compute.CreateVM(routerInstance, instancePath)
	if err != nil {
		// Handle error
		fmt.Printf("Error creating Tier 1 Router: %s\n", err)
		return
	}
	fmt.Printf("Create Tier 1 Router Output: %v\n", out)
	vncPort, err := compute.GetVNCPort(routerInstance.ID)
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}
	routerInstance.VNCPort = vncPort
	db.Save(&routerInstance)
}
