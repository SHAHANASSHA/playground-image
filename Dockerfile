FROM ubuntu:22.04

LABEL maintainer="yourname@example.com"
LABEL description="Playground container with 2 microVMs using Firecracker"

# Install dependencies
RUN apt-get update && apt-get install -y \
    curl \
    tar \
    iproute2 \
    iputils-ping \
    socat \
    net-tools \
    qemu-utils \
    unzip \
    python3 \
    bridge-utils \
    screen \
    ca-certificates \
    && apt-get clean

# Create working dir
WORKDIR /opt/firecracker

# Download Firecracker binary and jailer
RUN curl -LOJ https://github.com/firecracker-microvm/firecracker/releases/download/v1.5.0/firecracker-v1.5.0-x86_64.tgz \
    && tar xvf firecracker-v1.5.0-x86_64.tgz \
    && mv release-v1.5.0-x86_64/firecracker /usr/local/bin/firecracker \
    && chmod +x /usr/local/bin/firecracker

# Copy kernel and rootfs files into container (you will mount them in prod)
COPY ./vmlinux1 ./vmlinux2 ./rootfs1.ext4 ./rootfs2.ext4 ./
COPY run.sh .

RUN chmod +x run.sh

# Entrypoint runs the script that starts both VMs
ENTRYPOINT ["./run.sh"]

