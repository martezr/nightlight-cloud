#!/bin/sh
profile_nightlight() {
    profile_standard
    # Customize kernel options, e.g., for specific hardware or boot parameters
    #kernel_cmdline="console=tty0 console=ttyS0,115200 intel_iommu=on iommu=pt"
    kernel_addons="" # Example: Add ZFS support
	kernel_flavors="lts"
    #kernel_cmdline="quiet loglevel=0 console=tty0"
    kernel_cmdline="quiet console=tty0 nomodeset alpine_devtmpfs=1"
    # Add desired packages
#    apks="$apks alpine-base linux-virt openvswitch libvirt-daemon qemu-img qemu ovmf qemu-modules openrc libvirt openssh swtpm edk2 nfs-utils e2fsprogs"
    apks="$apks alpine-base linux-lts openvswitch libvirt libvirt-daemon qemu-img ovmf swtpm edk2 openrc openssh nfs-utils e2fsprogs"

    # Specify your custom overlay script
    apkovl="nightlight"
}