package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/insomniacslk/dhcp/dhcpv4"
	"github.com/insomniacslk/dhcp/dhcpv4/server4"
)

type DHCPPayload struct {
	IPAddress  string   `json:"ipAddress"`
	Netmask    string   `json:"netmask"`
	Gateway    string   `json:"gateway"`
	DNSServers []string `json:"dnsServers"`
	DomainName string   `json:"domainName"`
	NTPServers []string `json:"ntpServers"`
}

var (
	// 	ipPoolStart  = net.IP{10, 0, 0, 15}
	// 	ipPoolEnd    = net.IP{10, 0, 0, 17}
	// 	subnetMask   = net.IP{255, 255, 255, 0}
	// 	router       = net.IP{10, 0, 0, 1}
	staticRoutes = []struct {
		Dest    net.IP
		Mask    net.IP
		Gateway net.IP
	}{}
	//{
	// 		{net.IP{10, 0, 0, 0}, net.IP{255, 0, 0, 0}, net.IP{10, 0, 0, 1}},
	// 	}
	leaseDuration = 12 * time.Hour
)

func handler(conn net.PacketConn, peer net.Addr, m *dhcpv4.DHCPv4) {
	f, err := os.OpenFile("/var/log/dhcpagent.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Failed to open log file: %v", err)
	} else {
		mw := io.MultiWriter(os.Stdout, f)
		log.SetOutput(mw)
		defer f.Close()
	}
	log.Println("Received DHCP message type:", m.MessageType())
	switch m.MessageType() {
	case dhcpv4.MessageTypeDiscover:
		// Add logic to check with the dhcp backend to see if the source mac is valid
		// remove colons from the mac address
		mac := m.ClientHWAddr.String()
		// if the mac address doesn't start with 52 then ignore it
		if !strings.HasPrefix(mac, "52") {
			log.Printf("Ignoring DHCP Discover from non-KVM MAC: %s", mac)
			return
		}
		payload, validMac := FetchDHCPPayload(mac)
		if !validMac {
			log.Printf("No DHCP payload found for MAC: %s", mac)
			return
		}
		log.Printf("Received DHCP Discover from MAC: %s", mac)
		log.Printf("DHCP Payload for MAC %s: %+v", mac, payload)
		offer, err := dhcpv4.NewReplyFromRequest(m)
		if err != nil {
			log.Println("Error creating DHCP offer:", err)
			return
		}
		offer.YourIPAddr = net.ParseIP(payload.IPAddress)
		offer.ServerIPAddr = net.IPv4(169, 254, 169, 253)
		offer.UpdateOption(dhcpv4.OptMessageType(dhcpv4.MessageTypeOffer))
		maskIP := net.ParseIP(payload.Netmask).To4()
		if maskIP == nil {
			panic("invalid netmask")
		}

		mask := net.IPv4Mask(maskIP[0], maskIP[1], maskIP[2], maskIP[3])
		offer.UpdateOption(dhcpv4.OptSubnetMask(mask))
		offer.UpdateOption(dhcpv4.OptRouter(net.ParseIP(payload.Gateway)))
		offer.UpdateOption(dhcpv4.OptIPAddressLeaseTime(leaseDuration))
		offer.UpdateOption(dhcpv4.OptDomainName(payload.DomainName))
		if len(payload.DNSServers) > 0 {
			var dnsIPs []net.IP
			for _, dns := range payload.DNSServers {
				dnsIPs = append(dnsIPs, net.ParseIP(dns))
			}
			offer.UpdateOption(dhcpv4.OptDNS(dnsIPs...))
		}
		offer.UpdateOption(dhcpv4.OptServerIdentifier(net.IPv4(169, 254, 169, 253)))
		_, ip, _ := net.ParseCIDR("169.254.0.0/16")
		test := &dhcpv4.Route{
			Dest:   ip,
			Router: net.IP{0, 0, 0, 0},
		}
		//offer.UpdateOption(dhcpv4.OptBootFileName("http://10.0.0.237/api/v1/ipxe"))
		offer.UpdateOption(dhcpv4.Option{Code: dhcpv4.OptionBootfileName, Value: dhcpv4.String("http://10.0.0.237/api/v1/ipxe")})
		opt1 := dhcpv4.OptClasslessStaticRoute(test)
		offer.UpdateOption(opt1)
		//		log.Println("Sending DHCP offer for IP:", currentIP)
		log.Println("DHCP Offer:", offer)
		log.Printf("DHCP Offer Options: %+v", offer.Options)
		_, err = conn.WriteTo(offer.ToBytes(), peer)
		if err != nil {
			log.Println("Error sending DHCP offer:", err)
		}
	// Handle DHCP Request
	case dhcpv4.MessageTypeRequest:
		ack, err := dhcpv4.NewReplyFromRequest(m)
		if err != nil {
			return
		}
		mac := m.ClientHWAddr.String()
		log.Printf("Received DHCP Request from MAC: %s", mac)
		payload, validMac := FetchDHCPPayload(mac)
		if !validMac {
			log.Printf("No DHCP payload found for MAC: %s", mac)
			return
		}
		ack.YourIPAddr = net.ParseIP(payload.IPAddress)
		ack.UpdateOption(dhcpv4.OptMessageType(dhcpv4.MessageTypeAck))
		maskIP := net.ParseIP(payload.Netmask).To4()
		if maskIP == nil {
			panic("invalid netmask")
		}
		mask := net.IPv4Mask(maskIP[0], maskIP[1], maskIP[2], maskIP[3])
		ack.UpdateOption(dhcpv4.OptSubnetMask(mask))
		ack.UpdateOption(dhcpv4.OptRouter(net.ParseIP(payload.Gateway)))
		ack.UpdateOption(dhcpv4.OptIPAddressLeaseTime(leaseDuration))
		ack.UpdateOption(dhcpv4.OptDomainName(payload.DomainName))
		if len(payload.DNSServers) > 0 {
			var dnsIPs []net.IP
			for _, dns := range payload.DNSServers {
				dnsIPs = append(dnsIPs, net.ParseIP(dns))
			}
			ack.UpdateOption(dhcpv4.OptDNS(dnsIPs...))
		}
		ack.UpdateOption(dhcpv4.OptServerIdentifier(net.IPv4(169, 254, 169, 253)))
		// Add static routes (option 121)
		var routes []byte
		for _, r := range staticRoutes {
			routes = append(routes, 16) // 16 bits mask for 10.0.0.0/8
			routes = append(routes, r.Dest[0])
			routes = append(routes, r.Gateway...)
		}
		_, ip, _ := net.ParseCIDR("169.254.0.0/16")
		test := &dhcpv4.Route{
			Dest:   ip,
			Router: net.IP{0, 0, 0, 0},
		}
		//offer.UpdateOption(dhcpv4.OptBootFileName("http://10.0.0.237/api/v1/ipxe"))
		ack.UpdateOption(dhcpv4.Option{Code: dhcpv4.OptionBootfileName, Value: dhcpv4.String("http://10.0.0.237/api/v1/ipxe")})
		opt1 := dhcpv4.OptClasslessStaticRoute(test)
		ack.UpdateOption(opt1)
		//ack.UpdateOption(dhcpv4.Option{Code: 121, Value: dhcpv4.OptionGeneric(routes)})
		log.Println("Sending DHCP ack for IP:", ack.YourIPAddr)
		log.Printf("DHCP Ack Options: %+v", ack.Options)
		_, err = conn.WriteTo(ack.ToBytes(), peer)
		if err != nil {
			log.Println("Error sending DHCP ack:", err)
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

func FetchDHCPPayload(mac string) (DHCPPayload, bool) {
	socketPath := "/opt/nightlight/dhcp.sock"

	// Create a custom DialContext for Unix Domain Sockets
	udsDialer := func(ctx context.Context, _, _ string) (net.Conn, error) {
		return net.Dial("unix", socketPath)
	}

	// Create a custom Transport with the UDS dialer
	transport := &http.Transport{
		DialContext: udsDialer,
	}

	// Create an HTTP client using the custom Transport
	client := &http.Client{
		Transport: transport,
		Timeout:   5 * time.Second, // Optional: set a timeout
	}

	// Take the mac address and remove colons
	cleanMac := strings.ReplaceAll(mac, ":", "")

	// Make a request to the UDS HTTP server
	resp, err := client.Get("http://unix-socket/" + cleanMac) // The host name here is arbitrary, as it's ignored
	if err != nil {
		fmt.Printf("Error making request: %v\n", err)
		return DHCPPayload{}, false
	}
	defer resp.Body.Close()

	fmt.Printf("Response Status: %s\n", resp.Status)

	// Read and print the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
		return DHCPPayload{}, false
	}
	fmt.Printf("Response Body: %s\n", string(body))

	var payload DHCPPayload
	err = json.Unmarshal(body, &payload)
	if err != nil {
		fmt.Printf("Error unmarshaling response body: %v\n", err)
		return DHCPPayload{}, false
	}
	return payload, true
}

func PlainTextUnixClient(path string) string {
	socketPath := "/opt/nightlight/dhcp.sock"

	// Create a custom DialContext for Unix Domain Sockets
	udsDialer := func(ctx context.Context, _, _ string) (net.Conn, error) {
		return net.Dial("unix", socketPath)
	}

	// Create a custom Transport with the UDS dialer
	transport := &http.Transport{
		DialContext: udsDialer,
	}

	// Create an HTTP client using the custom Transport
	client := &http.Client{
		Transport: transport,
		Timeout:   5 * time.Second, // Optional: set a timeout
	}

	// Make a request to the UDS HTTP server
	resp, err := client.Get("http://unix-socket" + path) // The host name here is arbitrary, as it's ignored
	if err != nil {
		fmt.Printf("Error making request: %v\n", err)
		return ""
	}
	defer resp.Body.Close()

	fmt.Printf("Response Status: %s\n", resp.Status)

	// Read and print the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
		return ""
	}
	fmt.Printf("Response Body: %s\n", string(body))
	return string(body)
}
