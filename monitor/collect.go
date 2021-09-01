package monitor

import (
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"time"
)

var (
	metricUp = prometheus.NewDesc(
		prometheus.BuildFQName("webmon", "site", "up"),
		"Set to 1 if the site is up",
		[]string{"site_url", "site_name"},
		nil,
	)
	metricLatency = prometheus.NewDesc(
		prometheus.BuildFQName("webmon", "site", "latency_seconds"),
		"Time to check the site, in seconds",
		[]string{"site_url", "site_name"},
		nil,
	)
	metricCertAge = prometheus.NewDesc(
		prometheus.BuildFQName("webmon", "certificate", "expiry"),
		"Number of days before the HTTPS certificate expires",
		[]string{"site_url", "site_name"},
		nil,
	)
)

// Describe implements the prometheus collector Describe interface
func (monitor *Monitor) Describe(ch chan<- *prometheus.Desc) {
	ch <- metricUp
	ch <- metricLatency
	ch <- metricCertAge
}

// Collect implements the prometheus collector Collect interface
func (monitor *Monitor) Collect(ch chan<- prometheus.Metric) {
	start := time.Now()
	monitor.lock.RLock()
	defer monitor.lock.RUnlock()

	for url, entry := range monitor.sites {
		if entry.State != nil && entry.State.LastCheck.IsZero() == false {
			name := entry.Spec.Name
			if name == "" {
				name = url
			}
			if entry.State.Up {
				ch <- prometheus.MustNewConstMetric(metricUp, prometheus.GaugeValue, 1.0, url, name)
				ch <- prometheus.MustNewConstMetric(metricLatency, prometheus.GaugeValue, entry.State.Latency.Seconds(), url, name)
				if entry.State.IsTLS {
					ch <- prometheus.MustNewConstMetric(metricCertAge, prometheus.GaugeValue, entry.State.CertificateAge, url, name)
				}
			} else {
				ch <- prometheus.MustNewConstMetric(metricUp, prometheus.GaugeValue, 0.0, url, name)
			}
		}
	}

	log.WithField("duration", time.Now().Sub(start)).Debug("prometheus scrape done")
}
