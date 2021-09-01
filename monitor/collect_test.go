package monitor_test

import (
	"context"
	"github.com/clambin/gotools/metrics"
	"github.com/clambin/webmon/monitor"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestCollector_Describe(t *testing.T) {
	m := monitor.New([]string{"localhost"})

	ch := make(chan *prometheus.Desc)
	go m.Describe(ch)

	for _, name := range []string{
		"webmon_site_up",
		"webmon_site_latency_seconds",
		"webmon_certificate_expiry",
	} {
		metric := <-ch
		assert.Contains(t, metric.String(), "\""+name+"\"")
	}
}

func TestCollector_Collect(t *testing.T) {
	stub := &serverStub{}
	testServer := httptest.NewServer(http.HandlerFunc(stub.Handle))
	defer testServer.Close()

	m := monitor.New([]string{testServer.URL})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	m.CheckSites(ctx)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		err2 := m.Run(ctx, 1*time.Minute)
		assert.NoError(t, err2)
		wg.Done()
	}()

	ch := make(chan prometheus.Metric)
	go m.Collect(ch)

	metric := <-ch
	assert.Equal(t, 1.0, metrics.MetricValue(metric).GetGauge().GetValue())
	assert.Equal(t, testServer.URL, metrics.MetricLabel(metric, "site_url"))
	assert.Equal(t, testServer.URL, metrics.MetricLabel(metric, "site_name"))
	metric = <-ch
	assert.NotZero(t, metrics.MetricValue(metric).GetGauge().GetValue())
	// metric = <-ch
	// assert.Zero(t, metricValue(metric).GetGauge().GetValue())

	cancel()

	wg.Wait()
}

func TestCollector_Collect_TLS(t *testing.T) {
	stub := &serverStub{}
	testServer := httptest.NewTLSServer(http.HandlerFunc(stub.Handle))
	defer testServer.Close()

	m := monitor.New([]string{testServer.URL})
	// allow the client to recognize the server during HTTPS TLS handshake
	m.HTTPClient = testServer.Client()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	m.CheckSites(ctx)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		err2 := m.Run(ctx, 1*time.Minute)
		assert.NoError(t, err2)
		wg.Done()
	}()

	ch := make(chan prometheus.Metric)
	go m.Collect(ch)

	metric := <-ch
	assert.Equal(t, 1.0, metrics.MetricValue(metric).GetGauge().GetValue())
	metric = <-ch
	assert.Less(t, metrics.MetricValue(metric).GetGauge().GetValue(), 0.1)
	metric = <-ch
	assert.NotZero(t, metrics.MetricValue(metric).GetGauge().GetValue())

	cancel()

	wg.Wait()
}

func TestCollector_Collect_StatusCodes(t *testing.T) {
	stub := &serverStub{}
	testServer := httptest.NewTLSServer(http.HandlerFunc(stub.Handle))
	defer testServer.Close()

	m := monitor.New([]string{testServer.URL})
	// allow the client to recognize the server during HTTPS TLS handshake
	m.HTTPClient = testServer.Client()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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

		m.CheckSites(ctx)

		// use a buffered channel so Collect doesn't block when we don't read all metrics
		ch := make(chan prometheus.Metric, 3)
		go m.Collect(ch)

		up := metrics.MetricValue(<-ch).GetGauge().GetValue()

		assert.Equal(t, testCase.up, up)
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
