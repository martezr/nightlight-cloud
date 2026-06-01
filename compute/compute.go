package compute

import (
	"crypto/rand"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"

	"libvirt.org/go/libvirtxml"

	"github.com/digitalocean/go-libvirt"
	"github.com/google/uuid"
	"github.com/martezr/nightlight-cloud/utils"
)

type InstanceType struct {
	InstanceType string       `json:"instanceType"`
	InstanceSize InstanceSize `json:"instanceSize"`
}

type InstanceSize struct {
	CPU    uint `json:"cpu"`
	Memory uint `json:"memory"`
}

// generateInstanceUUID generates a random id for instances
func generateInstanceUUID() (output string) {
	input := uuid.New()
	firstOutput := replaceAtIndex(input.String(), 'e', 0)
	secondOutput := replaceAtIndex(firstOutput, 'c', 1)
	finalOutput := replaceAtIndex(secondOutput, '2', 2)
	return finalOutput
}

func replaceAtIndex(in string, r rune, i int) string {
	out := []rune(in)
	out[i] = r
	return string(out)
}

func copyFile(src, dest string) {
	sourceFile := src
	destinationFile := dest

	source, err := os.Open(sourceFile) //open the source file
	if err != nil {
		panic(err)
	}
	defer source.Close()

	destination, err := os.Create(destinationFile) //create the destination file
	if err != nil {
		panic(err)
	}
	defer destination.Close()
	_, err = io.Copy(destination, source) //copy the contents of source to destination file
	if err != nil {
		panic(err)
	}
}

