package main

import (
	"context"
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
	ctx := context.Background()
	appInstance, err := initApp(ctx)
	if err != nil {
		panic(err)
	}
	defer appInstance.syncFs()

	if appInstance.dbConn != nil {
		defer func(dbConn *pgx.Conn, ctx context.Context) {
			err := dbConn.Close(ctx)
			if err != nil {
				panic("error in close dbConn")
			}
		}(appInstance.dbConn, ctx)
	}

	internal.InitLogger()
	r := appInstance.initRouters(ctx)

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
		if err = appInstance.fs.SyncByInterval(appInstance.memStorage, syncChan); err != nil {
			panic(err)
		}
	}()

	wg.Wait()
}
