package handlers

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"github.com/jackc/pgx/v5"
	"github.com/sotavant/yandex-metrics/internal/server"
	"github.com/sotavant/yandex-metrics/internal/server/config"
	"github.com/sotavant/yandex-metrics/internal/server/midleware"
	"github.com/sotavant/yandex-metrics/internal/server/repository/memory"
	"github.com/sotavant/yandex-metrics/internal/server/repository/postgres"
	"github.com/sotavant/yandex-metrics/internal/server/repository/postgres/test"
	"github.com/sotavant/yandex-metrics/internal/server/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func Test_updateJsonHandler(t *testing.T) {
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

	handler := UpdateJSONHandler(appInstance)

	type want struct {
		status int
		body   string
	}

	tests := []struct {
		name string
		body string
		want want
	}{
		{
			name: `newGaugeValue`,
			body: `{"id": "ss","type":"gauge","value":-33.345345}`,
			want: struct {
				status int
				body   string
			}{status: 200, body: `{"id":"ss","type":"gauge","value":-33.345345}
`},
		},
		{
			name: `newCounterValue`,
			body: `{"id": "ss","type":"counter","delta":3}`,
			want: struct {
				status int
				body   string
			}{status: 200, body: `{"id":"ss","type":"counter","delta":3}
`},
		},
		{
			name: `repeatCounterValue`,
			body: `{"id": "ss","type":"counter","delta":3}`,
			want: struct {
				status int
				body   string
			}{status: 200, body: `{"id":"ss","type":"counter","delta":6}
`},
		},
		{
			name: `badJson`,
			body: `{id": "ss","type":"counter","delta":3}`,
			want: struct {
				status int
				body   string
			}{status: http.StatusBadRequest, body: `internal server error`},
		},
		{
			name: `badType`,
			body: `{"id": "ss","type":"counterBad","delta":3}`,
			want: struct {
				status int
				body   string
			}{status: http.StatusBadRequest, body: `internal server error`},
		},
		{
			name: `emptyValue`,
			body: `{"id": "ss","type":"counter","value":3}`,
			want: struct {
				status int
				body   string
			}{status: http.StatusBadRequest, body: `internal server error`},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/update", strings.NewReader(tt.body))
			w := httptest.NewRecorder()

			handler(w, request)
			result := w.Result()
			defer func() {
				err := result.Body.Close()
				assert.NoError(t, err)
			}()

			body, err := io.ReadAll(result.Body)

			assert.NoError(t, err)

			assert.Equal(t, tt.want.status, result.StatusCode)
			if result.StatusCode == http.StatusOK {
				assert.Equal(t, tt.want.body, string(body))
			}
		})
	}
}

func Test_getValueJsonHandler(t *testing.T) {
	appInstance := &server.App{
		Config:  nil,
		Storage: memory.NewMetricsRepository(),
		Fs:      nil,
	}

	err := appInstance.Storage.AddGaugeValue(context.Background(), "ss", -3444)
	assert.NoError(t, err)
	err = appInstance.Storage.AddCounterValue(context.Background(), "ss", 3)
	assert.NoError(t, err)
	handler := GetValueJSONHandler(appInstance)

	type want struct {
		status int
		body   string
	}

	tests := []struct {
		name string
		body string
		want want
	}{
		{
			name: `getGaugeValue`,
			body: `{"id": "ss","type":"gauge","value":-33.345345}`,
			want: struct {
				status int
				body   string
			}{status: 200, body: `{"id":"ss","type":"gauge","value":-3444}
`},
		},
		{
			name: `getCounterValue`,
			body: `{"id": "ss","type":"counter","delta":6}`,
			want: struct {
				status int
				body   string
			}{status: 200, body: `{"id":"ss","type":"counter","delta":3}
`},
		},
		{
			name: `badJson`,
			body: `{id": "ss","type":"counter","delta":3}`,
			want: struct {
				status int
				body   string
			}{status: http.StatusBadRequest, body: `internal server error`},
		},
		{
			name: `badType`,
			body: `{"id": "ss","type":"counterBad","delta":3}`,
			want: struct {
				status int
				body   string
			}{status: http.StatusBadRequest, body: `internal server error`},
		},
		{
			name: `emptyValue`,
			body: `{"id": "ssww","type":"counter","value":3}`,
			want: struct {
				status int
				body   string
			}{status: http.StatusNotFound, body: `internal server error`},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/value", strings.NewReader(tt.body))
			w := httptest.NewRecorder()

			handler(w, request)
			result := w.Result()
			defer func() {
				err := result.Body.Close()
				assert.NoError(t, err)
			}()

			body, err := io.ReadAll(result.Body)

			assert.NoError(t, err)

			assert.Equal(t, tt.want.status, result.StatusCode)
			if result.StatusCode == http.StatusOK {
				assert.Equal(t, tt.want.body, string(body))
			}
		})
	}
}

