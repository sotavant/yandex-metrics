package handlers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sotavant/yandex-metrics/internal"
	"github.com/sotavant/yandex-metrics/internal/server"
	"github.com/sotavant/yandex-metrics/internal/server/config"
	"github.com/sotavant/yandex-metrics/internal/server/repository"
	"github.com/sotavant/yandex-metrics/internal/server/repository/memory"
	"github.com/sotavant/yandex-metrics/internal/server/repository/postgres"
	"github.com/sotavant/yandex-metrics/internal/server/repository/postgres/test"
	"github.com/sotavant/yandex-metrics/internal/server/storage"
	"github.com/stretchr/testify/assert"
)

func Test_getValueHandler(t *testing.T) {
	type want struct {
		value  string
		status int
	}

	ctx := context.Background()
	conn, tableName, DNS, err := test.InitConnection(ctx, t)
	assert.NoError(t, err)

	if conn != nil {
		defer func(ctx context.Context, conn *pgxpool.Pool, tableName string) {
			err = test.DropTable(ctx, conn, tableName)
			assert.NoError(t, err)

			conn.Close()
		}(ctx, conn, tableName)
	}

	tests := []struct {
		name          string
		mType         string
		mName         string
		ctxMetricName string
		request       string
		want          want
		gaugeValue    float64
		memory        bool
		counterValue  int64
	}{
		{
			name:          "getExistCounterValue",
			mType:         internal.CounterType,
			mName:         `ss`,
			ctxMetricName: `ss`,
			counterValue:  3,
			want: struct {
				value  string
				status int
			}{status: http.StatusOK, value: `3`},
			request: `/value/counter/ss`,
			memory:  true,
		},
		{
			name:          "getAbsentCounterValue",
			mType:         internal.CounterType,
			mName:         `ss`,
			ctxMetricName: `sss`,
			want: struct {
				value  string
				status int
			}{status: http.StatusNotFound},
			request: `/value/counter/ss`,
			memory:  true,
		},
		{
			name:          "getExistGaugeValue",
			mType:         internal.GaugeType,
			mName:         `ss`,
			ctxMetricName: `ss`,
			counterValue:  3,
			gaugeValue:    3,
			want: struct {
				value  string
				status int
			}{status: http.StatusOK, value: `3`},
			request: `/value/gauge/ss`,
			memory:  true,
		},
		{
			name:          "getAbsentGaugeValue",
			mType:         internal.GaugeType,
			mName:         `ss`,
			ctxMetricName: `sss`,
			want: struct {
				value  string
				status int
			}{status: http.StatusNotFound},
			request: `/value/gauge/ss`,
			memory:  true,
		},
		{
			name:          "getExistCounterValuePG",
			mType:         internal.CounterType,
			mName:         `ss`,
			ctxMetricName: `ss`,
			counterValue:  3,
			want: struct {
				value  string
				status int
			}{status: http.StatusOK, value: `3`},
			request: `/value/counter/ss`,
			memory:  false,
		},
		{
			name:          "getAbsentCounterValuePG",
			mType:         internal.CounterType,
			mName:         `ss`,
			ctxMetricName: `sss`,
			want: struct {
				value  string
				status int
			}{status: http.StatusNotFound},
			request: `/value/counter/ss`,
			memory:  false,
		},
		{
			name:          "getExistGaugeValuePG",
			mType:         internal.GaugeType,
			mName:         `ss`,
			ctxMetricName: `ss`,
			counterValue:  3,
			gaugeValue:    3,
			want: struct {
				value  string
				status int
			}{status: http.StatusOK, value: `3`},
			request: `/value/gauge/ss`,
			memory:  false,
		},
		{
			name:          "getAbsentGaugeValuePG",
			mType:         internal.GaugeType,
			mName:         `ss`,
			ctxMetricName: `sss`,
			want: struct {
				value  string
				status int
			}{status: http.StatusNotFound},
			request: `/value/gauge/ss`,
			memory:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var storage repository.Storage
			request := httptest.NewRequest(http.MethodPost, tt.request, nil)
			w := httptest.NewRecorder()

			if conn != nil && !tt.memory {
				storage, err = postgres.NewMemStorage(ctx, conn, tableName, DNS)
				assert.NoError(t, err)
			} else if !tt.memory && conn == nil {
				return
			} else {
				storage = memory.NewMetricsRepository()
			}

			appInstance := &server.App{
				Config:  nil,
				Storage: storage,
				Fs:      nil,
			}

			if tt.mType == internal.CounterType {
				err := appInstance.Storage.AddCounterValue(request.Context(), tt.mName, tt.counterValue)
				assert.NoError(t, err)
			} else {
				err := appInstance.Storage.AddGaugeValue(request.Context(), tt.mName, tt.gaugeValue)
				assert.NoError(t, err)
			}

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add(`type`, tt.mType)
			rctx.URLParams.Add(`name`, tt.ctxMetricName)

			request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))

			h := http.HandlerFunc(GetValueHandler(appInstance))
			h(w, request)
			result := w.Result()
			defer func() {
				err := result.Body.Close()
				assert.NoError(t, err)
			}()

			assert.Equal(t, tt.want.status, result.StatusCode)

			if !strings.Contains(tt.name, `getAbsentCounterValue`) && !strings.Contains(tt.name, `getAbsentGaugeValue`) {
				bodyBytes, err := io.ReadAll(result.Body)
				assert.NoError(t, err)
				assert.Equal(t, tt.want.value, string(bodyBytes))
			}
		})
	}
}

