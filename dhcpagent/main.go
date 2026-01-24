package main

import (
	"log"
	"net"
	"time"

	"github.com/insomniacslk/dhcp/dhcpv4"
	"github.com/insomniacslk/dhcp/dhcpv4/server4"
)

var (
	ipPoolStart  = net.IP{192, 168, 1, 100}
	ipPoolEnd    = net.IP{192, 168, 1, 200}
	subnetMask   = net.IP{255, 255, 255, 0}
	router       = net.IP{192, 168, 1, 1}
	staticRoutes = []struct {
		Dest    net.IP
		Mask    net.IP
		Gateway net.IP
	}{
		{net.IP{10, 0, 0, 0}, net.IP{255, 0, 0, 0}, net.IP{192, 168, 1, 1}},
	}
	leaseDuration = 12 * time.Hour
)

var currentIP = ipPoolStart

func nextIP(ip net.IP) net.IP {
	ip = append(net.IP(nil), ip...)
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] != 0 {
			break
		}
	}
	return ip
}

func ipInRange(ip, start, end net.IP) bool {
	for i := 0; i < len(ip); i++ {
		if ip[i] < start[i] || ip[i] > end[i] {
			return false
		}
	}
	return true
}

func handler(conn net.PacketConn, peer net.Addr, m *dhcpv4.DHCPv4) {
	switch m.MessageType() {
	case dhcpv4.MessageTypeDiscover:
		offer, err := dhcpv4.NewReplyFromRequest(m)
		if err != nil {
			log.Println("Error creating DHCP offer:", err)
			return
		}
		offer.YourIPAddr = currentIP
		offer.UpdateOption(dhcpv4.OptMessageType(dhcpv4.MessageTypeOffer))
		offer.UpdateOption(dhcpv4.OptSubnetMask(net.IPMask(subnetMask)))
		offer.UpdateOption(dhcpv4.OptRouter(router))
		offer.UpdateOption(dhcpv4.OptIPAddressLeaseTime(leaseDuration))
		// Add static routes (option 121)
		var routes []byte
		for _, r := range staticRoutes {
			routes = append(routes, 8) // 8 bits mask for 10.0.0.0/8
			routes = append(routes, r.Dest[0])
			routes = append(routes, r.Gateway...)
		}
		//offer.UpdateOption(dhcpv4.Option{Code: 121, Value: dhcpv4.OptionGeneric(routes)})
	case dhcpv4.MessageTypeRequest:
		ack, err := dhcpv4.NewReplyFromRequest(m)
		if err != nil {
			return
		}
		ack.YourIPAddr = currentIP
		ack.UpdateOption(dhcpv4.OptMessageType(dhcpv4.MessageTypeAck))
		ack.UpdateOption(dhcpv4.OptSubnetMask(net.IPMask(subnetMask)))
		ack.UpdateOption(dhcpv4.OptRouter(router))
		ack.UpdateOption(dhcpv4.OptIPAddressLeaseTime(leaseDuration))
		// Add static routes (option 121)
		var routes []byte
		for _, r := range staticRoutes {
			routes = append(routes, 8) // 8 bits mask for 10.0.0.0/8
			routes = append(routes, r.Dest[0])
			routes = append(routes, r.Gateway...)
		}
		//ack.UpdateOption(dhcpv4.Option{Code: 121, Value: dhcpv4.OptionGeneric(routes)})
		// Move to next IP for next client
		currentIP = nextIP(currentIP)
		if !ipInRange(currentIP, ipPoolStart, ipPoolEnd) {
			currentIP = ipPoolStart
		}
	default:
		log.Println("Unhandled DHCP message type:", m.MessageType())
	}
}

func main() {
	laddr := &net.UDPAddr{
		IP:   net.ParseIP("0.0.0.0"),
		Port: dhcpv4.ServerPort,
	}

	log.Printf("Starting DHCP server on %s", laddr)
	server, err := server4.NewServer("", laddr, handler)
	if err != nil {
		log.Fatal(err)
	}

	server.Serve()
}
