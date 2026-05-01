package main

import (
	"fmt"

	"github.com/martezr/go-openvswitch/ovs"
)

var (
	//	MetadataServiceIP   = "169.254.169.254"
	//	MetadataServiceMAC  = "32:6b:ce:89:41:42"
	//	MetadataServicePort = 80
	TableL2Rewrite = 10
)

func InstallDefaultFlows(bridge string, metadataOfPort int) error {
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

	err = ovsClient.OpenFlow.AddFlow(bridge, &ovs.Flow{
		Priority: 0,
		Table:    TableL2Rewrite,
		Actions: []ovs.Action{
			ovs.Normal(),
		},
	})
	if err != nil {
		return err
	}

	// Map mac address to ofPort
	err = ovsClient.OpenFlow.AddFlow(bridge, &ovs.Flow{
		Priority: 100,
		Matches: []ovs.Match{
			ovs.DataLinkDestination("32:6b:ce:89:41:42"),
		},
		Table: TableL2Rewrite,
		Actions: []ovs.Action{
			ovs.Output(metadataOfPort),
		},
	})

	if err != nil {
		fmt.Println("Error adding flow:", err)
		return err
	}

	return nil
}
