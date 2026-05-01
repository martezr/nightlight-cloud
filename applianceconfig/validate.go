package main

import (
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/vishvananda/netlink"
)

// EnsureInterfaceOperational verifies interface state and IP after OVS migration
func EnsureInterfaceOperational(
	ifaceName string,
	expectedCIDR string,
	pingTarget string,
	timeout time.Duration,
) error {
	hclog.Default().Named("network").Info("Validating interface", "interface", ifaceName, "expected_ip", expectedCIDR)
	link, err := netlink.LinkByName(ifaceName)
	if err != nil {
		return fmt.Errorf("interface %s not found: %w", ifaceName, err)
	}

	// Ensure link is UP
	if link.Attrs().OperState != netlink.OperUp {
		if err := netlink.LinkSetUp(link); err != nil {
			return fmt.Errorf("failed to bring interface up: %w", err)
		}
	}

	// Parse expected IP
	ipNet, err := netlink.ParseIPNet(expectedCIDR)
	if err != nil {
		return fmt.Errorf("invalid CIDR: %w", err)
	}

	addr := &netlink.Addr{
		IPNet: ipNet,
	}

	// Ensure IP address exists
	addrs, err := netlink.AddrList(link, netlink.FAMILY_ALL)
	if err != nil {
		return fmt.Errorf("failed to list addresses: %w", err)
	}

	hasIP := false
	for _, a := range addrs {
		if a.IPNet.String() == ipNet.String() {
			hasIP = true
			break
		}
	}

	if !hasIP {
		if err := netlink.AddrAdd(link, addr); err != nil {
			return fmt.Errorf("failed to add IP address: %w", err)
		}
	}

	// Wait for interface to become operational
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		link, err = netlink.LinkByName(ifaceName)
		if err != nil {
			return err
		}

		if link.Attrs().OperState == netlink.OperUp {
			break
		}

		time.Sleep(500 * time.Millisecond)
	}

	if link.Attrs().OperState != netlink.OperUp {
		return errors.New("interface failed to become operational")
	}

	// Optional connectivity check
	if pingTarget != "" {
		conn, err := net.DialTimeout("ip4:icmp", pingTarget, 2*time.Second)
		if err != nil {
			return fmt.Errorf("connectivity test failed: %w", err)
		}
		conn.Close()
	}

	hclog.Default().Named("network").Info("Interface is operational", "interface", ifaceName, "ip", expectedCIDR)
	return nil
}
