docker run --rm -it \
  --platform=linux/amd64 \
  --privileged \
  -v "$PWD/out:/out" \
  alpine:3.23 sh -c '
set -e

apk add --no-cache \
  alpine-sdk alpine-conf git xorriso squashfs-tools mtools syslinux grub-bios busybox openssl

adduser -D builder

su builder -c "
set -e

# Initialize abuild environment (THIS FIXES update-kernel)
abuild-keygen -a -n

cd /home/builder
git clone --depth=1 https://github.com/alpinelinux/aports.git
cd aports

# IMPORTANT: run mkimage through proper environment
abuild -F checksum

cd scripts

./mkimage.sh iso standard \
  --repository https://dl-cdn.alpinelinux.org/alpine/v3.23/main \
  --repository https://dl-cdn.alpinelinux.org/alpine/v3.23/community \
  --outdir /out
"
'