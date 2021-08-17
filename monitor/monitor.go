package monitor

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"net/http"
	"sync"
	"time"
)

// DefaultMaxConcurrentChecks specifies the default maximum number of parallel checks
const DefaultMaxConcurrentChecks = 5

// A Monitor checks a list of website, either on a continuous basis through the Run() function, or on demand via the CheckSites method.
// See the Entry structure for attributes of a site that are checked.
type Monitor struct {
	// The Register channel is used to add a host to the monitor
	Register chan SiteSpec
	// The Unregister channel is used to remove a host from the monitor
	Unregister chan SiteSpec
	// HTTPClient is the http.Client that will be used to check sites.
	// Under normal circumstances, this can be left blank and Monitor will create the required client.
	HTTPClient *http.Client
	// MaxConcurrentChecks limits the number of sites that are checked in parallel. Default: DefaultMaxConcurrentChecks
	MaxConcurrentChecks int64

	sites         map[string]Entry
	lock          sync.RWMutex
	metricUp      *prometheus.Desc
	metricLatency *prometheus.Desc
	metricCertAge *prometheus.Desc
}

// New creates a new Monitor instance for the specified list of sites
func New(hosts []string) (monitor *Monitor) {
	monitor = &Monitor{
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		Register:   make(chan SiteSpec),
		Unregister: make(chan SiteSpec),
		sites:      make(map[string]Entry),
		metricUp: prometheus.NewDesc(
			prometheus.BuildFQName("webmon", "site", "up"),
			"Set to 1 if the site is up",
			[]string{"site_url", "site_name"},
			nil,
		),
		metricLatency: prometheus.NewDesc(
			prometheus.BuildFQName("webmon", "site", "latency_seconds"),
			"Time to check the site, in seconds",
			[]string{"site_url", "site_name"},
			nil,
		),
		metricCertAge: prometheus.NewDesc(
			prometheus.BuildFQName("webmon", "certificate", "expiry"),
			"Number of days before the HTTPS certificate expires",
			[]string{"site_url", "site_name"},
			nil,
		),
	}

	for _, host := range hosts {
		monitor.sites[host] = Entry{Spec: SiteSpec{URL: host}}
	}

	return
}

// Run calls CheckSites on a recurring basis, based on the specified interval.
// To be able to stop this function, call it with a context obtained by context.WithCancel()
// and then call cancel() when required.
func (monitor *Monitor) Run(ctx context.Context, interval time.Duration) (err error) {
	log.Info("monitor started")

	ticker := time.NewTicker(interval)
	for running := true; running; {
		select {
		case <-ctx.Done():
			running = false
		case <-ticker.C:
			monitor.CheckSites(ctx)
		case site := <-monitor.Register:
			monitor.register(site)
		case site := <-monitor.Unregister:
			monitor.unregister(site)
		}
	}
	ticker.Stop()

	log.Info("monitor stopped")
	return
}

func (monitor *Monitor) register(site SiteSpec) {
	monitor.lock.Lock()
	defer monitor.lock.Unlock()

	entry, ok := monitor.sites[site.URL]
	if ok == false {
		log.WithField("url", site.URL).Info("registering new url")
	}
	entry.Spec = site
	monitor.sites[site.URL] = entry
}

func (monitor *Monitor) unregister(site SiteSpec) {
	monitor.lock.Lock()
	defer monitor.lock.Unlock()

	log.WithField("url", site.URL).Info("unregistering url")
	delete(monitor.sites, site.URL)
}

// GetEntry returns the monitor's entry for the specified site URL.
// Should only be used for testing purposes
func (monitor *Monitor) GetEntry(url string) (entry Entry, ok bool) {
	monitor.lock.RLock()
	defer monitor.lock.RUnlock()

	entry, ok = monitor.sites[url]
	return
}
