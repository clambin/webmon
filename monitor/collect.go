package monitor

import (
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"time"
)

// Describe implements the prometheus collector Describe interface
func (monitor *Monitor) Describe(ch chan<- *prometheus.Desc) {
	ch <- monitor.metricUp
	ch <- monitor.metricLatency
	ch <- monitor.metricCertAge
}

// Collect implements the prometheus collector Collect interface
func (monitor *Monitor) Collect(ch chan<- prometheus.Metric) {
	start := time.Now()
	monitor.lock.RLock()
	defer monitor.lock.RUnlock()

	for name, entry := range monitor.sites {
		if entry.LastCheck.IsZero() == false {
			if entry.Up {
				ch <- prometheus.MustNewConstMetric(monitor.metricUp, prometheus.GaugeValue, 1.0, name)
				ch <- prometheus.MustNewConstMetric(monitor.metricLatency, prometheus.GaugeValue, entry.Latency.Seconds(), name)
				ch <- prometheus.MustNewConstMetric(monitor.metricCertAge, prometheus.GaugeValue, entry.CertificateAge, name)
			} else {
				ch <- prometheus.MustNewConstMetric(monitor.metricUp, prometheus.GaugeValue, 0.0, name)
			}
		}
	}

	log.WithField("duration", time.Now().Sub(start)).Debug("prometheus scrape done")
}
