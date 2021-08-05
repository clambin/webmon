package webmon_test

import (
	"github.com/clambin/webmon/webmon"
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

	assert.Equal(t, []string{"aaa", "bbb", "ccc"}, webmon.Unique(input))
}
