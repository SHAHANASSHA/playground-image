FROM ubuntu:22.04
 
# Install required packages
RUN apt-get update && apt-get install -y \
    qemu-kvm \
    libvirt-daemon-system \
    libvirt-clients \
    curl \
    wget \
    tar \
    iptables \ 
    dnsmasq \
    bridge-utils \
    iproute2 \
    procps \
    ca-certificates \
    git \
    build-essential \
    python3 \ 
    python3-pip \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Download and extract firecracker
RUN wget https://github.com/firecracker-microvm/firecracker/releases/download/v1.11.0/firecracker-v1.11.0-x86_64.tgz
RUN tar -xf  firecracker-v1.11.0-x86_64.tgz

# Install Firecracker
RUN mv release-v1.11.0-x86_64/firecracker-v1.11.0-x86_64 /usr/local/bin/firecracker \
    && chmod +x /usr/local/bin/firecracker

# Inatll Jailer
RUN mv release-v1.11.0-x86_64/jailer-v1.11.0-x86_64 /usr/local/bin/jailer \
    && chmod +x /usr/local/bin/jailer

# Create necessary directories
RUN mkdir -p /dev/netdev
# RUN if [ ! -c /dev/net/tun ]; then mknod /dev/net/tun c 10 200; fi

# Make sure KVM device has correct permissions
RUN chmod 666 /dev/kvm || true

# Create directories for Firecracker
# RUN mkdir -p /root/firecracker/kernels
# COPY ./kernel/vmlinux-6.1.102 /root/firecracker/kernels/vmlinux

# RUN mkdir -p /root/firecracker/rootfs
# COPY ./rootfs/rootfs.ext4 /root/firecracker/rootfs/rootfs.ext4

# RUN mkdir -p /root/firecracker/overlayfs
# COPY ./overlayfs/overlay.ext4 /root/firecracker/overlayfs/overlay.ext4

RUN mkdir -p /root/firecracker/keys
COPY ./keys/id_rsa /root/firecracker/keys/ubuntu-24.04.id_rsa

COPY ./firestarter/firestarter /root/firecracker/firestarter
COPY ./firestarter/ws-server /root/firecracker/ws-server

COPY ./firestarter/runner.sh /root/firecracker/runner.sh
RUN chmod +x /root/firecracker/runner.sh

WORKDIR /root/firecracker

# Entry point
# CMD ["sleep", "infinity"]
CMD ["/root/firecracker/runner.sh"]
