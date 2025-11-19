# Setup Guide

## Prerequisites
- Go 1.19 or later
- Docker (optional, for containerized deployment)
- Access to an MQTT broker (we use `mqtt.wilkywayre.com:1883`)
- Samsung SmartThings account with smart switches/plugs

## Step 1: Get SmartThings API Token

1. Visit https://account.smartthings.com/tokens
2. Sign in with your Samsung account
3. Click **+ Generate token**
4. Enter a name: `camera_observer` (or your preference)
5. Select **Devices** scope with **read** and **control** permissions
6. Click **Generate**
7. **IMPORTANT**: Copy the token immediately as it won't be shown again
8. Save it somewhere safe temporarily

## Step 2: List Your SmartThings Devices

First, we need to identify the device IDs for your smart switches/plugs.

```bash
go run cmd/list-devices/main.go --token "YOUR_TOKEN_HERE"
```

Replace `YOUR_TOKEN_HERE` with the actual token from Step 1.

**Example Output:**
```
=== SmartThings Devices ===

Device ID: 12a3b4c5-d678-9e01-fg23-456789h0ijkl
  Name: Front Door Camera Plug
  Label: Front Door Switch

Device ID: 87x6y5z4-w321-vut0-sr98-765432qpomno
  Name: Backyard Camera Plug
  Label: Backyard Switch
```

Copy the **Device IDs** that correspond to your camera power switches.

## Step 3: Configure the Application

### Option A: Edit config.yaml

1. Open `config.yaml` in your editor
2. Update the SmartThings token:
   ```yaml
   smartthings:
     api_token: "YOUR_ACTUAL_TOKEN_HERE"
   ```
3. Add your camera mappings using device IDs from Step 2:
   ```yaml
   mapping:
     "front_door": "12a3b4c5-d678-9e01-fg23-456789h0ijkl"
     "backyard": "87x6y5z4-w321-vut0-sr98-765432qpomno"
   ```

The camera names (e.g., "front_door", "backyard") must match exactly what appears in your Frigate `stats` JSON payload.

### Option B: Use Environment Variables

If running in Docker or prefer environment variables:
```bash
export MQTT_BROKER="tcp://mqtt.wilkywayre.com:1883"
export MQTT_TOPIC="frigate/stats"
export MQTT_CLIENT_ID="camera_observer"
export SMARTTHINGS_API_TOKEN="YOUR_TOKEN_HERE"
```

## Step 4: Run Locally

```bash
go run cmd/camera-observer/main.go
```

You should see output like:
```
2025/11/19 15:30:45 ========================================
2025/11/19 15:30:45   Camera Observer Service
2025/11/19 15:30:45 ========================================
2025/11/19 15:30:46 [INIT] Loading configuration from config.yaml...
2025/11/19 15:30:46 [INIT] Configuration loaded successfully
2025/11/19 15:30:46 [INIT] Found 2 camera-to-switch mappings
2025/11/19 15:30:46 [INIT] Initializing SmartThings API client...
2025/11/19 15:30:46 [INIT] SmartThings client initialized
2025/11/19 15:30:46 [INIT] Initializing Observer Manager...
2025/11/19 15:30:46 [INIT] Observer Manager initialized
2025/11/19 15:30:46 [INIT] Connecting to MQTT broker: tcp://mqtt.wilkywayre.com:1883
2025/11/19 15:30:47 [INIT] MQTT broker connected successfully
2025/11/19 15:30:47 [INIT] Subscribing to MQTT topic: frigate/stats
2025/11/19 15:30:47 [INIT] Subscribed to topic successfully
2025/11/19 15:30:47 ========================================
2025/11/19 15:30:47   SERVICE RUNNING
2025/11/19 15:30:47 ========================================
```

## Step 5: Verify the Configuration

1. Check that the service is connected to MQTT
2. Trigger a camera failure in Frigate (or simulate one by setting camera FPS to 0)
3. Watch the logs for restart activity:
   ```
   [Observer] ALERT: DETECTED FAILURE (0 FPS) for Camera: front_door (PID: 1234)
   [Observer] ACTION: Initiating restart sequence for camera 'front_door' (Switch Device: 12a3b4c5-d678-9e01-fg23-456789h0ijkl)
   [SmartThings] ACTION: Turning OFF device 12a3b4c5-d678-9e01-fg23-456789h0ijkl
   [SmartThings] WAIT: Waiting 10 seconds before turning ON...
   [SmartThings] ACTION: Turning ON device 12a3b4c5-d678-9e01-fg23-456789h0ijkl
   [Observer] SUCCESS: Successfully completed restart sequence for camera 'front_door'
   ```

## Step 6: Docker Deployment

### Build the Image
```bash
docker build -t camera-observer:latest .
```

### Run with Docker
```bash
docker run -d \
  --name camera-observer \
  -v $(pwd)/config.yaml:/app/config.yaml:ro \
  camera-observer:latest
```

### Or Use Docker Compose
```bash
docker-compose up -d
```

View logs:
```bash
docker logs -f camera-observer
```

## Step 7: Production Deployment

For permanent deployment (e.g., on a NAS or server):

1. Copy the entire project to your target server
2. Build the Docker image:
   ```bash
   docker build -t camera-observer:latest .
   ```
3. Store your config.yaml securely (not in git!)
4. Run with docker-compose or your preferred orchestration
5. Set up log rotation if needed

## Troubleshooting

### Cannot connect to MQTT broker
- Verify the broker URL: `tcp://mqtt.wilkywayre.com:1883`
- Check your network connectivity
- Verify firewall rules allow outbound on port 1883

### SmartThings API token invalid
- Generate a new token at https://account.smartthings.com/tokens
- Ensure you have the correct scope permissions
- Check that the token hasn't expired (tokens are usually permanent until revoked)

### Camera not restarting
- Verify device IDs match exactly from `list-devices` output
- Check that camera names in `mapping` match Frigate stats JSON keys
- Verify the smart switch is powered on and accessible from your network
- Check SmartThings app to ensure the device responds to manual commands

### No activity in logs
- Service is running correctly - it only logs when failures are detected
- Simulate a failure by creating a test camera with 0 FPS in Frigate
- Manually toggle a switch in SmartThings app to verify API connectivity

## Security Notes
- **Never commit config.yaml with real tokens to version control**
- Store tokens in a `.gitignore`d file or use environment variables
- SmartThings tokens should be treated as passwords
- Use Docker secrets or environment variables in production
- Keep your MQTT broker access restricted if on a public network

## Monitoring

### Health Checks
Monitor these log patterns:
- `[INIT] Subscribed to topic successfully` - Service started
- `[Observer] Received stats` - Normal operation (repeats every 60s from Frigate)
- `[Observer] ALERT: DETECTED FAILURE` - Problem detected
- `[Observer] SUCCESS:` - Restart completed

### Setting Up Log Aggregation
If running in Docker Swarm or Kubernetes, configure log drivers to send logs to:
- ELK Stack
- Splunk
- CloudWatch
- Any syslog-compatible service

