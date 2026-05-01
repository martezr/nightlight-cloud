package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/hashicorp/go-hclog"
	"github.com/martezr/go-openvswitch/ovs"
	"github.com/martezr/nightlight-cloud/compute"
	"github.com/martezr/nightlight-cloud/network"
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
	if err := json.NewDecoder(r.Body).Decode(&instance); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	outputInstance := instance
	outputInstance.ID = "i-" + utils.IDGenerator(10)

	hclog.Default().Named("core").Info(fmt.Sprintf("Creating instance: %+v", instance))
	if instance.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	if instance.DatastoreId == "" {
		http.Error(w, "datastoreId is required", http.StatusBadRequest)
		return
	}

	if len(instance.Devices.NetworkInterfaces) == 0 {
		http.Error(w, "at least one network interface is required", http.StatusBadRequest)
		return
	}

	if instance.Devices.NetworkInterfaces[0].BridgeName == "" {
		http.Error(w, "vpcId is required", http.StatusBadRequest)
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
	vncPort, err := compute.GetVNCPort(outputInstance.ID)
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}
	outputInstance.VNCPort = vncPort

	// Allocate primary IP address for the instance's first network interface
	hclog.Default().Named("core").Info("Allocating IP address for instance:", outputInstance.ID)
	primaryIP, err := AllocateIPAddress(instance.Devices.NetworkInterfaces[0].VPCId, outputInstance.ID, instance.Devices.NetworkInterfaces[0].MacAddress)
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	outputInstance.PrimaryIPAddress = primaryIP
	outputInstance.PrimaryMacAddress = instance.Devices.NetworkInterfaces[0].MacAddress
	outputInstance.CreatedAt = time.Now().GoString()
	db.Save(&outputInstance)
	c := ovs.New()
	ports, err := c.VSwitch.ListPorts("nightlight")
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}
	var ofPort int
	var metadataOfPort int
	var dhcpOfPort int
	for _, port := range ports {
		portDetails, err := c.VSwitch.Get.Port(port)
		fmt.Printf("Existing port: %s - Port Number: %s\n", port, portDetails.OFPort)
		if err != nil {
			hclog.Default().Named("core").Error(err.Error())
		}
		if portDetails.Name == "mddefaultvpc" {
			metadataOfPort, err = strconv.Atoi(portDetails.OFPort)
			if err != nil {
				hclog.Default().Named("core").Error(err.Error())
			}
			continue
		}
		if portDetails.Name == "dhdefaultvpc" {
			dhcpOfPort, err = strconv.Atoi(portDetails.OFPort)
			if err != nil {
				hclog.Default().Named("core").Error(err.Error())
			}
			continue
		}
		if portDetails.ExternalIds.AttachedMac == outputInstance.Devices.NetworkInterfaces[0].MacAddress {
			fmt.Println("Port already exists for MAC:", outputInstance.Devices.NetworkInterfaces[0].MacAddress)
			iPort, err := strconv.Atoi(portDetails.OFPort)
			if err != nil {
				hclog.Default().Named("core").Error(err.Error())
			} else {
				ofPort = iPort
			}
		}
	}

	hclog.Default().Named("core").Info(fmt.Sprintf("Adding OVS flows for instance %s on port %d (metadata port: %d, DHCP port: %d)", outputInstance.ID, ofPort, metadataOfPort, dhcpOfPort))
	network.AddVMFlows("nightlight", outputInstance.Devices.NetworkInterfaces[0].MacAddress, ofPort, metadataOfPort, dhcpOfPort)
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
	// Release allocated IP addressfor the instance's first network interface
	hclog.Default().Named("core").Info("Releasing IP address for instance:", instance.ID)

	err = ReleaseIPAddress(instance.Devices.NetworkInterfaces[0].VPCId, instance.ID)
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}

	// Remove instance directory
	instancePath := fmt.Sprintf("%s/%s", datastore.LocalPath, instance.ID)
	err = os.RemoveAll(instancePath)
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}

	// Remove OVS flows
	c := ovs.New()
	ports, err := c.VSwitch.ListPorts("nightlight")
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}
	var ofPort int
	for _, port := range ports {
		portDetails, err := c.VSwitch.Get.Port(port)
		if err != nil {
			hclog.Default().Named("core").Error(err.Error())
		}
		if portDetails.ExternalIds.AttachedMac == instance.Devices.NetworkInterfaces[0].MacAddress {
			iPort, err := strconv.Atoi(portDetails.OFPort)
			if err != nil {
				hclog.Default().Named("core").Error(err.Error())
			} else {
				ofPort = iPort
			}
		}
	}

	network.RemoveVMFlows("nightlight", ofPort)
}

