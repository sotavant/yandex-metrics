package repository

import (
	"context"
	"github.com/sotavant/yandex-metrics/internal"
)

type Storage interface {
	AddGaugeValue(ctx context.Context, key string, value float64) error
	AddCounterValue(ctx context.Context, key string, value int64) error
	GetValue(ctx context.Context, mType string, key string) (interface{}, error)
	GetGauge(ctx context.Context) (map[string]float64, error)
	GetCounters(ctx context.Context) (map[string]int64, error)
	GetCounterValue(ctx context.Context, key string) (int64, error)
	GetGaugeValue(ctx context.Context, key string) (float64, error)
	KeyExist(ctx context.Context, mType string, key string) (bool, error)
	AddValue(ctx context.Context, m internal.Metrics) error
	AddValues(ctx context.Context, m []internal.Metrics) error
	GetValues(ctx context.Context) ([]internal.Metrics, error)
}
