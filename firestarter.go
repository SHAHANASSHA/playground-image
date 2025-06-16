package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"

	firecracker "github.com/firecracker-microvm/firecracker-go-sdk"
	"github.com/firecracker-microvm/firecracker-go-sdk/client/models"
)

func main() {
	// Setup context with signal handling
	// ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	// log.Println(100)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	configureNetWork("172.16.0.1/24", "172.16.0.0/24")

	go startMicroVM(ctx, "vm1", "tap0", "172.16.0.2", "172.16.0.1", "AA:FC:00:00:00:01")
	go startMicroVM(ctx, "vm2", "tap1", "172.16.0.3", "172.16.0.1", "AA:FC:00:00:00:02")
	select {}
}

func createNetworkInterface(tapName, ipAddr, gateway, macAddr string) firecracker.NetworkInterface {
	return firecracker.NetworkInterface{
		StaticConfiguration: &firecracker.StaticNetworkConfiguration{
			HostDevName: tapName,
			MacAddress:  macAddr,
			IPConfiguration: &firecracker.IPConfiguration{
				IPAddr: net.IPNet{
					IP:   net.ParseIP(ipAddr),
					Mask: net.CIDRMask(24, 32),
				},
				Gateway:     net.ParseIP(gateway),
				Nameservers: []string{"8.8.8.8", "1.1.1.1"},
				IfName:      "eth0",
			},
		},
	}
}

func startMicroVM(ctx context.Context, vmID, tapName, ipAddr, gateway, macAddr string) {
	// Configure VM
	kernelImagePath := "/assets/kernel/vmlinux-6.1.102"
	// kernelImagePath := "/home/john/resourses/kernel/vmlinux-6.1.102"
	// kernelImagePath := "/home/devaraj/run-firecracker/kernel/vmlinux-6.1.102"
	rootfsPath := "/assets/ubuntu24.04/rootfs/rootfs.ext4"
	// rootfsPath := "/home/john/resourses/ubuntu24.04/rootfs/rootfs.ext4"
	// rootfsPath := "/home/john/resourses/synnefo/rootfs/rootfs.ext4"
	// rootfsPath := "/home/devaraj/run-firecracker/rootfs/rootfs.ext4"
	overlayfsPath := "/root/firecracker/overlayfs/" + vmID + "-overlay.ext4"
	// overlayfsPath := "/home/john/resourses/ubuntu24.04/overlay/" + vmID + "-overlay.ext4"
	// overlayfsPath := "/home/john/resourses/synnefo/overlay/" + vmID + "-overlay.ext4"
	// overlayfsPath := "/home/devaraj/run-firecracker/overlay/" + vmID + "-overlay.ext4"

	socketPath := fmt.Sprintf("/tmp/firecracker-%s.sock", vmID)

	// Check if socket exists and remove it
	if _, err := os.Stat(socketPath); err == nil {
		if err := os.Remove(socketPath); err != nil {
			log.Fatalf("Failed to remove existing socket: %v", err)
		}
	}

	// Clean up existing FIFOs
	fifoFiles := []string{"/tmp/firecracker.out.fifo", "/tmp/firecracker.metrics.fifo"}
	for _, fifo := range fifoFiles {
		if _, err := os.Stat(fifo); err == nil {
			if err := os.Remove(fifo); err != nil {
				log.Fatalf("Failed to remove existing FIFO %s: %v", fifo, err)
			}
			log.Printf("Removed existing FIFO: %s", fifo)
		}
	}

	// mmdsData := map[string]interface{}{
	// 	"latest": map[string]interface{}{
	// 		"meta-data": map[string]interface{}{
	// 			"ami-id":          "ami-12345678",
	// 			"reservation-id":  "r-fea54097",
	// 			"local-hostname":  "ip-10-251-50-12.ec2.internal",
	// 			"public-hostname": "ec2-203-0-113-25.compute-1.amazonaws.com",
	// 		},
	// 	},
	// }

	cfg := firecracker.Config{
		SocketPath:      socketPath,
		KernelImagePath: kernelImagePath,
		KernelArgs:      "console=ttyS0 reboot=k panic=1 pci=off vm_hostname=" + vmID + " overlay_root=vdb init=/sbin/overlay-init",
		Drives: []models.Drive{
			{
				DriveID:      firecracker.String("rootfs"),
				PathOnHost:   firecracker.String(rootfsPath),
				CacheType:    firecracker.String(models.DriveCacheTypeUnsafe),
				IsRootDevice: firecracker.Bool(true),
				IsReadOnly:   firecracker.Bool(true),
				RateLimiter:  nil,
			},
			{
				DriveID:      firecracker.String("overlayfs"),
				PathOnHost:   firecracker.String(overlayfsPath),
				CacheType:    firecracker.String(models.DriveCacheTypeUnsafe),
				IsRootDevice: firecracker.Bool(false),
				IsReadOnly:   firecracker.Bool(false),
				RateLimiter:  nil,
			},
		},
		MachineCfg: models.MachineConfiguration{
			VcpuCount:  firecracker.Int64(2),
			MemSizeMib: firecracker.Int64(2048),
		},
		NetworkInterfaces: []firecracker.NetworkInterface{
			createNetworkInterface(tapName, ipAddr, gateway, macAddr),
		},
		VMID:     vmID,
		LogLevel: "Debug",
		LogPath:  filepath.Join(os.TempDir(), fmt.Sprintf("firecracker-%s.log", vmID)),
	}

	// Let's use a simpler approach without FIFOs
	cmd := firecracker.VMCommandBuilder{}.
		WithBin("firecracker").
		WithSocketPath(socketPath).
		// WithStdin(os.Stdin).
		// WithStdout(os.Stdout).
		WithStderr(os.Stderr).
		Build(ctx)

	m, err := firecracker.NewMachine(ctx, cfg, firecracker.WithProcessRunner(cmd))
	if err != nil {
		log.Fatalf("Failed to create machine: %v", err)
	}

	// if err := m.SetMetadata(ctx, mmdsData); err != nil {
	// 	log.Fatalf("Failed to set metadata: %v", err)
	// }

	// if err = preloadRootFS(rootfsPath); err != nil {
	// 	log.Fatalf("Failed to start machine: %v", err)
	// }

	// Start the VM
	log.Println("Starting Firecracker VM...")
	go func() {
		if err := m.Start(ctx); err != nil {
			log.Fatalf("Failed to start machine: %v", err)
		}
	}()

	// log.Println("Firecracker VM started successfully")

	// Wait for VM to finish or context to be canceled
	// if err := m.Wait(ctx); err != nil {
	// 	log.Fatalf("Wait returned an error: %v", err)
	// }

	// Ensure VM shuts down properly
	// if err := m.Shutdown(ctx); err != nil {
	// 	time.Sleep(500 * time.Millisecond) // Brief pause before forcing
	// 	if err := m.StopVMM(); err != nil {
	// 		log.Fatalf("Failed to stop VM: %v", err)
	// 	}
	// }
	// log.Println("Firecracker VM shut down successfully")
}

