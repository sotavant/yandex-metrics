package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"github.com/go-resty/resty/v2"
	"github.com/sotavant/yandex-metrics/internal"
)

const (
	URL             = `/update/`
	counterType     = `counter`
	gaugeType       = `gauge`
	poolCounterName = `PollCount`
)

func reportMetric(ms *MetricsStorage) {
	sendGauge(ms)
	sendCounter(ms)
}

func sendGauge(ms *MetricsStorage) {
	for k, v := range ms.Metrics {
		m := internal.Metrics{
			ID:    k,
			MType: gaugeType,
			Value: &v,
		}
		sendRequest(m)
	}
}

func sendCounter(ms *MetricsStorage) {
	m := internal.Metrics{
		ID:    poolCounterName,
		MType: counterType,
		Delta: &ms.PollCount,
	}
	sendRequest(m)
}

func sendRequest(metrics internal.Metrics) {
	jsonData, err := json.Marshal(metrics)
	if err != nil {
		logger.Infoln("marshall error", err)
	}

	client := resty.New()
	_, err = client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetBody(getCompressedData(jsonData)).
		Post("http://" + Config.addr + URL)

	if err != nil {
		logger.Infoln("error in request", err)
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
