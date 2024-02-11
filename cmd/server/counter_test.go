package main

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_handleCounter(t *testing.T) {
	type want struct {
		contentType int
		value       int64
	}

	tests := []struct {
		name    string
		request string
		storage *MemStorage
		want    want
	}{
		{
			name:    `newValue`,
			request: `/update/counter/newValue/1`,
			storage: NewMemStorage(),
			want: struct {
				contentType int
				value       int64
			}{contentType: http.StatusOK, value: 1},
		},
		{
			name:    `updateValue`,
			request: `/update/counter/updateValue/3`,
			storage: NewMemStorage(),
			want: struct {
				contentType int
				value       int64
			}{contentType: http.StatusOK, value: 6},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, tt.request, nil)
			w := httptest.NewRecorder()
			h := http.HandlerFunc(handleCounter(tt.storage))
			h(w, request)
			result := w.Result()

			switch tt.name {
			case `newValue`:
				assert.Equal(t, tt.want.contentType, result.StatusCode)
				assert.Equal(t, tt.want.value, tt.storage.Counter[tt.name])
			case `updateValue`:
				h(w, request)
				assert.Equal(t, tt.want.contentType, result.StatusCode)
				assert.Equal(t, tt.want.value, tt.storage.Counter[tt.name])
			}
		})
	}
}
