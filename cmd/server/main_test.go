package main

import (
	"github.com/sotavant/yandex-metrics/internal/server"
	"github.com/sotavant/yandex-metrics/internal/server/handlers"
	"github.com/sotavant/yandex-metrics/internal/server/repository/memory"
	storage2 "github.com/sotavant/yandex-metrics/internal/server/storage"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_badTypeHandler(t *testing.T) {
	type wants struct {
		responseStatus int
	}

	conf := &server.Config{
		Addr:            "",
		StoreInterval:   0,
		FileStoragePath: "/tmp/fs_test",
		Restore:         false,
	}
	fs, _ := storage2.NewFileStorage(*conf)

	storage := memory.NewMetricsRepository()

	appInstanse := &server.App{
		Config:  conf,
		Storage: storage,
		Fs:      fs,
	}

	tests := []struct {
		name    string
		request string
		wants   wants
	}{
		{
			name:    `badType`,
			request: `/update/badType/asdf/sdff`,
			wants: wants{
				responseStatus: http.StatusBadRequest,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, tt.request, nil)
			w := httptest.NewRecorder()
			h := http.HandlerFunc(handlers.UpdateHandler(appInstanse))
			h(w, request)
			result := w.Result()
			defer func() {
				err := result.Body.Close()
				assert.NoError(t, err)
			}()

			res := w.Result()
			defer func() {
				err := res.Body.Close()
				assert.NoError(t, err)
			}()
			assert.Equal(t, tt.wants.responseStatus, res.StatusCode)
		})
	}
}
