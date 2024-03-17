package main

import (
	"context"
	"github.com/sotavant/yandex-metrics/internal"
	"github.com/sotavant/yandex-metrics/internal/server"
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

	r := appInstance.InitRouters()

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
		if err = appInstance.Fs.SyncByInterval(ctx, appInstance, syncChan); err != nil {
			panic(err)
		}
	}()

	wg.Wait()
}
