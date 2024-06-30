package client

import (
	"bytes"
	"encoding/gob"
	"os"

	"github.com/sotavant/yandex-metrics/internal"
	"github.com/sotavant/yandex-metrics/internal/agent/config"
	"github.com/sotavant/yandex-metrics/internal/agent/storage"
	"github.com/sotavant/yandex-metrics/internal/utils"
	pb "github.com/sotavant/yandex-metrics/proto"
	"google.golang.org/grpc/metadata"
)

type GRPCReporter struct {
	c pb.MetricsClient
}

func NewGRPCReporter(c pb.MetricsClient) *GRPCReporter {
	return &GRPCReporter{
		c: c,
	}
}

// ReportMetric отправляет метрики по протоколу gRPC.
// На вход принимает хранилище и количество воркеров (параллельных процессов)
func (r *GRPCReporter) ReportMetric(ms *storage.MetricsStorage, workerCount int, sigs chan os.Signal) bool {
	for {
		r.sendMetricsByGRPCWorkers(ms, workerCount)
		select {
		case <-sigs:
			return true
		default:
			return false
		}
	}
}

func (r *GRPCReporter) sendMetricsByGRPCWorkers(ms *storage.MetricsStorage, workersCount int) {
	m := collectMetrics(ms)
	if len(m) == 0 {
		return
	}

	jobs := make(chan internal.Metrics, len(m))

	for w := 0; w < workersCount; w++ {
		go r.gRPCWorker(jobs)
	}

	for _, metric := range m {
		jobs <- metric
	}

	close(jobs)
}

func (r *GRPCReporter) gRPCWorker(jobs <-chan internal.Metrics) {
	for j := range jobs {
		r.sendGRPCRequest(j)
	}
}

func (r *GRPCReporter) sendGRPCRequest(m internal.Metrics) {
	intervals := utils.GetRetryWaitTimes()
	retries := len(intervals)
	retries++
	//counter := 1

	//md := r.SetMetadata(m)
	//ctx := metadata.NewOutgoingContext(context.Background(), md)
}

func (r *GRPCReporter) SetMetadata(m internal.Metrics) metadata.MD {
	ip, err := utils.GetLocalIP()
	if err != nil {
		internal.Logger.Fatalw("get local ip error", "err", err)
	}

	md := metadata.Pairs("X-Real-IP", ip.String())
	md = r.addHashMetadata(m, md)

	return md
}

func (r *GRPCReporter) addHashMetadata(m internal.Metrics, md metadata.MD) metadata.MD {
	if config.AppConfig.HashKey == "" {
		return md
	}

	var metricsBuf bytes.Buffer
	enc := gob.NewEncoder(&metricsBuf)
	err := enc.Encode(m)
	if err != nil {
		internal.Logger.Fatalw("encode metrics error", "err", err)
	}

	hash, err := utils.GetHash(metricsBuf.Bytes(), config.AppConfig.HashKey)
	if err != nil {
		internal.Logger.Fatalw("get hash error", "err", err)
	}

	md.Set(utils.HasherHeaderKey, hash)
	return md
}
