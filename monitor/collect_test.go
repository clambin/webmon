package monitor_test

import (
	"context"
	"github.com/clambin/webmon/monitor"
	"github.com/prometheus/client_golang/prometheus"
	pcg "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestCollector_Describe(t *testing.T) {
	m := monitor.New([]string{"localhost"})

	metrics := make(chan *prometheus.Desc)
	go m.Describe(metrics)

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

	// TODO: check labels
	metric := <-ch
	assert.Equal(t, 1.0, metricValue(metric).GetGauge().GetValue())
	metric = <-ch
	assert.NotZero(t, metricValue(metric).GetGauge().GetValue())
	metric = <-ch
	assert.Zero(t, metricValue(metric).GetGauge().GetValue())

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
	assert.Equal(t, 1.0, metricValue(metric).GetGauge().GetValue())
	metric = <-ch
	assert.Less(t, metricValue(metric).GetGauge().GetValue(), 0.1)
	metric = <-ch
	assert.NotZero(t, metricValue(metric).GetGauge().GetValue())

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

	go func() {
		err2 := m.Run(ctx, 10*time.Millisecond)
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
			go m.Collect(ch)

			metric := <-ch
			up := metricValue(metric).GetGauge().GetValue()

			if up == 1.0 {
				_ = <-ch
				_ = <-ch
			}
			return up == testCase.up
		}, 500*time.Millisecond, 10*time.Millisecond)
	}
}

func BenchmarkMonitor_CheckSites(b *testing.B) {
	stub := &serverStub{}
	testServer := httptest.NewTLSServer(http.HandlerFunc(stub.Handle))
	defer testServer.Close()

	m := monitor.New([]string{testServer.URL})
	// allow the client to recognize the server during HTTPS TLS handshake
	m.HTTPClient = testServer.Client()

	ctx := context.Background()
	b.ResetTimer()

	for i := 0; i < 5000; i++ {
		m.CheckSites(ctx)
	}
}

func BenchmarkMonitor_Parallel(b *testing.B) {
	stub := &serverStub{}
	var testServers []*httptest.Server
	var urls []string

	for i := 0; i < 10; i++ {
		testServer := httptest.NewServer(http.HandlerFunc(stub.Handle))
		testServers = append(testServers, testServer)
		urls = append(urls, testServer.URL)
	}

	m := monitor.New(urls)
	// m.MaxConcurrentChecks = 3

	ctx := context.Background()
	b.ResetTimer()

	for i := 0; i < 1000; i++ {
		m.CheckSites(ctx)
	}

	for _, testServer := range testServers {
		testServer.Close()
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
