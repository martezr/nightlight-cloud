# create custom alpine linux iso using aports
FROM alpine:3.22
RUN apk update && apk upgrade
RUN apk add --no-cache bash curl xorriso syslinux mtools abuild alpine-conf squashfs-tools grub git doas openssh coreutils
RUN git clone --depth=1 https://gitlab.alpinelinux.org/alpine/aports.git
RUN abuild-keygen -a -n
RUN chmod +x aports/scripts/mkimage.sh

COPY mkimg.nightlight.sh aports/scripts
RUN mkdir /iso

# Set TMPDIR if /tmp is too small (e.g., in a default Docker container)
# Adjust the path if you prefer a different location
ENV TMPDIR=/root/tmp
RUN mkdir -pv ${TMPDIR}

COPY genapkovl-mkimgoverlay.sh aports/scripts/genapkovl-mkimgoverlay.sh
RUN chmod +x aports/scripts/genapkovl-mkimgoverlay.sh

#COPY mkimg.base.sh aports/scripts/mkimg.base.sh
#RUN chmod +x aports/scripts/mkimg.base.sh

COPY nightlight-cloud aports/scripts/nightlight-cloud
COPY dhcpagent aports/scripts/dhcpagent
COPY metadataagent aports/scripts/metadataagent
COPY flowmonitoragent aports/scripts/flowmonitoragent
COPY nightlight-config aports/scripts/nightlight-config
COPY OVMF_CODE.secboot.fd aports/scripts/OVMF_CODE.secboot.fd
COPY OVMF_VARS.secboot.fd aports/scripts/OVMF_VARS.secboot.fd
COPY OVMF_CODE_4M.ms.fd aports/scripts/OVMF_CODE_4M.ms.fd
COPY OVMF_VARS_4M.ms.fd aports/scripts/OVMF_VARS_4M.ms.fd
COPY virtio-net.rom aports/scripts/virtio-net.rom

CMD ["sh", "aports/scripts/mkimage.sh", "--tag", "v3.22", "--arch", "x86_64", "--outdir", "/iso", "--repository", "http://alpine-mirror.grt.local/v3.22/main", "--repository","http://alpine-mirror.grt.local/v3.22/community", "--profile", "nightlight"]