package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
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

func TestCrypto_Handler(t *testing.T) {
	var cryptoMiddlware *Crypto
	internal.InitLogger()

	privateKeyPath := os.Getenv("TEST_CRYPT_PRIV_KEY")
	publicKeyPath := os.Getenv("TEST_CRYPT_PUB_KEY")

	if privateKeyPath == "" || publicKeyPath == "" {
		t.Skip("no test credentials")
	}

	requestBody := `{"value":3,"id":"ss","type":"gauge"}`

	conf := config.Config{
		Addr:            "",
		StoreInterval:   0,
		FileStoragePath: "/tmp/fs_test",
		Restore:         false,
		CryptoKeyPath:   privateKeyPath,
	}

	st := memory.NewMetricsRepository()
	fs, err := storage.NewFileStorage(conf.FileStoragePath, conf.Restore, conf.StoreInterval)
	assert.NoError(t, err)

	ch, err := utils.NewCipher(privateKeyPath, publicKeyPath)
	assert.NoError(t, err)

	encryptedText, err := ch.Encrypt([]byte(requestBody))
	assert.NoError(t, err)

	appInstance := &server.App{
		Config:  &conf,
		Storage: st,
		Fs:      fs,
	}

	tests := []struct {
		name       string
		body       []byte
		wantStatus int
	}{
		{
			"correctAnswer",
			encryptedText,
			http.StatusOK,
		},
		{
			"wrongHash",
			[]byte(requestBody),
			http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cryptoMiddlware, err = NewCrypto(privateKeyPath)
			assert.NoError(t, err)

			r := chi.NewRouter()
			r.Use(cryptoMiddlware.Handler)
			r.Post("/update/", handlers.UpdateJSONHandler(appInstance))

			w := httptest.NewRecorder()

			reqFunc := func() *http.Request {
				req := httptest.NewRequest(http.MethodPost, "/update/", strings.NewReader(string(tt.body)))
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
