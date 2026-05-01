#!/bin/sh -e

HOSTNAME="$1"
if [ -z "$HOSTNAME" ]; then
	echo "usage: $0 hostname"
	exit 1
fi

cleanup() {
	rm -rf "$tmp"
}

makefile() {
	OWNER="$1"
	PERMS="$2"
	FILENAME="$3"
	cat > "$FILENAME"
	chown "$OWNER" "$FILENAME"
	chmod "$PERMS" "$FILENAME"
}

rc_add() {
	mkdir -p "$tmp"/etc/runlevels/"$2"
	ln -sf /etc/init.d/"$1" "$tmp"/etc/runlevels/"$2"/"$1"
}

tmp="$(mktemp -d)"
trap cleanup EXIT

mkdir -p "$tmp"/etc
makefile root:root 0644 "$tmp"/etc/hostname <<EOF
$HOSTNAME
EOF

cp /aports/scripts/nightlight-cloud "$tmp"/etc/nightlight-cloud
cp /aports/scripts/metadataagent "$tmp"/etc/metadataagent
cp /aports/scripts/dhcpagent "$tmp"/etc/dhcpagent
cp /aports/scripts/flowmonitoragent "$tmp"/etc/flowmonitoragent
cp /aports/scripts/nightlight-config "$tmp"/etc/nightlight-config
cp /aports/scripts/OVMF_CODE.secboot.fd "$tmp"/etc/OVMF_CODE.secboot.fd
cp /aports/scripts/OVMF_VARS.secboot.fd "$tmp"/etc/OVMF_VARS.secboot.fd
cp /aports/scripts/OVMF_CODE_4M.ms.fd "$tmp"/etc/OVMF_CODE_4M.ms.fd
cp /aports/scripts/OVMF_VARS_4M.ms.fd "$tmp"/etc/OVMF_VARS_4M.ms.fd
cp /aports/scripts/virtio-net.rom "$tmp"/etc/virtio-net.rom

chmod 755 "$tmp"/etc/nightlight-cloud
chmod 755 "$tmp"/etc/metadataagent
chmod 755 "$tmp"/etc/dhcpagent
chmod 755 "$tmp"/etc/flowmonitoragent
chmod 755 "$tmp"/etc/nightlight-config

# create a service file for nightlight-cloud
mkdir -p "$tmp"/etc/init.d
makefile root:root 0755 "$tmp"/etc/init.d/nightlight-cloud <<'EOF'	
#!/sbin/openrc-run

command="/etc/nightlight-cloud"
command_args="&"
pidfile="/var/run/nightlight-cloud.pid"
name="nightlight-cloud"
description="Nightlight Cloud Service"
depend() {
	after nightlight-config
}
EOF

makefile root:root 0755 "$tmp"/etc/init.d/nightlight-config <<'EOF'	
#!/sbin/openrc-run

command="/etc/nightlight-config"
command_args="&"
pidfile="/var/run/nightlight-config.pid"
name="nightlight-config"
description="Nightlight Config Service"
depend() {
	after net
}
EOF

makefile root:root 0755 "$tmp"/etc/init.d/metadataagent <<'EOF'	
#!/sbin/openrc-run
#/etc/init.d/metadataagent
name="metadataagent"
command="/sbin/ip"
command_args="netns exec mddefaultvpc /etc/metadataagent"
command_user="root"
pidfile="/var/run/metadataagent.pid"
command_background="yes"
depend() {
	after net
}
EOF

## Setup DHCP service
makefile root:root 0755 "$tmp"/etc/init.d/dhcpagent <<'EOF'	
#!/sbin/openrc-run
#/etc/init.d/dhcpagent
name="dhcpagent"
command="/sbin/ip"
command_args="netns exec dhdefaultvpc /etc/dhcpagent"
command_user="root"
pidfile="/var/run/dhcpagent.pid"
command_background="yes"
depend() {
	after net
}
EOF

## Setup Flow Monitor service
makefile root:root 0755 "$tmp"/etc/init.d/flowmonitoragent <<'EOF'	
#!/sbin/openrc-run
#/etc/init.d/flowmonitoragent
name="flowmonitoragent"
command="/sbin/ip"
command_args="netns exec fmdefaultvpc /etc/flowmonitoragent"
command_user="root"
pidfile="/var/run/flowmonitoragent.pid"
command_background="yes"
depend() {
	after net
}
EOF

mkdir -p "$tmp"/etc/network
makefile root:root 0644 "$tmp"/etc/network/interfaces <<EOF
auto lo
iface lo inet loopback

auto eth0
iface eth0 inet dhcp
EOF

mkdir -p "$tmp"/etc/modules-load.d
makefile root:root 0644 "$tmp"/etc/modules-load.d/kvm.conf <<EOF
kvm_intel
EOF
	
mkdir -p "$tmp"/etc/apk
makefile root:root 0644 "$tmp"/etc/apk/world <<EOF
alpine-base
iproute2
tcpdump
supervisor
openvswitch
libvirt
libvirt-daemon
qemu-hw-usb-host
qemu-img
qemu-system-x86_64
ovmf
qemu-modules 
openrc
openssh
swtpm
edk2
qemu-system-i386
nfs-utils
EOF

#modprobe kvm_intel
#chmod 777 /dev/kvm

rc_add devfs sysinit
rc_add dmesg sysinit
rc_add mdev sysinit
rc_add hwdrivers sysinit
rc_add modloop sysinit

rc_add hwclock boot
rc_add modules boot
rc_add sysctl boot
rc_add hostname boot
rc_add bootmisc boot
rc_add syslog boot
rc_add sshd boot
rc_add libvirtd boot
rc_add nightlight-config boot
rc_add nightlight-cloud boot
#rc_add iptables boot

# Open vSwitch services
rc_add ovs-modules boot
rc_add ovsdb-server boot
rc_add ovs-vswitchd boot

rc_add mount-ro shutdown
rc_add killprocs shutdown
rc_add savecache shutdown

tar -c -C "$tmp" etc | gzip -9n > $HOSTNAME.apkovl.tar.gz