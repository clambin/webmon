package webmon

import (
	"context"
	"crypto/tls"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func (monitor *Monitor) checkSite(ctx context.Context, site *url.URL) (entry Entry){
	log.WithField("site", site).Debug("checking site")

	start := time.Now()
	entry.Up = monitor.pingSite(ctx, site)

	if entry.Up {
		entry.Latency = time.Now().Sub(start)

		var err error
		entry.CertificateAge, err = monitor.getCertificateLifetime(site)

		if err != nil {
			log.WithError(err).Warning("failed to get HTTPS certificate expiry")
		}
	}
	entry.LastCheck = time.Now()

	return
}

func (monitor *Monitor) pingSite(ctx context.Context, site *url.URL) (up bool) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, site.String(), nil)

	if err == nil {
		var resp *http.Response
		resp, err = monitor.HTTPClient.Do(req)

		if err == nil {
			up =  validStatusCode(resp.StatusCode)
			_ = resp.Body.Close()
		}
	}

	log.WithError(err).WithFields(log.Fields{"site": site.String(), "up": up}).Debug("pingSite")
	return
}

var validHTTPStatusCodes = []int {
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

func (monitor *Monitor) getCertificateLifetime(site *url.URL) (expiry float64, err error) {
	siteHost :=  strings.Split(site.Host, ":")

	if len(siteHost) == 1 {
		switch site.Scheme {
		case "http":
			siteHost = append(siteHost, "80")
		case "https":
			siteHost = append(siteHost, "443")
		default:
			return 0.0, fmt.Errorf("unsupported scheme: %s", site.Scheme)
		}
	}

	var config *tls.Config
	if monitor.RootCAs != nil {
		config = &tls.Config{
			RootCAs: monitor.RootCAs,
		}
	}

	var conn *tls.Conn
	conn, err = tls.Dial("tcp", siteHost[0] + ":" + siteHost[1], config)

	if err == nil {
		expiry = conn.ConnectionState().PeerCertificates[0].NotAfter.Sub(time.Now()).Hours() / 24
	}

	log.WithError(err).WithFields(log.Fields{"name": siteHost, "expiry": expiry}).Debug("getCertificateLifetime")

	return
}

