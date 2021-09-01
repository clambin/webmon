package monitor_test

import (
	"context"
	"github.com/clambin/webmon/monitor"
	"net/http"
	"net/http/httptest"
	"testing"
)

func BenchmarkMonitor_CheckSites(b *testing.B) {
	stub := &serverStub{}
	testServer := httptest.NewTLSServer(http.HandlerFunc(stub.Handle))
	defer testServer.Close()

	m := monitor.New([]string{testServer.URL})
	// allow the client to recognize the server during HTTPS TLS handshake
	m.HTTPClient = testServer.Client()

	ctx := context.Background()
	b.ResetTimer()

	for i := 0; i < 100; i++ {
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

	for i := 0; i < 100; i++ {
		m.CheckSites(ctx)
	}

	for _, testServer := range testServers {
		testServer.Close()
	}
}
