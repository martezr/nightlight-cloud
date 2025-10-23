package network

import (
	"fmt"
	"net"

	"github.com/martezr/go-openvswitch/ovs"
)

func InstallDefaultFlows(bridge string) error {
	ovsClient := ovs.New()

	err := ovsClient.OpenFlow.AddFlow(bridge, &ovs.Flow{
		Priority: 0,
		Actions: []ovs.Action{
			ovs.Normal(),
		},
	})
	if err != nil {
		return err
	}

	return nil
}

func AddVMFlows(bridge string, vmMac string, ofPort int, metadataOfPort int) error {
	ovsClient := ovs.New()

	// convert ofPort to two ip address octets
	vmNatIP := fmt.Sprintf("100.127.%d.%d", (ofPort>>8)&0xff, ofPort&0xff)

	metadataIpAddress := "169.254.169.254"
	metadataMacAddress := "32:6b:ce:89:41:42"
	// convert string to hardware address
	metadataMacHardwareAddress, err := net.ParseMAC(metadataMacAddress)
	if err != nil {
		fmt.Printf("Error parsing MAC address: %v\n", err)
	}
	vmMacHardwareAddress, err := net.ParseMAC(vmMac)
	if err != nil {
		fmt.Printf("Error parsing VM MAC address: %v\n", err)
	}

	// DHCP responder
	/*	err := ovsClient.OpenFlow.AddFlow(bridge, &ovs.Flow{
			Cookie:   0x1,
			Priority: 100,
			Protocol: ovs.ProtocolUDP,
			InPort:   ofPort,
			Matches: []ovs.Match{
				ovs.UDPDestinationPort(67),
				ovs.DataLinkSource(vmMac),
			},
			Table: 0,
			Actions: []ovs.Action{
				ovs.Move("OXM_OF_ETH_SRC[]", "OXM_OF_ETH_DST[]"),
				ovs.SetField(metadataMacAddress, "eth_src"),
				ovs.Move("OXM_OF_IP_SRC[]", "OXM_OF_IP_DST[]"),
				ovs.SetField(metadataIpAddress, "ip_src"),
				ovs.SetField(67, "udp_src"),
				ovs.SetField(68, "udp_dst"),
				ovs.InPort(),
			},
		})
		if err != nil {
			return err
		}
	*/
	// VM to Metadata ARP responder
	err = ovsClient.OpenFlow.AddFlow(bridge, &ovs.Flow{
		Cookie:   0x1,
		Priority: 100,
		Protocol: ovs.ProtocolARP,
		InPort:   ofPort,
		Matches: []ovs.Match{
			ovs.ARPOperation(1), // ARP Request
			ovs.ARPTargetProtocolAddress(metadataIpAddress),
			ovs.DataLinkSource(vmMac),
		},
		Table: 0,
		Actions: []ovs.Action{
			ovs.Load("0x2", "OXM_OF_ARP[]"), // ARP Reply
			ovs.ModDataLinkSource(metadataMacHardwareAddress),
			ovs.ModDataLinkDestination(vmMacHardwareAddress),
			ovs.SetField(metadataMacAddress, "arp_sha"),
			ovs.SetField(vmMac, "arp_tha"),
			ovs.Move("OXM_OF_ARP_SPA[]", "OXM_OF_ARP_TPA[]"),
			ovs.SetField(metadataIpAddress, "arp_spa"),
			ovs.InPort(),
		},
	})

	if err != nil {
		return err
	}

	// Metadata to VM ARP responder
	err = ovsClient.OpenFlow.AddFlow(bridge, &ovs.Flow{
		Cookie:   0x1,
		Priority: 110,
		Protocol: ovs.ProtocolARP,
		InPort:   metadataOfPort,
		Matches: []ovs.Match{
			ovs.ARPOperation(1), // ARP Request
			ovs.ARPTargetProtocolAddress(vmNatIP),
			ovs.ARPSourceProtocolAddress(metadataIpAddress),
			ovs.DataLinkSource(metadataMacAddress),
		},
		Table: 0,
		Actions: []ovs.Action{
			ovs.Load("0x2", "OXM_OF_ARP[]"), // ARP Reply

			ovs.ModDataLinkSource(vmMacHardwareAddress),
			ovs.ModDataLinkDestination(metadataMacHardwareAddress),

			ovs.SetField(vmMac, "arp_sha"),
			ovs.SetField(metadataMacAddress, "arp_tha"),
			ovs.SetField(vmNatIP, "arp_spa"),
			ovs.SetField(metadataIpAddress, "arp_tpa"),
			ovs.InPort(),
		},
	})

	if err != nil {
		return err
	}

	// Nat VM metadata requests
	err = ovsClient.OpenFlow.AddFlow(bridge, &ovs.Flow{
		Cookie:   0x1,
		Priority: 120,
		Protocol: ovs.ProtocolTCPv4,
		InPort:   ofPort,
		Matches: []ovs.Match{
			ovs.NetworkDestination(metadataIpAddress),
			ovs.TransportDestinationPort(80),
			ovs.DataLinkSource(vmMac),
		},
		Table: 0,
		Actions: []ovs.Action{
			ovs.ConnectionTracking(fmt.Sprintf("zone=%d,commit,nat(src=%s),exec(set_field:%d->ct_mark)", ofPort, vmNatIP, 5)),
			ovs.Resubmit(0, 1),
		},
	})

	if err != nil {
		return err
	}

	// Nat Metadata responses to VM
	err = ovsClient.OpenFlow.AddFlow(bridge, &ovs.Flow{
		Cookie:   0x1,
		Priority: 130,
		Protocol: ovs.ProtocolTCPv4,
		InPort:   metadataOfPort,
		Matches: []ovs.Match{
			ovs.Metadata(0),
			ovs.NetworkSource(metadataIpAddress),
			ovs.NetworkDestination(vmNatIP),
			ovs.DataLinkDestination(vmMac),
			ovs.TransportSourcePort(80),
		},
		Table: 0,
		Actions: []ovs.Action{
			ovs.ModifyIPv4Source(metadataIpAddress),
			ovs.ModifyDataLinkDestination(vmMacHardwareAddress),
			ovs.Output(ofPort),
		},
	})

	if err != nil {
		return err
	}

	return nil
}
