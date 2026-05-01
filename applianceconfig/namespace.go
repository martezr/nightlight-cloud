package main

import (
	"fmt"
	"net"
	"runtime"
	"time"

	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

func CreateNetworkNamespace(name string, macAddress string, ipAddress string) error {
	// 🔒 Lock to a single OS thread (namespaces are thread-local)
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// Save original namespace
	origns, err := netns.Get()
	if err != nil {
		return fmt.Errorf("failed to get current namespace: %v", err)
	}
	defer origns.Close()

	// ALWAYS restore original namespace
	defer func() {
		if err := netns.Set(origns); err != nil {
			fmt.Printf("failed to restore original namespace: %v\n", err)
		}
	}()

	// Check if namespace exists
	if _, err := netns.GetFromName(name); err == nil {
		fmt.Printf("Network namespace %s already exists\n", name)
		return nil
	}

	fmt.Println("Creating network namespace:", name)

	// Create new namespace
	newns, err := netns.NewNamed(name)
	if err != nil {
		return fmt.Errorf("failed to create namespace: %v", err)
	}
	defer newns.Close()

	// Switch back to original namespace before moving interfaces
	if err := netns.Set(origns); err != nil {
		return fmt.Errorf("failed to switch back to original namespace: %v", err)
	}

	// Debug: list root namespace interfaces
	oldifaces, _ := net.Interfaces()
	fmt.Printf("Root Namespace Interfaces: %v\n", oldifaces)

	// Wait for link (race-safe)
	link, err := waitForLink(name, 5*time.Second)
	if err != nil {
		return err
	}

	// Move interface into new namespace
	if err := netlink.LinkSetNsFd(link, int(newns)); err != nil {
		return fmt.Errorf("error moving link to namespace: %v", err)
	}

	// Switch into new namespace
	if err := netns.Set(newns); err != nil {
		return fmt.Errorf("failed to enter new namespace: %v", err)
	}

	// Debug: list interfaces in new namespace
	ifaces, _ := net.Interfaces()
	fmt.Printf("New Namespace Interfaces: %v\n", ifaces)

	namespaceLink, err := netlink.LinkByName(name)
	if err != nil {
		return fmt.Errorf("error getting link in new ns: %v", err)
	}

	// Parse MAC
	hwAddr, err := net.ParseMAC(macAddress)
	if err != nil {
		return fmt.Errorf("invalid MAC address %q: %v", macAddress, err)
	}

	// Set MAC
	if err := netlink.LinkSetHardwareAddr(namespaceLink, hwAddr); err != nil {
		return fmt.Errorf("error setting MAC: %v", err)
	}

	// Set IP
	addr := &netlink.Addr{
		IPNet: &net.IPNet{
			IP:   net.ParseIP(ipAddress),
			Mask: net.CIDRMask(24, 32),
		},
	}

	if err := netlink.AddrAdd(namespaceLink, addr); err != nil {
		return fmt.Errorf("error adding IP: %v", err)
	}

	// Bring interface up
	if err := netlink.LinkSetUp(namespaceLink); err != nil {
		return fmt.Errorf("error setting link up: %v", err)
	}

	// Default route
	route := &netlink.Route{
		LinkIndex: namespaceLink.Attrs().Index,
		Dst: &net.IPNet{
			IP:   net.IPv4zero,
			Mask: net.CIDRMask(0, 32),
		},
	}

	if err := netlink.RouteAdd(route); err != nil {
		return fmt.Errorf("error adding default route: %v", err)
	}

	// Debug print interfaces
	for _, iface := range ifaces {
		fmt.Printf("Interface: %v\n", iface.Name)
	}

	fmt.Println("Namespace setup complete:", name)
	return nil
}

func waitForLink(name string, timeout time.Duration) (netlink.Link, error) {
	deadline := time.Now().Add(timeout)

	for {
		link, err := netlink.LinkByName(name)
		if err == nil {
			return link, nil
		}

		if time.Now().After(deadline) {
			return nil, fmt.Errorf("link %s not found after %v", name, timeout)
		}

		time.Sleep(100 * time.Millisecond)
	}
}
