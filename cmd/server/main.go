package main

import (
	"github.com/go-chi/chi/v5"
	"net/http"
)

const (
	// const contentType string = `text/plain`
	gaugeType     = "gauge"
	counterType   = "counter"
	serverAddress = "localhost:8080"
)

func main() {
	r := chi.NewRouter()
	mem := NewMemStorage()
	config := new(config)

	config.parseFlags()

	r.Post("/update/{type}/{name}/{value}", updateHandler(mem))
	r.Get("/value/{type}/{name}", getValueHandler(mem))
	r.Get("/", getValuesHandler(mem))

	err := http.ListenAndServe(config.addr, r)
	if err != nil {
		panic(err)
	}
}
