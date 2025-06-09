#!/bin/bash
set -e

echo "Starting microVM 1..."
screen -dmS fc1 firecracker --api-sock /tmp/firecracker1.socket
sleep 1
curl -s --unix-socket /tmp/firecracker1.socket -i \
    -X PUT 'http://localhost/boot-source' \
    -H 'Accept: application/json' \
    -H 'Content-Type: application/json' \
    -d '{
        "kernel_image_path": "vmlinux1",
        "boot_args": "console=ttyS0 reboot=k panic=1 pci=off"
    }'

curl -s --unix-socket /tmp/firecracker1.socket -i \
    -X PUT 'http://localhost/rootfs' \
    -H 'Accept: application/json' \
    -H 'Content-Type: application/json' \
    -d '{
        "drive_id": "rootfs",
        "path_on_host": "rootfs1.ext4",
        "is_root_device": true,
        "is_read_only": false
    }'

curl -s --unix-socket /tmp/firecracker1.socket -i \
    -X PUT 'http://localhost/actions' \
    -H 'Accept: application/json' \
    -d '{
        "action_type": "InstanceStart"
    }'

echo "MicroVM 1 started."

echo "Starting microVM 2..."
screen -dmS fc2 firecracker --api-sock /tmp/firecracker2.socket
sleep 1
curl -s --unix-socket /tmp/firecracker2.socket -i \
    -X PUT 'http://localhost/boot-source' \
    -H 'Accept: application/json' \
    -H 'Content-Type: application/json' \
    -d '{
        "kernel_image_path": "vmlinux2",
        "boot_args": "console=ttyS0 reboot=k panic=1 pci=off"
    }'

curl -s --unix-socket /tmp/firecracker2.socket -i \
    -X PUT 'http://localhost/rootfs' \
    -H 'Accept: application/json' \
    -H 'Content-Type: application/json' \
    -d '{
        "drive_id": "rootfs",
        "path_on_host": "rootfs2.ext4",
        "is_root_device": true,
        "is_read_only": false
    }'

curl -s --unix-socket /tmp/firecracker2.socket -i \
    -X PUT 'http://localhost/actions' \
    -H 'Accept: application/json' \
    -d '{
        "action_type": "InstanceStart"
    }'

echo "MicroVM 2 started."

# Keep container alive
tail -f /dev/null

