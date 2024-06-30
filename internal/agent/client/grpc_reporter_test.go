package client

import (
	"bytes"
	"encoding/gob"
	"testing"

	"github.com/sotavant/yandex-metrics/internal"
	"github.com/sotavant/yandex-metrics/internal/agent/config"
	"github.com/sotavant/yandex-metrics/internal/utils"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
)

func TestGRPCReporter_addHashMetadata(t *testing.T) {
	var val float64 = 1
	var delta int64 = 1

	internal.InitLogger()

	m := internal.Metrics{
		Value: &val,
		Delta: &delta,
		ID:    "ddd",
		MType: "gauge",
	}

	tests := []struct {
		name      string
		key       string
		wantCount int
	}{
		{
			name:      "with key",
			key:       "ddd",
			wantCount: 1,
		},
		{
			name:      "without key",
			key:       "",
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config.AppConfig = &config.Config{
				HashKey: tt.key,
			}

			r := NewGRPCReporter(nil)

			md := metadata.Pairs()
			md = r.addHashMetadata(m, md)

			assert.Len(t, md.Get(utils.HasherHeaderKey), tt.wantCount)

			if tt.wantCount == 0 {
				return
			}

			var encodedMD bytes.Buffer
			enc := gob.NewEncoder(&encodedMD)
			err := enc.Encode(m)
			assert.NoError(t, err)

			hash, err := utils.GetHash(encodedMD.Bytes(), tt.key)
			assert.NoError(t, err)

			assert.Equal(t, hash, md.Get(utils.HasherHeaderKey)[0])
		})
	}
}
