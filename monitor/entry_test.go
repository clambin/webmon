package monitor_test

import (
	"encoding/json"
	"github.com/clambin/webmon/monitor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestSiteState_Marshal(t *testing.T) {
	state := monitor.SiteState{
		Up:             true,
		CertificateAge: 52.0,
		Latency:        monitor.Duration{Duration: 1250 * time.Millisecond},
		LastCheck:      time.Date(2021, 8, 9, 15, 30, 0, 0, time.UTC),
	}

	out, err := json.Marshal(&state)
	require.NoError(t, err)
	assert.Equal(t, `{"up":true,"certificate_age":52,"latency":"1.25s","last_check":"2021-08-09T15:30:00Z"}`, string(out))
}

func TestSiteState_Unmarshal(t *testing.T) {
	input := []byte(`{
	"up": true,
    "certificate_age": 7.0,
	"latency": "250ms",
    "last_check": "2021-08-09T15:30:00Z"
}`)

	var output monitor.SiteState
	err := json.Unmarshal(input, &output)
	require.NoError(t, err)
	assert.True(t, output.Up)
	assert.Equal(t, 7.0, output.CertificateAge)
	assert.Equal(t, 0.25, output.Latency.Seconds())
	assert.Equal(t, time.Date(2021, 8, 9, 15, 30, 0, 0, time.UTC), output.LastCheck)
}
