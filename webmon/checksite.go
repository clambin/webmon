package webmon

import (
	"context"
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"time"
)

func (monitor *Monitor) checkSite(ctx context.Context, site *url.URL) (entry Entry) {
	log.WithField("site", site).Debug("checking site")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, site.String(), nil)

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
		"site":    site.String(),
		"up":      entry.Up,
		"certAge": entry.CertificateAge,
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
