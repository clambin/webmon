package monitor

import (
	"context"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/semaphore"
	"net/http"
	"time"
)

// CheckSites checks each site. The site's status isn't reported here, but is kept internally to be scraped by Prometheus
// using the Collect function.
func (monitor *Monitor) CheckSites(ctx context.Context) {
	monitor.lock.Lock()
	defer monitor.lock.Unlock()

	if monitor.MaxConcurrentChecks == 0 {
		monitor.MaxConcurrentChecks = DefaultMaxConcurrentChecks
	}
	maxJobs := semaphore.NewWeighted(monitor.MaxConcurrentChecks)

	responses := make(map[string]chan *SiteState)
	for site := range monitor.sites {
		responses[site] = make(chan *SiteState)

		_ = maxJobs.Acquire(ctx, 1)
		go func(ch chan *SiteState, site string) {
			state := monitor.checkSite(ctx, site)
			maxJobs.Release(1)
			ch <- state
		}(responses[site], site)
	}

	for site, ch := range responses {
		entry, _ := monitor.sites[site]
		entry.State = <-ch
		monitor.sites[site] = entry
	}
}

func (monitor *Monitor) checkSite(ctx context.Context, site string) (state *SiteState) {
	log.WithField("site", site).Debug("checking site")

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, site, nil)

	start := time.Now()
	resp, err := monitor.HTTPClient.Do(req)

	state = &SiteState{}
	if err != nil {
		state.LastError = err.Error()
		return
	}

	state.HTTPCode = resp.StatusCode
	state.Up = goodStatusCode(resp.StatusCode)
	state.Latency = Duration{Duration: time.Now().Sub(start)}

	if resp.TLS != nil && len(resp.TLS.PeerCertificates) > 0 {
		state.IsTLS = true
		state.CertificateAge = resp.TLS.PeerCertificates[0].NotAfter.Sub(time.Now()).Hours() / 24
	}

	_ = resp.Body.Close()

	state.LastCheck = time.Now()

	log.WithError(err).WithFields(log.Fields{
		"site":    site,
		"up":      state.Up,
		"certAge": state.CertificateAge,
		"latency": state.Latency,
	}).Debug("checkSite")
	return
}

var goodHTTPStatusCodes = []int{
	http.StatusOK,
	http.StatusUnauthorized,
	http.StatusTemporaryRedirect,
	http.StatusFound,
}

func goodStatusCode(statusCode int) bool {
	for _, code := range goodHTTPStatusCodes {
		if statusCode == code {
			return true
		}
	}
	return false
}