func Test_getValuesHandler(t *testing.T) {
	type want struct {
		value  string
		status int
	}

	type values []struct {
		key   string
		value float64
	}

	tests := []struct {
		name    string
		request string
		values  values
		want    want
	}{
		{
			name:   "emptyValue",
			values: values{},
			want: struct {
				value  string
				status int
			}{status: http.StatusOK, value: `no value`},
		},
		{
			name: "oneValue",
			values: values{
				{
					key:   `ss`,
					value: 134.456,
				},
			},
			want: struct {
				value  string
				status int
			}{status: http.StatusOK, value: `<p>ss: 134.456</p>`},
		},
		{
			name: "twoValue",
			values: values{
				{
					key:   `ss`,
					value: 134.456,
				},
				{
					key:   `pp`,
					value: -456.123,
				},
			},
			want: struct {
				value  string
				status int
				//}{status: http.StatusOK, value: `<p>pp: -456.123</p><p>ss: 134.456</p>`},
			}{status: http.StatusOK, value: `<p>pp: -456.123</p><p>ss: 134.456</p>`},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()
			appInstance := &server.App{
				Config:  nil,
				Storage: memory.NewMetricsRepository(),
				Fs:      nil,
			}

			for _, v := range tt.values {
				err := appInstance.Storage.AddGaugeValue(request.Context(), v.key, v.value)
				assert.NoError(t, err)
			}

			h := http.HandlerFunc(GetValuesHandler(appInstance))
			h(w, request)
			result := w.Result()
			defer func() {
				err := result.Body.Close()
				assert.NoError(t, err)
			}()

			assert.Equal(t, tt.want.status, result.StatusCode)

			bodyBytes, err := io.ReadAll(result.Body)
			assert.NoError(t, err)
			assert.Equal(t, tt.want.value, string(bodyBytes))
		})
	}
}

func Test_pingDBHandler(t *testing.T) {
	ctx := context.Background()
	internal.InitLogger()
	conf := config.InitConfig()

	if conf.DatabaseDSN == "" {
		return
	}

	tests := []struct {
		conf       *config.Config
		name       string
		wantStatus int
	}{
		{
			name:       "emptyDsn",
			conf:       &config.Config{DatabaseDSN: ""},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "goodDsn",
			conf:       conf,
			wantStatus: http.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbConn, _ := storage.InitDB(ctx, tt.conf.DatabaseDSN)
			h := http.HandlerFunc(PingDBHandler(dbConn))
			request := httptest.NewRequest(http.MethodGet, "/ping", nil)
			w := httptest.NewRecorder()
			h(w, request)
			result := w.Result()
			defer func() {
				err := result.Body.Close()
				assert.NoError(t, err)
			}()

			assert.Equal(t, tt.wantStatus, result.StatusCode)
		})
	}
}

func ExampleUpdateHandler() {
	conf := config.Config{
		Addr:          "",
		StoreInterval: 0,
	}
	ctx := context.Background()

	appInstance := &server.App{
		Config:  &conf,
		Storage: memory.NewMetricsRepository(),
	}

	request := httptest.NewRequest(http.MethodPost, "/", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(`type`, internal.GaugeType)
	rctx.URLParams.Add(`name`, "test")
	rctx.URLParams.Add(`value`, "134134")

	request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))
	h := http.HandlerFunc(UpdateHandler(appInstance))
	h(w, request)
	result := w.Result()
	defer func() {
		_ = result.Body.Close()
	}()
	fmt.Println(result.StatusCode)
	value, _ := appInstance.Storage.GetGaugeValue(ctx, "test")
	fmt.Println(value)

	// Output:
	// 200
	// 134134
}

func ExampleGetValueHandler() {
	var metricValue float64 = 134134
	metricName := "test"

	conf := config.Config{
		Addr:          "",
		StoreInterval: 0,
	}
	ctx := context.Background()

	appInstance := &server.App{
		Config:  &conf,
		Storage: memory.NewMetricsRepository(),
	}

	_ = appInstance.Storage.AddGaugeValue(ctx, metricName, metricValue)

	request := httptest.NewRequest(http.MethodPost, "/", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(`type`, internal.GaugeType)
	rctx.URLParams.Add(`name`, metricName)

	request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))
	h := http.HandlerFunc(GetValueHandler(appInstance))
	h(w, request)
	result := w.Result()
	defer func() {
		_ = result.Body.Close()
	}()
	response, _ := io.ReadAll(result.Body)
	fmt.Println(result.StatusCode)
	fmt.Println(string(response))

	// Output:
	// 200
	// 134134
}

func ExampleGetValuesHandler() {
	conf := config.Config{
		Addr:          "",
		StoreInterval: 0,
	}
	ctx := context.Background()

	appInstance := &server.App{
		Config:  &conf,
		Storage: memory.NewMetricsRepository(),
	}

	_ = appInstance.Storage.AddGaugeValue(ctx, "test", 134134)
	_ = appInstance.Storage.AddGaugeValue(ctx, "next", 1)

	request := httptest.NewRequest(http.MethodPost, "/", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()

	request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))
	h := http.HandlerFunc(GetValuesHandler(appInstance))
	h(w, request)
	result := w.Result()
	defer func() {
		_ = result.Body.Close()
	}()
	response, _ := io.ReadAll(result.Body)
	fmt.Println(result.StatusCode)
	fmt.Println(string(response))

	// Output:
	// 200
	// <p>next: 1</p><p>test: 134134</p>
}

func ExamplePingDBHandler() {
	request := httptest.NewRequest(http.MethodPost, "/", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()

	request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))
	h := http.HandlerFunc(PingDBHandler(nil))
	h(w, request)
	result := w.Result()
	defer func() {
		_ = result.Body.Close()
	}()
	fmt.Println(result.StatusCode)

	// Output:
	// 500
}
