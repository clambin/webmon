package monitor

import (
	"encoding/json"
	"time"
)

// The Entry structure holds all information on one host to be checked
type Entry struct {
	// Spec contains the site's specification. See SiteSpec
	Spec SiteSpec `json:"spec"`
	// State contains the site's state. See SiteState
	State *SiteState `json:"state,omitempty"`
}

// A SiteSpec to monitor
type SiteSpec struct {
	// URL of the site
	URL string `json:"url"`
	// Name of the site
	Name string `json:"name,omitempty"`
}

// The SiteState structure holds the attributes that will be checked
type SiteState struct {
	// Up indicates if the site is up or down
	Up bool `json:"up"`
	// LastError is the last error received when checking the site
	LastError string `json:"last_error,omitempty"`
	// HTTPCode is the last HTTP Code received when checking the site
	HTTPCode int `json:"http_code,omitempty"`
	// CertificateAge contains the number of days that the site's TLS certificate is still valid
	// IsTLS indicates the site is using encryption (i.e. TLS)
	IsTLS bool `json:"is_tls"`
	// For HTTP sites, this will be zero.
	CertificateAge float64 `json:"certificate_age,omitempty"`
	// Latency contains the time it took to check the site
	Latency Duration `json:"latency,omitempty"`
	// LastCheck is the timestamp the site was last checked. Before there first check, this is zero
	LastCheck time.Time `json:"last_check,omitempty"`
}

// Duration datatype. Equivalent to time.Duration, but allows us to marshal/unmarshal Entry data structure to/from json
type Duration struct {
	time.Duration
}

// MarshalJSON marshals Duration to bytes
func (d Duration) MarshalJSON() (out []byte, err error) {
	return json.Marshal(d.String())
}

// UnmarshalJSON unmarshals bytes to a Duration
func (d *Duration) UnmarshalJSON(b []byte) (err error) {
	var v interface{}
	err = json.Unmarshal(b, &v)

	if err == nil {
		d.Duration, err = time.ParseDuration(v.(string))
	}
	return
}
