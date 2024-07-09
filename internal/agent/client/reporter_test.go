package client

import (
	"compress/gzip"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"testing"

	"github.com/sotavant/yandex-metrics/internal"
	"github.com/sotavant/yandex-metrics/internal/agent/config"
	storage2 "github.com/sotavant/yandex-metrics/internal/agent/storage"
	"github.com/sotavant/yandex-metrics/internal/utils"
	"github.com/stretchr/testify/assert"
)

func BenchmarkReportMetric(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	r := NewReporter(nil)
	storage := storage2.NewStorage()
	storage.UpdateValues()
	storage.UpdateAdditionalValues()
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	config.AppConfig = &config.Config{
		Addr:      strings.TrimPrefix(server.URL, "http://"),
		HashKey:   "",
		RateLimit: 10,
	}
	internal.InitLogger()

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		r.ReportMetric(storage, config.AppConfig.RateLimit, sigs)
	}
}

func ExampleReporter_ReportMetric() {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	r := NewReporter(nil)
	ms := storage2.NewStorage()
	ms.UpdateValues()
	ms.UpdateAdditionalValues()

	config.AppConfig = &config.Config{
		Addr:      strings.TrimPrefix(server.URL, "http://"),
		HashKey:   "",
		RateLimit: 10,
	}
	internal.InitLogger()
	r.ReportMetric(ms, config.AppConfig.RateLimit, sigs)
}

func Test_getCompressedData(t *testing.T) {
	var s []byte
	var err error
	var gz *gzip.Reader
	tests := []struct {
		name string
		data string
	}{
		{
			name: "compress data",
			data: "test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cd := getCompressedData([]byte(tt.data))
			assert.NotNil(t, cd)
			gz, err = gzip.NewReader(cd)

			assert.NoError(t, err)
			s, err = io.ReadAll(gz)

			assert.Equal(t, tt.data, string(s))
		})
	}
}

func TestReporter_sendRequestWithIP(t *testing.T) {
	internal.InitLogger()
	ip, err := utils.GetLocalIP()
	wrongIP := net.ParseIP("127.0.0.1")
	assert.NoError(t, err)

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		header := req.Header.Get("X-Real-IP")
		assert.Equal(t, ip.String(), header)
		assert.NotEqual(t, wrongIP.String(), header)
	}))

	defer server.Close()

	r := NewReporter(nil)
	ms := storage2.NewStorage()
	ms.UpdateValues()
	ms.UpdateAdditionalValues()

	config.AppConfig = &config.Config{
		Addr:      strings.TrimPrefix(server.URL, "http://"),
		HashKey:   "",
		RateLimit: 1,
	}

	r.sendRequest([]byte("[]"), "/")
}