func RestartInstance(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var instance utils.Instance
	err := db.One("ID", id, &instance)
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}
	compute.RestartVM(instance.ID)
	vncPort, err := compute.GetVNCPort(instance.ID)
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}
	instance.VNCPort = vncPort
	db.Save(&instance)
	json.NewEncoder(w).Encode(utils.NilSliceToEmptySlice(instance))
}

func ShutdownInstance(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var instance utils.Instance
	err := db.One("ID", id, &instance)
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}
	compute.ShutdownVM(instance.ID)
	json.NewEncoder(w).Encode(utils.NilSliceToEmptySlice(instance))
}

func StopInstance(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var instance utils.Instance
	err := db.One("ID", id, &instance)
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}
	compute.StopVM(instance.ID)
	json.NewEncoder(w).Encode(utils.NilSliceToEmptySlice(instance))
}

func StartInstance(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var instance utils.Instance
	err := db.One("ID", id, &instance)
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}
	compute.StartVM(instance.ID)
	vncPort, err := compute.GetVNCPort(instance.ID)
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}
	instance.VNCPort = vncPort
	db.Save(&instance)
	json.NewEncoder(w).Encode(utils.NilSliceToEmptySlice(instance))
}

type DHCPPayload struct {
	IPAddress  string   `json:"ipAddress"`
	Netmask    string   `json:"netmask"`
	Gateway    string   `json:"gateway"`
	DNSServers []string `json:"dnsServers"`
	DomainName string   `json:"domainName"`
	NTPServers []string `json:"ntpServers"`
}

func FindInstanceByMacAddress(mac string) (dhcpRecord DHCPPayload, found bool) {
	var instances []utils.Instance
	err := db.All(&instances)
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}
	for _, instance := range instances {
		for _, nic := range instance.Devices.NetworkInterfaces {
			if nic.MacAddress == mac {
				vpc := FindVPCByID(nic.VPCId) // touch VPC to update last used timestamp for IP allocation logic
				var payload DHCPPayload
				payload.IPAddress = instance.PrimaryIPAddress
				payload.Gateway = vpc.Gateway // Set the gateway if available
				payload.DNSServers = vpc.DNSServers
				payload.DomainName = vpc.DomainName
				payload.NTPServers = vpc.NTPServers
				// convert vpc.CIDRBlock from CIDR notation to netmask
				if _, ipnet, err := net.ParseCIDR(vpc.CIDRBlock); err == nil {
					ones, bits := ipnet.Mask.Size()
					mask := net.CIDRMask(ones, bits)
					payload.Netmask = fmt.Sprintf("%d.%d.%d.%d", mask[0], mask[1], mask[2], mask[3])
				} else {
					hclog.Default().Named("core").Error(fmt.Sprintf("invalid CIDR %s: %v", vpc.CIDRBlock, err))
				}

				return payload, true
			}
		}
	}
	return DHCPPayload{}, false
}

////////////////////////////////////////////////////////////
// TYPES
////////////////////////////////////////////////////////////

type KeyPress struct {
	Code  key.Code
	Shift bool
}

type ActionType int

const (
	ActionKey ActionType = iota
	ActionCombo
	ActionWait
)

type Action struct {
	Type  ActionType
	Keys  []KeyPress
	Delay time.Duration
}

////////////////////////////////////////////////////////////
// KEY MAP
////////////////////////////////////////////////////////////