func CreateVM(instanceDef utils.Instance, instancePath string) (utils.Instance, error) {
	c, err := net.DialTimeout("unix", "/var/run/libvirt/libvirt-sock", 10*time.Second)
	if err != nil {
		return utils.Instance{}, fmt.Errorf("failed to dial libvirt: %v", err)
	}

	l := libvirt.New(c)
	if err := l.Connect(); err != nil {
		c.Close()
		return utils.Instance{}, fmt.Errorf("failed to connect to libvirt: %v", err)
	}

	vmUUID := generateInstanceUUID()
	var top libvirtxml.DomainSysInfo
	var test libvirtxml.DomainSysInfoSMBIOS
	var demo libvirtxml.DomainSysInfoSystem
	var t1 libvirtxml.DomainSysInfoEntry
	var t2 libvirtxml.DomainSysInfoEntry
	t2.Name = "serial"
	t2.Value = vmUUID
	t1.Name = "uuid"
	t1.Value = vmUUID
	demo.Entry = []libvirtxml.DomainSysInfoEntry{t1, t2}
	test.System = &demo
	top.SMBIOS = &test
	memory := instanceDef.MemoryMB
	domainDef := libvirtxml.Domain{
		UUID:     vmUUID,
		SysInfo:  []libvirtxml.DomainSysInfo{top},
		Metadata: &libvirtxml.DomainMetadata{},
		Memory: &libvirtxml.DomainMemory{
			Unit:  "MiB",
			Value: uint(memory),
		},
		VCPU: &libvirtxml.DomainVCPU{
			Placement: "static",
			Value:     uint(instanceDef.CPUCores * instanceDef.CPUSockets),
		},
		CPU: &libvirtxml.DomainCPU{
			Mode:       "host-passthrough",
			Check:      "none",
			Migratable: "on",
			Topology: &libvirtxml.DomainCPUTopology{
				Sockets: instanceDef.CPUSockets,
				Cores:   instanceDef.CPUCores,
				Threads: 1,
			},
		},
		Devices: &libvirtxml.DomainDeviceList{
			Emulator: "/usr/bin/qemu-system-x86_64",
			Controllers: []libvirtxml.DomainController{
				{
					Type:  "pci",
					Model: "pcie-root",
					Index: new(uint),
					Alias: &libvirtxml.DomainAlias{
						Name: "pcie.0",
					},
				},
			},
			Consoles: []libvirtxml.DomainConsole{
				{
					Target: &libvirtxml.DomainConsoleTarget{
						Type: "serial",
					},
				},
			},
			Serials: []libvirtxml.DomainSerial{
				{
					Target: &libvirtxml.DomainSerialTarget{
						Type: "isa-serial",
						Model: &libvirtxml.DomainSerialTargetModel{
							Name: "isa-serial",
						},
					},
				},
			},
			Graphics: []libvirtxml.DomainGraphic{
				{
					VNC: &libvirtxml.DomainGraphicVNC{
						AutoPort: "yes",
						Listen:   "0.0.0.0",
					},
				},
			},
			Channels: []libvirtxml.DomainChannel{
				{
					Source: &libvirtxml.DomainChardevSource{
						UNIX: &libvirtxml.DomainChardevSourceUNIX{},
					},
					Target: &libvirtxml.DomainChannelTarget{
						VirtIO: &libvirtxml.DomainChannelTargetVirtIO{
							Name: "org.qemu.guest_agent.0",
						},
					},
				},
			},
		},
		Features: &libvirtxml.DomainFeatureList{
			PAE:  &libvirtxml.DomainFeature{},
			ACPI: &libvirtxml.DomainFeature{},
			APIC: &libvirtxml.DomainFeatureAPIC{},
		},
	}

	domainDef.Name = instanceDef.ID
	domainDef.Type = "kvm"

	// Bootloader
	if instanceDef.BootType == "uefi" {
		domainDef.OS = &libvirtxml.DomainOS{
			Firmware: "efi",
			Type: &libvirtxml.DomainOSType{
				Type: "hvm",
			},
			SMBios: &libvirtxml.DomainSMBios{
				Mode: "sysinfo",
			},
			Loader: &libvirtxml.DomainLoader{
				Secure:   "no",
				Readonly: "yes",
				Type:     "pflash",
				Path:     "/usr/share/qemu/edk2-x86_64-code.fd",
			},
		}
	} else {
		domainDef.OS = &libvirtxml.DomainOS{
			//Firmware: "bios",
			Type: &libvirtxml.DomainOSType{
				Type: "hvm",
			},
			SMBios: &libvirtxml.DomainSMBios{
				Mode: "sysinfo",
			},
		}
	}
	domainDef.OS.Type.Arch = "x86_64"
	domainDef.OS.Type.Machine = "pc-q35-10.0"

	if instanceDef.SecureBoot {
		domainDef.OS.FirmwareInfo = &libvirtxml.DomainOSFirmwareInfo{
			Features: []libvirtxml.DomainOSFirmwareFeature{
				{Name: "secure-boot", Enabled: "yes"},
				{Name: "enrolled-keys", Enabled: "no"},
			},
		}
		domainDef.OS.Loader.Secure = "yes"
		domainDef.OS.Loader.Readonly = "yes"
		domainDef.OS.Loader.Type = "pflash"
		domainDef.OS.Loader.Path = "/etc/OVMF_CODE_4M.ms.fd"
		domainDef.OS.Loader.Format = "raw"
		// NVRAM Template
		domainDef.OS.NVRam = &libvirtxml.DomainNVRam{
			Template: "/etc/OVMF_VARS_4M.ms.fd",
			NVRam:    fmt.Sprintf("/var/lib/libvirt/qemu/nvram/%s_VARS.fd", instanceDef.ID),
			Format:   "raw",
		}
		domainDef.Features = &libvirtxml.DomainFeatureList{
			PAE:  &libvirtxml.DomainFeature{},
			ACPI: &libvirtxml.DomainFeature{},
			APIC: &libvirtxml.DomainFeatureAPIC{},
			SMM: &libvirtxml.DomainFeatureSMM{
				State: "on",
			},
		}
	}

	if instanceDef.TPM {
		domainDef.Devices.TPMs = []libvirtxml.DomainTPM{
			{
				Model: "tpm-tis",
				Backend: &libvirtxml.DomainTPMBackend{
					Emulator: &libvirtxml.DomainTPMBackendEmulator{
						Version: "2.0",
					},
				},
			},
		}
	}

	// Add network interfaces
	nics := instanceDef.Devices.NetworkInterfaces
	for i, nic := range nics {
		mac, err := randomMACAddress()
		if err != nil {
			fmt.Errorf("error generating mac address: %w", err)
		}
		if i == 0 {
			instanceDef.PrimaryMacAddress = mac
		}
		nics[i].MacAddress = mac

		netIface := libvirtxml.DomainInterface{
			VirtualPort: &libvirtxml.DomainInterfaceVirtualPort{
				Params: &libvirtxml.DomainInterfaceVirtualPortParams{
					OpenVSwitch: &libvirtxml.DomainInterfaceVirtualPortParamsOpenVSwitch{},
				},
			},
			Model: &libvirtxml.DomainInterfaceModel{
				Type: nic.Model,
			},
			ROM: &libvirtxml.DomainROM{},
			MAC: &libvirtxml.DomainInterfaceMAC{
				Address: mac,
			},
			Source: &libvirtxml.DomainInterfaceSource{
				Bridge: &libvirtxml.DomainInterfaceSourceBridge{
					Bridge: nic.BridgeName,
				},
			},
		}

		if nic.BootOrder > 0 {
			netIface.Boot = &libvirtxml.DomainDeviceBoot{
				Order: uint(nic.BootOrder),
			}
		}

		if nic.Model == "virtio" {
			romFile := "/etc/virtio-net.rom"
			netIface.ROM = &libvirtxml.DomainROM{
				File: &romFile,
				Bar:  "off",
			}
		}

		domainDef.Devices.Interfaces = append(domainDef.Devices.Interfaces, netIface)
	}

	virtioDisks := []string{"vda", "vdb", "vdc", "vdd", "vde", "vdf", "vdg", "vdh", "vdi", "vdj"}
	sataDisks := []string{"sda", "sdb", "sdc", "sdd", "sde", "sdf", "sdg", "sdh", "sdi", "sdj"}

	// Add storage disks
	storageDisks := instanceDef.Devices.StorageDisks
	virtioIndex := 0
	sataIndex := 0
	for _, disk := range storageDisks {
		var diskTarget string
		switch disk.BusType {
		case "virtio":
			diskTarget = virtioDisks[virtioIndex]
			virtioIndex++
		case "sata":
			diskTarget = sataDisks[sataIndex]
			sataIndex++
		default:
			fmt.Printf("unsupported bus type: %s\n", disk.BusType)
			continue
		}
		//diskPath := fmt.Sprintf("%s/%s_disk_%s.qcow2", instancePath, instanceDef.ID, diskTarget)
		if disk.ExistingPath != "" {
			fmt.Println("Using existing disk image:", disk.ExistingPath)
			// Check if the existing path is a qcow2 file
			if !strings.HasSuffix(disk.ExistingPath, ".qcow2") {
				fmt.Println("Found non-qcow2 disk image, converting...")
				ConvertDiskImage(disk.ExistingPath, disk.Path, disk.SizeGB)
			} else {
				fmt.Println("Found qcow2 disk image, copying...")
				// Copy existing disk image
				copyFile(disk.ExistingPath, disk.Path)
			}
		}

		storageDisk := libvirtxml.DomainDisk{
			Boot: &libvirtxml.DomainDeviceBoot{
				Order: uint(disk.BootOrder),
			},
			Driver: &libvirtxml.DomainDiskDriver{
				Name: "qemu",
				Type: "qcow2",
			},
			Device: "disk",
			Target: &libvirtxml.DomainDiskTarget{
				Dev: diskTarget,
				Bus: disk.BusType,
			},
			Source: &libvirtxml.DomainDiskSource{
				File: &libvirtxml.DomainDiskSourceFile{
					File: disk.Path,
				},
			},
		}
		domainDef.Devices.Disks = append(domainDef.Devices.Disks, storageDisk)
	}

	cdroms := instanceDef.Devices.CDROMs
	for _, cd := range cdroms {
		diskTarget := sataDisks[sataIndex]
		sataIndex++
		cdromDevice := libvirtxml.DomainDisk{
			Boot: &libvirtxml.DomainDeviceBoot{
				Order: uint(cd.BootOrder),
			},
			Device: "cdrom",
			Target: &libvirtxml.DomainDiskTarget{
				Dev: diskTarget,
				Bus: "sata",
			},
			Source: &libvirtxml.DomainDiskSource{
				File: &libvirtxml.DomainDiskSourceFile{
					File: cd.Path,
				},
			},
			ReadOnly: &libvirtxml.DomainDiskReadOnly{},
		}
		domainDef.Devices.Disks = append(domainDef.Devices.Disks, cdromDevice)
	}

	if instanceDef.WinAutoattend != "" {
		// create floppy directory in instance path
		err := os.MkdirAll(fmt.Sprintf("%s/floppy", instancePath), os.ModePerm)
		if err != nil {
			fmt.Println(err)
		}
		// write win autattend file to instance path
		err = os.WriteFile(fmt.Sprintf("%s/floppy/autounattend.xml", instancePath), []byte(instanceDef.WinAutoattend), 0644)
		if err != nil {
			fmt.Println(err)
		}
		// add floppy drive with autounattend.xml
		floppyDevice := libvirtxml.DomainDisk{
			ReadOnly: &libvirtxml.DomainDiskReadOnly{},
			Driver: &libvirtxml.DomainDiskDriver{
				Name: "qemu",
				Type: "fat",
			},
			Device: "floppy",
			Target: &libvirtxml.DomainDiskTarget{
				Dev: "fda",
				//	Bus: "fdc",
			},
			Source: &libvirtxml.DomainDiskSource{
				Dir: &libvirtxml.DomainDiskSourceDir{
					Dir: fmt.Sprintf("%s/floppy", instancePath),
				},
			},
		}
		domainDef.Devices.Disks = append(domainDef.Devices.Disks, floppyDevice)
	}

	xmldoc, err := domainDef.Marshal()
	if err != nil {
		fmt.Println(err)
	}

	// save the domain xml to a file for debugging
	fmt.Printf("Compute instance path: %s\n", instancePath)
	err = os.WriteFile(fmt.Sprintf("%s/%s.xml", instancePath, instanceDef.ID), []byte(xmldoc), 0644)
	if err != nil {
		fmt.Println(err)
	}

	// define and start the domain
	domain, err := l.DomainDefineXML(xmldoc)
	if err != nil {
		out := fmt.Sprintf("error defining libvirt domain: %s", err)
		fmt.Println(out)
	}

	errOut := l.DomainCreate(domain)
	if errOut != nil {
		fmt.Println(errOut)
	}

	if err := l.Disconnect(); err != nil {
		fmt.Printf("warning: failed to disconnect from libvirt: %v\n", err)
	}

	return instanceDef, nil
}

