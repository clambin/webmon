package monitor_test

import (
	"context"
	"encoding/json"
	"github.com/clambin/webmon/monitor"
	"github.com/stretchr/testify/assert"
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

	if assert.Equal(t, http.StatusOK, resp.StatusCode) {
		body, _ := io.ReadAll(resp.Body)

		var parsed map[string]interface{}
		err := json.Unmarshal(body, &parsed)

		if assert.NoError(t, err) {
			entry, ok := parsed[testServer.URL]
			if assert.True(t, ok) {
				assert.True(t, entry.(map[string]interface{})["up"].(bool))
				assert.NotZero(t, entry.(map[string]interface{})["certificate_age"].(float64))
				assert.NotZero(t, entry.(map[string]interface{})["latency"].(float64))
				assert.NotEqual(t, "0001-01-01T00:00:00Z", entry.(map[string]interface{})["last_check"].(string))
			}
		}
		_ = resp.Body.Close()
	}
}
