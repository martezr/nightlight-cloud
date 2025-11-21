docker build -f Dockerfile -t fedora-ovmf --output out .
cp out/OVMF_CODE.secboot.fd ..
cp out/OVMF_VARS.secboot.fd ..