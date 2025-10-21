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

func CreateVM(instanceDef utils.Instance, instancePath string) (macAddress string) {
	c, err := net.DialTimeout("unix", "/var/run/libvirt/libvirt-sock", 10*time.Second)
	if err != nil {
		log.Fatalf("failed to dial libvirt: %v", err)
	}

	l := libvirt.New(c)
	if err := l.Connect(); err != nil {
		log.Fatalf("failed to connect: %v", err)
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
	cpu, memory := instanceDef.CPUSockets, instanceDef.MemoryMB
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
			Value:     uint(cpu),
		},
		CPU: &libvirtxml.DomainCPU{},
		Devices: &libvirtxml.DomainDeviceList{
			/*			Disks: []libvirtxml.DomainDisk{
							{
								Boot: &libvirtxml.DomainDeviceBoot{
									Order: 1,
								},
								Driver: &libvirtxml.DomainDiskDriver{
									Name: "qemu",
									Type: "qcow2",
								},
								Device: "disk",
								Target: &libvirtxml.DomainDiskTarget{
									Dev: "vda",
									Bus: "virtio",
								},
								Source: &libvirtxml.DomainDiskSource{
									File: &libvirtxml.DomainDiskSourceFile{
										File: imagePath,
									},
								},
							},
						},
			*/
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
	domainDef.OS.Type.Machine = "pc-q35-6.2"

	mac, err := randomMACAddress()
	if err != nil {
		fmt.Errorf("error generating mac address: %w", err)
	}

	// Add network interfaces
	nics := instanceDef.Devices.NetworkInterfaces
	for _, nic := range nics {
		netIface := libvirtxml.DomainInterface{
			VirtualPort: &libvirtxml.DomainInterfaceVirtualPort{
				Params: &libvirtxml.DomainInterfaceVirtualPortParams{
					OpenVSwitch: &libvirtxml.DomainInterfaceVirtualPortParamsOpenVSwitch{},
				},
			},
			Model: &libvirtxml.DomainInterfaceModel{
				Type: nic.Model,
			},
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
		if disk.BusType == "virtio" {
			diskTarget = virtioDisks[virtioIndex]
			virtioIndex++
		} else if disk.BusType == "sata" {
			diskTarget = sataDisks[sataIndex]
			sataIndex++
		} else {
			fmt.Errorf("unsupported bus type: %s", disk.BusType)
			continue
		}
		diskPath := fmt.Sprintf("%s/%s_disk_%s.qcow2", instancePath, instanceDef.ID, diskTarget)
		if disk.ExistingPath != "" {
			// Copy existing disk image
			copyFile(disk.ExistingPath, diskPath)
		} else {
			// Create new disk image
			err := createDiskImage(diskPath, disk.SizeGB)
			if err != nil {
				fmt.Errorf("error creating disk image: %w", err)
				continue
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
					File: diskPath,
				},
			},
		}
		domainDef.Devices.Disks = append(domainDef.Devices.Disks, storageDisk)
	}

	cdroms := instanceDef.Devices.CDROMs
	for _, cd := range cdroms {
		cdromDevice := libvirtxml.DomainDisk{
			Boot: &libvirtxml.DomainDeviceBoot{
				Order: uint(cd.BootOrder),
			},
			Device: "cdrom",
			Target: &libvirtxml.DomainDiskTarget{
				Dev: "hdc",
				Bus: "ide",
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

	xmldoc, err := domainDef.Marshal()
	if err != nil {
		fmt.Println(err)
	}

	// save the domain xml to a file for debugging
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
		log.Fatalf("failed to disconnect: %v", err)
	}
	return mac
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
	var rebootFlags libvirt.DomainRebootFlagValues
	destroyErr := l.DomainReboot(dom, rebootFlags)
	if err != nil {
		fmt.Println(destroyErr)
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
	if err != nil {
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

func SendConsoleKeyEvent(vmId string, keycode uint32) {
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
	sendErr := l.DomainSendKey(dom, libvirt.DOMAIN_SEND_KEY_FLAGS_RELEASE, []uint32{keycode}, 1, 0)
	if err != nil {
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

func createDiskImage(imagePath string, sizeGB int) error {
	cmd := fmt.Sprintf("qemu-img create -f qcow2 %s %dG", imagePath, sizeGB)
	fmt.Println(cmd)
	out := exec.Command(cmd)
	err := out.Run()
	if err != nil {
		return err
	}
	return nil
}
