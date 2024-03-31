package storage

import (
	"context"
	"github.com/sotavant/yandex-metrics/internal"
	"github.com/sotavant/yandex-metrics/internal/server/config"
	"github.com/sotavant/yandex-metrics/internal/server/repository/memory"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"testing"
)

func TestFileStorage_Restore(t *testing.T) {
	conf := config.Config{
		Addr:            "",
		StoreInterval:   0,
		FileStoragePath: "/tmp/fs_test",
		Restore:         false,
	}

	tests := []struct {
		name        string
		data        []string
		wantData    []internal.Metrics
		needRestore bool
	}{
		{
			name:        "oneData",
			data:        []string{`{"id":"s","value":111,"type":"gauge"}`},
			needRestore: true,
			wantData: []internal.Metrics{
				{
					ID:    "s",
					Value: getFloat64Pointer(111),
				},
			},
		},
		{
			name:        "twoData",
			data:        []string{`{"id":"s","value":111,"type":"gauge"}{"id":"p","value":13,"type":"gauge"}`},
			needRestore: true,
			wantData: []internal.Metrics{
				{
					ID:    "s",
					Value: getFloat64Pointer(111),
				},
				{
					ID:    "p",
					Value: getFloat64Pointer(13),
				},
			},
		},
		{
			name:        "dataWithNewLine",
			data:        []string{`{"id":"s","value":111,"type":"gauge"}`, `{"id":"p","value":13,"type":"gauge"}`},
			needRestore: true,
			wantData: []internal.Metrics{
				{
					ID:    "s",
					Value: getFloat64Pointer(111),
				},
				{
					ID:    "p",
					Value: getFloat64Pointer(13),
				},
			},
		},
		{
			name:        "noNeedRestore",
			data:        []string{`{"id":"s","value":111,"type":"gauge"}`, `{"id":"p","value":13,"type":"gauge"}`},
			needRestore: false,
			wantData: []internal.Metrics{
				{
					ID:    "s",
					Value: getFloat64Pointer(0),
				},
				{
					ID:    "p",
					Value: getFloat64Pointer(0),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs, _ := NewFileStorage(conf.FileStoragePath, conf.Restore, conf.StoreInterval)
			fs.needRestore = tt.needRestore
			ms := memory.NewMetricsRepository()
			ctx := context.Background()

			defer func(file *os.File) {
				err := file.Close()
				assert.NoError(t, err)

				err = os.Remove(conf.FileStoragePath)
				assert.NoError(t, err)
			}(fs.File)

			for _, v := range tt.data {
				_, err := fs.File.WriteString(v)
				assert.NoError(t, err)
			}

			err := fs.File.Sync()
			assert.NoError(t, err)

			err = fs.Restore(ctx, ms)
			assert.NoError(t, err)
			for _, v := range tt.wantData {
				val, err := ms.GetGaugeValue(ctx, v.ID)
				assert.NoError(t, err)
				assert.Equal(t, *v.Value, val)
			}
		})
	}
}

func getFloat64Pointer(num float64) *float64 {
	return &num
}

func TestFileStorage_Sync(t *testing.T) {
	conf := config.Config{
		Addr:            "",
		StoreInterval:   0,
		FileStoragePath: "/tmp/fs_test",
		Restore:         false,
	}

	ctx := context.Background()

	t.Run("checkWriteToFile", func(t *testing.T) {
		fs, _ := NewFileStorage(conf.FileStoragePath, conf.Restore, conf.StoreInterval)
		ms := memory.MetricsRepository{
			Gauge: map[string]float64{
				"s": 111,
			},
			Counter: map[string]int64{
				"c": 13,
			},
		}
		want := []string{`{"id":"s","type":"gauge","value":111}`, `{"id":"c","type":"counter","delta":13}`}

		defer func(file *os.File) {
			err := file.Close()
			assert.NoError(t, err)

			err = os.Remove(conf.FileStoragePath)
			assert.NoError(t, err)
		}(fs.File)

		err := fs.Sync(ctx, &ms)
		assert.NoError(t, err)

		_, err = fs.File.Seek(0, io.SeekStart)
		assert.NoError(t, err)

		data, err := os.ReadFile(fs.File.Name())
		assert.NoError(t, err)
		for _, str := range want {
			assert.Contains(t, string(data), str)
		}
	})
}
