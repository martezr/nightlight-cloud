package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/martezr/go-openvswitch/ovs"
)

func main() {
	baseConfiguration()
	SetupBaseNetworking()
	//configureDefaultNetworking()
}

func baseConfiguration() {
	// make /opt/nightlight directory
	os.MkdirAll("/opt/nightlight/volumes", 0755)

	err := setHostname("nightlight-cloud")
	if err != nil {
		log.Fatalf("Error setting hostname: %v", err)
	}
	err = setRootPassword("nightlight")
	if err != nil {
		log.Fatalf("Error setting root password: %v", err)
	}
	err = configureRootSSH()
	if err != nil {
		log.Fatalf("Error configuring root SSH: %v", err)
	}

	err = enableSSHPasswordAuth()
	if err != nil {
		log.Fatalf("Error enabling SSH password authentication: %v", err)
	}

	UpdateUEFIImages()
}

// Set the system hostname
func setHostname(hostname string) error {
	err := os.WriteFile("/etc/hostname", []byte(hostname), 0644)
	if err != nil {
		return err
	}

	// Set the hostname immediately
	err = os.WriteFile("/proc/sys/kernel/hostname", []byte(hostname), 0644)
	if err != nil {
		return err
	}

	return nil
}

// set root password
func setRootPassword(password string) error {
	cmd := exec.Command("sh", "-c", fmt.Sprintf("echo 'root:%s' | chpasswd", password))
	return cmd.Run()
}

// configure root ssh access
func configureRootSSH() error {
	sshDir := "/root/.ssh"
	err := os.MkdirAll(sshDir, 0700)
	if err != nil {
		return err
	}

	// For demo purposes, we use a hardcoded public key. In production, consider generating a new key pair or using a secure method.
	publicKey := ""
	err = os.WriteFile(sshDir+"/authorized_keys", []byte(publicKey), 0600)
	if err != nil {
		return err
	}

	return nil
}

// enable ssh password authentication
func enableSSHPasswordAuth() error {
	// check if /etc/ssh/sshd_config.d/10-nightlight.conf exists, if not create it
	if _, err := os.Stat("/etc/ssh/sshd_config.d/10-nightlight.conf"); os.IsNotExist(err) {
		err = os.WriteFile("/etc/ssh/sshd_config.d/10-nightlight.conf", []byte("PasswordAuthentication yes\nPermitRootLogin yes\n"), 0644)
		if err != nil {
			return err
		}
		restartErr := restartSSHService()
		if restartErr != nil {
			return restartErr
		}
		return nil
	}
	return nil
}

// restart ssh service
func restartSSHService() error {
	cmd := exec.Command("sh", "-c", "rc-service sshd restart")
	return cmd.Run()
}

func configureDefaultNetworking() {
	// Create dhcp network namespace and OVS interface
	ovsClient := ovs.New()
	ovsClient.VSwitch.AddPort("nl-transit", "dhdefaultvnet")
	ovsClient.VSwitch.Set.Interface("dhdefaultvnet", ovs.InterfaceOptions{
		Type: "internal",
		ExternalIds: map[string]string{
			"iface-id":     "dhdefaultvnet",
			"attached-mac": "32:6b:ce:89:41:43",
		},
	})

	time.Sleep(1 * time.Second) // wait for OVS to create the interface before trying to use it

	err := CreateNetworkNamespace("dhdefaultvnet", "32:6b:ce:89:41:43", "169.254.169.253")
	if err != nil {
		log.Fatalf("Error creating network namespace: %v", err)
	}
	err = startService("dhcpagent")
	if err != nil {
		log.Fatalf("Error starting DHCP agent service: %v", err)
	}

	// Create metadata network namespace and OVS interface
	ovsClient.VSwitch.AddPort("nl-transit", "mddefaultvnet")
	ovsClient.VSwitch.Set.Interface("mddefaultvnet", ovs.InterfaceOptions{
		Type: "internal",
		ExternalIds: map[string]string{
			"iface-id":     "mddefaultvnet",
			"attached-mac": "32:6b:ce:89:41:42",
		},
	})

	time.Sleep(1 * time.Second) // wait for OVS to create the interface before trying to use it

	err = CreateNetworkNamespace("mddefaultvnet", "32:6b:ce:89:41:42", "169.254.169.254")
	if err != nil {
		log.Fatalf("Error creating network namespace: %v", err)
	}
	err = startService("metadataagent")
	if err != nil {
		log.Fatalf("Error starting metadata agent service: %v", err)
	}

	//
	// Create flow monitor network namespace and OVS interface
	//
	time.Sleep(5 * time.Second) // wait for OVS to be ready before adding mirror and port
	ovsClient.VSwitch.AddPort("nl-transit", "fmdefaultvnet")
	ovsClient.VSwitch.Set.Interface("fmdefaultvnet", ovs.InterfaceOptions{
		Type: "internal",
		ExternalIds: map[string]string{
			"iface-id":     "fmdefaultvnet",
			"attached-mac": "32:6b:ce:89:41:44",
		},
	})

	err = CreateNetworkNamespace("fmdefaultvnet", "32:6b:ce:89:41:44", "169.254.169.252")
	if err != nil {
		log.Fatalf("Error creating network namespace: %v", err)
	}
	err = startService("flowmonitoragent")
	if err != nil {
		log.Fatalf("Error starting flow monitor agent service: %v", err)
	}

	time.Sleep(20 * time.Second) // wait for OVS to create the interface before trying to use it

	c := ovs.New()
	ports, err := c.VSwitch.ListPorts("nl-transit")
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}
	var metadataOfPort int

	c.VSwitch.AddMirror("nl-transit", ovs.MirrorOptions{
		Name:       "mirror_fmdefaultvnet",
		SelectAll:  true,
		OutputPort: "fmdefaultvnet",
	})

	for _, port := range ports {
		portDetails, err := c.VSwitch.Get.Port(port)
		fmt.Printf("Existing port: %s - Port Number: %d\n", port, portDetails.OFPort)
		if err != nil {
			hclog.Default().Named("core").Error(err.Error())
		}
		if portDetails.Name == "mddefaultvnet" {
			metadataOfPort, err = strconv.Atoi(portDetails.OFPort)
			if err != nil {
				hclog.Default().Named("core").Error(err.Error())
			}
			continue
		}
	}

	InstallDefaultFlows("nl-transit", metadataOfPort)
}

func UpdateUEFIImages() {
	text := `
	{
    "description": "UEFI firmware for x86_64, with Secure Boot and SMM",
    "interface-types": [
        "uefi"
    ],
    "mapping": {
        "device": "flash",
        "executable": {
            "filename": "/etc/OVMF_CODE_4M.ms.fd",
            "format": "raw"
        },
        "nvram-template": {
            "filename": "/etc/OVMF_VARS_4M.ms.fd",
            "format": "raw"
        }
    },
    "targets": [
        {
            "architecture": "x86_64",
            "machines": [
                "pc-q35-*"
            ]
        }
    ],
    "features": [
        "acpi-s3",
        "amd-sev",
        "requires-smm",
        "secure-boot",
        "verbose-dynamic"
    ],
    "tags": [

    ]
}`
	os.WriteFile("/usr/share/qemu/firmware/50-edk2-x86_64-secure.json", []byte(text), 0644)

}

// start openrc service
func startService(service string) error {
	cmd := exec.Command("sh", "-c", fmt.Sprintf("rc-service %s start", service))
	return cmd.Run()
}
