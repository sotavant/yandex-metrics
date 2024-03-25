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
	URL             = `/update/`
	batchURL        = `/updates/`
	counterType     = `counter`
	gaugeType       = `gauge`
	poolCounterName = `PollCount`
)

func ReportMetric(ms *storage.MetricsStorage) {
	//sendGauge(ms)
	//sendCounter(ms)
	sendBatchMetrics(ms)
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

		sendRequest(jsonData, URL)
	}
}

func sendBatchMetrics(ms *storage.MetricsStorage) {
	if len(ms.Metrics) == 0 {
		return
	}

	metricLen := len(ms.Metrics)
	if ms.PollCount != 0 {
		metricLen += 1
	}

	m := make([]internal.Metrics, 0, metricLen)

	for k := range ms.Metrics {
		val := ms.Metrics[k]
		m = append(m, internal.Metrics{
			ID:    k,
			MType: gaugeType,
			Value: &val,
		})
	}

	m = append(m, internal.Metrics{
		ID:    poolCounterName,
		MType: counterType,
		Delta: &ms.PollCount,
	})

	jsonData, err := json.Marshal(m)
	if err != nil {
		internal.Logger.Infoln("marshall error", err)
		return
	}

	sendRequest(jsonData, batchURL)
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
	sendRequest(jsonData, URL)
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
