package client

import (
	"github.com/sotavant/yandex-metrics/internal"
	"github.com/sotavant/yandex-metrics/internal/agent/config"
	storage2 "github.com/sotavant/yandex-metrics/internal/agent/storage"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func BenchmarkReportMetric(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	storage := storage2.NewStorage()
	storage.UpdateValues()
	storage.UpdateAdditionalValues()

	config.AppConfig = &config.Config{
		Addr:      strings.TrimPrefix(server.URL, "http://"),
		HashKey:   "",
		RateLimit: 10,
	}
	internal.InitLogger()

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		ReportMetric(storage, config.AppConfig.RateLimit)
	}
}