func RegisterInstance(instance utils.Instance) {
	// Implementation for registering an instance
	instancePath := fmt.Sprintf("/opt/nightlight/volumes/%s", instance.DatastoreId)
	instanceXMLPath := fmt.Sprintf("%s/%s.xml", instancePath, instance.ID)
	if _, err := os.Stat(instanceXMLPath); os.IsNotExist(err) {
		fmt.Printf("Instance XML file does not exist at path: %s\n", instanceXMLPath)
		return
	}

	xmlData, err := os.ReadFile(instanceXMLPath)
	if err != nil {
		fmt.Printf("Error reading instance XML file: %v\n", err)
		return
	}

	var domainDef libvirtxml.Domain
	err = xml.Unmarshal(xmlData, &domainDef)
	if err != nil {
		fmt.Printf("Error unmarshalling instance XML: %v\n", err)
		return
	}
	libvirtDomainDef, err := domainDef.Marshal()
	if err != nil {
		fmt.Printf("Error marshalling domain definition: %v\n", err)
		return
	}

	c, err := net.DialTimeout("unix", "/var/run/libvirt/libvirt-sock", 10*time.Second)
	if err != nil {
		log.Fatalf("failed to dial libvirt: %v", err)
	}

	l := libvirt.New(c)
	if err := l.Connect(); err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer l.Disconnect()

	domain, err := l.DomainDefineXML(libvirtDomainDef)
	if err != nil {
		fmt.Printf("Error defining domain in libvirt: %v\n", err)
		return
	}

	err = l.DomainCreate(domain)
	if err != nil {
		fmt.Printf("Error starting domain in libvirt: %v\n", err)
		return
	}

	fmt.Printf("Instance %s registered and started successfully.\n", instance.ID)
}

