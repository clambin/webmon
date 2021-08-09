package monitor_test

import (
	"context"
	"github.com/clambin/webmon/monitor"
	"github.com/stretchr/testify/assert"
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
		assert.NoError(t, err)
	}()

	m.Register <- testServer.URL

	assert.Eventually(t, func() bool {
		entry, ok := m.GetEntry(testServer.URL)
		return ok && entry.Up
	}, 500*time.Millisecond, 10*time.Millisecond)

	m.Unregister <- testServer.URL

	assert.Eventually(t, func() bool {
		_, ok := m.GetEntry(testServer.URL)
		return !ok
	}, 500*time.Millisecond, 10*time.Millisecond)
}
