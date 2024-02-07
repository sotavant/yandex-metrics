package main

import (
	"github.com/sotavant/yandex-metrics/internal"
	"net/http"
)

func counterHandler(res http.ResponseWriter, req *http.Request) {
	if !internal.RequestCheck(res, req, contentType) {
		return
	}

	res.WriteHeader(http.StatusOK)
}
