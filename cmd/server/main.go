package main

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/sotavant/yandex-metrics/internal"
	"github.com/sotavant/yandex-metrics/internal/server"
	"github.com/sotavant/yandex-metrics/internal/server/handlers"
	"github.com/sotavant/yandex-metrics/internal/server/midleware"
	"net/http"
	"sync"
)

func main() {
	ctx := context.Background()
	internal.InitLogger()

	appInstance, err := server.InitApp(ctx)
	if err != nil {
		panic(err)
	}
	defer appInstance.SyncFs(ctx)

	r := initRouters(appInstance)

	httpChan := make(chan bool)
	syncChan := make(chan bool)
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		err = http.ListenAndServe(appInstance.Config.Addr, r)

		if err != nil {
			close(httpChan)
			panic(err)
		}
	}()

	go func() {
		if appInstance.Fs == nil {
			close(syncChan)
			return
		}
		if err = appInstance.Fs.SyncByInterval(ctx, appInstance.Storage, syncChan); err != nil {
			panic(err)
		}
	}()

	wg.Wait()
}

func initRouters(app *server.App) *chi.Mux {
	r := chi.NewRouter()

	hasher := midleware.NewHasher(app.Config.HashKey)

	r.Use(hasher.Handler)
	r.Use(midleware.GzipMiddleware)
	r.Use(midleware.WithLogging)

	r.Post("/update/{type}/{name}/{value}", handlers.UpdateHandler(app))
	r.Get("/value/{type}/{name}", handlers.GetValueHandler(app))
	r.Post("/update/", handlers.UpdateJSONHandler(app))
	r.Post("/updates/", handlers.UpdateBatchJSONHandler(app))
	r.Post("/value/", handlers.GetValueJSONHandler(app))
	r.Get("/", handlers.GetValuesHandler(app))
	r.Get("/ping", handlers.PingDBHandler(app.DBConn))

	return r
}
