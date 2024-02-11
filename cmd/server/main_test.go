package main

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_badTypeHandler(t *testing.T) {
	type wants struct {
		responseStatus int
	}

	storage := NewMemStorage()

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
			h := http.HandlerFunc(updateHandler(storage))
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