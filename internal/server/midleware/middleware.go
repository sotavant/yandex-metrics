package midleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/sotavant/yandex-metrics/internal"
	"github.com/sotavant/yandex-metrics/internal/server"
)

const AcceptableEncoding = "gzip"

func getTypesForEncoding() [3]string {
	return [3]string{
		"html/text",
		"text/html",
		"application/json",
	}
}

func WithLogging(h http.Handler) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		uri := r.RequestURI
		method := r.Method
		rData := &server.ResponseData{
			Status: 0,
			Size:   0,
		}
		lw := server.LoggingResponseWriter{
			ResponseWriter: w,
			ResponseData:   rData,
		}

		h.ServeHTTP(&lw, r)

		duration := time.Since(start)

		internal.Logger.Infow(
			"Request info",
			"uri", uri,
			"method", method,
			"duration", duration,
		)

		internal.Logger.Infow(
			"Response info",
			"status", rData.Status,
			"size", rData.Size,
		)
	}

	return http.HandlerFunc(logFn)
}

func GzipMiddleware(h http.Handler) http.Handler {
	gzipFn := func(w http.ResponseWriter, r *http.Request) {
		ow := w
		if isNeedEncoding(r) {
			contentEncoding := r.Header.Get("Content-Encoding")
			sendsGzip := strings.Contains(contentEncoding, AcceptableEncoding)
			if sendsGzip {
				cr, err := server.NewCompressReader(r.Body)
				if err != nil {
					internal.Logger.Infow("compressReaderError", "err", err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				r.Body = cr
				defer func(cr *server.CompressReader) {
					err = cr.Close()
					if err != nil {
						internal.Logger.Infow("compressReaderCloseError", "err", err)
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
				}(cr)
			}

			acceptEncoding := r.Header.Get("Accept-Encoding")
			supportGzip := strings.Contains(acceptEncoding, "gzip")

			if supportGzip {
				cw := server.NewCompressWriter(w)
				cw.Header().Set("Content-Encoding", "gzip")
				ow = cw
				defer func(cw *server.CompressWriter) {
					err := cw.Close()
					if err != nil {
						internal.Logger.Infow("compressWriterCloseError", "err", err)
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
				}(cw)
			}
		}

		h.ServeHTTP(ow, r)
	}

	return http.HandlerFunc(gzipFn)
}

func isNeedEncoding(r *http.Request) bool {
	contentType := r.Header.Get("Accept")
	for _, t := range getTypesForEncoding() {
		if strings.Contains(contentType, t) {
			return true
		}
	}

	return false
}