var keyMap = map[rune]KeyPress{
	'a': {key.CodeA, false}, 'b': {key.CodeB, false}, 'c': {key.CodeC, false},
	'd': {key.CodeD, false}, 'e': {key.CodeE, false}, 'f': {key.CodeF, false},
	'g': {key.CodeG, false}, 'h': {key.CodeH, false}, 'i': {key.CodeI, false},
	'j': {key.CodeJ, false}, 'k': {key.CodeK, false}, 'l': {key.CodeL, false},
	'm': {key.CodeM, false}, 'n': {key.CodeN, false}, 'o': {key.CodeO, false},
	'p': {key.CodeP, false}, 'q': {key.CodeQ, false}, 'r': {key.CodeR, false},
	's': {key.CodeS, false}, 't': {key.CodeT, false}, 'u': {key.CodeU, false},
	'v': {key.CodeV, false}, 'w': {key.CodeW, false}, 'x': {key.CodeX, false},
	'y': {key.CodeY, false}, 'z': {key.CodeZ, false},

	'A': {key.CodeA, true}, 'B': {key.CodeB, true}, 'C': {key.CodeC, true},
	'D': {key.CodeD, true}, 'E': {key.CodeE, true}, 'F': {key.CodeF, true},
	'G': {key.CodeG, true}, 'H': {key.CodeH, true}, 'I': {key.CodeI, true},
	'J': {key.CodeJ, true}, 'K': {key.CodeK, true}, 'L': {key.CodeL, true},
	'M': {key.CodeM, true}, 'N': {key.CodeN, true}, 'O': {key.CodeO, true},
	'P': {key.CodeP, true}, 'Q': {key.CodeQ, true}, 'R': {key.CodeR, true},
	'S': {key.CodeS, true}, 'T': {key.CodeT, true}, 'U': {key.CodeU, true},
	'V': {key.CodeV, true}, 'W': {key.CodeW, true}, 'X': {key.CodeX, true},
	'Y': {key.CodeY, true}, 'Z': {key.CodeZ, true},

	'0': {key.Code0, false}, '1': {key.Code1, false}, '2': {key.Code2, false},
	'3': {key.Code3, false}, '4': {key.Code4, false}, '5': {key.Code5, false},
	'6': {key.Code6, false}, '7': {key.Code7, false}, '8': {key.Code8, false},
	'9': {key.Code9, false},

	')': {key.Code0, true}, '!': {key.Code1, true}, '@': {key.Code2, true},
	'#': {key.Code3, true}, '$': {key.Code4, true}, '%': {key.Code5, true},
	'^': {key.Code6, true}, '&': {key.Code7, true}, '*': {key.Code8, true},
	'(': {key.Code9, true},

	' ': {key.CodeSpacebar, false},

	'-': {key.CodeHyphenMinus, false}, '_': {key.CodeHyphenMinus, true},
	'=': {key.CodeEqualSign, false}, '+': {key.CodeEqualSign, true},

	'[': {key.CodeLeftSquareBracket, false}, '{': {key.CodeLeftSquareBracket, true},
	']': {key.CodeRightSquareBracket, false}, '}': {key.CodeRightSquareBracket, true},
	'\\': {key.CodeBackslash, false}, '|': {key.CodeBackslash, true},

	';': {key.CodeSemicolon, false}, ':': {key.CodeSemicolon, true},
	'\'': {key.CodeApostrophe, false}, '"': {key.CodeApostrophe, true},

	',': {key.CodeComma, false}, '<': {key.CodeComma, true},
	'.': {key.CodeFullStop, false}, '>': {key.CodeFullStop, true},
	'/': {key.CodeSlash, false}, '?': {key.CodeSlash, true},

	'`': {key.CodeGraveAccent, false}, '~': {key.CodeGraveAccent, true},
}

////////////////////////////////////////////////////////////
// TOKENIZER
////////////////////////////////////////////////////////////

func tokenize(input string) []string {
	var tokens []string
	var current strings.Builder
	inEscape := false

	for _, r := range input {
		switch r {
		case '<':
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
			inEscape = true
			current.WriteRune(r)

		case '>':
			current.WriteRune(r)
			tokens = append(tokens, current.String())
			current.Reset()
			inEscape = false

		default:
			if inEscape {
				current.WriteRune(r)
			} else {
				tokens = append(tokens, string(r))
			}
		}
	}

	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}

	return tokens
}

////////////////////////////////////////////////////////////
// PARSER
////////////////////////////////////////////////////////////

