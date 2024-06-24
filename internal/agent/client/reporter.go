// Package client Данный пакет служит для отправки собранных метрик на сервер.
package client

import (
	"bytes"
	"compress/gzip"
	"errors"
	"os"
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

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type Reporter struct {
	ch *utils.Cipher
}

func NewReporter(ch *utils.Cipher) *Reporter {
	return &Reporter{
		ch: ch,
	}
}

// ReportMetric отправляет метрики.
// На вход принимает хранилище и количество воркеров (параллельных процессов)
func (r *Reporter) ReportMetric(ms *storage.MetricsStorage, workerCount int, sigs chan os.Signal) bool {
	//sendGauge(ms)
	//sendCounter(ms)
	//sendBatchMetrics(ms)
	for {
		r.sendMetricsByWorkers(ms, workerCount)
		select {
		case <-sigs:
			return true
		default:
			return false
		}
	}
}

func (r *Reporter) sendGauge(ms *storage.MetricsStorage) {
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

		r.sendRequest(jsonData, updateURL)
	}
}

func (r *Reporter) sendBatchMetrics(ms *storage.MetricsStorage) {
	m := collectMetrics(ms)
	if len(m) == 0 {
		return
	}

	jsonData, err := json.Marshal(m)
	if err != nil {
		internal.Logger.Infoln("marshall error", err)
		return
	}

	r.sendRequest(jsonData, batchUpdateURL)
}

func (r *Reporter) sendCounter(ms *storage.MetricsStorage) {
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
	r.sendRequest(jsonData, updateURL)
}

func (r *Reporter) sendMetricsByWorkers(ms *storage.MetricsStorage, workersCount int) {
	m := collectMetrics(ms)
	if len(m) == 0 {
		return
	}

	jobs := make(chan []byte, len(m))

	for w := 0; w < workersCount; w++ {
		go r.worker(jobs)
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

func (r *Reporter) worker(jobs <-chan []byte) {
	for j := range jobs {
		r.sendRequest(j, updateURL)
	}
}

func (r *Reporter) sendRequest(jsonData []byte, url string) {
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
	req = r.addCipheredData(req, data)

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

func (r *Reporter) addCipheredData(req *resty.Request, buf *bytes.Buffer) *resty.Request {
	if r.ch.IsPublicKeyExist() {
		cryptedData, err := r.ch.Encrypt(buf.Bytes())
		if err != nil {
			internal.Logger.Infoln("error in encrypt buf", err)
			panic(err)
		}
		req.SetBody(cryptedData)
		return req
	}
	req.SetBody(buf)
	return req
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
