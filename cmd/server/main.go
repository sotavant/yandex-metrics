package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/sotavant/yandex-metrics/internal"
	"go.uber.org/zap"
	"net/http"
	"sync"
)

const (
	gaugeType          = "gauge"
	counterType        = "counter"
	serverAddress      = "localhost:8080"
	acceptableEncoding = "gzip"
)

var logger zap.SugaredLogger

func main() {
	config := new(config)
	config.parseFlags()

	r := chi.NewRouter()
	mem := NewMemStorage()
	fs, err := NewFileStorage(*config)
	defer func(fs *FileStorage) {
		err := fs.Sync(mem)
		if err != nil {
			panic(err)
		}

		err = fs.file.Close()
		if err != nil {
			panic(err)
		}
	}(fs)

	if err != nil {
		panic(err)
	}

	if err = fs.Restore(mem); err != nil {
		panic(err)
	}

	logger = internal.InitLogger()

	r.Post("/update/{type}/{name}/{value}", withLogging(gzipMiddleware(updateHandler(mem, fs))))
	r.Get("/value/{type}/{name}", withLogging(gzipMiddleware(getValueHandler(mem))))
	r.Post("/update/", withLogging(gzipMiddleware(updateJSONHandler(mem, fs))))
	r.Post("/value/", withLogging(gzipMiddleware(getValueJSONHandler(mem))))
	r.Get("/", withLogging(gzipMiddleware(getValuesHandler(mem))))

	httpChan := make(chan bool)
	syncChan := make(chan bool)
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		err = http.ListenAndServe(config.addr, r)
		if err != nil {
			close(httpChan)
			panic(err)
		}
	}()

	go func() {
		if err = fs.SyncByInterval(mem, syncChan); err != nil {
			panic(err)
		}
	}()

	wg.Wait()
}
