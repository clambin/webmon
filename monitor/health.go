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

	sites := make(map[string]Entry)
	var lastUpdate time.Time
	for url, entry := range monitor.sites {
		if entry.LastCheck.IsZero() == false {
			sites[url] = entry
			if entry.LastCheck.After(lastUpdate) {
				lastUpdate = entry.LastCheck
			}
		}
	}

	health := struct {
		Count      int              `json:"count"`
		LastUpdate time.Time        `json:"last_update,omitempty"`
		Sites      map[string]Entry `json:"sites"`
	}{
		Count:      len(monitor.sites),
		LastUpdate: lastUpdate,
		Sites:      sites,
	}

	out, _ := json.MarshalIndent(&health, "", "  ")
	_, _ = w.Write(out)
}
