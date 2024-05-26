package storage

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sotavant/yandex-metrics/internal"
	"github.com/sotavant/yandex-metrics/internal/utils"
)

func InitDB(ctx context.Context, DSN string) (*pgxpool.Pool, error) {
	if DSN == "" {
		return nil, nil
	}

	config, err := pgxpool.ParseConfig(DSN)
	if err != nil {
		internal.Logger.Infow("bad database config", "err", err)
		return nil, err
	}

	dbConn, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		internal.Logger.Infow("Unable to connect to database", "err", err)
	}

	return dbConn, nil
}

func CheckConnection(ctx context.Context, pool *pgxpool.Pool) bool {
	intervals := utils.GetRetryWaitTimes()
	retries := len(intervals) + 1
	counter := 1

	for counter <= retries {
		err := pool.Ping(ctx)
		if err != nil {
			internal.Logger.Infow("attempt connect err", "err", err)
			counter++
			time.Sleep(time.Duration(intervals[counter-1]) * time.Second)
			continue
		}

		return true
	}

	return false
}
