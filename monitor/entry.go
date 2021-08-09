package monitor

import (
	"encoding/json"
	"time"
)

// The Entry structure holds the attributes that will be checked
type Entry struct {
	// Up indicates if the site is up or down
	Up bool
	// CertificateAge contains the number of days that the site's TLS certificate is still valid
	// For HTTP sites, this will be zero.
	CertificateAge float64
	// Latency contains the time it took to check the site
	Latency Duration
	// LastCheck is the timestamp the site was last checked. Before there first check, this is zero
	LastCheck time.Time
}

type Duration time.Duration

func (d Duration) Seconds() float64 {
	return time.Duration(d).Seconds()
}

func (d Duration) MarshalJSON() (out []byte, err error) {
	return json.Marshal(d.Seconds())
}

func (d *Duration) UnmarshalJSON(b []byte) (err error) {
	var duration time.Duration
	err = json.Unmarshal(b, &duration)

	if err != nil {
		*d = Duration(duration)
	}

	return
}
