package main

import (
	"flag"
	"fmt"
	"log"

	"camera_observer/internal/smartthings"
)

func main() {
	token := flag.String("token", "", "SmartThings API Token")
	flag.Parse()

	if *token == "" {
		log.Fatal("Please provide --token flag with your SmartThings API token")
	}

	client := smartthings.NewClient(*token)
	devices, err := client.ListDevices()
	if err != nil {
		log.Fatalf("Failed to list devices: %v", err)
	}

	fmt.Println("\n=== SmartThings Devices ===\n")
	for _, device := range devices {
		fmt.Printf("Device ID: %s\n", device.DeviceID)
		fmt.Printf("  Name: %s\n", device.Name)
		fmt.Printf("  Label: %s\n\n", device.Label)
	}
}
