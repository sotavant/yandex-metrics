package main

import (
	"github.com/sotavant/yandex-metrics/internal/server"
	"net/http"
	"strconv"
	"strings"
)

func handleCounter(storage Storage) func(res http.ResponseWriter, req *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		if !server.RequestCheck(res, req, contentType) {
			return
		}

		key, value := server.ParseURL(req.RequestURI)
		intVal, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
		if err != nil {
			http.Error(res, "bad request", http.StatusBadRequest)
			return
		}
		storage.AddCounterValue(key, intVal)

		res.WriteHeader(http.StatusOK)
	}
}
