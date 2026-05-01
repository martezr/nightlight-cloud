package main

import (
	"fmt"
	"log"
	"net"

	"github.com/hashicorp/go-hclog"
	"github.com/lorenzosaino/go-sysctl"
	"github.com/martezr/go-openvswitch/ovs"
	"github.com/vishvananda/netlink"
)

func SetupBaseNetworking() {
	hclog.Default().Named("network").Info("Setting up networking")
	err := sysctl.Set("net.ipv4.ip_forward", "1")
	if err != nil {
		fmt.Println(err)
	}
	c := ovs.New()
	ConfigureManagementNetwork(c)
	ConfigureDefaultVPCNetworking(c)
}

func ConfigureManagementNetwork(c *ovs.Client) {

	// Check if bridge "nl-external" already exists
	bridgeExists := false
	bridges, err := netlink.LinkList()
	if err != nil {
		log.Println("Error listing links:", err)
	} else {
		for _, l := range bridges {
			if l.Type() == "bridge" && l.Attrs().Name == "nl-external" {
				bridgeExists = true
				break
			}
		}
	}

	if bridgeExists {
		return
	}

	c.VSwitch.AddBridge("nl-external")
	c.VSwitch.AddPort("nl-external", "eth0")

	eth0, err := netlink.LinkByName("eth0")
	if err != nil {
		log.Println("Error getting eth0:", err)
		return
	}
	eth0Addrs, err := netlink.AddrList(eth0, netlink.FAMILY_V4)
	if err != nil {
		log.Println("Error getting eth0 addr:", err)
		return
	}
	if len(eth0Addrs) > 0 {
		for _, addr := range eth0Addrs {
			netlink.AddrDel(eth0, &addr)
		}
	}

	// Assign IP to nl-external bridge
	link, err := netlink.LinkByName("nl-external")
	if err != nil {
		log.Println("Error getting link:", err)
		return
	}
	netlink.AddrAdd(link, &netlink.Addr{IPNet: &net.IPNet{
		IP:   net.ParseIP("10.0.0.237"),
		Mask: net.CIDRMask(24, 32),
	}})

	netlink.LinkSetUp(link)
	// Set default route via eth0 gateway
	gw := net.ParseIP("10.0.0.1")
	route := &netlink.Route{
		LinkIndex: link.Attrs().Index,
		Gw:        gw,
		Dst:       &net.IPNet{IP: net.IPv4zero, Mask: net.CIDRMask(0, 32)},
	}
	netlink.RouteAdd(route)
}

func ConfigureDefaultVPCNetworking(c *ovs.Client) {
	c.VSwitch.AddBridge("nightlight")
	c.VSwitch.AddPort("nl-external", "nightlight-patch")
	c.VSwitch.AddPort("nightlight", "nl-external-patch")

	c.VSwitch.Set.Interface("nl-external-patch", ovs.InterfaceOptions{
		Type: "patch",
		Peer: "nightlight-patch",
	})

	c.VSwitch.Set.Interface("nightlight-patch", ovs.InterfaceOptions{
		Type: "patch",
		Peer: "nl-external-patch",
	})
}
