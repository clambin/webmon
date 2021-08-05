package webmon

import (
	"encoding/json"
	"net/http"
	"time"
)

type jsonEntry struct {
	Up             bool      `json:"up"`
	Latency        float64   `json:"latency"`
	CertificateAge float64   `json:"certificate_age,omitempty"`
	LastCheck      time.Time `json:"last_check"`
}

func convert(input map[string]Entry) map[string]jsonEntry {
	output := make(map[string]jsonEntry)

	for site, entry := range input {
		output[site] = jsonEntry{
			Up:             entry.Up,
			Latency:        entry.Latency.Seconds(),
			CertificateAge: entry.CertificateAge,
			LastCheck:      entry.LastCheck,
		}
	}
	return output
}

// Health reports back all configured sites and their state
func (monitor *Monitor) Health(w http.ResponseWriter, _ *http.Request) {
	monitor.lock.RLock()
	defer monitor.lock.RUnlock()

	running := false
	for _, entry := range monitor.sites {
		if entry.LastCheck.IsZero() == false {
			running = true
			break
		}
	}

	if running == false {
		http.Error(w, "Server not running yet", http.StatusNotFound)
		return
	}

	out, err := json.MarshalIndent(convert(monitor.sites), "", "  ")

	if err == nil {
		_, _ = w.Write(out)
	} else {
		http.Error(w, "unable to create response: "+err.Error(), http.StatusInternalServerError)
	}
}
