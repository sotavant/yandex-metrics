package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/sotavant/yandex-metrics/internal"
	"go.uber.org/zap"
	"net/http"
)

const (
	gaugeType          = "gauge"
	counterType        = "counter"
	serverAddress      = "localhost:8080"
	acceptableEncoding = "gzip"
)

var logger zap.SugaredLogger

func main() {
	r := chi.NewRouter()
	mem := NewMemStorage()
	config := new(config)
	logger = internal.InitLogger()

	config.parseFlags()

	r.Post("/update/{type}/{name}/{value}", withLogging(gzipMiddleware(updateHandler(mem))))
	r.Get("/value/{type}/{name}", withLogging(gzipMiddleware(getValueHandler(mem))))
	r.Post("/update/", withLogging(gzipMiddleware(updateJSONHandler(mem))))
	r.Post("/value/", withLogging(gzipMiddleware(getValueJSONHandler(mem))))
	r.Get("/", withLogging(gzipMiddleware(getValuesHandler(mem))))

	err := http.ListenAndServe(config.addr, r)
	if err != nil {
		panic(err)
	}
}
