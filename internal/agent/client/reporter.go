// Package client Данный пакет служит для отправки собранных метрик на сервер.
package client

import (
	"bytes"
	"compress/gzip"
	"errors"
	"syscall"
	"time"

	"github.com/go-resty/resty/v2"
	jsoniter "github.com/json-iterator/go"
	"github.com/sotavant/yandex-metrics/internal"
	"github.com/sotavant/yandex-metrics/internal/agent/config"
	"github.com/sotavant/yandex-metrics/internal/agent/storage"
	"github.com/sotavant/yandex-metrics/internal/utils"
)

// настройки
const (
	updateURL       = `/update/`  // адрес для отправки одного значения
	batchUpdateURL  = `/updates/` // адрес для отправки пакета со всеми метриками
	counterType     = `counter`   // название типа со счетчиком
	gaugeType       = `gauge`     // название типа с метриками
	poolCounterName = `PollCount`
)

type Semaphore struct {
	semaCh chan struct{}
}

var json = jsoniter.ConfigCompatibleWithStandardLibrary

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

// ReportMetric отправляет метрики.
// На вход принимает хранилище и количество воркеров (параллельных процессов)
func ReportMetric(ms *storage.MetricsStorage, workerCount int, cipher *utils.Cipher) {
	//sendGauge(ms)
	//sendCounter(ms)
	//sendBatchMetrics(ms)
	sendMetricsByWorkers(ms, workerCount, cipher)
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
	retries := len(intervals)
	retries++
	counter := 1
	data := getCompressedData(jsonData)

	client := resty.New()
	req := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip")

	req = addHashData(req, data)
	req = addCipheredData(req, data)

	for counter <= retries {
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
	zb.Reset(buf)
	_, err := zb.Write(data)

	if err != nil {
		panic(err)
	}

	if err = zb.Close(); err != nil {
		panic(err)
	}

	return buf
}

func addCipheredData(req *resty.Request, data []byte) (*resty.Request, error) {

	req.SetBody(data)
	return req, nil
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
