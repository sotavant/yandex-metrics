package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/sotavant/yandex-metrics/internal"
	"github.com/sotavant/yandex-metrics/internal/server"
	"github.com/sotavant/yandex-metrics/internal/server/config"
	"github.com/sotavant/yandex-metrics/internal/server/handlers"
	"github.com/sotavant/yandex-metrics/internal/server/repository/memory"
	"github.com/sotavant/yandex-metrics/internal/server/storage"
	"github.com/stretchr/testify/assert"
)

func TestIPChecker_CheckIP_Handler(t *testing.T) {
	internal.InitLogger()

	requestBody := `{"value":3,"id":"ss","type":"gauge"}`

	conf := config.Config{
		Addr:            "",
		StoreInterval:   0,
		FileStoragePath: "/tmp/fs_test",
		Restore:         false,
		TrustedSubnet:   "192.168.1.0/24",
	}

	st := memory.NewMetricsRepository()
	fs, err := storage.NewFileStorage(conf.FileStoragePath, conf.Restore, conf.StoreInterval)
	assert.NoError(t, err)

	appInstance := &server.App{
		Config:  &conf,
		Storage: st,
		Fs:      fs,
	}

	tests := []struct {
		name       string
		ip         string
		wantStatus int
	}{
		{
			"good IP",
			"192.168.1.130",
			http.StatusOK,
		},
		{
			"bad IP",
			"132.132.132.132",
			http.StatusForbidden,
		},
		{
			"empty IP",
			"",
			http.StatusForbidden,
		},
		{
			"not correct IP",
			"sdfsdf.sdfsd.sdfsdf",
			http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ipMiddleware := NewIPChecker(conf.TrustedSubnet)
			assert.NotNil(t, ipMiddleware)

			r := chi.NewRouter()
			r.Use(ipMiddleware.CheckIP)
			r.Post("/update/", handlers.UpdateJSONHandler(appInstance))

			w := httptest.NewRecorder()

			reqFunc := func() *http.Request {
				req := httptest.NewRequest(http.MethodPost, "/update/", strings.NewReader(requestBody))
				req.Header.Set("Accept", "application/json")
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("X-Real-IP", tt.ip)

				return req
			}

			r.ServeHTTP(w, reqFunc())
			result := w.Result()
			defer func() {
				err = result.Body.Close()
				assert.NoError(t, err)
			}()

			assert.Equal(t, tt.wantStatus, result.StatusCode)
		})
	}
}
