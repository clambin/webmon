package webmon_test

import (
	"context"
	"crypto/x509"
	"github.com/clambin/webmon/webmon"
	"github.com/prometheus/client_golang/prometheus"
	pcg "github.com/prometheus/client_model/go"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestCollector_Describe(t *testing.T) {
	monitor, _ := webmon.New([]string{"localhost"})

	metrics := make(chan *prometheus.Desc)
	go monitor.Describe(metrics)

	for _, name := range []string{
		"webmon_site_up",
		"webmon_site_latency_seconds",
		"webmon_certificate_expiry",
	} {
		metric := <-metrics
		assert.Contains(t, metric.String(), "\""+name+"\"")
	}
}

func TestCollector_Collect(t *testing.T) {
	stub := &serverStub{}
	testServer := httptest.NewServer(http.HandlerFunc(stub.Handle))
	defer testServer.Close()

	monitor, err := webmon.New([]string{testServer.URL})
	assert.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	monitor.CheckSites(ctx)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		err2 := monitor.Run(ctx, 1*time.Minute)
		assert.NoError(t, err2)
		wg.Done()
	}()

	ch := make(chan prometheus.Metric)
	go monitor.Collect(ch)

	// TODO: check labels
	m := <-ch
	assert.Equal(t, 1.0, metricValue(m).GetGauge().GetValue())
	m = <-ch
	assert.NotZero(t, metricValue(m).GetGauge().GetValue())
	m = <-ch
	assert.Zero(t, metricValue(m).GetGauge().GetValue())

	cancel()

	wg.Wait()
}

func TestCollector_Collect_TLS(t *testing.T) {
	stub := &serverStub{}
	testServer := httptest.NewTLSServer(http.HandlerFunc(stub.Handle))
	defer testServer.Close()

	monitor, err := webmon.New([]string{testServer.URL})
	assert.NoError(t, err)
	// allow the client to recognize the server during HTTPS TLS handshake
	monitor.HTTPClient = testServer.Client()
	// allow the client to recognize the server during tls.Dial TSL handshake
	pool := x509.NewCertPool()
	pool.AddCert(testServer.Certificate())
	monitor.RootCAs = pool

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	monitor.CheckSites(ctx)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		err2 := monitor.Run(ctx, 1*time.Minute)
		assert.NoError(t, err2)
		wg.Done()
	}()

	ch := make(chan prometheus.Metric)
	go monitor.Collect(ch)

	m := <-ch
	assert.Equal(t, 1.0, metricValue(m).GetGauge().GetValue())
	m = <-ch
	assert.Less(t, metricValue(m).GetGauge().GetValue(), 0.1)
	m = <-ch
	assert.NotZero(t, metricValue(m).GetGauge().GetValue())

	cancel()

	wg.Wait()
}

func TestCollector_Collect_StatusCodes(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	stub := &serverStub{}
	testServer := httptest.NewTLSServer(http.HandlerFunc(stub.Handle))
	defer testServer.Close()

	monitor, err := webmon.New([]string{testServer.URL})
	assert.NoError(t, err)
	// allow the client to recognize the server during HTTPS TLS handshake
	monitor.HTTPClient = testServer.Client()
	// allow the client to recognize the server during tls.Dial TSL handshake
	pool := x509.NewCertPool()
	pool.AddCert(testServer.Certificate())
	monitor.RootCAs = pool

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		err2 := monitor.Run(ctx, 10*time.Millisecond)
		assert.NoError(t, err2)
	}()

	type testCaseStruct struct {
		statusCode int
		up         float64
	}

	testCases := []testCaseStruct{
		{statusCode: http.StatusOK, up: 1.0},
		{statusCode: http.StatusNotFound, up: 0.0},
		{statusCode: http.StatusTemporaryRedirect, up: 1.0},
	}

	for _, testCase := range testCases {
		stub.StatusCode(testCase.statusCode)

		assert.Eventually(t, func() bool {
			ch := make(chan prometheus.Metric)
			go monitor.Collect(ch)

			m := <-ch
			_ = <-ch
			_ = <-ch
			return metricValue(m).GetGauge().GetValue() == testCase.up
		}, 500*time.Millisecond, 10*time.Millisecond)
	}
}

func BenchmarkMonitor_CheckSites(b *testing.B) {
	stub := &serverStub{}
	testServer := httptest.NewTLSServer(http.HandlerFunc(stub.Handle))
	defer testServer.Close()

	monitor, err := webmon.New([]string{testServer.URL})
	assert.NoError(b, err)
	// allow the client to recognize the server during HTTPS TLS handshake
	monitor.HTTPClient = testServer.Client()
	// allow the client to recognize the server during tls.Dial TSL handshake
	pool := x509.NewCertPool()
	pool.AddCert(testServer.Certificate())
	monitor.RootCAs = pool

	ctx := context.Background()
	b.ResetTimer()

	for i := 0; i < 1000; i++ {
		monitor.CheckSites(ctx)
	}

}

type serverStub struct {
	statusCode int
	lock       sync.RWMutex
}

func (stub *serverStub) StatusCode(statusCode int) {
	stub.lock.Lock()
	defer stub.lock.Unlock()
	stub.statusCode = statusCode
}

func (stub *serverStub) Handle(w http.ResponseWriter, _ *http.Request) {
	stub.lock.RLock()
	defer stub.lock.RUnlock()

	if stub.statusCode == 0 {
		stub.statusCode = http.StatusOK
	}

	w.WriteHeader(stub.statusCode)
}

// metricValue checks that a prometheus metric has a specified value
func metricValue(metric prometheus.Metric) *pcg.Metric {
	m := new(pcg.Metric)
	if metric.Write(m) != nil {
		panic("failed to parse metric")
	}

	return m
}
