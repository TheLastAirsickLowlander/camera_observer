# Deployment Guide - Camera Observer

This guide walks you through deploying the Camera Observer service to production.

## Prerequisites

- Docker and Docker Compose installed on target system
- Network access to MQTT broker (`mqtt.wilkywayre.com:1883`)
- Samsung SmartThings account with API token
- SmartThings devices (smart switches/plugs) for camera power control

## Pre-Deployment Tasks

### 1. Get SmartThings Token
1. Visit https://account.smartthings.com/tokens
2. Generate a token with **Devices** scope (read/write)
3. Copy the token securely

### 2. Identify Your Devices
```bash
# Run locally first to list devices
go run cmd/list-devices/main.go --token "YOUR_TOKEN_HERE"
```

Output will show device IDs for your smart switches.

### 3. Create Production config.yaml
```yaml
mqtt:
  broker: "tcp://mqtt.wilkywayre.com:1883"
  topic: "frigate/stats"
  client_id: "camera_observer"

smartthings:
  api_token: "YOUR_ACTUAL_TOKEN_HERE"

mapping:
  "front_door": "actual-device-id-001"
  "backyard": "actual-device-id-002"
```

**IMPORTANT**: Do NOT commit this file to version control if it contains real tokens.

## Deployment Options

### Option 1: Direct Docker Compose (Recommended)

#### Step 1: Copy Project to Server
```bash
scp -r camera_observer user@your-server:/opt/
```

#### Step 2: Set Up Configuration
```bash
ssh user@your-server
cd /opt/camera_observer
# Edit config.yaml with your real settings
nano config.yaml
```

#### Step 3: Start Service
```bash
docker-compose up -d
```

#### Step 4: Verify
```bash
docker logs -f camera-observer
```

Look for:
```
[INIT] Subscribed to topic successfully
SERVICE RUNNING
```

### Option 2: Pre-Built Docker Image (For NAS)

If your NAS supports Docker:

```bash
# Build on development machine
docker build -t camera-observer:1.0 .
docker save camera-observer:1.0 | gzip > camera-observer.tar.gz

# Transfer to NAS
scp camera-observer.tar.gz user@nas:/tmp/

# On NAS, import image
docker load < /tmp/camera-observer.tar.gz

# Run container
docker run -d \
  --name camera-observer \
  --restart unless-stopped \
  -v /path/to/config.yaml:/root/config.yaml \
  camera-observer:1.0
```

### Option 3: Manual Deployment (No Docker)

If Docker is not available:

```bash
# Copy project to server
scp -r camera_observer user@your-server:/opt/

# SSH into server
ssh user@your-server
cd /opt/camera_observer

# Build
go build -o camera-observer cmd/camera-observer/main.go

# Create systemd service file
sudo tee /etc/systemd/system/camera-observer.service > /dev/null <<EOF
[Unit]
Description=Camera Observer Service
After=network.target

[Service]
Type=simple
User=nobody
WorkingDirectory=/opt/camera_observer
ExecStart=/opt/camera_observer/camera-observer
Restart=unless-stopped
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

# Enable and start
sudo systemctl daemon-reload
sudo systemctl enable camera-observer
sudo systemctl start camera-observer

# Check status
sudo systemctl status camera-observer
```

## Configuration Methods

### Method 1: config.yaml (Simple)
- Edit `config.yaml` directly
- Best for simple deployments
- Default method

### Method 2: Environment Variables (Secure)
```bash
export MQTT_BROKER="tcp://mqtt.wilkywayre.com:1883"
export MQTT_TOPIC="frigate/stats"
export MQTT_CLIENT_ID="camera_observer"
export SMARTTHINGS_API_TOKEN="your-token-here"

go run cmd/camera-observer/main.go
```

### Method 3: Docker Secrets (Most Secure)
For Docker Swarm:

```bash
echo "your-token-here" | docker secret create smartthings_token -

docker service create \
  --name camera-observer \
  --secret smartthings_token \
  -e SMARTTHINGS_API_TOKEN_FILE=/run/secrets/smartthings_token \
  camera-observer:latest
```

## Monitoring & Troubleshooting

### Check Logs
```bash
# Docker
docker logs -f camera-observer

# Systemd
sudo journalctl -u camera-observer -f
```

### Common Issues

#### Connection Failed
```
[ERROR] Failed to connect to MQTT broker
```
- Verify broker URL is correct
- Check firewall allows outbound 1883
- Verify network connectivity

#### Invalid SmartThings Token
```
[ERROR] Failed to restart device
api returned non-200 status: 401
```
- Generate new token at https://account.smartthings.com/tokens
- Verify token permissions (Devices scope)

#### No Events Detected
This is normal! The service only logs when:
- A camera fails (FPS = 0)
- A restart is triggered
- An error occurs

To verify it's running, watch for periodic log entries or trigger a test failure in Frigate.

### Logs to Watch

**Normal Operation:**
```
[INIT] Subscribed to topic successfully
[Observer] Received stats for 3 cameras
```

