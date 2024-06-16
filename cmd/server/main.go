package main

import (
	"context"
	"net/http"
	"net/http/pprof"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/sotavant/yandex-metrics/internal"
	"github.com/sotavant/yandex-metrics/internal/server"
	"github.com/sotavant/yandex-metrics/internal/server/handlers"
	"github.com/sotavant/yandex-metrics/internal/server/middleware"
)

// Build info.
// Need define throw ldflags:
//
//	go build -ldflags "-X main.buildVersion=0.1 -X 'main.buildDate=$(date +'%Y/%m/%d')' -X 'main.buildCommit=$(git rev-parse --short HEAD)'"
var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	internal.PrintBuildInfo(buildVersion, buildDate, buildCommit)
	const chanCount = 2
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
	wg.Add(chanCount)

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

	hasher := middleware.NewHasher(app.Config.HashKey)
	crypto, err := middleware.NewCrypto(app.Config.CryptoKeyPath)
	if err != nil {
		internal.Logger.Infow("crypto initialization failed", "error", err)
	}

	r.Use(crypto.Handler)
	r.Use(hasher.Handler)
	r.Use(middleware.GzipMiddleware)
	r.Use(middleware.WithLogging)

	r.Post("/update/{type}/{name}/{value}", handlers.UpdateHandler(app))
	r.Get("/value/{type}/{name}", handlers.GetValueHandler(app))
	r.Post("/update/", handlers.UpdateJSONHandler(app))
	r.Post("/updates/", handlers.UpdateBatchJSONHandler(app))
	r.Post("/value/", handlers.GetValueJSONHandler(app))
	r.Get("/", handlers.GetValuesHandler(app))
	r.Get("/ping", handlers.PingDBHandler(app.DBConn))

	initProfiling(r)

	return r
}

func initProfiling(r *chi.Mux) {
	r.HandleFunc("/pprof/*", pprof.Index)
	r.Handle("/pprof/heap", pprof.Handler("heap"))
}
