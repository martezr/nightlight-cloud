package network

import (
	"fmt"
	"log"
	"net"
	"os/exec"

	"github.com/lorenzosaino/go-sysctl"
	"github.com/martezr/go-openvswitch/ovs"
	"github.com/vishvananda/netlink"
)

func SetupBaseNetworking() {
	log.Println("Setting up networking")
	err := sysctl.Set("net.ipv4.ip_forward", "1")
	if err != nil {
		fmt.Println(err)
	}
	c := ovs.New()
	ConfigureManagementNetwork(c)
}

func ConfigureManagementNetwork(c *ovs.Client) {
	c.VSwitch.AddBridge("nightlight")
	c.VSwitch.AddPort("nightlight", "eth0")

	/*eth0, err := netlink.LinkByName("eth0")
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
	*/
	flushcmd := exec.Command("ip", "addr", "flush", "dev", "eth0")
	err := flushcmd.Run()
	if err != nil {
		log.Println("Error flushing eth0 addr:", err)
		return
	}

	link, err := netlink.LinkByName("nightlight")
	if err != nil {
		log.Println("Error getting link:", err)
		return
	}
	netlink.AddrAdd(link, &netlink.Addr{IPNet: &net.IPNet{
		IP:   net.ParseIP("10.0.0.235"),
		Mask: net.CIDRMask(24, 32),
	}})

	// Set default route via eth0 gateway
	gw := net.ParseIP("10.0.0.1")
	route := &netlink.Route{
		LinkIndex: link.Attrs().Index,
		Gw:        gw,
	}
	netlink.RouteAdd(route)
	netlink.LinkSetUp(link)
}
