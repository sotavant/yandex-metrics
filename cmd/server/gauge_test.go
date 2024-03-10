package main

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/sotavant/yandex-metrics/internal"
	"github.com/sotavant/yandex-metrics/internal/server/repository/in_memory"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_handleGauge(t *testing.T) {
	type want struct {
		contentType int
		value       float64
	}

	conf := config{
		addr:            "",
		storeInterval:   0,
		fileStoragePath: "/tmp/fs_test",
		restore:         false,
	}

	tests := []struct {
		name    string
		request string
		storage *in_memory.MetricsRepository
		want    want
		mName   string
		mValue  string
	}{
		{
			name:    `newValue`,
			request: `/update/gauge/newValue/1`,
			mName:   `newValue`,
			mValue:  `1`,
			storage: in_memory.NewMetricsRepository(),
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
			storage: in_memory.NewMetricsRepository(),
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
			storage: in_memory.NewMetricsRepository(),
			want: struct {
				contentType int
				value       float64
			}{contentType: http.StatusBadRequest, value: 0},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs, err := NewFileStorage(conf)
			assert.NoError(t, err)

			appInstance := &app{
				config:     &conf,
				memStorage: tt.storage,
				fs:         fs,
			}

			request := httptest.NewRequest(http.MethodPost, tt.request, nil)
			w := httptest.NewRecorder()

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add(`type`, internal.GaugeType)
			rctx.URLParams.Add(`name`, tt.mName)
			rctx.URLParams.Add(`value`, tt.mValue)

			request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))

			h := http.HandlerFunc(updateHandler(appInstance))
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
