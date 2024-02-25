package main

import (
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func Test_updateJsonHandler(t *testing.T) {
	st := NewMemStorage()
	handler := updateJsonHandler(st)

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
	st := NewMemStorage()
	st.AddGaugeValue("ss", -3444)
	st.AddCounterValue("ss", 3)
	handler := getValueJsonHandler(st)

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
