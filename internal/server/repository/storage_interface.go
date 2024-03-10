package repository

import (
	"context"
	"github.com/sotavant/yandex-metrics/internal"
)

type Storage interface {
	AddGaugeValue(ctx context.Context, key string, value float64) error
	AddCounterValue(ctx context.Context, key string, value int64) error
	GetValue(ctx context.Context, mType string, key string) interface{}
	GetGauge(ctx context.Context) map[string]float64
	GetCounters(ctx context.Context) map[string]int64
	GetCounterValue(ctx context.Context, key string) int64
	GetGaugeValue(ctx context.Context, key string) float64
	KeyExist(ctx context.Context, mType string, key string) bool
	AddValue(ctx context.Context, m internal.Metrics) error
}
