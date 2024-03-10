package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"github.com/sotavant/yandex-metrics/internal/server/repository/in_memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func Test_updateJsonHandler(t *testing.T) {
	conf := config{
		addr:            "",
		storeInterval:   0,
		fileStoragePath: "/tmp/fs_test",
		restore:         false,
	}
	st := in_memory.NewMetricsRepository()
	fs, err := NewFileStorage(conf)
	assert.NoError(t, err)

	appInstance := &app{
		config:     &conf,
		memStorage: st,
		fs:         fs,
	}

	handler := updateJSONHandler(appInstance)

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
	appInstance := &app{
		config:     nil,
		memStorage: in_memory.NewMetricsRepository(),
		fs:         nil,
	}

	err := appInstance.memStorage.AddGaugeValue(context.Background(), "ss", -3444)
	assert.NoError(t, err)
	err = appInstance.memStorage.AddCounterValue(context.Background(), "ss", 3)
	assert.NoError(t, err)
	handler := getValueJSONHandler(appInstance)

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
	appInstance := &app{
		config:     nil,
		memStorage: in_memory.NewMetricsRepository(),
		fs:         nil,
	}
	err := appInstance.memStorage.AddGaugeValue(context.Background(), "ss", -3444)
	assert.NoError(t, err)
	err = appInstance.memStorage.AddCounterValue(context.Background(), "ss", 3)
	assert.NoError(t, err)
	requestBody := `{"id":"ss","type":"counter","delta":3}
`
	htmlResponse := `<p>ss: -3444</p>`

	t.Run("sends_gzip", func(t *testing.T) {
		handler := gzipMiddleware(getValueJSONHandler(appInstance))

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
		handler := gzipMiddleware(getValuesHandler(appInstance))

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
