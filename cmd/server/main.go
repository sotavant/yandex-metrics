package main

import (
	"github.com/go-chi/chi/v5"
	"net/http"
)

// const contentType string = `text/plain`
const gaugeType = "gauge"
const counterType = "counter"

func main() {
	r := chi.NewRouter()
	mem := NewMemStorage()

	r.Post("/update/{type}/{name}/{value}", updateHandler(mem))
	r.Get("/value/{type}/{name}", getValueHandler(mem))
	r.Get("/", getValuesHandler(mem))

	err := http.ListenAndServe(`:8080`, r)
	if err != nil {
		panic(err)
	}
}
