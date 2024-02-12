package main

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"strconv"
)

const (
	URL             = `%s/update/%s/%s/%s`
	counterType     = `counter`
	gaugeType       = `gauge`
	poolCounterName = `poolCounter`
)

func reportMetric(ms *MetricsStorage) {
	sendGauge(ms)
	sendCounter(ms)
}

func getURL(mType, name, value string) string {
	return fmt.Sprintf(URL, serverAddress, mType, name, value)
}

func sendGauge(ms *MetricsStorage) {
	for k, v := range ms.Metrics {
		sendRequest(gaugeType, k, fmt.Sprintf(`%f`, v))
	}
}

func sendCounter(ms *MetricsStorage) {
	sendRequest(counterType, poolCounterName, strconv.FormatInt(ms.PollCount, 10))
}

func sendRequest(mType, name, value string) {
	url := getURL(mType, name, value)
	client := resty.New()
	_, err := client.R().Post(url)

	if err != nil {
		fmt.Println(fmt.Errorf("error in %s request: %v", mType, err))
	}
}
