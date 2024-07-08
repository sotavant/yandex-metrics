package metric

import (
	"context"
	"errors"

	"github.com/sotavant/yandex-metrics/internal"
	"github.com/sotavant/yandex-metrics/internal/server/repository"
)

var (
	ErrIDAbsent        = errors.New("ID is absent")
	ErrBadType         = errors.New("Bad metric type")
	ErrValueAbsent     = errors.New("Value is absent")
	ErrAddGaugeValue   = errors.New("error in add gauge value")
	ErrAddCounterValue = errors.New("error in add counter value")
)

type MetricService struct {
	storage repository.Storage
}

func NewMetricService(st repository.Storage) *MetricService {
	return &MetricService{
		storage: st,
	}
}

func (ms *MetricService) Upsert(ctx context.Context, m internal.Metrics) (internal.Metrics, error) {
	if m.ID == "" {
		return internal.Metrics{}, ErrIDAbsent
	}

	switch m.MType {
	case internal.GaugeType:
		if m.Value == nil {
			return internal.Metrics{}, ErrValueAbsent
		}

		err := ms.storage.AddGaugeValue(ctx, m.ID, *m.Value)
		if err != nil {
			return internal.Metrics{}, ErrAddGaugeValue
		}
	case internal.CounterType:
		if m.Delta == nil {
			return internal.Metrics{}, ErrValueAbsent
		}

		err := ms.storage.AddCounterValue(ctx, m.ID, *m.Delta)
		if err != nil {
			return internal.Metrics{}, ErrAddCounterValue
		}
	default:
		return internal.Metrics{}, ErrBadType
	}

	return GetMetricsStruct(ctx, ms.storage, m)
}

func GetMetricsStruct(ctx context.Context, storage repository.Storage, before internal.Metrics) (internal.Metrics, error) {
	var err error
	var gValue float64
	var cValue int64
	m := before

	switch m.MType {
	case internal.GaugeType:
		gValue, err = storage.GetGaugeValue(ctx, m.ID)
		if err != nil {
			return m, err
		}
		m.Value = &gValue
	case internal.CounterType:
		cValue, err = storage.GetCounterValue(ctx, m.ID)
		if err != nil {
			return m, err
		}
		m.Delta = &cValue
	}

	return m, err
}
