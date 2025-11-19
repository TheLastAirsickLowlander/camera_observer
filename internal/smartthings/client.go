package smartthings

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

const BaseURL = "https://api.smartthings.com/v1"

type Client struct {
	Token      string
	HttpClient *http.Client
}

func NewClient(token string) *Client {
	return &Client{
		Token: token,
		HttpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type CommandRequest struct {
	Commands []Command `json:"commands"`
}

type Command struct {
	Component  string        `json:"component"`
	Capability string        `json:"capability"`
	Command    string        `json:"command"`
	Arguments  []interface{} `json:"arguments"`
}

// Device represents a SmartThings device
type Device struct {
	DeviceID string `json:"deviceId"`
	Name     string `json:"name"`
	Label    string `json:"label"`
}

// DevicesResponse represents the API response when listing devices
type DevicesResponse struct {
	Items []Device `json:"items"`
}

// sendCommand executes a command against a specific device
func (c *Client) sendCommand(deviceID, cmd string) error {
	payload := CommandRequest{
		Commands: []Command{
			{
				Component:  "main",
				Capability: "switch",
				Command:    cmd,
				Arguments:  []interface{}{},
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal command: %w", err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/devices/%s/commands", BaseURL, deviceID), bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return fmt.Errorf("network error sending command: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("api returned non-200 status: %d", resp.StatusCode)
	}

	return nil
}

// ListDevices retrieves all SmartThings devices
func (c *Client) ListDevices() ([]Device, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/devices", BaseURL), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("network error listing devices: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("api returned non-200 status: %d", resp.StatusCode)
	}

	var devicesResp DevicesResponse
	if err := json.NewDecoder(resp.Body).Decode(&devicesResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return devicesResp.Items, nil
}

// RestartDevice turns the switch OFF, waits, and turns it ON.
func (c *Client) RestartDevice(deviceID string) error {
	log.Printf("[SmartThings] ACTION: Turning OFF device %s", deviceID)
	if err := c.sendCommand(deviceID, "off"); err != nil {
		return err
	}

	log.Println("[SmartThings] WAIT: Waiting 10 seconds before turning ON...")
	time.Sleep(10 * time.Second)

	log.Printf("[SmartThings] ACTION: Turning ON device %s", deviceID)
	if err := c.sendCommand(deviceID, "on"); err != nil {
		return err
	}

	return nil
}
