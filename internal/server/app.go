// Package server Пакет для получения и хранения метрик
package server

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sotavant/yandex-metrics/internal/server/config"
	"github.com/sotavant/yandex-metrics/internal/server/repository"
	"github.com/sotavant/yandex-metrics/internal/server/repository/memory"
	"github.com/sotavant/yandex-metrics/internal/server/repository/postgres"
	"github.com/sotavant/yandex-metrics/internal/server/storage"
)

// App структура для хранения текущего состояния, конфигурация, соединения с базой данных
type App struct {
	Config  *config.Config
	Storage repository.Storage
	Fs      *storage.FileStorage
	DBConn  *pgxpool.Pool
}

// InitApp Инициализация приложения
func InitApp(ctx context.Context) (*App, error) {
	var err error

	conf := config.InitConfig()
	dbConn, err := storage.InitDB(ctx, conf.DatabaseDSN)
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
	appInstance.DBConn = dbConn

	return appInstance, nil
}

// SyncFs Метод для синхронизация значения в памяти и в файле. В том случае, если используется in-memory хранилище
func (app *App) SyncFs(ctx context.Context) {
	fmt.Println("syncing fs")
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
	if err != nil && !errors.Is(err, os.ErrClosed) {
		panic(err)
	}
}
