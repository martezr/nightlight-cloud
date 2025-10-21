package network

import "github.com/martezr/go-openvswitch/ovs"

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
