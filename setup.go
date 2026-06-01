package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/hashicorp/go-hclog"
)

func formatAndMount(device, mountPoint string) error {
	if err := os.MkdirAll(mountPoint, 0755); err != nil {
		return err
	}

	if err := exec.Command("mkfs.ext4", device).Run(); err != nil {
		return err
	}

	if err := exec.Command("mount", device, mountPoint).Run(); err != nil {
		return err
	}

	return nil
}

func mountDisk(device, mountPoint string) error {
	if err := os.MkdirAll(mountPoint, 0755); err != nil {
		return err
	}

	if err := exec.Command("mount", device, mountPoint).Run(); err != nil {
		return err
	}

	return nil
}

func checkDiskFormatted(diskName string) bool {
	// Run blkid command to get filesystem info
	cmd := exec.Command("blkid")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Error executing blkid command: %s", err)
	}

	// Convert output to string and check for the disk name and filesystem
	outputStr := string(output)
	if strings.Contains(outputStr, diskName) {
		// Check if the disk has a filesystem (FSTYPE)
		if strings.Contains(outputStr, "TYPE=") {
			return true
		}
	}
	return false
}

// check if a disk is mounted at a given mount point
func isDiskMounted(mountPoint string) bool {
	cmd := exec.Command("mountpoint", "-q", mountPoint)
	err := cmd.Run()
	return err == nil
}

func initialSetup() {
	diskStatus := checkDiskFormatted("/dev/sda")
	if diskStatus {
		hclog.Default().Named("core").Info("Disk /dev/sda is already formatted.")
		if isDiskMounted("/opt/nightlight") {
			hclog.Default().Named("core").Info("Disk /dev/sda is already mounted at /opt/nightlight.")
			return
		}
		hclog.Default().Named("core").Info("Disk /dev/sda is not mounted. Attempting to mount...")
		err := mountDisk("/dev/sda", "/opt/nightlight")
		if err != nil {
			hclog.Default().Named("core").Error(fmt.Sprintf("Error mounting disk: %v", err))
		} else {
			hclog.Default().Named("core").Info("Successfully mounted disk to /opt/nightlight")
		}
	} else {
		hclog.Default().Named("core").Info("Disk /dev/sda is not formatted. Formatting and mounting...")
		// Create necessary directories
		dirs := []string{"/opt/nightlight"}
		for _, dir := range dirs {
			err := os.MkdirAll(dir, 0755)
			if err != nil {
				hclog.Default().Named("core").Error(fmt.Sprintf("Error creating directory %s: %v", dir, err))
			} else {
				hclog.Default().Named("core").Info(fmt.Sprintf("Successfully created directory: %s", dir))
			}
		}
		// Format and mount the disk
		err := formatAndMount("/dev/sda", "/opt/nightlight")
		if err != nil {
			hclog.Default().Named("core").Error(fmt.Sprintf("Error formatting and mounting disk: %v", err))
		} else {
			hclog.Default().Named("core").Info("Successfully formatted and mounted disk to /opt/nightlight")
		}
		// write the current version to a file
		err = os.WriteFile("/opt/nightlight/version.json", []byte(`{"version":"0.0.1"}`), 0644)
		if err != nil {
			hclog.Default().Named("core").Error(fmt.Sprintf("Error writing version file: %v", err))
		} else {
			hclog.Default().Named("core").Info("Successfully wrote version file")
		}
	}
}
