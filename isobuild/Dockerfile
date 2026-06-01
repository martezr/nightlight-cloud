ARG ALPINE_IMAGE_VERSION=3.23
ARG APORTS_TAG=v3.23.4
ARG BINFMT_IMAGE=tonistiigi/binfmt:latest

FROM ${BINFMT_IMAGE} AS binfmt

FROM alpine:${ALPINE_IMAGE_VERSION}
RUN apk add alpine-sdk build-base apk-tools alpine-conf busybox \
  fakeroot xorriso squashfs-tools sudo \
  mtools dosfstools grub-efi git

# syslinux is missing for aarch64
ARG TARGETARCH
RUN if [ "${TARGETARCH}" = "amd64" ]; then apk add syslinux; fi

COPY --from=binfmt /usr/bin /binfmt

RUN addgroup root abuild
RUN abuild-keygen -i -a -n
RUN apk update

RUN echo "APORTS_TAG='${APORTS_TAG}'"

RUN git clone \
    --depth=1 \
    --branch v3.23.4 \
    https://github.com/alpinelinux/aports.git \
    /aports

#ADD src/aports /home/build/aports
# Remove --no-chown which is deprecated in apk 3.0 as alias for --usermode (disallowed as root)
RUN sed -i 's/--initdb --no-chown/--initdb/' /aports/scripts/mkimage.sh
WORKDIR /aports/scripts

COPY mkimg.nightlight.sh /aports/scripts/
COPY mkimg.base.sh /aports/scripts/
COPY mkimg.standard.sh /aports/scripts/

RUN chmod +x /aports/scripts/mkimg.nightlight.sh
RUN chmod +x /aports/scripts/mkimg.base.sh
RUN chmod +x /aports/scripts/mkimg.standard.sh 
#COPY metadataagent /aports/scripts/
RUN ls -l /aports/scripts/

#ENTRYPOINT ["sh", "./mkimage.sh"]
#CMD ["sh", "./mkimage.sh", "--tag", "v3.23", "--arch", "x86_64", "--outdir", "/iso", "--repository", "http://alpine-mirror.grt.local/v3.23/main", "--repository", "http://alpine-mirror.grt.local/v3.23/community", "--profile", "nightlight"]