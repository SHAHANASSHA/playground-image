#!/bin/sh

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

EXECUTABLE="$SCRIPT_DIR/firestarter"
WSS_EXECUTABLE="$SCRIPT_DIR/ws-server"

# echo $(date +%H:%M:%S)

mkdir -p "/root/firecracker/overlayfs"
dd if=/dev/zero of=/root/firecracker/overlayfs/vm1-overlay.ext4 conv=sparse bs=1M count=5120 && mkfs.ext4 /root/firecracker/overlayfs/vm1-overlay.ext4
dd if=/dev/zero of=/root/firecracker/overlayfs/vm2-overlay.ext4 conv=sparse bs=1M count=5120 && mkfs.ext4 /root/firecracker/overlayfs/vm2-overlay.ext4
# cp "/assets/ubuntu24.04/overlay/overlay.ext4" "/root/firecracker/overlayfs/vm1-overlay.ext4"
# cp "/assets/ubuntu24.04/overlay/overlay.ext4" "/root/firecracker/overlayfs/vm2-overlay.ext4"
# rm "/root/firecracker/overlayfs/overlay.ext4"

# echo $(date +%H:%M:%S)

"$WSS_EXECUTABLE" &

until "$EXECUTABLE"; do
  echo "Command failed. Retrying in 1 seconds..."
  sleep 1
done

echo "Command succeeded."

# /root/firecracker/overlayfs