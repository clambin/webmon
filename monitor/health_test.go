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
	"time"
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
				assert.True(t, entry.(map[string]interface{})["Up"].(bool))
				assert.NotZero(t, entry.(map[string]interface{})["CertificateAge"].(float64))
				assert.NotZero(t, entry.(map[string]interface{})["Latency"].(string))

				var lastCheck time.Time
				lastCheck, err = time.Parse(time.RFC3339, entry.(map[string]interface{})["LastCheck"].(string))
				assert.NoError(t, err)
				assert.NotZero(t, lastCheck)
			}
		}
		_ = resp.Body.Close()
	}
}