func parseToken(token string) ([]Action, error) {
	if strings.HasPrefix(token, "<") && strings.HasSuffix(token, ">") {
		content := strings.Trim(token, "<>")

		// WAIT
		switch content {
		case "wait":
			return []Action{{Type: ActionWait, Delay: 500 * time.Millisecond}}, nil
		case "wait5s":
			return []Action{{Type: ActionWait, Delay: 5 * time.Second}}, nil
		case "wait10s":
			return []Action{{Type: ActionWait, Delay: 10 * time.Second}}, nil
		}

		// COMBO
		if strings.Contains(content, "+") {
			parts := strings.Split(content, "+")
			var keys []KeyPress

			for _, p := range parts {
				kp, ok := lookupSpecialOrChar(p)
				if !ok {
					return nil, fmt.Errorf("unknown key: %s", p)
				}
				keys = append(keys, kp)
			}

			return []Action{{Type: ActionCombo, Keys: keys}}, nil
		}

		// SINGLE SPECIAL
		kp, ok := lookupSpecialOrChar(content)
		if !ok {
			return nil, fmt.Errorf("unknown key: %s", content)
		}

		return []Action{{Type: ActionKey, Keys: []KeyPress{kp}}}, nil
	}

	// NORMAL TEXT
	var actions []Action
	for _, r := range token {
		kp, ok := keyMap[r]
		if !ok {
			continue
		}
		actions = append(actions, Action{
			Type: ActionKey,
			Keys: []KeyPress{kp},
		})
	}

	return actions, nil
}

////////////////////////////////////////////////////////////
// LOOKUP
////////////////////////////////////////////////////////////

func lookupSpecialOrChar(s string) (KeyPress, bool) {
	s = strings.ToLower(s)

	switch s {
	case "ctrl":
		return KeyPress{key.CodeLeftControl, false}, true
	case "shift":
		return KeyPress{key.CodeLeftShift, false}, true
	case "alt":
		return KeyPress{key.CodeLeftAlt, false}, true
	case "enter":
		return KeyPress{key.CodeReturnEnter, false}, true
	case "tab":
		return KeyPress{key.CodeTab, false}, true
	case "esc":
		return KeyPress{key.CodeEscape, false}, true
	case "space":
		return KeyPress{key.CodeSpacebar, false}, true
	case "up":
		return KeyPress{key.CodeUpArrow, false}, true
	case "down":
		return KeyPress{key.CodeDownArrow, false}, true
	case "left":
		return KeyPress{key.CodeLeftArrow, false}, true
	case "right":
		return KeyPress{key.CodeRightArrow, false}, true
	case "backspace":
		return KeyPress{key.CodeDeleteBackspace, false}, true
	case "delete":
		return KeyPress{key.CodeDeleteBackspace, false}, true
	case "pgup":
		return KeyPress{key.CodePageUp, false}, true
	case "pgdn":
		return KeyPress{key.CodePageDown, false}, true
	}

	if len(s) == 1 {
		kp, ok := keyMap[rune(s[0])]
		return kp, ok
	}

	return KeyPress{}, false
}

////////////////////////////////////////////////////////////
// EXECUTION
////////////////////////////////////////////////////////////

func executeActions(instanceID string, actions []Action) {
	for _, action := range actions {
		switch action.Type {
		case ActionWait:
			time.Sleep(action.Delay)
		case ActionKey, ActionCombo:
			send(instanceID, action.Keys)
		}
	}
}

func send(instanceID string, keys []KeyPress) {
	var codes []uint32

	for _, k := range keys {
		if k.Shift {
			codes = append(codes, uint32(key.CodeLeftShift))
		}
		codes = append(codes, uint32(k.Code))
	}

	compute.SendConsoleKeyEvent(instanceID, codes)
}

////////////////////////////////////////////////////////////
// HTTP HANDLER
////////////////////////////////////////////////////////////

func SendInstanceConsoleKeys(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var instance utils.Instance
	if err := db.One("ID", id, &instance); err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}

	type Command struct {
		KeyCode    string `json:"keyCode"`
		RawMapping bool   `json:"rawMapping"`
		RawKeyCode uint32 `json:"rawKeyCode"`
	}

	var cmd Command
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}

	// RAW MODE (unchanged behavior)
	if cmd.RawMapping {
		compute.SendConsoleKeyEvent(instance.ID, []uint32{cmd.RawKeyCode})
		json.NewEncoder(w).Encode(utils.NilSliceToEmptySlice(instance))
		return
	}

	fmt.Println("Sending keycode:", cmd.KeyCode)

	// DSL MODE
	tokens := tokenize(cmd.KeyCode)

	var actions []Action

	for _, token := range tokens {
		acts, err := parseToken(token)
		if err != nil {
			hclog.Default().Named("core").Error(err.Error())
			continue
		}
		actions = append(actions, acts...)
	}
	fmt.Println("Sending actions:", actions)

	executeActions(instance.ID, actions)

	json.NewEncoder(w).Encode(utils.NilSliceToEmptySlice(instance))
}
