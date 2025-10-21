package main

// upload iso file to datastore using govmomi library
import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/units"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"
)

type extraConfig []types.BaseOptionValue

func main() {
	ctx := context.Background()

	// Set these variables as needed
	vsphereServer := "grtvcenter01.grt.local"
	vsphereUsername := "administrator@vsphere.local"
	vspherePassword := "Password123#"
	datacenterName := "GRT"
	datastoreName := "Local"
	isoLocalPath := "../isobuild/iso/alpine-nightlight-v3.22-x86_64.iso"
	isoDatastorePath := "ISO/alpine-nightlight-v3.22-x86_64.iso"

	vcenterURL, err := url.Parse(fmt.Sprintf("https://%v/sdk", vsphereServer))
	if err != nil {
		log.Println(err)
	}
	credentials := url.UserPassword(vsphereUsername, vspherePassword)
	vcenterURL.User = credentials

	c, err := govmomi.NewClient(ctx, vcenterURL, true)
	if err != nil {
		fmt.Println("vCenter connection error:", err)
		return
	}
	defer c.Logout(ctx)

	// Find datacenter and datastore
	finder := find.NewFinder(c.Client, true)
	dc, err := finder.Datacenter(ctx, datacenterName)
	if err != nil {
		fmt.Println("Datacenter error:", err)
		return
	}
	finder.SetDatacenter(dc)
	ds, err := finder.Datastore(ctx, datastoreName)
	if err != nil {
		fmt.Println("Datastore error:", err)
		return
	}

	// Delete existing VM if it exists
	deleteVM(c, ctx, finder, "nldemo")

	// Upload ISO
	err = uploadISO(ctx, ds, isoLocalPath, isoDatastorePath)
	if err != nil {
		fmt.Println("Upload error:", err)
		return
	}
	fmt.Println("ISO uploaded successfully")

	// Create VM
	createVM(c, ctx, finder)
}

func uploadISO(ctx context.Context, ds *object.Datastore, isoPath, dsPath string) error {
	f, err := os.Open(isoPath)
	if err != nil {
		return err
	}
	defer f.Close()

	p := soap.DefaultUpload
	err = ds.Upload(ctx, f, dsPath, &p)
	if err != nil {
		return err
	}
	return nil
}

func createVM(client *govmomi.Client, ctx context.Context, finder *find.Finder) {
	var devices object.VirtualDeviceList
	pool, err := finder.ResourcePoolOrDefault(ctx, "")
	if err != nil {
		log.Println(err)
	}

	host, err := finder.HostSystemOrDefault(ctx, "")
	if err != nil {
		log.Println(err)
	}

	folder, err := finder.FolderOrDefault(ctx, "")
	if err != nil {
		log.Println(err)
	}

	spec := &types.VirtualMachineConfigSpec{
		Name:            "nldemo",
		NumCPUs:         4,
		MemoryMB:        8192,
		Annotation:      "nightlight",
		NestedHVEnabled: types.NewBool(true),
		Firmware:        string(types.GuestOsDescriptorFirmwareTypeBios),
		Version:         "vmx-19",
		GuestId:         string(types.VirtualMachineGuestOsIdentifierOther5xLinux64Guest),
		Files: &types.VirtualMachineFileInfo{
			VmPathName: "[Local]",
		},
	}

	var settings extraConfig
	settings = append(settings, &types.OptionValue{Key: "guestinfo.rms.ipaddress", Value: "10.0.0.33"})
	settings = append(settings, &types.OptionValue{Key: "guestinfo.rms.mask", Value: "255.255.255.0"})
	settings = append(settings, &types.OptionValue{Key: "guestinfo.rms.gw", Value: "10.0.0.1"})
	authSpec := types.VirtualMachineConfigSpec{
		ExtraConfig: settings,
	}
	spec.ExtraConfig = authSpec.ExtraConfig
	network, err := finder.NetworkOrDefault(ctx, "VM Network")
	if err != nil {
		log.Println(err)
	}

	backing, err := network.EthernetCardBackingInfo(context.TODO())
	if err != nil {
		log.Println(err)
	}

	device, err := object.EthernetCardTypes().CreateEthernetCard("vmxnet3", backing)
	if err != nil {
		log.Println(err)
	}

	devices = append(devices, device)

	// Add storage controller
	scsi, err := devices.CreateSCSIController("scsi")
	if err != nil {
		log.Println(err)
	}
	devices = append(devices, scsi)

	controller, err := devices.FindDiskController("scsi")
	if err != nil {
		log.Println(err)
	}

	var b units.ByteSize
	b.Set("10GB")

	disk := &types.VirtualDisk{
		VirtualDevice: types.VirtualDevice{
			Key: devices.NewKey(),
			Backing: &types.VirtualDiskFlatVer2BackingInfo{
				DiskMode:        string(types.VirtualDiskModePersistent),
				ThinProvisioned: types.NewBool(true),
			},
		},
		CapacityInKB: int64(b) / 1024,
	}

	devices.AssignController(disk, controller)
	devices = append(devices, disk)

	// Add CD-ROM device with ISO

	// create an ide controller
	ide, err := devices.CreateIDEController()
	if err != nil {
		log.Println(err)
	}
	devices = append(devices, ide)

	// find ide controller
	ideController, err := devices.FindIDEController("")
	if err != nil {
		log.Println(err)
	}

	cdrom, err := devices.CreateCdrom(ideController)
	if err != nil {
		log.Println(err)
	}

	//devices.AssignController(cdrom, ideController)
	devices.InsertIso(cdrom, "[Local] ISO/alpine-nightlight-v3.22-x86_64.iso")

	devices = append(devices, cdrom)

	deviceChange, err := devices.ConfigSpec(types.VirtualDeviceConfigSpecOperationAdd)
	if err != nil {
		log.Println(err)
	}

	spec.DeviceChange = deviceChange

	task, err := folder.CreateVM(ctx, *spec, pool, host)
	if err != nil {
		log.Println(err)
	}

	info, err := task.WaitForResult(ctx)
	if err != nil {
		log.Println(err)
	}

	vm := object.NewVirtualMachine(client.Client, info.Result.(types.ManagedObjectReference))
	name, err := vm.ObjectName(ctx)
	if err != nil {
		log.Println(err)
	}
	log.Println("Created VM:", name)

	task, err = vm.PowerOn(ctx)
	if err != nil {
		log.Println(err)
	}

	_, err = task.WaitForResult(ctx)
	if err != nil {
		log.Println(err)
	}

	log.Println("VM created: ", info.Result.(types.ManagedObjectReference).Value)
}

// find virtual machine by name and delete it
func deleteVM(client *govmomi.Client, ctx context.Context, finder *find.Finder, vmName string) {
	vm, err := finder.VirtualMachine(ctx, vmName)
	if err != nil {
		log.Println(err)
		return
	}

	// check if vm is powered on, if so power it off
	state, err := vm.PowerState(ctx)
	if err != nil {
		log.Println(err)
		return
	}

	if state == types.VirtualMachinePowerStatePoweredOn {

		task, err := vm.PowerOff(ctx)
		if err != nil {
			log.Println(err)
			return
		}

		err = task.Wait(ctx)
		if err != nil {
			log.Println(err)
			return
		}
	}

	task, err := vm.Destroy(ctx)
	if err != nil {
		log.Println(err)
		return
	}

	err = task.Wait(ctx)
	if err != nil {
		log.Println(err)
		return
	}

	log.Println("VM deleted: ", vmName)
}
