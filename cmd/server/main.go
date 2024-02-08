package main

import (
	"net/http"
)

var contentType string = `text/plain`

func badTypeHandler(res http.ResponseWriter, req *http.Request) {
	http.Error(res, "bad request", http.StatusBadRequest)
}

func defaultHandler(res http.ResponseWriter, req *http.Request) {
	http.Error(res, "not found", http.StatusNotFound)
}

func main() {
	mux := http.NewServeMux()
	mem := NewMemStorage()

	mux.HandleFunc("/update/gauge/", handleGauge(mem))
	mux.HandleFunc("/update/counter/", handleCounter(mem))
	mux.HandleFunc(`/update/`, badTypeHandler)
	mux.HandleFunc(`/`, defaultHandler)

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
