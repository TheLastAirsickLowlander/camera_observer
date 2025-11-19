package observer

import (
	"camera_observer/internal/smartthings"
	"encoding/json"
	"log"
	"sync"
	"time"
)

// CameraStats represents the specific metrics we care about from Frigate.
type CameraStats struct {
	CameraFPS float64 `json:"camera_fps"`
	PID       int     `json:"pid"`
}

// StatsPayload represents the "frigate/stats" JSON structure.
type StatsPayload struct {
	Cameras map[string]CameraStats `json:"cameras"`
}

type Manager struct {
	stClient  *smartthings.Client
	mapping   map[string]string
	cooldowns map[string]time.Time
	mu        sync.Mutex
}

func NewManager(st *smartthings.Client, mapping map[string]string) *Manager {
	return &Manager{
		stClient:  st,
		mapping:   mapping,
		cooldowns: make(map[string]time.Time),
	}
}

// HandleMessage processes the MQTT message from "frigate/stats".
func (m *Manager) HandleMessage(topic string, payload []byte) {
	// 1. Parse the Stats Payload
	var data StatsPayload
	if err := json.Unmarshal(payload, &data); err != nil {
		log.Printf("[Observer] ERROR: Failed to parse stats payload: %v", err)
		return
	}

	log.Printf("[Observer] Received stats for %d cameras", len(data.Cameras))

	// 2. Iterate over all cameras reported in the stats
	for cameraID, stats := range data.Cameras {
		// Check for stream failure (0 FPS)
		if stats.CameraFPS > 0 {
			continue
		}

		log.Printf("[Observer] ALERT: DETECTED FAILURE (0 FPS) for Camera: %s (PID: %d)", cameraID, stats.PID)

		// 3. Find corresponding Switch ID
		switchID, exists := m.mapping[cameraID]
		if !exists {
			// Only log this once in a while to avoid spam?
			// For now, we'll just log it. The user might have cameras they don't want to auto-restart.
			// log.Printf("[Observer] WARNING: No switch mapping found for camera: %s", cameraID)
			continue
		}

		// 4. Check Cooldown (prevent spamming restarts)
		m.mu.Lock()
		lastRun, found := m.cooldowns[cameraID]
		if found && time.Since(lastRun) < 5*time.Minute {
			// Extended cooldown to 5 minutes since stats come every minute
			// We don't want to cycle power too fast if it's a persistent hardware issue.
			log.Printf("[Observer] SKIPPED: Restart for %s is on cooldown (last restart: %v ago)", cameraID, time.Since(lastRun).Round(time.Second))
			m.mu.Unlock()
			continue
		}
		m.cooldowns[cameraID] = time.Now()
		m.mu.Unlock()

		// 5. Trigger Restart (Async)
		go func(cam, dev string) {
			log.Printf("[Observer] ACTION: Initiating restart sequence for camera '%s' (Switch Device: %s)", cam, dev)
			err := m.stClient.RestartDevice(dev)
			if err != nil {
				log.Printf("[Observer] ERROR: Failed to restart device %s for camera %s: %v", dev, cam, err)
			} else {
				log.Printf("[Observer] SUCCESS: Successfully completed restart sequence for camera '%s'", cam)
			}
		}(cameraID, switchID)
	}
}
