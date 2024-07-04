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
	"github.com/sotavant/yandex-metrics/internal/server/handlers"
	"github.com/sotavant/yandex-metrics/internal/server/repository/memory"
	"github.com/sotavant/yandex-metrics/internal/server/storage"
	pb "github.com/sotavant/yandex-metrics/proto_test"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024
const trustedSubnet = "192.168.1.0/24"

var lis *bufconn.Listener

func TestIPChecker_CheckIP_Handler(t *testing.T) {
	internal.InitLogger()

	requestBody := `{"value":3,"id":"ss","type":"gauge"}`

	conf := config.Config{
		Addr:            "",
		StoreInterval:   0,
		FileStoragePath: "/tmp/fs_test",
		Restore:         false,
		TrustedSubnet:   trustedSubnet,
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

type TestGRPCServer struct {
	pb.UnimplementedTestServer
}

func (t *TestGRPCServer) SetXRealIP(ctx context.Context, req *pb.SetXRealIPRequest) (*pb.SetXRealIPResponse, error) {
	return nil, nil
}

func TestIPChecker_CheckIPInterceptor(t *testing.T) {
	internal.InitLogger()
	var ctx context.Context

	tests := []struct {
		name       string
		ip         string
		wantStatus codes.Code
	}{
		{
			"good IP",
			"192.168.1.130",
			codes.OK,
		},
		{
			"bad IP",
			"132.132.132.132",
			codes.Unauthenticated,
		},
		{
			"empty IP",
			"",
			codes.Unauthenticated,
		},
		{
			"not correct IP",
			"sdfsdf.sdfsd.sdfsdf",
			codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ipMiddleware := NewIPChecker(trustedSubnet)
			assert.NotNil(t, ipMiddleware)

			lis = bufconn.Listen(bufSize)
			s := grpc.NewServer(grpc.UnaryInterceptor(ipMiddleware.CheckIPInterceptor))
			pb.RegisterTestServer(s, &TestGRPCServer{})
			go func() {
				err := s.Serve(lis)
				assert.NoError(t, err)
			}()

			conn, err := grpc.NewClient("passthrough://bufnet", grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
			assert.NoError(t, err)
			defer func(conn *grpc.ClientConn) {
				err = conn.Close()
				assert.NoError(t, err)
			}(conn)

			if tt.ip != "" {
				md := metadata.Pairs("X-Real-IP", tt.ip)
				ctx = metadata.NewOutgoingContext(context.Background(), md)
			} else {
				ctx = context.Background()
			}

			client := pb.NewTestClient(conn)
			_, err = client.SetXRealIP(ctx, &pb.SetXRealIPRequest{IP: tt.ip})
			if tt.wantStatus == codes.OK {
				assert.NoError(t, err)
			} else {
				assert.Equal(t, status.Code(err), tt.wantStatus)
			}
		})
	}
}

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}
