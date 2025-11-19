# Camera Observer

## Overview
Camera Observer is a Go-based microservice designed to monitor Frigate MQTT stats for camera failures. Upon detecting a stream failure (0 FPS), it interacts with the Samsung SmartThings API to restart the corresponding IoT power switch (OFF -> Wait 10s -> ON), effectively rebooting the camera.

## Features
- **Frigate Integration**: Monitors `frigate/stats` MQTT topic for camera failures
- **Failure Detection**: Automatically detects when a camera has 0 FPS
- **SmartThings Control**: Toggles smart power switches to restart failed cameras
- **Cooldown Protection**: Prevents excessive power cycling (5-minute cooldown per camera)
- **Configurable Mapping**: Map specific cameras to specific SmartThings Device IDs
- **Docker Ready**: Fully containerized for easy deployment
- **Comprehensive Logging**: Structured logging for debugging and monitoring

## Project Structure
This project follows the [Standard Go Project Layout](https://github.com/golang-standards/project-layout).

- `cmd/camera-observer/`: Main entry point for the application
- `cmd/list-devices/`: Utility to list available SmartThings devices
- `internal/`: Private application and library code
  - `config/`: Configuration loading (YAML & Env Vars)
  - `mqtt/`: MQTT client and subscription logic
  - `smartthings/`: Client for Samsung SmartThings API
  - `observer/`: Core business logic connecting MQTT events to actions
- `docs/`: Detailed documentation

## Configuration
The application is configured via `config.yaml` or Environment Variables.

### Example `config.yaml`
```yaml
mqtt:
  broker: "tcp://mqtt.wilkywayre.com:1883"
  topic: "frigate/stats"
  client_id: "camera_observer"

smartthings:
  api_token: "YOUR_SMARTTHINGS_API_TOKEN"

# Map the Camera Name (from Frigate stats) to the SmartThings Switch Device ID
mapping:
  "front_door": "device-id-001"
  "backyard": "device-id-002"
```

## Quick Start

### 1. Get Your SmartThings API Token
1. Go to https://account.smartthings.com/tokens
2. Click **Generate token** 
3. Give it a name like "Camera Observer"
4. Select scope: **Devices** (read/write)
5. Copy the generated token

### 2. List Available SmartThings Devices
```bash
go run cmd/list-devices/main.go --token "YOUR_TOKEN_HERE"
```

This will output all your SmartThings devices with their IDs. Look for the smart switch/plug devices you've connected to your cameras.

### 3. Update `config.yaml`
Replace the placeholder token and add your camera-to-device mappings:
```yaml
mqtt:
  broker: "tcp://mqtt.wilkywayre.com:1883"
  topic: "frigate/stats"
  client_id: "camera_observer"

smartthings:
  api_token: "YOUR_ACTUAL_TOKEN"

mapping:
  "front_door": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"  # Device ID from list-devices
  "backyard": "yyyyyyyy-yyyy-yyyy-yyyy-yyyyyyyyyyyy"    # Device ID from list-devices
```

## Running Locally
1. Ensure you have Go 1.19+ installed
2. Set up your `config.yaml` with your token and device mappings
3. Run the application:
   ```bash
   go run cmd/camera-observer/main.go
   ```

## Running with Docker
1. Build the image:
   ```bash
   docker build -t camera-observer .
   ```

2. Run the container:
   ```bash
   docker run -d \
     --name camera-observer \
     -v /path/to/config.yaml:/app/config.yaml \
     camera-observer
   ```

3. Or use docker-compose:
   ```bash
   docker-compose up -d
   ```

## Environment Variables
You can override config values with environment variables:
- `MQTT_BROKER`: MQTT broker URL
- `MQTT_TOPIC`: MQTT topic to subscribe to
- `MQTT_CLIENT_ID`: MQTT client ID
- `SMARTTHINGS_API_TOKEN`: SmartThings API token

## Logging
The application uses structured logging with the following prefixes:
- `[INIT]`: Initialization messages
- `[Observer]`: Observer/detection events
- `[SmartThings]`: SmartThings API interactions
- `[ERROR]`: Error messages
- `[ACTION]`: Actions being taken
- `[SUCCESS]`: Successful operations
- `[SKIPPED]`: Skipped actions (e.g., cooldown)
- `[ALERT]`: Alert/warning messages

## Architecture
See `docs/ARCHITECTURE.md` for detailed information about the system design and data flow.

