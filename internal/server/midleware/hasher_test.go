package midleware

import (
	"context"
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
	"github.com/sotavant/yandex-metrics/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestRequestHasherMiddleware(t *testing.T) {
	key := "hashKey"
	wrongKey := "wrongHashKey"
	requestBody := `{"id":"ss","type":"gauge","value":3}`
	hash, err := utils.GetHash([]byte(requestBody), key)
	assert.NoError(t, err)
	internal.InitLogger()

	conf := config.Config{
		Addr:            "",
		StoreInterval:   0,
		FileStoragePath: "/tmp/fs_test",
		Restore:         false,
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
		name,
		key string
		wantStatus int
	}{
		{
			"correctHash",
			key,
			http.StatusOK,
		},
		{
			"wrongHash",
			wrongKey,
			http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasherMiddlware := NewHasher(tt.key)

			r := chi.NewRouter()
			r.Use(hasherMiddlware.Handler)
			r.Post("/update/", handlers.UpdateJSONHandler(appInstance))

			w := httptest.NewRecorder()

			reqFunc := func() *http.Request {
				req := httptest.NewRequest(http.MethodPost, "/update/", strings.NewReader(requestBody))
				req.Header.Set(utils.HasherHeaderKey, hash)
				req.Header.Set("Accept", "application/json")
				req.Header.Set("Content-Type", "application/json")

				return req
			}

			r.ServeHTTP(w, reqFunc())
			result := w.Result()
			defer func() {
				err := result.Body.Close()
				assert.NoError(t, err)
			}()

			assert.Equal(t, tt.wantStatus, result.StatusCode)
		})
	}
}

func TestResponseHasherMiddleware(t *testing.T) {
	key := "hashKey"
	requestBody := `{"id":"ss","type":"gauge","value":3}`
	ctx := context.Background()
	hash, err := utils.GetHash([]byte(requestBody), key)
	assert.NoError(t, err)
	internal.InitLogger()

	conf := config.Config{
		Addr:            "",
		StoreInterval:   0,
		FileStoragePath: "/tmp/fs_test",
		Restore:         false,
	}
	st := memory.NewMetricsRepository()
	fs, err := storage.NewFileStorage(conf.FileStoragePath, conf.Restore, conf.StoreInterval)
	assert.NoError(t, err)

	err = st.AddGaugeValue(ctx, "ss", 3)
	assert.NoError(t, err)

	appInstance := &server.App{
		Config:  &conf,
		Storage: st,
		Fs:      fs,
	}

	tests := []struct {
		name,
		wantHash string
		wantStatus int
	}{
		{
			"correctHash",
			hash,
			http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasherMiddleware := NewHasher(key)

			r := chi.NewRouter()
			r.Use(hasherMiddleware.Handler)
			r.Post("/value/", handlers.GetValueJSONHandler(appInstance))

			w := httptest.NewRecorder()

			reqFunc := func() *http.Request {
				req := httptest.NewRequest(http.MethodPost, "/value/", strings.NewReader(requestBody))
				req.Header.Set(utils.HasherHeaderKey, hash)
				req.Header.Set("Accept", "application/json")
				req.Header.Set("Content-Type", "application/json")

				return req
			}

			r.ServeHTTP(w, reqFunc())
			result := w.Result()
			defer func() {
				err := result.Body.Close()
				assert.NoError(t, err)
			}()

			assert.Equal(t, hash, result.Header.Get(utils.HasherHeaderKey))
			assert.Equal(t, tt.wantStatus, result.StatusCode)
		})
	}
}
