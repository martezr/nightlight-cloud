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
chmod 755 "$tmp"/etc/nightlight-cloud

# create a service file for nightlight-cloud
mkdir -p "$tmp"/etc/init.d
makefile root:root 0755 "$tmp"/etc/init.d/nightlight-cloud <<'EOF'	
#!/sbin/openrc-run

command="/etc/nightlight-cloud"
#command_args="&"
pidfile="/var/run/nightlight-cloud.pid"
name="nightlight-cloud"
description="RMS Router Service"
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
EOF

#modprobe vhost_net

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
#rc_add nightlight-cloud boot
#rc_add iptables boot

# Open vSwitch services
rc_add ovs-modules boot
rc_add ovsdb-server boot
rc_add ovs-vswitchd boot

rc_add mount-ro shutdown
rc_add killprocs shutdown
rc_add savecache shutdown

tar -c -C "$tmp" etc | gzip -9n > $HOSTNAME.apkovl.tar.gz