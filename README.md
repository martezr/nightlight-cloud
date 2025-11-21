# Nightlight Cloud

Nightlight cloud is a lightweight virtualization solution that is intended for homelab environments and lacks many of the standard enterprise grade virtualization features like clustering, live migration, and more. The goal is to enable rapid expirementation 

/usr/share/qemu/firmware/50-edk2-x86_64-secure.json

{
    "description": "UEFI firmware for x86_64, with Secure Boot and SMM",
    "interface-types": [
        "uefi"
    ],
    "mapping": {
        "device": "flash",
        "executable": {
            "filename": "/etc/OVMF_CODE.secboot.fd",
            "format": "raw"
        },
        "nvram-template": {
            "filename": "/etc/OVMF_VARS.secboot.fd",
            "format": "raw"
        }
    },
    "targets": [
        {
            "architecture": "x86_64",
            "machines": [
                "pc-q35-*"
            ]
        }
    ],
    "features": [
        "acpi-s3",
        "amd-sev",
        "requires-smm",
        "secure-boot",
        "verbose-dynamic"
    ],
    "tags": [

    ]
}


{
    "description": "UEFI firmware for x86_64, with Secure Boot and SMM",
    "interface-types": [
        "uefi"
    ],
    "mapping": {
        "device": "flash",
        "executable": {
            "filename": "/etc/OVMF_CODE_4M.ms.fd",
            "format": "raw"
        },
        "nvram-template": {
            "filename": "/etc/OVMF_VARS_4M.ms.fd",
            "format": "raw"
        }
    },
    "targets": [
        {
            "architecture": "x86_64",
            "machines": [
                "pc-q35-*"
            ]
        }
    ],
    "features": [
        "acpi-s3",
        "amd-sev",
        "requires-smm",
        "secure-boot",
        "verbose-dynamic"
    ],
    "tags": [

    ]
}