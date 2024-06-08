package storage

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetricsStorage_updateValues(t *testing.T) {
	s := NewStorage()

	s.Metrics[`Alloc`] = 0

	s.UpdateValues()

	assert.NotEqual(t, 0, s.Metrics[`Alloc`])
	assert.Equal(t, int64(1), s.PollCount)

	s.UpdateValues()
	assert.Equal(t, int64(2), s.PollCount)
}

func BenchmarkMetricsStorage_UpdateValues(b *testing.B) {
	s := NewStorage()
	for n := 0; n < b.N; n++ {
		s.UpdateValues()
	}
}

func BenchmarkMetricsStorage_UpdateAdditionalValues(b *testing.B) {
	s := NewStorage()
	for n := 0; n < b.N; n++ {
		s.UpdateAdditionalValues()
	}
}

func ExampleMetricsStorage_UpdateValues() {
	s := NewStorage()
	s.UpdateValues()

	fmt.Println(s.PollCount)
	// Output:
	// 1
}