**Failure Detected:**
```
[Observer] ALERT: DETECTED FAILURE (0 FPS) for Camera: front_door
[Observer] ACTION: Initiating restart sequence
[SmartThings] ACTION: Turning OFF device xxx
[SmartThings] ACTION: Turning ON device xxx
[Observer] SUCCESS: Successfully completed restart sequence
```

**On Cooldown:**
```
[Observer] SKIPPED: Restart for front_door is on cooldown
```

## Health Checks

### Manual Health Check
```bash
# Check container is running
docker ps | grep camera-observer

# Check logs for errors
docker logs camera-observer | grep ERROR
```

### Automated Monitoring

Set up alerts for these patterns:
- `[ERROR]` - Any errors
- `Failed to restart` - Restart failures
- `connection` error - MQTT connectivity issues

## Backup & Recovery

### Backup Configuration
```bash
# Backup config.yaml (keep secure!)
scp user@server:/opt/camera_observer/config.yaml ~/backups/config.yaml.bak
```

### Recovery
```bash
# Restore config
scp ~/backups/config.yaml.bak user@server:/opt/camera_observer/config.yaml

# Restart service
docker restart camera-observer
# OR
sudo systemctl restart camera-observer
```

## Performance

### Resource Usage
- Memory: ~20-30 MB
- CPU: Minimal (event-driven)
- Network: ~1KB per camera per minute (stats polling)

### Optimization Tips
- Use official MQTT broker for reliability
- Place service near MQTT broker for low latency
- Monitor SmartThings API response times in logs

## Security Hardening

### 1. Secure Token Storage
```bash
# Use environment file instead of config.yaml
sudo nano /etc/camera-observer.env
# Add: SMARTTHINGS_API_TOKEN=your-token-here
# Set permissions
sudo chmod 600 /etc/camera-observer.env
```

### 2. Restrict Network Access
```bash
# Only allow MQTT broker IP
sudo ufw allow from MQTT_IP to any port 1883
```

### 3. Use MQTT Authentication
If your broker supports authentication:
```yaml
mqtt:
  broker: "tcp://username:password@mqtt.wilkywayre.com:1883"
```

### 4. Container Security
```bash
# Run as non-root (already in Dockerfile)
# Read-only filesystem
docker run --read-only ...

# Drop capabilities
docker run --cap-drop=ALL ...
```

## Scaling

For multiple instances/cameras:

1. **Load Balancing**: Connect multiple observers to same MQTT broker
2. **Separation of Concerns**: Run separate instances per building/zone
3. **Configuration Management**: Use environment variables for flexibility

## Maintenance

### Log Rotation
```bash
# Docker - automatic via log driver
# Systemd
sudo nano /etc/logrotate.d/camera-observer
```

### Updates
```bash
# Pull latest code
git pull origin main

# Rebuild image
docker build -t camera-observer:2.0 .

# Stop old container
docker stop camera-observer
docker rm camera-observer

# Start new container
docker-compose up -d
```

### Health Verification (Monthly)
- [ ] Check logs for errors
- [ ] Verify cameras are restarting properly
- [ ] Test manual restart to verify SmartThings connectivity
- [ ] Check memory usage isn't increasing

## Disaster Recovery

### Complete System Failure
```bash
# If everything is lost but config.yaml backup exists:

# 1. Set up new server with Docker
# 2. Copy Camera Observer project
# 3. Restore config.yaml from backup
# 4. Run docker-compose up -d
# 5. Verify in logs
```

### Partial Failure (Service Crashed)
```bash
# Docker auto-restart will handle this (restart: unless-stopped)
# If manual restart needed:
docker restart camera-observer
```

## Support & Debugging

### Enable Debug Logging
Add to `config.yaml`:
```yaml
debug: true
```

### Common Commands
```bash
# View all logs
docker logs camera-observer

# View last 100 lines
docker logs --tail 100 camera-observer

# Follow logs in real-time
docker logs -f camera-observer

# Export logs
docker logs camera-observer > logs.txt

# Restart service
docker restart camera-observer

# Stop service
docker stop camera-observer

# Remove stopped container
docker rm camera-observer
```

## Rollback Plan

If new version causes issues:

```bash
# 1. Stop current version
docker stop camera-observer

# 2. Remove current container
docker rm camera-observer

# 3. Use previous image
docker-compose up -d --force-recreate

# 4. Verify in logs
docker logs -f camera-observer
```

## Success Indicators

After deployment, verify:

- ✅ Service starts without errors
- ✅ Service stays running (check after 24 hours)
- ✅ Logs show "SERVICE RUNNING"
- ✅ MQTT subscription successful
- ✅ SmartThings API connectivity works
- ✅ No memory leaks (memory stable over time)
- ✅ Cameras restart when failures occur

## Next Steps

1. **Monitor** - Set up log aggregation and alerting
2. **Document** - Record your device IDs and token location
3. **Test** - Intentionally trigger a camera failure to verify restart
4. **Optimize** - Adjust cooldown times if needed
5. **Backup** - Backup your config.yaml securely

---

**Questions?** Check the README.md, SETUP.md, or ARCHITECTURE.md files for more detailed information.

