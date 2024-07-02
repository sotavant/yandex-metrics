package middleware

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/sotavant/yandex-metrics/internal"
	"github.com/sotavant/yandex-metrics/internal/server"
	"github.com/sotavant/yandex-metrics/internal/server/config"
	grpc2 "github.com/sotavant/yandex-metrics/internal/server/grpc"
	"github.com/sotavant/yandex-metrics/internal/server/handlers"
	"github.com/sotavant/yandex-metrics/internal/server/repository/memory"
	"github.com/sotavant/yandex-metrics/internal/server/storage"
	pb "github.com/sotavant/yandex-metrics/proto"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

var lis *bufconn.Listener

func TestIPChecker_CheckIP_Handler(t *testing.T) {
	internal.InitLogger()

	requestBody := `{"value":3,"id":"ss","type":"gauge"}`

	conf := config.Config{
		Addr:            "",
		StoreInterval:   0,
		FileStoragePath: "/tmp/fs_test",
		Restore:         false,
		TrustedSubnet:   "192.168.1.0/24",
	}

	st := memory.NewMetricsRepository()
	fs, err := storage.NewFileStorage(conf.FileStoragePath, conf.Restore, conf.StoreInterval)
	assert.NoError(t, err)

	appInstance := &server.App{
		Config:  &conf,
		Storage: st,
		Fs:      fs,
	}

	tests := []struct {
		name       string
		ip         string
		wantStatus int
	}{
		{
			"good IP",
			"192.168.1.130",
			http.StatusOK,
		},
		{
			"bad IP",
			"132.132.132.132",
			http.StatusForbidden,
		},
		{
			"empty IP",
			"",
			http.StatusForbidden,
		},
		{
			"not correct IP",
			"sdfsdf.sdfsd.sdfsdf",
			http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ipMiddleware := NewIPChecker(conf.TrustedSubnet)
			assert.NotNil(t, ipMiddleware)

			r := chi.NewRouter()
			r.Use(ipMiddleware.CheckIP)
			r.Post("/update/", handlers.UpdateJSONHandler(appInstance))

			w := httptest.NewRecorder()

			reqFunc := func() *http.Request {
				req := httptest.NewRequest(http.MethodPost, "/update/", strings.NewReader(requestBody))
				req.Header.Set("Accept", "application/json")
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("X-Real-IP", tt.ip)

				return req
			}

			r.ServeHTTP(w, reqFunc())
			result := w.Result()
			defer func() {
				err = result.Body.Close()
				assert.NoError(t, err)
			}()

			assert.Equal(t, tt.wantStatus, result.StatusCode)
		})
	}
}

func TestIPChecker_CheckIPInterceptor(t *testing.T) {

	internal.InitLogger()

	requestBody := `{"value":3,"id":"ss","type":"gauge"}`

	conf := config.Config{
		Addr:            "",
		StoreInterval:   0,
		FileStoragePath: "/tmp/fs_test",
		Restore:         false,
		TrustedSubnet:   "192.168.1.0/24",
	}

	st := memory.NewMetricsRepository()
	fs, err := storage.NewFileStorage(conf.FileStoragePath, conf.Restore, conf.StoreInterval)
	assert.NoError(t, err)

	appInstance := &server.App{
		Config:  &conf,
		Storage: st,
		Fs:      fs,
	}

	tests := []struct {
		name       string
		ip         string
		wantStatus int
	}{
		{
			"good IP",
			"192.168.1.130",
			http.StatusOK,
		},
		{
			"bad IP",
			"132.132.132.132",
			http.StatusForbidden,
		},
		{
			"empty IP",
			"",
			http.StatusForbidden,
		},
		{
			"not correct IP",
			"sdfsdf.sdfsd.sdfsdf",
			http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ipMiddleware := NewIPChecker(conf.TrustedSubnet)
			assert.NotNil(t, ipMiddleware)

			lis = bufconn.Listen(bufSize)
			s := grpc.NewServer(grpc.UnaryInterceptor(ipMiddleware.CheckIPInterceptor))
			pb.RegisterMetricsServer(s, &grpc2.MetricServer{})
			go func() {
				if err = s.Serve(lis); err != nil {
					assert.NoError(t, err)
				}
			}()

			reqFunc := func() *http.Request {
				req := httptest.NewRequest(http.MethodPost, "/update/", strings.NewReader(requestBody))
				req.Header.Set("Accept", "application/json")
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("X-Real-IP", tt.ip)

				return req
			}

			r.ServeHTTP(w, reqFunc())
			result := w.Result()
			defer func() {
				err = result.Body.Close()
				assert.NoError(t, err)
			}()

			assert.Equal(t, tt.wantStatus, result.StatusCode)
		})
	}
}

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}
