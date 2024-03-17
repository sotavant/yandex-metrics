package server

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/sotavant/yandex-metrics/internal"
	"github.com/sotavant/yandex-metrics/internal/server/handlers"
	"github.com/sotavant/yandex-metrics/internal/server/midleware"
	"github.com/sotavant/yandex-metrics/internal/server/repository"
	"github.com/sotavant/yandex-metrics/internal/server/repository/memory"
	"github.com/sotavant/yandex-metrics/internal/server/repository/postgres"
	"github.com/sotavant/yandex-metrics/internal/server/storage"
)

type App struct {
	Config  *Config
	Storage repository.Storage
	Fs      *storage.FileStorage
	dbConn  *pgx.Conn
}

func InitApp(ctx context.Context) (*App, error) {
	var err error

	conf := InitConfig()
	dbConn := InitDB(ctx, *conf)
	appInstance := new(App)

	if dbConn == nil {
		appInstance.Storage = memory.NewMetricsRepository()
		appInstance.Fs, err = storage.NewFileStorage(*conf)

		if err != nil {
			panic(err)
		}

		if err = appInstance.Fs.Restore(ctx, appInstance.Storage); err != nil {
			panic(err)
		}
	} else {
		appInstance.Storage, err = postgres.NewMemStorage(ctx, dbConn, conf.TableName)

		if err != nil {
			panic(err)
		}
	}

	appInstance.Config = conf
	appInstance.dbConn = dbConn

	return appInstance, nil
}

func (app *App) SyncFs(ctx context.Context) {
	needSync := false

	switch app.Storage.(type) {
	case *memory.MetricsRepository:
		needSync = true
	}

	if !needSync {
		return
	}

	err := app.Fs.Sync(ctx, app.Storage)
	if err != nil {
		panic(err)
	}

	err = app.Fs.File.Close()
	if err != nil {
		panic(err)
	}
}

func InitDB(ctx context.Context, conf Config) *pgx.Conn {
	if conf.DatabaseDSN == "" {
		return nil
	}

	dbConn, err := pgx.Connect(ctx, conf.DatabaseDSN)
	if err != nil {
		internal.Logger.Panicw("Unable to connect to database", "err", err)
	}

	return dbConn
}

func (app *App) InitRouters() *chi.Mux {
	r := chi.NewRouter()

	r.Post("/update/{type}/{name}/{value}", midleware.WithLogging(midleware.GzipMiddleware(handlers.UpdateHandler(app))))
	r.Get("/value/{type}/{name}", midleware.WithLogging(midleware.GzipMiddleware(handlers.GetValueHandler(app))))
	r.Post("/update/", midleware.WithLogging(midleware.GzipMiddleware(handlers.UpdateJSONHandler(app))))
	r.Post("/updates/", midleware.WithLogging(midleware.GzipMiddleware(handlers.UpdateBatchJSONHandler(app))))
	r.Post("/value/", midleware.WithLogging(midleware.GzipMiddleware(handlers.GetValueJSONHandler(app))))
	r.Get("/", midleware.WithLogging(midleware.GzipMiddleware(handlers.GetValuesHandler(app))))
	r.Get("/ping", midleware.WithLogging(midleware.GzipMiddleware(handlers.PingDBHandler(app.dbConn))))

	return r
}
