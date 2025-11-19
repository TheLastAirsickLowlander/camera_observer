package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type MQTTConfig struct {
	Broker   string `yaml:"broker"`
	Topic    string `yaml:"topic"`
	ClientID string `yaml:"client_id"`
}

type SmartThingsConfig struct {
	APIToken string `yaml:"api_token"`
}

type Config struct {
	MQTT        MQTTConfig        `yaml:"mqtt"`
	SmartThings SmartThingsConfig `yaml:"smartthings"`
	Mapping     map[string]string `yaml:"mapping"`
}

// Load reads the configuration from the given yaml file path.
// It also allows environment variable overrides for critical secrets.
func Load(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer f.Close()

	var cfg Config
	decoder := yaml.NewDecoder(f)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("failed to decode config file: %w", err)
	}

	// Environment Variable Overrides (Docker friendly)
	if envToken := os.Getenv("SMARTTHINGS_TOKEN"); envToken != "" {
		cfg.SmartThings.APIToken = envToken
	}
	if envBroker := os.Getenv("MQTT_BROKER"); envBroker != "" {
		cfg.MQTT.Broker = envBroker
	}

	return &cfg, nil
}
