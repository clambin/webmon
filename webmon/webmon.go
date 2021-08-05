package webmon

import (
	"context"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type Monitor struct {
	HTTPClient    *http.Client
	sites         map[*url.URL]Entry
	lock          sync.RWMutex
	metricUp      *prometheus.Desc
	metricLatency *prometheus.Desc
	metricCertAge *prometheus.Desc
}

type Entry struct {
	Up             bool
	CertificateAge float64
	Latency        time.Duration
	LastCheck      time.Time
}

func New(hosts []string) (monitor *Monitor, err error) {
	sites := make(map[*url.URL]Entry)

	for _, host := range hosts {
		var parsedURL *url.URL
		parsedURL, err = url.Parse(host)

		if err != nil {
			return nil, fmt.Errorf("invalid URL '%s': %v", host, err)
		}
		sites[parsedURL] = Entry{}
	}

	monitor = &Monitor{
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		sites: sites,
		metricUp: prometheus.NewDesc(
			prometheus.BuildFQName("webmon", "site", "up"),
			"Set to 1 if the site is up",
			[]string{"site_url"},
			nil,
		),
		metricLatency: prometheus.NewDesc(
			prometheus.BuildFQName("webmon", "site", "latency_seconds"),
			"Time to check the site in seconds",
			[]string{"site_url"},
			nil,
		),
		metricCertAge: prometheus.NewDesc(
			prometheus.BuildFQName("webmon", "certificate", "expiry"),
			"Measures how long until a certificate expires",
			[]string{"site_url"},
			nil,
		),
	}

	return
}

func (monitor *Monitor) Run(ctx context.Context, interval time.Duration) (err error) {
	monitor.CheckSites(ctx)

	ticker := time.NewTicker(interval)
	for running := true; running; {
		select {
		case <-ctx.Done():
			running = false
		case <-ticker.C:
			monitor.CheckSites(ctx)
		}
	}

	ticker.Stop()
	return
}

func (monitor *Monitor) CheckSites(ctx context.Context) {
	monitor.lock.Lock()
	defer monitor.lock.Unlock()

	// TODO: probably want to limit how many sites we poll concurrently
	responses := make(map[*url.URL]chan Entry)
	for site := range monitor.sites {
		responses[site] = make(chan Entry)

		go func(ch chan Entry, site *url.URL) {
			ch <- monitor.checkSite(ctx, site)
		}(responses[site], site)
	}

	for site, ch := range responses {
		monitor.sites[site] = <-ch
	}
}
