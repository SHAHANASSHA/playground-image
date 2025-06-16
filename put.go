package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	firecracker "github.com/firecracker-microvm/firecracker-go-sdk"
)

func main() {
	vmID := "vm1"
	socketPath := fmt.Sprintf("/tmp/firecracker-%s.sock", vmID)
	client := firecracker.NewClient(socketPath, nil, false)

	mmdsData := map[string]interface{}{
		"latest": map[string]interface{}{
			"meta-data": map[string]interface{}{
				"ami-id":          vmID + "-12345678",
				"reservation-id":  "r-fea54097",
				"local-hostname":  vmID,
				"public-hostname": "ec2-203-0-113-25.compute-1.amazonaws.com",
			},
		},
	}

	mmdsJSON, err := json.Marshal(mmdsData)
	if err != nil {
		log.Fatalf("failed to marshal MMDS data: %v", err)
	}

	if _, err := client.PutMmds(context.Background(), mmdsJSON); err != nil {
		log.Fatalf("failed to set MMDS data: %v", err)
	}
	log.Fatalf("Completed!")
}
