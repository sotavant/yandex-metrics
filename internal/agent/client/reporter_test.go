package client

import (
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
