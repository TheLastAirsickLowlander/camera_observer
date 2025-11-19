package main

import (
	"fmt"
	"net"
	"time"
)

func main() {
	timeout := 5 * time.Second
	conn, err := net.DialTimeout("tcp", "mqtt.wilkywayre.com:1883", timeout)
	if err != nil {
		fmt.Printf("MQTT Connection FAILED: %v\n", err)
		return
	}
	defer conn.Close()
	fmt.Println("MQTT Connection SUCCESSFUL (Port 1883 is open)")
}
