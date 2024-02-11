package main

import "testing"
import "github.com/stretchr/testify/assert"

func TestMetricsStorage_updateValues(t *testing.T) {
	s := NewStorage()

	s.Metrics[`Alloc`] = 0

	s.updateValues()

	assert.NotEqual(t, 0, s.Metrics[`Alloc`])
	assert.Equal(t, int64(1), s.PollCount)

	s.updateValues()
	assert.Equal(t, int64(2), s.PollCount)
}
