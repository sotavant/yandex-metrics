package main

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/sotavant/yandex-metrics/internal"
	"net/http"
	"sync"
)

const (
	counterType        = internal.CounterType
	serverAddress      = "localhost:8080"
	acceptableEncoding = "gzip"
	tableName          = "metric"
)

func main() {
	ctx := context.Background()
	internal.InitLogger()

	appInstance, err := initApp(ctx)
	if err != nil {
		panic(err)
	}
	defer appInstance.syncFs(ctx)

	if appInstance.dbConn != nil {
		defer func(dbConn *pgx.Conn, ctx context.Context) {
			err := dbConn.Close(ctx)
			if err != nil {
				panic("error in close dbConn")
			}
		}(appInstance.dbConn, ctx)
	}

	r := appInstance.initRouters()

	httpChan := make(chan bool)
	syncChan := make(chan bool)
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		err = http.ListenAndServe(appInstance.config.addr, r)
		if err != nil {
			close(httpChan)
			panic(err)
		}
	}()

	go func() {
		if err = appInstance.fs.SyncByInterval(ctx, appInstance, syncChan); err != nil {
			panic(err)
		}
	}()

	wg.Wait()
}
