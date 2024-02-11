package main

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_handleGauge(t *testing.T) {
	type want struct {
		contentType int
		value       float64
	}

	tests := []struct {
		name    string
		request string
		storage *MemStorage
		want    want
	}{
		{
			name:    `newValue`,
			request: `/update/gauge/newValue/1`,
			storage: NewMemStorage(),
			want: struct {
				contentType int
				value       float64
			}{contentType: http.StatusOK, value: 1},
		},
		{
			name:    `updateValue`,
			request: `/update/gauge/updateValue/3`,
			storage: NewMemStorage(),
			want: struct {
				contentType int
				value       float64
			}{contentType: http.StatusOK, value: 3},
		},
		{
			name:    `badValue`,
			request: `/update/gauge/badValue/sdfsdfsdf`,
			storage: NewMemStorage(),
			want: struct {
				contentType int
				value       float64
			}{contentType: http.StatusBadRequest, value: 0},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, tt.request, nil)
			w := httptest.NewRecorder()
			h := http.HandlerFunc(handleGauge(tt.storage))
			h(w, request)
			result := w.Result()

			switch tt.name {
			case `updateValue`:
				h(w, request)
				assert.Equal(t, tt.want.contentType, result.StatusCode)
				assert.Equal(t, tt.want.value, tt.storage.Gauge[tt.name])
			default:
				assert.Equal(t, tt.want.contentType, result.StatusCode)
				assert.Equal(t, tt.want.value, tt.storage.Gauge[tt.name])
			}
		})
	}
}
