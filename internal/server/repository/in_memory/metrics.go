package in_memory

import (
	"context"
	"fmt"
	"github.com/sotavant/yandex-metrics/internal"
)

type MetricsRepository struct {
	Gauge   map[string]float64
	Counter map[string]int64
}

func (m *MetricsRepository) AddGaugeValue(ctx context.Context, key string, value float64) error {
	m.Gauge[key] = value

	return nil
}

func (m *MetricsRepository) AddCounterValue(ctx context.Context, key string, value int64) error {
	m.Counter[key] += value

	return nil
}

func (m *MetricsRepository) AddValue(ctx context.Context, metric internal.Metrics) error {
	var err error

	switch metric.MType {
	case internal.GaugeType:
		err = m.AddGaugeValue(ctx, metric.ID, *metric.Value)
	case internal.CounterType:
		err = m.AddCounterValue(ctx, metric.ID, *metric.Delta)
	default:
		err = fmt.Errorf("undefinde type: %s", metric.MType)
	}

	return err
}

func (m *MetricsRepository) GetValue(ctx context.Context, mType, key string) (interface{}, error) {
	switch mType {
	case internal.GaugeType:
		val, ok := m.Gauge[key]
		if ok {
			return val, nil
		}
	case internal.CounterType:
		val, ok := m.Counter[key]
		if ok {
			return val, nil
		}
	}

	return nil, nil
}

func (m *MetricsRepository) KeyExist(ctx context.Context, mType, key string) (bool, error) {
	switch mType {
	case internal.GaugeType:
		_, ok := m.Gauge[key]
		if ok {
			return true, nil
		}
	case internal.CounterType:
		_, ok := m.Counter[key]
		if ok {
			return true, nil
		}
	}

	return false, nil
}

func (m *MetricsRepository) GetGauge(ctx context.Context) map[string]float64 {
	return m.Gauge
}

func (m *MetricsRepository) GetGaugeValue(ctx context.Context, key string) float64 {
	return m.Gauge[key]
}

func (m *MetricsRepository) GetCounters(ctx context.Context) map[string]int64 {
	return m.Counter
}

func (m *MetricsRepository) GetCounterValue(ctx context.Context, key string) int64 {
	return m.Counter[key]
}

func NewMetricsRepository() *MetricsRepository {
	var m MetricsRepository
	m.Gauge = make(map[string]float64)
	m.Counter = make(map[string]int64)

	return &m
}