func configureNetWork(ipAdd string, ipAddB string) {
	// Create bridge
	if err := createBridge(ipAdd); err != nil {
		log.Fatalf("Failed to create bridge: %v", err)
	}

	// Create TAP interfaces (e.g., tap0 and tap1)
	tapInterfaces := []string{"tap0", "tap1"}
	for _, tap := range tapInterfaces {
		if err := createTapInterface(tap); err != nil {
			log.Fatalf("Failed to create TAP interface %s: %v", tap, err)
		}
	}

	// Set up NAT
	if err := setupNAT(ipAddB); err != nil {
		log.Fatalf("Failed to set up NAT: %v", err)
	}

	log.Println("Network setup completed successfully.")
}

func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error running %s %v: %v\nOutput: %s", name, args, err, output)
	}
	return nil
}

func preloadRootFS(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(io.Discard, f) // read entire file into kernel page cache
	return err
}

// createBridge sets up the bridge interface br0 with the specified IP address.
func createBridge(ipAddr string) error {
	// Delete existing bridge if it exists
	runCommand("ip", "link", "set", "br0", "down")
	runCommand("ip", "link", "delete", "br0", "type", "bridge")

	// Create bridge
	if err := runCommand("ip", "link", "add", "name", "br0", "type", "bridge"); err != nil {
		return err
	}

	// Assign IP address to bridge
	if err := runCommand("ip", "addr", "add", ipAddr, "dev", "br0"); err != nil {
		return err
	}

	// Bring up the bridge interface
	if err := runCommand("ip", "link", "set", "br0", "up"); err != nil {
		return err
	}

	return nil
}

// createTapInterface creates a TAP interface with the given name and attaches it to br0.
func createTapInterface(tapName string) error {
	// Delete existing TAP interface if it exists
	runCommand("ip", "link", "set", tapName, "down")
	runCommand("ip", "link", "delete", tapName)

	// Create TAP interface
	if err := runCommand("ip", "tuntap", "add", "dev", tapName, "mode", "tap"); err != nil {
		return err
	}

	// Attach TAP interface to bridge
	if err := runCommand("ip", "link", "set", tapName, "master", "br0"); err != nil {
		return err
	}

	// Bring up the TAP interface
	if err := runCommand("ip", "link", "set", tapName, "up"); err != nil {
		return err
	}

	return nil
}

// setupNAT configures NAT using iptables for internet access.
func setupNAT(ipAddr string) error {
	// Enable IP forwarding
	if err := runCommand("sysctl", "-w", "net.ipv4.ip_forward=1"); err != nil {
		return err
	}

	// Set up iptables rules
	if err := runCommand("iptables", "-t", "nat", "-A", "POSTROUTING", "-s", ipAddr, "-o", "eth0", "-j", "MASQUERADE"); err != nil {
		return err
	}
	if err := runCommand("iptables", "-A", "FORWARD", "-i", "br0", "-o", "eth0", "-j", "ACCEPT"); err != nil {
		return err
	}
	if err := runCommand("iptables", "-A", "FORWARD", "-i", "eth0", "-o", "br0", "-m", "state", "--state", "RELATED,ESTABLISHED", "-j", "ACCEPT"); err != nil {
		return err
	}

	return nil
}
