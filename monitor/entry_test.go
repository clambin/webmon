package monitor_test

import (
	"encoding/json"
	"github.com/clambin/webmon/monitor"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestEntry_Marshal(t *testing.T) {
	entry := monitor.Entry{
		Up:             true,
		CertificateAge: 52.0,
		Latency:        monitor.Duration{Duration: 1250 * time.Millisecond},
		LastCheck:      time.Date(2021, 8, 9, 15, 30, 0, 0, time.UTC),
	}

	out, err := json.Marshal(&entry)

	assert.NoError(t, err)
	assert.Equal(t, `{"Up":true,"CertificateAge":52,"Latency":"1.25s","LastCheck":"2021-08-09T15:30:00Z"}`, string(out))
}

func TestEntry_Unmarshal(t *testing.T) {
	input := []byte(`{
	"Up": true,
    "CertificateAge": 7.0,
	"Latency": "250ms",
    "LastCheck": "2021-08-09T15:30:00Z"
}`)

	var output monitor.Entry
	err := json.Unmarshal(input, &output)

	assert.NoError(t, err)
	assert.True(t, output.Up)
	assert.Equal(t, 7.0, output.CertificateAge)
	assert.Equal(t, 0.25, output.Latency.Seconds())
	assert.Equal(t, time.Date(2021, 8, 9, 15, 30, 0, 0, time.UTC), output.LastCheck)
}