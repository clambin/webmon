package webmon

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

	responses := make(map[string]chan Entry)
	for site := range monitor.sites {
		responses[site] = make(chan Entry)

		_ = maxJobs.Acquire(ctx, 1)
		go func(ch chan Entry, site string) {
			entry := monitor.checkSite(ctx, site)
			maxJobs.Release(1)
			ch <- entry
		}(responses[site], site)
	}

	for site, ch := range responses {
		monitor.sites[site] = <-ch
	}
}

func (monitor *Monitor) checkSite(ctx context.Context, site string) (entry Entry) {
	log.WithField("site", site).Debug("checking site")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, site, nil)

	if err == nil {
		var resp *http.Response
		start := time.Now()
		resp, err = monitor.HTTPClient.Do(req)

		if err == nil {
			entry.Up = validStatusCode(resp.StatusCode)
			entry.Latency = time.Now().Sub(start)

			if resp.TLS != nil && len(resp.TLS.PeerCertificates) > 0 {
				entry.CertificateAge = resp.TLS.PeerCertificates[0].NotAfter.Sub(time.Now()).Hours() / 24
			}

			_ = resp.Body.Close()
		}
	}

	entry.LastCheck = time.Now()

	log.WithError(err).WithFields(log.Fields{
		"site":    site,
		"up":      entry.Up,
		"certAge": entry.CertificateAge,
		"latency": entry.Latency,
	}).Debug("checkSite")
	return
}

var validHTTPStatusCodes = []int{
	http.StatusOK,
	http.StatusUnauthorized,
	http.StatusTemporaryRedirect,
}

func validStatusCode(statusCode int) bool {
	for _, code := range validHTTPStatusCodes {
		if statusCode == code {
			return true
		}
	}
	return false
}
