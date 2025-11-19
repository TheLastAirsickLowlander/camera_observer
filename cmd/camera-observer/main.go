package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"camera_observer/internal/config"
	"camera_observer/internal/mqtt"
	"camera_observer/internal/observer"
	"camera_observer/internal/smartthings"
)

func main() {
	log.Println("========================================")
	log.Println("  Camera Observer Service")
	log.Println("========================================")

	// 1. Load Configuration
	log.Println("[INIT] Loading configuration from config.yaml...")
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("[ERROR] Failed to load configuration: %v", err)
	}
	log.Printf("[INIT] Configuration loaded successfully")
	log.Printf("[INIT] Found %d camera-to-switch mappings", len(cfg.Mapping))

	// 2. Initialize SmartThings Client
	log.Println("[INIT] Initializing SmartThings API client...")
	stClient := smartthings.NewClient(cfg.SmartThings.APIToken)
	log.Println("[INIT] SmartThings client initialized")

	// 3. Initialize Observer Manager
	log.Println("[INIT] Initializing Observer Manager...")
	manager := observer.NewManager(stClient, cfg.Mapping)
	log.Println("[INIT] Observer Manager initialized")

	// 4. Initialize MQTT Client
	log.Printf("[INIT] Connecting to MQTT broker: %s", cfg.MQTT.Broker)
	mqttClient, err := mqtt.NewClient(cfg.MQTT.Broker, cfg.MQTT.ClientID)
	if err != nil {
		log.Fatalf("[ERROR] Failed to connect to MQTT broker: %v", err)
	}
	defer mqttClient.Disconnect()
	log.Println("[INIT] MQTT broker connected successfully")

	// 5. Subscribe to Topic
	log.Printf("[INIT] Subscribing to MQTT topic: %s", cfg.MQTT.Topic)
	err = mqttClient.Subscribe(cfg.MQTT.Topic, manager.HandleMessage)
	if err != nil {
		log.Fatalf("[ERROR] Failed to subscribe to topic %s: %v", cfg.MQTT.Topic, err)
	}
	log.Println("[INIT] Subscribed to topic successfully")

	log.Println("========================================")
	log.Println("  SERVICE RUNNING")
	log.Println("========================================")

	// 6. Wait for Shutdown Signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("========================================")
	log.Println("  Shutting down gracefully...")
	log.Println("========================================")
}
