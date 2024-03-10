package main

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/sotavant/yandex-metrics/internal"
)

type app struct {
	config     *config
	memStorage *MemStorage
	fs         *FileStorage
	dbConn     *pgx.Conn
}

func initApp(ctx context.Context) (*app, error) {
	conf := new(config)
	conf.parseFlags()

	mem := NewMemStorage()
	fs, err := NewFileStorage(*conf)

	if err != nil {
		panic(err)
	}

	if err = fs.Restore(mem); err != nil {
		panic(err)
	}

	dbConn, _ := initDB(ctx, *conf)

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

func initDB(ctx context.Context, conf config) (*pgx.Conn, error) {
	if conf.databaseDSN == "" {
		return nil, nil
	}

	dbConn, err := pgx.Connect(ctx, conf.databaseDSN)
	if err != nil {
		internal.Logger.Infow("Unable to connect to database", "err", err)
		return nil, err
	}

	return dbConn, nil
}

func (app *app) initRouters(ctx context.Context) *chi.Mux {
	r := chi.NewRouter()

	r.Post("/update/{type}/{name}/{value}", withLogging(gzipMiddleware(updateHandler(app))))
	r.Get("/value/{type}/{name}", withLogging(gzipMiddleware(getValueHandler(app))))
	r.Post("/update/", withLogging(gzipMiddleware(updateJSONHandler(app))))
	r.Post("/value/", withLogging(gzipMiddleware(getValueJSONHandler(app))))
	r.Get("/", withLogging(gzipMiddleware(getValuesHandler(app))))
	r.Get("/ping", withLogging(gzipMiddleware(pingDBHandler(ctx, app.dbConn))))

	return r
}
