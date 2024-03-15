package main

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/sotavant/yandex-metrics/internal"
	"github.com/sotavant/yandex-metrics/internal/server/repository/in_memory"
	"github.com/sotavant/yandex-metrics/internal/server/repository/postgres"
)

type app struct {
	config     *config
	memStorage *in_memory.MetricsRepository
	dbStorage  *postgres.MetricsRepository
	fs         *FileStorage
	dbConn     *pgx.Conn
}

func initApp(ctx context.Context) (*app, error) {
	var err error

	conf := new(config)
	conf.parseFlags()
	dbConn := initDB(ctx, *conf)
	appInstance := new(app)

	if dbConn == nil {
		appInstance.memStorage = in_memory.NewMetricsRepository()
		appInstance.fs, err = NewFileStorage(*conf)

		if err != nil {
			panic(err)
		}

		if err = appInstance.fs.Restore(ctx, appInstance.memStorage); err != nil {
			panic(err)
		}
	} else {
		appInstance.dbStorage, err = postgres.NewMemStorage(ctx, dbConn, conf.tableName)

		if err != nil {
			panic(err)
		}
	}

	return appInstance, nil
}

func (app *app) syncFs(ctx context.Context) {
	if app.memStorage == nil {
		return
	}

	err := app.fs.Sync(ctx, app.memStorage)
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
