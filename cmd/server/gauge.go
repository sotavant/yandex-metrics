package main

import (
	"github.com/sotavant/yandex-metrics/internal"
	"net/http"
	"strconv"
	"strings"
)

func handleGauge(storage internal.Storage) func(res http.ResponseWriter, req *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		if !internal.RequestCheck(res, req, contentType) {
			return
		}

		key, value := internal.ParseUrl(req.RequestURI)
		floatVal, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
		if err != nil {
			http.Error(res, "bad request", http.StatusBadRequest)
			return
		}
		storage.AddGaugeValue(key, floatVal)

		res.WriteHeader(http.StatusOK)
	}
}
