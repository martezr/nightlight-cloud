package network

import (
	"fmt"
	"net"
	"runtime"

	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

func CreateNetworkNamespace(name string, macAddress string, ipAddress string) error {
	// Lock the OS Thread so we don't accidentally switch namespaces
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// check if namespace already exists
	nsList, err := netns.GetFromName(name)
	if err == nil {
		fmt.Printf("Network namespace %s already exists\n", name)
		nsList.Close()
		return nil
	}

	// Save the current network namespace
	origns, _ := netns.Get()
	defer origns.Close()

	// Create a new network namespace
	newns, _ := netns.NewNamed(name)

	// Switch back to the original namespace and add interface to the new namespace
	netns.Set(origns)

	netlinkLink, err := netlink.LinkByName(name)
	if err != nil {
		return fmt.Errorf("error getting link %s: %v", name, err)
	}

	err = netlink.LinkSetNsFd(netlinkLink, int(newns))
	if err != nil {
		return fmt.Errorf("error setting ns fd: %v", err)
	}

	// Switch to the new namespace
	netns.Set(newns)

	// Do something with the network namespace
	ifaces, _ := net.Interfaces()
	fmt.Printf("Interfaces: %v\n", ifaces)

	namespaceLink, err := netlink.LinkByName(name)
	if err != nil {
		return fmt.Errorf("error getting link in new ns: %v", err)
	}

	// parse the provided MAC address
	hwAddr, err := net.ParseMAC(macAddress)
	if err != nil {
		return fmt.Errorf("invalid MAC address %q: %v", macAddress, err)
	}

	// set the hardware address on the link
	if err := netlink.LinkSetHardwareAddr(namespaceLink, hwAddr); err != nil {
		return fmt.Errorf("error setting hardware address: %v", err)
	}

	// set the link's IP address
	addr := &netlink.Addr{IPNet: &net.IPNet{
		IP:   net.ParseIP(ipAddress),
		Mask: net.CIDRMask(24, 32),
	}}
	err = netlink.AddrAdd(namespaceLink, addr)
	if err != nil {
		return fmt.Errorf("error adding address to link in new ns: %v", err)
	}

	err = netlink.LinkSetUp(namespaceLink)
	if err != nil {
		return fmt.Errorf("error setting link up in new ns: %v", err)
	}

	// set default route via interface (use nil Dst for default route)
	route := &netlink.Route{
		LinkIndex: namespaceLink.Attrs().Index,
		Gw:        nil,
		Dst:       &net.IPNet{IP: net.IPv4zero, Mask: net.CIDRMask(0, 32)},
	}
	if err := netlink.RouteAdd(route); err != nil {
		return fmt.Errorf("error adding route in new ns: %v", err)
	}

	for _, iface := range ifaces {
		fmt.Printf("Interface: %v\n", iface.Name)
	}

	newns.Close()
	return nil
}
