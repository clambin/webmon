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
	Register chan string
	// The Unregister channel is used to remove a host from the monitor
	Unregister chan string
	// HTTPClient is the http.Client that will be used to check sites.
	// Under normal circumstances, this can be left blank and Monitor will create the required client.
	HTTPClient *http.Client
	// MaxConcurrentChecks limits the number of sites that are checked in parallel. Default: DefaultMaxConcurrentChecks
	MaxConcurrentChecks int64
	sites               map[string]Entry
	lock                sync.RWMutex
	metricUp            *prometheus.Desc
	metricLatency       *prometheus.Desc
	metricCertAge       *prometheus.Desc
}

// The Entry structure holds the attributes that will be checked
type Entry struct {
	// Up indicates if the site is up or down
	Up bool
	// CertificateAge contains the number of days that the site's TLS certificate is still valid
	// For HTTP sites, this will be zero.
	CertificateAge float64
	// Latency contains the time it took to check the site
	Latency time.Duration
	// LastCheck is the timestamp the site was last checked. Before there first check, this is zero
	LastCheck time.Time
}

// New creates a new Monitor instance for the specified list of sites
func New(hosts []string) (monitor *Monitor) {
	monitor = &Monitor{
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		Register:   make(chan string),
		Unregister: make(chan string),
		sites:      make(map[string]Entry),
		metricUp: prometheus.NewDesc(
			prometheus.BuildFQName("webmon", "site", "up"),
			"Set to 1 if the site is up",
			[]string{"site_url"},
			nil,
		),
		metricLatency: prometheus.NewDesc(
			prometheus.BuildFQName("webmon", "site", "latency_seconds"),
			"Time to check the site, in seconds",
			[]string{"site_url"},
			nil,
		),
		metricCertAge: prometheus.NewDesc(
			prometheus.BuildFQName("webmon", "certificate", "expiry"),
			"Number of days before the HTTPS certificate expires",
			[]string{"site_url"},
			nil,
		),
	}

	for _, host := range hosts {
		monitor.sites[host] = Entry{}
	}

	return
}

// Run calls CheckSites on a recurring basis, based on the specified interval.
// To be able to stop this function, call it with a context obtained by context.WithCancel()
// and then call cancel() when required.
func (monitor *Monitor) Run(ctx context.Context, interval time.Duration) (err error) {
	monitor.CheckSites(ctx)

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
	return
}

func (monitor *Monitor) register(site string) {
	monitor.lock.Lock()
	defer monitor.lock.Unlock()

	if _, ok := monitor.sites[site]; ok == false {
		log.WithField("url", site).Info("registering new url")
		monitor.sites[site] = Entry{}
	}
}

func (monitor *Monitor) unregister(site string) {
	monitor.lock.Lock()
	defer monitor.lock.Unlock()

	log.WithField("url", site).Info("unregistering url")
	delete(monitor.sites, site)
}

// GetEntry returns the monitor's entry for the specified site name.
// Should only be used for testing purposes
func (monitor *Monitor) GetEntry(site string) (entry Entry, ok bool) {
	monitor.lock.RLock()
	defer monitor.lock.RUnlock()

	entry, ok = monitor.sites[site]

	return
}
