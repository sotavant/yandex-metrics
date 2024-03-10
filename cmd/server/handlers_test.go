package main

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/sotavant/yandex-metrics/internal"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_getValueHandler(t *testing.T) {
	type want struct {
		status int
		value  string
	}

	tests := []struct {
		name          string
		mType         string
		mName         string
		ctxMetricName string
		counterValue  int64
		gaugeValue    float64
		want          want
		request       string
	}{
		{
			name:          "getExistCounterValue",
			mType:         counterType,
			mName:         `ss`,
			ctxMetricName: `ss`,
			counterValue:  3,
			want: struct {
				status int
				value  string
			}{status: http.StatusOK, value: `3`},
			request: `/value/counter/ss`,
		},
		{
			name:          "getAbsentCounterValue",
			mType:         counterType,
			mName:         `ss`,
			ctxMetricName: `sss`,
			want: struct {
				status int
				value  string
			}{status: http.StatusNotFound},
			request: `/value/counter/ss`,
		},
		{
			name:          "getExistGaugeValue",
			mType:         gaugeType,
			mName:         `ss`,
			ctxMetricName: `ss`,
			counterValue:  3,
			gaugeValue:    3,
			want: struct {
				status int
				value  string
			}{status: http.StatusOK, value: `3`},
			request: `/value/gauge/ss`,
		},
		{
			name:          "getAbsentGaugeValue",
			mType:         gaugeType,
			mName:         `ss`,
			ctxMetricName: `sss`,
			want: struct {
				status int
				value  string
			}{status: http.StatusNotFound},
			request: `/value/gauge/ss`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, tt.request, nil)
			w := httptest.NewRecorder()
			appInstance := &app{
				config:     nil,
				memStorage: NewMemStorage(),
				fs:         nil,
			}

			if tt.mType == counterType {
				appInstance.memStorage.AddCounterValue(tt.mName, tt.counterValue)
			} else {
				appInstance.memStorage.AddGaugeValue(tt.mName, tt.gaugeValue)
			}

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add(`type`, tt.mType)
			rctx.URLParams.Add(`name`, tt.ctxMetricName)

			request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))

			h := http.HandlerFunc(getValueHandler(appInstance))
			h(w, request)
			result := w.Result()
			defer func() {
				err := result.Body.Close()
				assert.NoError(t, err)
			}()

			assert.Equal(t, tt.want.status, result.StatusCode)

			if tt.name != `getAbsentCounterValue` && tt.name != `getAbsentGaugeValue` {
				bodyBytes, err := io.ReadAll(result.Body)
				assert.NoError(t, err)
				assert.Equal(t, tt.want.value, string(bodyBytes))
			}
		})
	}
}

func Test_getValuesHandler(t *testing.T) {
	type want struct {
		status int
		value  string
	}

	type values []struct {
		key   string
		value float64
	}

	tests := []struct {
		name    string
		values  values
		want    want
		request string
	}{
		{
			name:   "emptyValue",
			values: values{},
			want: struct {
				status int
				value  string
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
				status int
				value  string
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
				status int
				value  string
				//}{status: http.StatusOK, value: `<p>pp: -456.123</p><p>ss: 134.456</p>`},
			}{status: http.StatusOK, value: `<p>ss: 134.456</p><p>pp: -456.123</p>`},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()
			appInstance := &app{
				config:     nil,
				memStorage: NewMemStorage(),
				fs:         nil,
			}

			for _, v := range tt.values {
				appInstance.memStorage.AddGaugeValue(v.key, v.value)
			}

			h := http.HandlerFunc(getValuesHandler(appInstance))
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
	conf := initConfig()

	if conf.databaseDSN == "" {
		return
	}

	tests := []struct {
		name       string
		conf       *config
		wantStatus int
	}{
		{
			name:       "emptyDsn",
			conf:       &config{databaseDSN: ""},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "badDsn",
			conf:       &config{databaseDSN: "postgres://username:password@localhost:5432/database_name"},
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
			dbConn, _ := initDB(ctx, *tt.conf)
			h := http.HandlerFunc(pingDBHandler(ctx, dbConn))
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
