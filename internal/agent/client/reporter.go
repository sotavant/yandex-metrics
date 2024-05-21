package client

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"github.com/go-resty/resty/v2"
	"github.com/sotavant/yandex-metrics/internal"
	"github.com/sotavant/yandex-metrics/internal/agent/config"
	"github.com/sotavant/yandex-metrics/internal/agent/storage"
	"github.com/sotavant/yandex-metrics/internal/utils"
	"syscall"
	"time"
)

const (
	updateURL       = `/update/`
	batchUpdateURL  = `/updates/`
	counterType     = `counter`
	gaugeType       = `gauge`
	poolCounterName = `PollCount`
)

type Semaphore struct {
	semaCh chan struct{}
}

func NewSemaphore(maxReq int) *Semaphore {
	return &Semaphore{
		semaCh: make(chan struct{}, maxReq),
	}
}

func (s *Semaphore) Acquire() {
	s.semaCh <- struct{}{}
}

func (s *Semaphore) Release() {
	<-s.semaCh
}

func ReportMetric(ms *storage.MetricsStorage, workerCount int) {
	//sendGauge(ms)
	//sendCounter(ms)
	//sendBatchMetrics(ms)
	sendMetricsByWorkers(ms, workerCount)
}

func sendGauge(ms *storage.MetricsStorage) {
	for k, v := range ms.Metrics {
		m := internal.Metrics{
			ID:    k,
			MType: gaugeType,
			Value: &v,
		}

		jsonData, err := json.Marshal(m)
		if err != nil {
			internal.Logger.Infoln("marshall error", err)
			return
		}

		sendRequest(jsonData, updateURL)
	}
}

func sendBatchMetrics(ms *storage.MetricsStorage) {
	m := collectMetrics(ms)
	if len(m) == 0 {
		return
	}

	jsonData, err := json.Marshal(m)
	if err != nil {
		internal.Logger.Infoln("marshall error", err)
		return
	}

	sendRequest(jsonData, batchUpdateURL)
}

func sendCounter(ms *storage.MetricsStorage) {
	m := internal.Metrics{
		ID:    poolCounterName,
		MType: counterType,
		Delta: &ms.PollCount,
	}

	jsonData, err := json.Marshal(m)
	if err != nil {
		internal.Logger.Infoln("marshall error", err)
		return
	}
	sendRequest(jsonData, updateURL)
}

func sendMetricsByWorkers(ms *storage.MetricsStorage, workersCount int) {
	m := collectMetrics(ms)
	if len(m) == 0 {
		return
	}

	jobs := make(chan []byte, len(m))

	for w := 0; w < workersCount; w++ {
		go worker(jobs)
	}

	for _, metric := range m {
		jsonData, err := json.Marshal(metric)
		if err != nil {
			internal.Logger.Infoln("marshall error", err)
			return
		}
		jobs <- jsonData
	}
	close(jobs)
}

func worker(jobs <-chan []byte) {
	for j := range jobs {
		sendRequest(j, updateURL)
	}
}

func sendRequest(jsonData []byte, url string) {
	intervals := utils.GetRetryWaitTimes()
	retries := len(intervals) + 1
	counter := 1
	data := getCompressedData(jsonData)

	for counter <= retries {
		client := resty.New()
		req := client.R().
			SetHeader("Content-Type", "application/json").
			SetHeader("Content-Encoding", "gzip").
			SetBody(data)

		req = addHashData(req, data)

		internal.Logger.Infoln("sending request", string(jsonData))
		_, err := req.Post("http://" + config.AppConfig.Addr + url)

		if err != nil {
			internal.Logger.Infoln("error in request", err)
			if errors.Is(err, syscall.ECONNREFUSED) {
				time.Sleep(time.Duration(intervals[counter]) * time.Second)
				counter++
			} else {
				break
			}
		} else {
			break
		}
	}
}

func getCompressedData(data []byte) *bytes.Buffer {
	buf := bytes.NewBuffer(nil)
	zb := gzip.NewWriter(buf)
	_, err := zb.Write(data)

	if err != nil {
		panic(err)
	}

	if err := zb.Close(); err != nil {
		panic(err)
	}

	return buf
}

func addHashData(req *resty.Request, buf *bytes.Buffer) *resty.Request {
	if config.AppConfig.HashKey == "" {
		return req
	}

	hash, err := utils.GetHash(buf.Bytes(), config.AppConfig.HashKey)
	if err != nil {
		internal.Logger.Infoln("error in get hash", err)
		panic(err)
	}

	req.SetHeader(utils.HasherHeaderKey, hash)
	return req
}

func collectMetrics(ms *storage.MetricsStorage) []internal.Metrics {
	ms.RWMutex.RLock()
	defer ms.RWMutex.RUnlock()

	var res []internal.Metrics

	if len(ms.Metrics) == 0 {
		return res
	}

	metricLen := len(ms.Metrics)
	if ms.PollCount != 0 {
		metricLen++
	}

	res = make([]internal.Metrics, 0, metricLen)

	for k := range ms.Metrics {
		val := ms.Metrics[k]
		res = append(res, internal.Metrics{
			ID:    k,
			MType: gaugeType,
			Value: &val,
		})
	}

	if ms.PollCount != 0 {
		res = append(res, internal.Metrics{
			ID:    poolCounterName,
			MType: counterType,
			Delta: &ms.PollCount,
		})
	}

	return res
}
