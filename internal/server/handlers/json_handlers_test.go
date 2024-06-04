package handlers

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sotavant/yandex-metrics/internal/server"
	"github.com/sotavant/yandex-metrics/internal/server/config"
	"github.com/sotavant/yandex-metrics/internal/server/middleware"
	"github.com/sotavant/yandex-metrics/internal/server/repository/memory"
	"github.com/sotavant/yandex-metrics/internal/server/repository/postgres"
	"github.com/sotavant/yandex-metrics/internal/server/repository/postgres/test"
	"github.com/sotavant/yandex-metrics/internal/server/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		body   string
		status int
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
				body   string
				status int
			}{status: 200, body: `{"value":-33.345345,"id":"ss","type":"gauge"}
`},
		},
		{
			name: `newCounterValue`,
			body: `{"id": "ss","type":"counter","delta":3}`,
			want: struct {
				body   string
				status int
			}{status: 200, body: `{"delta":3,"id":"ss","type":"counter"}
`},
		},
		{
			name: `repeatCounterValue`,
			body: `{"id": "ss","type":"counter","delta":3}`,
			want: struct {
				body   string
				status int
			}{status: 200, body: `{"delta":6,"id":"ss","type":"counter"}
`},
		},
		{
			name: `badJson`,
			body: `{id": "ss","type":"counter","delta":3}`,
			want: struct {
				body   string
				status int
			}{status: http.StatusBadRequest, body: `internal server error`},
		},
		{
			name: `badType`,
			body: `{"id": "ss","type":"counterBad","delta":3}`,
			want: struct {
				body   string
				status int
			}{status: http.StatusBadRequest, body: `internal server error`},
		},
		{
			name: `emptyValue`,
			body: `{"id": "ss","type":"counter","value":3}`,
			want: struct {
				body   string
				status int
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
		body   string
		status int
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
				body   string
				status int
			}{status: 200, body: `{"value":-3444,"id":"ss","type":"gauge"}
`},
		},
		{
			name: `getCounterValue`,
			body: `{"delta":6,"id": "ss","type":"counter"}`,
			want: struct {
				body   string
				status int
			}{status: 200, body: `{"delta":3,"id":"ss","type":"counter"}
`},
		},
		{
			name: `badJson`,
			body: `{id": "ss","type":"counter","delta":3}`,
			want: struct {
				body   string
				status int
			}{status: http.StatusBadRequest, body: `internal server error`},
		},
		{
			name: `badType`,
			body: `{"delta":3,"id": "ss","type":"counterBad"}`,
			want: struct {
				body   string
				status int
			}{status: http.StatusBadRequest, body: `internal server error`},
		},
		{
			name: `emptyValue`,
			body: `{"value":3,"id": "ssww","type":"counter"}`,
			want: struct {
				body   string
				status int
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
	requestBody := `{"delta":3,"id":"ss","type":"counter"}
`
	htmlResponse := `<p>ss: -3444</p>`

	t.Run("sends_gzip", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)
		zb := gzip.NewWriter(buf)
		_, err := zb.Write([]byte(requestBody))
		assert.NoError(t, err)
		err = zb.Close()
		assert.NoError(t, err)

		r := chi.NewRouter()
		r.Use(middleware.GzipMiddleware)
		r.Post("/value", GetValueJSONHandler(appInstance))

		w := httptest.NewRecorder()

		reqFunc := func() *http.Request {
			req := httptest.NewRequest(http.MethodPost, "/value", buf)
			req.Header.Set("Content-Encoding", "gzip")
			req.Header.Set("Accept-Encoding", "0")
			req.Header.Set("Accept", "application/json")
			req.Header.Set("Content-Type", "application/json")

			return req
		}

		r.ServeHTTP(w, reqFunc())
		//handler(w, r)
		result := w.Result()
		defer func() {
			err = result.Body.Close()
			assert.NoError(t, err)
		}()

		body, err := io.ReadAll(result.Body)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, result.StatusCode)

		assert.Equal(t, requestBody, string(body))

	})

	t.Run("accept_gzip", func(t *testing.T) {
		buf := bytes.NewBufferString(requestBody)
		w := httptest.NewRecorder()
		reqFunc := func() *http.Request {
			req := httptest.NewRequest(http.MethodPost, "/", buf)
			req.Header.Set("Accept-Encoding", "gzip")
			req.Header.Set("Accept", "html/text")

			return req
		}

		r := chi.NewRouter()
		r.Use(middleware.GzipMiddleware)
		r.Post("/", GetValuesHandler(appInstance))

		r.ServeHTTP(w, reqFunc())

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
		defer func(ctx context.Context, conn *pgxpool.Pool, tableName string) {
			err = test.DropTable(ctx, conn, tableName)
			assert.NoError(t, err)

			conn.Close()
			assert.NoError(t, err)
		}(ctx, conn, tableName)
	}

	type want struct {
		body   string
		status int
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
				body   string
				status int
			}{status: 200, body: `[{"id":"ss","type":"gauge","value":-33.345345}]`},
			inMemory: true,
		},
		{
			name: `newCounterValue`,
			body: `[{"id": "ss","type":"counter","delta":3}]`,
			want: struct {
				body   string
				status int
			}{status: 200, body: `[{"id":"ss","type":"counter","delta":3}, {"id":"ss","type":"gauge","value":-33.345345}]`},
			inMemory: true,
		},
		{
			name: `repeatCounterValue`,
			body: `[{"id": "ss","type":"counter","delta":3}]`,
			want: struct {
				body   string
				status int
			}{status: 200, body: `[{"id":"ss","type":"counter","delta":6}, {"id":"ss","type":"gauge","value":-33.345345}]`},
			inMemory: true,
		},
		{
			name: `newGaugeValueBD`,
			body: `[{"id": "ss","type":"gauge","value":-33.345345}]`,
			want: struct {
				body   string
				status int
			}{status: 200, body: `[{"id":"ss","type":"gauge","value":-33.345345}]`},
			inMemory: false,
		},
		{
			name: `newGaugeValuesBD`,
			body: `[{"id": "ss","type":"gauge","value":-33.345345},{"id": "pp","type":"gauge","value":-33.345345}]`,
			want: struct {
				body   string
				status int
			}{status: 200, body: `[{"id": "ss","type":"gauge","value":-33.345345},{"id": "pp","type":"gauge","value":-33.345345}]`},
			inMemory: false,
		},
		{
			name: `addVariousValuesBD`,
			body: `[{"id": "ss","type":"gauge","value":-33.345345},{"id": "pp","type":"gauge","value":-33.345345}, {"id": "ss","type":"counter","delta":3}]`,
			want: struct {
				body   string
				status int
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

func ExampleUpdateJSONHandler() {
	conf := config.Config{
		Addr:          "",
		StoreInterval: 0,
	}
	st := memory.NewMetricsRepository()

	appInstance := &server.App{
		Config:  &conf,
		Storage: st,
	}

	handler := UpdateJSONHandler(appInstance)

	request := httptest.NewRequest(http.MethodPost, "/update", strings.NewReader(`{"id":"ss","type":"gauge","value":-33.345345}`))
	w := httptest.NewRecorder()

	handler(w, request)
	result := w.Result()
	defer func() {
		_ = result.Body.Close()
	}()

	body, _ := io.ReadAll(result.Body)

	fmt.Println(result.StatusCode)
	fmt.Println(string(body))

	// Output:
	// 200
	// {"value":-33.345345,"id":"ss","type":"gauge"}
}

func ExampleUpdateBatchJSONHandler() {
	conf := config.Config{
		Addr:          "",
		StoreInterval: 0,
	}
	st := memory.NewMetricsRepository()

	appInstance := &server.App{
		Config:  &conf,
		Storage: st,
	}

	handler := UpdateBatchJSONHandler(appInstance)

	request := httptest.NewRequest(http.MethodPost, "/updates/", strings.NewReader(`[{"id":"a","type":"gauge","value":1},{"id":"b","type":"counter","delta":2}]`))
	w := httptest.NewRecorder()

	handler(w, request)
	result := w.Result()
	defer func() {
		_ = result.Body.Close()
	}()

	body, _ := io.ReadAll(result.Body)

	fmt.Println(result.StatusCode)
	fmt.Println(string(body))

	// Output:
	// 200
	// [{"value":1,"id":"a","type":"gauge"},{"delta":2,"id":"b","type":"counter"}]
}

func ExampleGetValueJSONHandler() {
	conf := config.Config{
		Addr:          "",
		StoreInterval: 0,
	}
	st := memory.NewMetricsRepository()
	ctx := context.Background()

	appInstance := &server.App{
		Config:  &conf,
		Storage: st,
	}

	_ = appInstance.Storage.AddCounterValue(ctx, "c", 1)

	handler := GetValueJSONHandler(appInstance)

	request := httptest.NewRequest(http.MethodPost, "/value/", strings.NewReader(`{"id":"c","type":"counter"}`))
	w := httptest.NewRecorder()

	handler(w, request)
	result := w.Result()
	defer func() {
		_ = result.Body.Close()
	}()

	body, _ := io.ReadAll(result.Body)

	fmt.Println(result.StatusCode)
	fmt.Println(string(body))

	// Output:
	// 200
	// {"delta":1,"id":"c","type":"counter"}
}
