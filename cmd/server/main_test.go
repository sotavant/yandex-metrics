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
			badTypeHandler(w, request)

			res := w.Result()
			assert.Equal(t, tt.wants.responseStatus, res.StatusCode)
		})
	}
}

func Test_defaultHandler(t *testing.T) {
	type wants struct {
		responseStatus int
	}

	tests := []struct {
		name    string
		request string
		wants   wants
	}{
		{
			name:    `badPoint`,
			request: `/updatesd/badType/asdf/sdff`,
			wants: wants{
				responseStatus: http.StatusNotFound,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, tt.request, nil)
			w := httptest.NewRecorder()

			defaultHandler(w, request)

			res := w.Result()
			assert.Equal(t, tt.wants.responseStatus, res.StatusCode)
		})
	}
}
