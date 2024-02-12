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

func reportMetric(ms *MetricsStorage, conf config) {
	sendGauge(ms, conf)
	sendCounter(ms, conf)
}

func getURL(mType, name, value string, conf config) string {
	return fmt.Sprintf(URL, conf.addr, mType, name, value)
}

func sendGauge(ms *MetricsStorage, conf config) {
	for k, v := range ms.Metrics {
		sendRequest(gaugeType, k, fmt.Sprintf(`%f`, v), conf)
	}
}

func sendCounter(ms *MetricsStorage, conf config) {
	sendRequest(counterType, poolCounterName, strconv.FormatInt(ms.PollCount, 10), conf)
}

func sendRequest(mType, name, value string, conf config) {
	url := getURL(mType, name, value, conf)
	client := resty.New()
	_, err := client.R().Post("http://" + url)

	if err != nil {
		fmt.Println(fmt.Errorf("error in %s request: %v", mType, err))
	}
}