func GetVNCPort(vmId string) (int, error) {
	c, err := net.DialTimeout("unix", "/var/run/libvirt/libvirt-sock", 10*time.Second)
	if err != nil {
		return 0, fmt.Errorf("failed to dial libvirt: %v", err)
	}
	defer c.Close()

	l := libvirt.New(c)
	if err := l.Connect(); err != nil {
		return 0, fmt.Errorf("failed to connect: %v", err)
	}
	defer l.Disconnect()

	dom, err := l.DomainLookupByName(vmId)
	if err != nil {
		return 0, fmt.Errorf("failed to lookup domain: %v", err)
	}

	xmlDesc, err := l.DomainGetXMLDesc(dom, 0)
	if err != nil {
		return 0, fmt.Errorf("failed to get domain XML: %v", err)
	}

	var domain libvirtxml.Domain
	if err := xml.Unmarshal([]byte(xmlDesc), &domain); err != nil {
		return 0, fmt.Errorf("failed to unmarshal domain XML: %v", err)
	}

	for _, g := range domain.Devices.Graphics {
		if g.VNC != nil && g.VNC.Port != -1 {
			fmt.Println("VNC port found:", g.VNC.Port)
			return g.VNC.Port, nil
		}
	}

	return 0, fmt.Errorf("VNC port not found for domain %s", vmId)
}

