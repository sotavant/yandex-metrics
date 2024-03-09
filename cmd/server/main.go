package main

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/sotavant/yandex-metrics/internal"
	"net/http"
	"sync"
)

const (
	gaugeType          = "gauge"
	counterType        = "counter"
	serverAddress      = "localhost:8080"
	acceptableEncoding = "gzip"
)

func main() {
	internal.InitLogger()
	conf := initConfig()
	ctx := context.Background()
	dbConn, _ := initDB(ctx, *conf)
	mem := NewMemStorage()
	fs, err := NewFileStorage(*conf)

	if dbConn != nil {
		defer func(dbConn *pgx.Conn, ctx context.Context) {
			err := dbConn.Close(ctx)
			if err != nil {
				panic("error in close dbConn")
			}
		}(dbConn, ctx)
	}

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

	r := initRouter(ctx, mem, fs, dbConn)

	httpChan := make(chan bool)
	syncChan := make(chan bool)
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		err = http.ListenAndServe(conf.addr, r)
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

func initRouter(ctx context.Context, mem Storage, fs *FileStorage, dbConn *pgx.Conn) *chi.Mux {
	r := chi.NewRouter()

	r.Post("/update/{type}/{name}/{value}", withLogging(gzipMiddleware(updateHandler(mem, fs))))
	r.Get("/value/{type}/{name}", withLogging(gzipMiddleware(getValueHandler(mem))))
	r.Post("/update/", withLogging(gzipMiddleware(updateJSONHandler(mem, fs))))
	r.Post("/value/", withLogging(gzipMiddleware(getValueJSONHandler(mem))))
	r.Get("/", withLogging(gzipMiddleware(getValuesHandler(mem))))
	r.Get("/ping", withLogging(gzipMiddleware(pingDBHandler(ctx, dbConn))))

	return r
}

func initDB(ctx context.Context, conf config) (*pgx.Conn, error) {
	if conf.databaseDSN == "" {
		return nil, nil
	}

	internal.Logger.Infow("databaseDSN", "dsn", conf.databaseDSN)
	dbConn, err := pgx.Connect(ctx, conf.databaseDSN)
	if err != nil {
		internal.Logger.Infow("Unable to connect to database", "err", err)
		return nil, err
	}

	return dbConn, nil
}
