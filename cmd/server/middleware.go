package main

import (
	"github.com/sotavant/yandex-metrics/internal/server"
	"net/http"
)

func withLogging(h http.HandlerFunc) http.HandlerFunc {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		//start := time.Now()
		//uri := r.RequestURI
		//method := r.Method
		rData := &server.ResponseData{
			Status: 0,
			Size:   0,
		}
		lw := server.LoggingResponseWriter{
			ResponseWriter: w,
			ResponseData:   rData,
		}

		h.ServeHTTP(&lw, r)

		//duration := time.Since(start)

		/*logger.Infow(
			"Request info",
			"uri", uri,
			"method", method,
			"duration", duration,
		)*/

		/*		logger.Infow(
				"Response info",
				"status", rData.Status,
				"size", rData.Size,
			)*/
	}

	return logFn
}
