module github.com/martezr/nightlight-cloud/applianceconfig

go 1.25.4

require (
	github.com/hashicorp/go-hclog v1.6.3
	github.com/lorenzosaino/go-sysctl v0.3.1
	github.com/martezr/go-openvswitch v0.0.0-00010101000000-000000000000
	github.com/vishvananda/netlink v1.3.1
	github.com/vishvananda/netns v0.0.5
)

require (
	github.com/BurntSushi/toml v1.1.0 // indirect
	github.com/fatih/color v1.13.0 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	golang.org/x/exp/typeparams v0.0.0-20220613132600-b0d781184e0d // indirect
	golang.org/x/lint v0.0.0-20210508222113-6edffad5e616 // indirect
	golang.org/x/mod v0.6.0-dev.0.20220419223038-86c51ed26bb4 // indirect
	golang.org/x/sys v0.10.0 // indirect
	golang.org/x/tools v0.1.11 // indirect
	honnef.co/go/tools v0.3.2 // indirect
)

replace github.com/martezr/go-openvswitch => ../../go-openvswitch
