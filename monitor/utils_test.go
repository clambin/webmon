package monitor_test

import (
	"github.com/clambin/webmon/monitor"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUnique(t *testing.T) {
	input := []string{
		"aaa",
		"bbb",
		"bbb",
		"ccc",
	}

	assert.Equal(t, []string{"aaa", "bbb", "ccc"}, monitor.Unique(input))
}
