package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/sotavant/yandex-metrics/internal"
)

type MetricsRepository struct {
	Gauge   map[string]float64
	Counter map[string]int64
	mutex   sync.RWMutex
}

func (m *MetricsRepository) AddGaugeValue(ctx context.Context, key string, value float64) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.Gauge[key] = value

	return nil
}

func (m *MetricsRepository) AddCounterValue(ctx context.Context, key string, value int64) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
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

func (m *MetricsRepository) AddValues(ctx context.Context, metrics []internal.Metrics) error {
	for _, v := range metrics {
		err := m.AddValue(ctx, v)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *MetricsRepository) GetValue(ctx context.Context, mType, key string) (interface{}, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

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

func (m *MetricsRepository) GetValues(ctx context.Context) ([]internal.Metrics, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	metrics := make([]internal.Metrics, 0, len(m.Gauge)+len(m.Counter))

	for k, v := range m.Gauge {
		metrics = append(metrics, internal.Metrics{
			ID:    k,
			MType: internal.GaugeType,
			Delta: nil,
			Value: &v,
		})
	}

	for k, v := range m.Counter {
		metrics = append(metrics, internal.Metrics{
			ID:    k,
			MType: internal.CounterType,
			Delta: &v,
			Value: nil,
		})
	}

	return metrics, nil
}

func (m *MetricsRepository) KeyExist(ctx context.Context, mType, key string) (bool, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

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

func (m *MetricsRepository) GetGauge(ctx context.Context) (map[string]float64, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.Gauge, nil
}

func (m *MetricsRepository) GetGaugeValue(ctx context.Context, key string) (float64, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.Gauge[key], nil
}

func (m *MetricsRepository) GetCounters(ctx context.Context) (map[string]int64, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.Counter, nil
}

func (m *MetricsRepository) GetCounterValue(ctx context.Context, key string) (int64, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.Counter[key], nil
}

func NewMetricsRepository() *MetricsRepository {
	var m MetricsRepository
	m.Gauge = make(map[string]float64)
	m.Counter = make(map[string]int64)

	return &m
}
