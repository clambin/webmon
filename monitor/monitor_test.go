package monitor_test

import (
	"context"
	"github.com/clambin/webmon/monitor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestMonitor_Register(t *testing.T) {
	stub := &serverStub{}
	testServer := httptest.NewServer(http.HandlerFunc(stub.Handle))
	defer testServer.Close()

	m := monitor.New([]string{})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		err := m.Run(ctx, 10*time.Millisecond)
		require.NoError(t, err)
	}()

	m.Register <- monitor.SiteSpec{URL: testServer.URL}

	assert.Eventually(t, func() bool {
		entry, ok := m.GetEntry(testServer.URL)
		return ok && entry.State != nil && entry.State.Up
	}, 500*time.Millisecond, 10*time.Millisecond)

	m.Unregister <- monitor.SiteSpec{URL: testServer.URL}

	assert.Eventually(t, func() bool {
		_, ok := m.GetEntry(testServer.URL)
		return !ok
	}, 500*time.Millisecond, 10*time.Millisecond)
}
