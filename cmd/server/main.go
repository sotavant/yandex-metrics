package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/sotavant/yandex-metrics/internal/server"
	"go.uber.org/zap"
	"net/http"
)

const (
	// const contentType string = `text/plain`
	gaugeType     = "gauge"
	counterType   = "counter"
	serverAddress = "localhost:8080"
)

var logger zap.SugaredLogger

func main() {
	r := chi.NewRouter()
	mem := NewMemStorage()
	config := new(config)
	logger = server.InitLogger()

	config.parseFlags()

	r.Post("/update/{type}/{name}/{value}", withLogging(updateHandler(mem)))
	r.Get("/value/{type}/{name}", withLogging(getValueHandler(mem)))
	r.Get("/", withLogging(getValuesHandler(mem)))

	err := http.ListenAndServe(config.addr, r)
	if err != nil {
		panic(err)
	}
}
