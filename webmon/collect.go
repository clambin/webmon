package webmon

import "github.com/prometheus/client_golang/prometheus"

func (monitor *Monitor) Describe(ch chan<- *prometheus.Desc) {
	ch <- monitor.metricUp
	ch <- monitor.metricLatency
	ch <- monitor.metricCertAge
}

func (monitor *Monitor) Collect(ch chan<- prometheus.Metric) {
	monitor.lock.RLock()
	defer monitor.lock.RUnlock()

	for name, entry := range monitor.sites {
		if entry.LastCheck.IsZero() == false {
			up := 0.0
			if entry.Up {
				up = 1.0
			}
			ch <- prometheus.MustNewConstMetric(monitor.metricUp, prometheus.GaugeValue, up, name.String())
			ch <- prometheus.MustNewConstMetric(monitor.metricLatency, prometheus.GaugeValue, entry.Latency.Seconds(), name.String())
			ch <- prometheus.MustNewConstMetric(monitor.metricCertAge, prometheus.GaugeValue, entry.CertificateAge, name.String())
		}
	}
}