func DeleteVM(vmId string, datastorePath string) {
	c, err := net.DialTimeout("unix", "/var/run/libvirt/libvirt-sock", 10*time.Second)
	if err != nil {
		log.Fatalf("failed to dial libvirt: %v", err)
	}

	l := libvirt.New(c)
	if err := l.Connect(); err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	dom, err := l.DomainLookupByName(vmId)
	if err != nil {
		fmt.Println(err)
	}
	destroyErr := l.DomainDestroy(dom)
	if destroyErr != nil {
		fmt.Println(destroyErr)
	}
	l.DomainUndefineFlags(dom, libvirt.DomainUndefineManagedSave)
	vmPath := fmt.Sprintf("%s/%s", datastorePath, vmId)

	log.Printf("Deleting virtual machine: %s", vmPath)
	removeErr := os.RemoveAll(vmPath)
	if removeErr != nil {
		fmt.Println(removeErr)
	}
}

// libvirtDomainState maps a libvirt domain state integer to a human-readable
// string consistent with the PowerState field on utils.Instance.
//
// Libvirt state values (virDomainState):
//
//	0 = no state, 1 = running, 2 = blocked, 3 = paused,
//	4 = shutting down, 5 = shut off, 6 = crashed, 7 = PM suspended
func libvirtDomainState(state int32) string {
	switch state {
	case 1:
		return "running"
	case 3:
		return "paused"
	case 4:
		return "shutting-down"
	case 5:
		return "stopped"
	case 6:
		return "crashed"
	default:
		return "unknown"
	}
}

// dialLibvirt opens a single libvirt connection.  Returns nil on failure so
// callers can treat a missing libvirt socket as "all domains unknown".
func dialLibvirt() *libvirt.Libvirt {
	c, err := net.DialTimeout("unix", "/var/run/libvirt/libvirt-sock", 5*time.Second)
	if err != nil {
		return nil
	}
	l := libvirt.New(c)
	if err := l.Connect(); err != nil {
		return nil
	}
	return l
}

// ListVMPowerStates opens one libvirt connection and returns the live power
// state for each VM ID in the provided slice.  Unknown or unreachable VMs
// default to "stopped".
func ListVMPowerStates(vmIds []string) map[string]string {
	states := make(map[string]string, len(vmIds))
	for _, id := range vmIds {
		states[id] = "unknown"
	}

	l := dialLibvirt()
	if l == nil {
		return states
	}

	for _, vmId := range vmIds {
		dom, err := l.DomainLookupByName(vmId)
		if err != nil {
			states[vmId] = "stopped"
			continue
		}
		state, _, err := l.DomainGetState(dom, 0)
		if err != nil {
			states[vmId] = "unknown"
			continue
		}
		states[vmId] = libvirtDomainState(state)
	}
	return states
}

