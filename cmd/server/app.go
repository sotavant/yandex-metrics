package main

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/sotavant/yandex-metrics/internal"
	"github.com/sotavant/yandex-metrics/internal/server/repository/in_memory"
)

type app struct {
	config     *config
	memStorage *in_memory.MetricsRepository
	fs         *FileStorage
	dbConn     *pgx.Conn
}

func initApp(ctx context.Context) (*app, error) {
	conf := new(config)
	conf.parseFlags()

	mem := in_memory.NewMetricsRepository()
	fs, err := NewFileStorage(*conf)

	if err != nil {
		panic(err)
	}

	if err = fs.Restore(mem); err != nil {
		panic(err)
	}

	dbConn := initDB(ctx, *conf)

	return &app{
		config:     conf,
		memStorage: mem,
		fs:         fs,
		dbConn:     dbConn,
	}, nil
}

func (app *app) syncFs() {
	err := app.fs.Sync(app.memStorage)
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
