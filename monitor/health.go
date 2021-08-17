package monitor

import (
	"encoding/json"
	"net/http"
	"time"
)

// Health reports back all configured sites and their state
func (monitor *Monitor) Health(w http.ResponseWriter, _ *http.Request) {
	monitor.lock.RLock()
	defer monitor.lock.RUnlock()

	var lastUpdate time.Time
	for _, entry := range monitor.sites {
		if entry.State != nil && entry.State.LastCheck.After(lastUpdate) {
			lastUpdate = entry.State.LastCheck
		}
	}

	health := struct {
		Count      int              `json:"count"`
		LastUpdate time.Time        `json:"last_update,omitempty"`
		Sites      map[string]Entry `json:"sites"`
	}{
		Count:      len(monitor.sites),
		LastUpdate: lastUpdate,
		Sites:      monitor.sites,
	}

	out, _ := json.MarshalIndent(&health, "", "  ")
	_, _ = w.Write(out)
}