// GetVMPowerState returns the live power state for a single VM.
func GetVMPowerState(vmId string) string {
	l := dialLibvirt()
	if l == nil {
		return "unknown"
	}
	dom, err := l.DomainLookupByName(vmId)
	if err != nil {
		return "stopped"
	}
	state, _, err := l.DomainGetState(dom, 0)
	if err != nil {
		return "unknown"
	}
	return libvirtDomainState(state)
}

func ShutdownVM(vmId string) {
	c, err := net.DialTimeout("unix", "/var/run/libvirt/libvirt-sock", 10*time.Second)
	if err != nil {
		log.Fatalf("failed to dial libvirt: %v", err)
	}

	l := libvirt.New(c)
	if err := l.Connect(); err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	dom, err := l.DomainLookupByName(vmId)
	if err != nil {
		fmt.Println(err)
	}
	shutdownErr := l.DomainShutdown(dom)
	if err != nil {
		fmt.Println(shutdownErr)
	}
}

func RestartVM(vmId string) {
	c, err := net.DialTimeout("unix", "/var/run/libvirt/libvirt-sock", 10*time.Second)
	if err != nil {
		log.Fatalf("failed to dial libvirt: %v", err)
	}

	l := libvirt.New(c)
	if err := l.Connect(); err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	v, err := l.Version()
	if err != nil {
		log.Fatalf("failed to retrieve libvirt version: %v", err)
	}
	fmt.Println("Version:", v)
	dom, err := l.DomainLookupByName(vmId)
	if err != nil {
		fmt.Println(err)
	}
	var rebootFlags libvirt.DomainRebootFlagValues
	rebootErr := l.DomainReboot(dom, rebootFlags)
	if err != nil {
		fmt.Println(rebootErr)
	}
}

func ResetVM(vmId string) {
	c, err := net.DialTimeout("unix", "/var/run/libvirt/libvirt-sock", 10*time.Second)
	if err != nil {
		log.Fatalf("failed to dial libvirt: %v", err)
	}

	l := libvirt.New(c)
	if err := l.Connect(); err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	dom, err := l.DomainLookupByName(vmId)
	if err != nil {
		fmt.Println(err)
	}
	var resetFlags libvirt.DomainResetArgs
	resetErr := l.DomainReset(dom, resetFlags.Flags)
	if err != nil {
		fmt.Println(resetErr)
	}
}

func StartVM(vmId string) {
	c, err := net.DialTimeout("unix", "/var/run/libvirt/libvirt-sock", 10*time.Second)
	if err != nil {
		log.Fatalf("failed to dial libvirt: %v", err)
	}

	l := libvirt.New(c)
	if err := l.Connect(); err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	v, err := l.Version()
	if err != nil {
		log.Fatalf("failed to retrieve libvirt version: %v", err)
	}
	fmt.Println("Version:", v)
	dom, err := l.DomainLookupByName(vmId)
	if err != nil {
		fmt.Println(err)
	}

	if err := l.DomainCreate(dom); err != nil {
		panic(err)
	}
}

func StopVM(vmId string) {
	c, err := net.DialTimeout("unix", "/var/run/libvirt/libvirt-sock", 10*time.Second)
	if err != nil {
		log.Fatalf("failed to dial libvirt: %v", err)
	}

	l := libvirt.New(c)
	if err := l.Connect(); err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	dom, err := l.DomainLookupByName(vmId)
	if err != nil {
		fmt.Println(err)
	}
	if err := l.DomainDestroy(dom); err != nil {
		panic(err)
	}
}

