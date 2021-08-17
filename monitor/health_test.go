package monitor_test

import (
	"context"
	"encoding/json"
	"github.com/clambin/webmon/monitor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMonitor_Health(t *testing.T) {
	stub := &serverStub{}
	testServer := httptest.NewTLSServer(http.HandlerFunc(stub.Handle))
	defer testServer.Close()

	m := monitor.New([]string{testServer.URL})
	// allow the client to recognize the server during HTTPS TLS handshake
	m.HTTPClient = testServer.Client()

	w := httptest.NewRecorder()
	m.Health(w, &http.Request{})
	resp := w.Result()

	m.CheckSites(context.Background())

	w = httptest.NewRecorder()
	m.Health(w, &http.Request{})
	resp = w.Result()

	require.Equal(t, http.StatusOK, resp.StatusCode)
	body, _ := io.ReadAll(resp.Body)

	var parsed map[string]interface{}
	err := json.Unmarshal(body, &parsed)
	require.NoError(t, err)

	assert.Equal(t, 1.0, parsed["count"].(float64))
	assert.NotZero(t, parsed["last_update"].(string))

	sites := parsed["sites"].(map[string]interface{})
	entry, ok := sites[testServer.URL]
	require.True(t, ok)

	spec := entry.(map[string]interface{})["spec"].(map[string]interface{})
	assert.Equal(t, testServer.URL, spec["url"].(string))

	state := entry.(map[string]interface{})["state"].(map[string]interface{})
	assert.True(t, state["up"].(bool))
	assert.NotZero(t, state["certificate_age"].(float64))
	assert.NotZero(t, state["latency"].(string))
	assert.NotZero(t, state["last_check"].(string))

	_ = resp.Body.Close()

}