func TestGzipCompression(t *testing.T) {
	appInstance := &server.App{
		Config:  nil,
		Storage: memory.NewMetricsRepository(),
		Fs:      nil,
	}
	err := appInstance.Storage.AddGaugeValue(context.Background(), "ss", -3444)
	assert.NoError(t, err)
	err = appInstance.Storage.AddCounterValue(context.Background(), "ss", 3)
	assert.NoError(t, err)
	requestBody := `{"id":"ss","type":"counter","delta":3}
`
	htmlResponse := `<p>ss: -3444</p>`

	t.Run("sends_gzip", func(t *testing.T) {
		handler := midleware.GzipMiddleware(GetValueJSONHandler(appInstance))

		buf := bytes.NewBuffer(nil)
		zb := gzip.NewWriter(buf)
		_, err := zb.Write([]byte(requestBody))
		assert.NoError(t, err)
		err = zb.Close()
		assert.NoError(t, err)

		r := httptest.NewRequest(http.MethodPost, "/value", buf)
		r.RequestURI = ""
		r.Header.Set("Content-Encoding", "gzip")
		r.Header.Set("Accept-Encoding", "0")
		r.Header.Set("Accept", "application/json")
		r.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()

		handler(w, r)
		result := w.Result()
		defer func() {
			err := result.Body.Close()
			assert.NoError(t, err)
		}()

		body, err := io.ReadAll(result.Body)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, result.StatusCode)

		assert.Equal(t, requestBody, string(body))

	})

	t.Run("accept_gzip", func(t *testing.T) {
		handler := midleware.GzipMiddleware(GetValuesHandler(appInstance))

		buf := bytes.NewBufferString(requestBody)
		r := httptest.NewRequest("POST", "/", buf)
		r.RequestURI = ""
		r.Header.Set("Accept-Encoding", "gzip")
		r.Header.Set("Accept", "html/text")

		w := httptest.NewRecorder()

		handler(w, r)
		result := w.Result()
		defer func() {
			err := result.Body.Close()
			assert.NoError(t, err)
		}()

		require.Equal(t, http.StatusOK, result.StatusCode)

		zr, err := gzip.NewReader(result.Body)
		require.NoError(t, err)

		b, err := io.ReadAll(zr)
		require.NoError(t, err)

		require.Equal(t, htmlResponse, string(b))
	})
}

func Test_updateBatchJSONHandler(t *testing.T) {
	conf := config.Config{
		Addr:            "",
		StoreInterval:   3,
		FileStoragePath: "/tmp/fs_test",
		Restore:         false,
	}
	st := memory.NewMetricsRepository()
	fs, err := storage.NewFileStorage(conf.FileStoragePath, conf.Restore, conf.StoreInterval)
	assert.NoError(t, err)

	ctx := context.Background()
	conn, tableName, DNS, err := test.InitConnection(ctx, t)
	assert.NoError(t, err)

	appInstance := &server.App{
		Config:  &conf,
		Storage: st,
		Fs:      fs,
	}

	if conn != nil {
		defer func(ctx context.Context, conn pgx.Conn, tableName string) {
			err = test.DropTable(ctx, conn, tableName)
			assert.NoError(t, err)

			err = conn.Close(ctx)
			assert.NoError(t, err)
		}(ctx, *conn, tableName)
	}

	type want struct {
		status int
		body   string
	}

	tests := []struct {
		name     string
		body     string
		want     want
		inMemory bool
	}{
		{
			name: `newGaugeValue`,
			body: `[{"id": "ss","type":"gauge","value":-33.345345}]`,
			want: struct {
				status int
				body   string
			}{status: 200, body: `[{"id":"ss","type":"gauge","value":-33.345345}]`},
			inMemory: true,
		},
		{
			name: `newCounterValue`,
			body: `[{"id": "ss","type":"counter","delta":3}]`,
			want: struct {
				status int
				body   string
			}{status: 200, body: `[{"id":"ss","type":"counter","delta":3}, {"id":"ss","type":"gauge","value":-33.345345}]`},
			inMemory: true,
		},
		{
			name: `repeatCounterValue`,
			body: `[{"id": "ss","type":"counter","delta":3}]`,
			want: struct {
				status int
				body   string
			}{status: 200, body: `[{"id":"ss","type":"counter","delta":6}, {"id":"ss","type":"gauge","value":-33.345345}]`},
			inMemory: true,
		},
		{
			name: `newGaugeValueBD`,
			body: `[{"id": "ss","type":"gauge","value":-33.345345}]`,
			want: struct {
				status int
				body   string
			}{status: 200, body: `[{"id":"ss","type":"gauge","value":-33.345345}]`},
			inMemory: false,
		},
		{
			name: `newGaugeValuesBD`,
			body: `[{"id": "ss","type":"gauge","value":-33.345345},{"id": "pp","type":"gauge","value":-33.345345}]`,
			want: struct {
				status int
				body   string
			}{status: 200, body: `[{"id": "ss","type":"gauge","value":-33.345345},{"id": "pp","type":"gauge","value":-33.345345}]`},
			inMemory: false,
		},
		{
			name: `addVariousValuesBD`,
			body: `[{"id": "ss","type":"gauge","value":-33.345345},{"id": "pp","type":"gauge","value":-33.345345}, {"id": "ss","type":"counter","delta":3}]`,
			want: struct {
				status int
				body   string
			}{status: 200, body: `[{"id": "ss","type":"gauge","value":-33.345345},{"id": "pp","type":"gauge","value":-33.345345},{"id": "ss","type":"counter","delta":3}]`},
			inMemory: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.inMemory && conn == nil {
				return
			}

			if !tt.inMemory {
				appInstance.Storage, err = postgres.NewMemStorage(ctx, conn, tableName, DNS)
				assert.NoError(t, err)
			}

			handler := UpdateBatchJSONHandler(appInstance)

			request := httptest.NewRequest(http.MethodPost, "/updates/", strings.NewReader(tt.body))
			w := httptest.NewRecorder()

			handler(w, request)
			result := w.Result()
			defer func() {
				err := result.Body.Close()
				assert.NoError(t, err)
			}()

			body, err := io.ReadAll(result.Body)

			assert.NoError(t, err)

			assert.Equal(t, tt.want.status, result.StatusCode)
			if result.StatusCode == http.StatusOK {
				var expected, actual []interface{}
				assert.NoError(t, json.Unmarshal([]byte(tt.want.body), &expected))
				assert.NoError(t, json.Unmarshal(body, &actual))
				assert.ElementsMatch(t, expected, actual)
			}
		})
	}
}
