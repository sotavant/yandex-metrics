package main

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_handleCounter(t *testing.T) {
	type want struct {
		status int
		value  int64
	}

	tests := []struct {
		name    string
		request string
		storage *MemStorage
		mName   string
		mValue  string
		want    want
	}{
		{
			name:    `newValue`,
			request: `/update/counter/newValue/1`,
			mName:   `newValue`,
			mValue:  `1`,
			storage: NewMemStorage(),
			want: struct {
				status int
				value  int64
			}{status: http.StatusOK, value: 1},
		},
		{
			name:    `updateValue`,
			request: `/update/counter/updateValue/3`,
			mName:   `updateValue`,
			mValue:  `3`,
			storage: NewMemStorage(),
			want: struct {
				status int
				value  int64
			}{status: http.StatusOK, value: 6},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, tt.request, nil)
			w := httptest.NewRecorder()

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add(`type`, counterType)
			rctx.URLParams.Add(`name`, tt.mName)
			rctx.URLParams.Add(`value`, tt.mValue)

			request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))

			h := http.HandlerFunc(updateHandler(tt.storage))
			h(w, request)
			result := w.Result()
			defer func() {
				err := result.Body.Close()
				assert.NoError(t, err)
			}()

			switch tt.name {
			case `newValue`:
				assert.Equal(t, tt.want.status, result.StatusCode)
				assert.Equal(t, tt.want.value, tt.storage.Counter[tt.name])
			case `updateValue`:
				h(w, request)
				assert.Equal(t, tt.want.status, result.StatusCode)
				assert.Equal(t, tt.want.value, tt.storage.Counter[tt.name])
			}
		})
	}
}
