package middleware

import (
	"context"
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
	"github.com/sotavant/yandex-metrics/internal/utils"
	pb "github.com/sotavant/yandex-metrics/proto"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

func TestRequestHasherMiddleware(t *testing.T) {
	key := "hashKey"
	wrongKey := "wrongHashKey"
	requestBody := `{"id":"ss","type":"gauge","value":3}`
	hash, err := utils.GetHash([]byte(requestBody), key)
	assert.NoError(t, err)
	internal.InitLogger()

	conf := config.Config{
		Addr:            "",
		StoreInterval:   0,
		FileStoragePath: "/tmp/fs_test",
		Restore:         false,
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
		name,
		key string
		wantStatus int
	}{
		{
			"correctHash",
			key,
			http.StatusOK,
		},
		{
			"wrongHash",
			wrongKey,
			http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasherMiddlware := NewHasher(tt.key)

			r := chi.NewRouter()
			r.Use(hasherMiddlware.Handler)
			r.Post("/update/", handlers.UpdateJSONHandler(appInstance))

			w := httptest.NewRecorder()

			reqFunc := func() *http.Request {
				req := httptest.NewRequest(http.MethodPost, "/update/", strings.NewReader(requestBody))
				req.Header.Set(utils.HasherHeaderKey, hash)
				req.Header.Set("Accept", "application/json")
				req.Header.Set("Content-Type", "application/json")

				return req
			}

			r.ServeHTTP(w, reqFunc())
			result := w.Result()
			defer func() {
				err := result.Body.Close()
				assert.NoError(t, err)
			}()

			assert.Equal(t, tt.wantStatus, result.StatusCode)
		})
	}
}

func TestResponseHasherMiddleware(t *testing.T) {
	key := "hashKey"
	requestBody := `{"value":3,"id":"ss","type":"gauge"}`
	ctx := context.Background()
	hash, err := utils.GetHash([]byte(requestBody), key)
	assert.NoError(t, err)
	internal.InitLogger()

	conf := config.Config{
		Addr:            "",
		StoreInterval:   0,
		FileStoragePath: "/tmp/fs_test",
		Restore:         false,
	}
	st := memory.NewMetricsRepository()
	fs, err := storage.NewFileStorage(conf.FileStoragePath, conf.Restore, conf.StoreInterval)
	assert.NoError(t, err)

	err = st.AddGaugeValue(ctx, "ss", 3)
	assert.NoError(t, err)

	appInstance := &server.App{
		Config:  &conf,
		Storage: st,
		Fs:      fs,
	}

	tests := []struct {
		name,
		wantHash string
		wantStatus int
	}{
		{
			"correctHash",
			hash,
			http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasherMiddleware := NewHasher(key)

			r := chi.NewRouter()
			r.Use(hasherMiddleware.Handler)
			r.Post("/value/", handlers.GetValueJSONHandler(appInstance))

			w := httptest.NewRecorder()

			reqFunc := func() *http.Request {
				req := httptest.NewRequest(http.MethodPost, "/value/", strings.NewReader(requestBody))
				req.Header.Set(utils.HasherHeaderKey, hash)
				req.Header.Set("Accept", "application/json")
				req.Header.Set("Content-Type", "application/json")

				return req
			}

			r.ServeHTTP(w, reqFunc())
			result := w.Result()
			defer func() {
				err := result.Body.Close()
				assert.NoError(t, err)
			}()

			assert.Equal(t, hash, result.Header.Get(utils.HasherHeaderKey))
			assert.Equal(t, tt.wantStatus, result.StatusCode)
		})
	}
}

func TestHasher_CheckHasInterceptor(t *testing.T) {
	internal.InitLogger()

	var ctx context.Context
	const hashKey = "hashKey"

	val := 1.333

	m := internal.Metrics{
		Value: &val,
		Delta: nil,
		ID:    "sss",
		MType: internal.GaugeType,
	}

	mHash, err := utils.GetMetricHash(m, hashKey)
	assert.NoError(t, err)

	tests := []struct {
		name       string
		key        string
		reqHash    string
		wantStatus codes.Code
	}{
		{
			"empty key",
			"",
			"",
			codes.OK,
		},
		{
			"without hash",
			hashKey,
			"",
			codes.InvalidArgument,
		},
		{
			"wrong hash",
			hashKey,
			"lksdfsldkfj",
			codes.InvalidArgument,
		},
		{
			"good hash",
			hashKey,
			mHash,
			codes.OK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hashMiddleware := NewHasher(tt.key)

			lis = bufconn.Listen(bufSize)
			s := grpc.NewServer(grpc.UnaryInterceptor(hashMiddleware.CheckHashInterceptor))
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
