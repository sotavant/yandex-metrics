package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/sotavant/yandex-metrics/internal"
	"github.com/sotavant/yandex-metrics/internal/server"
	"github.com/sotavant/yandex-metrics/internal/server/config"
	"github.com/sotavant/yandex-metrics/internal/server/repository/memory"
	"github.com/sotavant/yandex-metrics/internal/server/storage"
	"github.com/stretchr/testify/assert"
)

func Test_handleGauge(t *testing.T) {
	type want struct {
		contentType int
		value       float64
	}

	conf := config.Config{
		Addr:            "",
		StoreInterval:   0,
		FileStoragePath: "/tmp/fs_test",
		Restore:         false,
	}

	tests := []struct {
		storage *memory.MetricsRepository
		name    string
		request string
		mValue  string
		mName   string
		want    want
	}{
		{
			name:    `newValue`,
			request: `/update/gauge/newValue/1`,
			mName:   `newValue`,
			mValue:  `1`,
			storage: memory.NewMetricsRepository(),
			want: struct {
				contentType int
				value       float64
			}{contentType: http.StatusOK, value: 1},
		},
		{
			name:    `updateValue`,
			request: `/update/gauge/updateValue/3`,
			mName:   `updateValue`,
			mValue:  `3`,
			storage: memory.NewMetricsRepository(),
			want: struct {
				contentType int
				value       float64
			}{contentType: http.StatusOK, value: 3},
		},
		{
			name:    `badValue`,
			request: `/update/gauge/badValue/sdfsdfsdf`,
			mName:   `badValue`,
			mValue:  `sdfsdfsdf`,
			storage: memory.NewMetricsRepository(),
			want: struct {
				contentType int
				value       float64
			}{contentType: http.StatusBadRequest, value: 0},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs, err := storage.NewFileStorage(conf.FileStoragePath, conf.Restore, conf.StoreInterval)
			assert.NoError(t, err)

			appInstance := &server.App{
				Config:  &conf,
				Storage: tt.storage,
				Fs:      fs,
			}

			request := httptest.NewRequest(http.MethodPost, tt.request, nil)
			w := httptest.NewRecorder()

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add(`type`, internal.GaugeType)
			rctx.URLParams.Add(`name`, tt.mName)
			rctx.URLParams.Add(`value`, tt.mValue)

			request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))

			h := http.HandlerFunc(UpdateHandler(appInstance))
			h(w, request)
			result := w.Result()
			defer func() {
				err := result.Body.Close()
				assert.NoError(t, err)
			}()

			switch tt.name {
			case `updateValue`:
				h(w, request)
				assert.Equal(t, tt.want.contentType, result.StatusCode)
				assert.Equal(t, tt.want.value, tt.storage.Gauge[tt.name])
			default:
				assert.Equal(t, tt.want.contentType, result.StatusCode)
				assert.Equal(t, tt.want.value, tt.storage.Gauge[tt.name])
			}
		})
	}
}