func AttachCDROM(vmId string, filePath string) {
	c, err := net.DialTimeout("unix", "/var/run/libvirt/libvirt-sock", 10*time.Second)
	if err != nil {
		log.Fatalf("failed to dial libvirt: %v", err)
	}

	l := libvirt.New(c)
	if err := l.Connect(); err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	dom, err := l.DomainLookupByName(vmId)
	if err != nil {
		fmt.Println(err)
	}
	attachErr := l.DomainAttachDevice(dom, "")
	if attachErr != nil {
		fmt.Println(attachErr)
	}
}

func GetVM(vmId string) {
	c, err := net.DialTimeout("unix", "/var/run/libvirt/libvirt-sock", 10*time.Second)
	if err != nil {
		log.Fatalf("failed to dial libvirt: %v", err)
	}

	l := libvirt.New(c)
	if err := l.Connect(); err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	_, err = l.DomainLookupByName(vmId)
	if err != nil {
		fmt.Println(err)
	}
}

func SendConsoleKeyEvent(vmId string, keycodes []uint32) {
	c, err := net.DialTimeout("unix", "/var/run/libvirt/libvirt-sock", 10*time.Second)
	if err != nil {
		log.Fatalf("failed to dial libvirt: %v", err)
	}

	l := libvirt.New(c)
	if err := l.Connect(); err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	dom, err := l.DomainLookupByName(vmId)
	if err != nil {
		fmt.Println(err)
	}

	// create a uint32 slice with the keycode
	//keycodes := []uint32{keycode}

	sendErr := l.DomainSendKey(dom, uint32(libvirt.KeycodeSetUsb), 150, keycodes, 0)
	if sendErr != nil {
		fmt.Println(sendErr)
	}
}

// TerraformInstanceXML type
type TerraformInstanceXML struct {
	XMLName xml.Name          `xml:"https://terraform.io ovn"`
	Tags    []TerraformTagXML `xml:"tag"`
}

// TerraformTagXML type
type TerraformTagXML struct {
	Key   string `xml:"name,attr"`
	Value string `xml:",chardata"`
}

func tagsToXML(tags map[string]interface{}, metadata *TerraformInstanceXML) (out string, err error) {
	// Overwrite existing tags while keeping additional metadata
	metadata.Tags = []TerraformTagXML{}
	for key, value := range tags {
		metadata.Tags = append(metadata.Tags, TerraformTagXML{
			Key:   key,
			Value: value.(string),
		})
	}
	var bytesOut []byte
	if bytesOut, err = xml.Marshal(metadata); err != nil {
		return "", fmt.Errorf("Failed to marshal metadata XML: %s", err)
	}
	return string(bytesOut), nil
}

func randomMACAddress() (string, error) {
	buf := make([]byte, 3)
	//nolint:gosec // math.rand is enough for this
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	// set local bit and unicast
	buf[0] = (buf[0] | 2) & 0xfe
	// Set the local bit
	buf[0] |= 2

	// avoid libvirt-reserved addresses
	if buf[0] == 0xfe {
		buf[0] = 0xee
	}

	return fmt.Sprintf("52:54:00:%02x:%02x:%02x",
		buf[0], buf[1], buf[2]), nil
}

func CreateDiskImage(imagePath string, sizeGB int) error {
	cmd := fmt.Sprintf("qemu-img create -f qcow2 %s %dG", imagePath, sizeGB)
	runcmd := strings.Split(cmd, " ")
	out := exec.Command(runcmd[0], runcmd[1:]...)
	err := out.Run()
	if err != nil {
		return err
	}
	return nil
}

func ConvertDiskImage(srcPath string, destPath string, sizeGB int) error {
	cmd := fmt.Sprintf("qemu-img convert -f qcow2 -O qcow2 %s %s", srcPath, destPath)
	runcmd := strings.Split(cmd, " ")
	out := exec.Command(runcmd[0], runcmd[1:]...)
	err := out.Run()
	if err != nil {
		return err
	}

	resizecmd := fmt.Sprintf("qemu-img resize -f qcow2 %s %dG", destPath, sizeGB)
	runcmd = strings.Split(resizecmd, " ")
	out = exec.Command(runcmd[0], runcmd[1:]...)
	err = out.Run()
	if err != nil {
		return err
	}

	return nil
}
