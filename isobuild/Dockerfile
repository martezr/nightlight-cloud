# create custom alpine linux iso using aports
FROM alpine:latest
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

COPY nightlight-cloud aports/scripts/nightlight-cloud

CMD ["sh", "aports/scripts/mkimage.sh", "--tag", "v3.22", "--arch", "x86_64", "--outdir", "/iso", "--repository", "http://dl-cdn.alpinelinux.org/alpine/v3.22/main", "--repository","http://dl-cdn.alpinelinux.org/alpine/v3.22/community", "--profile", "nightlight"]