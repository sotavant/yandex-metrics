package main

import "net/http"

var contentType string = `text/plain`

func badTypeHandler(res http.ResponseWriter, req *http.Request) {
	http.Error(res, "bad request", http.StatusBadRequest)
}

func defaultHandler(res http.ResponseWriter, req *http.Request) {
	http.Error(res, "not found", http.StatusNotFound)
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/update/gauge/", gaugeHandler)
	mux.HandleFunc("/update/counter/", counterHandler)
	mux.HandleFunc(`/update/`, badTypeHandler)
	mux.HandleFunc(`/`, defaultHandler)

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
