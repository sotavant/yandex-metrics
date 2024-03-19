package server

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/sotavant/yandex-metrics/internal/server/config"
	"github.com/sotavant/yandex-metrics/internal/server/repository"
	"github.com/sotavant/yandex-metrics/internal/server/repository/memory"
	"github.com/sotavant/yandex-metrics/internal/server/repository/postgres"
	"github.com/sotavant/yandex-metrics/internal/server/storage"
)

type App struct {
	Config  *config.Config
	Storage repository.Storage
	Fs      *storage.FileStorage
	DbConn  *pgx.Conn
}

func InitApp(ctx context.Context) (*App, error) {
	var err error

	conf := config.InitConfig()
	dbConn, err := postgres.InitDB(ctx, conf.DatabaseDSN)
	if err != nil {
		panic(err)
	}
	appInstance := new(App)

	if dbConn == nil {
		appInstance.Storage = memory.NewMetricsRepository()
		appInstance.Fs, err = storage.NewFileStorage(conf.FileStoragePath, conf.Restore, conf.StoreInterval)

		if err != nil {
			panic(err)
		}

		if err = appInstance.Fs.Restore(ctx, appInstance.Storage); err != nil {
			panic(err)
		}
	} else {
		appInstance.Storage, err = postgres.NewMemStorage(ctx, dbConn, conf.TableName, conf.DatabaseDSN)

		if err != nil {
			panic(err)
		}
	}

	appInstance.Config = conf
	appInstance.DbConn = dbConn

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
