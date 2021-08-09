package monitor

import (
	"encoding/json"
	"net/http"
)

// Health reports back all configured sites and their state
func (monitor *Monitor) Health(w http.ResponseWriter, _ *http.Request) {
	monitor.lock.RLock()
	defer monitor.lock.RUnlock()

	out, err := json.MarshalIndent(monitor.sites, "", "  ")

	if err == nil {
		_, _ = w.Write(out)
	} else {
		http.Error(w, "unable to create response: "+err.Error(), http.StatusInternalServerError)
	}
}
