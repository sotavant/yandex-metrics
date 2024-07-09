package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetLocalIP(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "get local ip",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GetLocalIP()
			assert.NoError(t, err)
		})
	}
}
