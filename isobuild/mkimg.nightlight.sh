#!/bin/sh
profile_nightlight() {
    profile_standard
    # Customize kernel options, e.g., for specific hardware or boot parameters
    kernel_cmdline="console=tty0 console=ttyS0,115200 intel_iommu=on iommu=pt"
    kernel_addons="" # Example: Add ZFS support
	kernel_flavors="virt"

    # Add desired packages
    apks="$apks alpine-base supervisor iproute2 tcpdump openvswitch libvirt-daemon qemu-hw-usb-host qemu-img qemu-system-x86_64 qemu-system-i386 ovmf qemu-modules openrc libvirt openssh swtpm edk2"

    # Specify your custom overlay script
    apkovl="aports/scripts/genapkovl-mkimgoverlay.sh"
}