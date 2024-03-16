package main

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/sotavant/yandex-metrics/internal"
	"github.com/sotavant/yandex-metrics/internal/server/repository"
	"github.com/sotavant/yandex-metrics/internal/server/repository/memory"
	"github.com/sotavant/yandex-metrics/internal/server/repository/postgres"
)

type app struct {
	config  *config
	storage repository.Storage
	fs      *FileStorage
	dbConn  *pgx.Conn
}

func initApp(ctx context.Context) (*app, error) {
	var err error

	conf := new(config)
	conf.parseFlags()
	dbConn := initDB(ctx, *conf)
	appInstance := new(app)

	if dbConn == nil {
		appInstance.storage = memory.NewMetricsRepository()
		appInstance.fs, err = NewFileStorage(*conf)

		if err != nil {
			panic(err)
		}

		if err = appInstance.fs.Restore(ctx, appInstance.storage); err != nil {
			panic(err)
		}
	} else {
		appInstance.storage, err = postgres.NewMemStorage(ctx, dbConn, conf.tableName)

		if err != nil {
			panic(err)
		}
	}

	appInstance.config = conf
	appInstance.dbConn = dbConn

	return appInstance, nil
}

func (app *app) syncFs(ctx context.Context) {
	needSync := false

	switch app.storage.(type) {
	case *memory.MetricsRepository:
		needSync = true
	}

	if !needSync {
		return
	}

	err := app.fs.Sync(ctx, app.storage)
	if err != nil {
		panic(err)
	}

	err = app.fs.file.Close()
	if err != nil {
		panic(err)
	}
}

func initDB(ctx context.Context, conf config) *pgx.Conn {
	if conf.databaseDSN == "" {
		return nil
	}

	dbConn, err := pgx.Connect(ctx, conf.databaseDSN)
	if err != nil {
		internal.Logger.Panicw("Unable to connect to database", "err", err)
	}

	return dbConn
}

func (app *app) initRouters() *chi.Mux {
	r := chi.NewRouter()

	r.Post("/update/{type}/{name}/{value}", withLogging(gzipMiddleware(updateHandler(app))))
	r.Get("/value/{type}/{name}", withLogging(gzipMiddleware(getValueHandler(app))))
	r.Post("/update/", withLogging(gzipMiddleware(updateJSONHandler(app))))
	r.Post("/value/", withLogging(gzipMiddleware(getValueJSONHandler(app))))
	r.Get("/", withLogging(gzipMiddleware(getValuesHandler(app))))
	r.Get("/ping", withLogging(gzipMiddleware(pingDBHandler(app.dbConn))))

	return r
}
