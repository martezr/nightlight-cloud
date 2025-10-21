package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/DataDog/datadog-agent/pkg/gohai"
	"github.com/martezr/nightlight-cloud/utils"
)

type Host struct {
	Hostname        string                 `json:"hostname"`
	OperatingSystem string                 `json:"os"`
	Cores           int64                  `json:"cores"`
	Memory          int64                  `json:"memory"`
	Interfaces      []HostNetworkInterface `json:"interfaces"`
	IPAddress       string                 `json:"ipAddress"`
	MacAddress      string                 `json:"macAddress"`
}

type HostListResponse struct {
	Hosts []Host `json:"hosts"`
}

type HostGetResponse struct {
	Host Host `json:"host"`
}

type HostNetworkInterface struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}

func getSystemInfo() (out Host) {

	var payload Host
	output := gohai.GetPayload(false)

	cpuMap := output.Gohai.CPU.(map[string]interface{})
	n, err := strconv.Atoi(cpuMap["cpu_cores"].(string))
	if err != nil {
		fmt.Println(err)
	}
	payload.Cores = int64(n)
	memoryMap := output.Gohai.Memory.(map[string]interface{})
	n, err = strconv.Atoi(memoryMap["total"].(string))
	if err != nil {
		fmt.Println(err)
	}
	payload.Memory = int64(n)

	networkMap := output.Gohai.Network.(map[string]interface{})
	fmt.Sprintf("network map: %v", networkMap)
	payload.IPAddress = networkMap["ipaddress"].(string)
	payload.MacAddress = networkMap["macaddress"].(string)

	ints := networkMap["interfaces"].([]interface{})
	for i, nic := range ints {
		log.Println(nic)

		rec := ints[i].(map[string]interface{})
		dat := rec["ipv4"].([]string)
		var address string
		if len(dat) > 0 {
			address = dat[0]
		}
		row := &HostNetworkInterface{
			Name:    rec["name"].(string),
			Address: address,
		}
		payload.Interfaces = append(payload.Interfaces, *row)
	}

	myMap := output.Gohai.Platform.(map[string]interface{})
	payload.OperatingSystem = myMap["os"].(string)
	payload.Hostname = myMap["hostname"].(string)

	return payload
}

func ListHosts(w http.ResponseWriter, r *http.Request) {
	var hosts []Host
	hosts = append(hosts, getSystemInfo())
	payload := HostListResponse{Hosts: hosts}
	json.NewEncoder(w).Encode(utils.NilSliceToEmptySlice(payload))
}
