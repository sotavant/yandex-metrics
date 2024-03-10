package postgres

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/sotavant/yandex-metrics/internal"
)

const tableName = "metric"

type MetricsRepository struct {
	conn *pgx.Conn
}

func NewMemStorage(ctx context.Context, conn *pgx.Conn) (*MetricsRepository, error) {
	err := conn.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("error in ping: %s", err)
	}

	err = createTable(ctx, *conn)
	if err != nil {
		return nil, fmt.Errorf("error in creating table: %s", err)
	}

	return &MetricsRepository{conn}, nil
}

func createTable(ctx context.Context, conn pgx.Conn) error {
	query := `create table if not exists $1
		(
			id    varchar not null,
			type  varchar not null,
			delta integer,
			value double precision,
			constraint metric_pk
				unique (id, type)
		);`

	_, err := conn.Exec(ctx, query, tableName)

	return err
}

func (m *MetricsRepository) AddGaugeValue(ctx context.Context, key string, value float64) error {

}

func (m *MetricsRepository) AddCounterValue(ctx context.Context, key string, value int64) {
	m.Counter[key] += value
}

func (m *MetricsRepository) upsert(ctx context.Context, metric internal.Metrics) error {
	return nil
}

func (m *MetricsRepository) AddValue(metric internal.Metrics) error {
	switch metric.MType {
	case internal.GaugeType:
		m.AddGaugeValue(metric.ID, *metric.Value)
	case internal.CounterType:
		m.AddCounterValue(metric.ID, *metric.Delta)
	default:
		return fmt.Errorf("undefinde type: %s", metric.MType)
	}

	return nil
}

func (m *MetricsRepository) GetValue(mType, key string) interface{} {
	switch mType {
	case internal.GaugeType:
		val, ok := m.Gauge[key]
		if ok {
			return val
		}
	case internal.CounterType:
		val, ok := m.Counter[key]
		if ok {
			return val
		}
	}

	return nil
}

func (m *MetricsRepository) KeyExist(mType, key string) bool {
	switch mType {
	case internal.GaugeType:
		_, ok := m.Gauge[key]
		if ok {
			return true
		}
	case internal.CounterType:
		_, ok := m.Counter[key]
		if ok {
			return true
		}
	}

	return false
}

func (m *MetricsRepository) GetGauge() map[string]float64 {
	return m.Gauge
}

func (m *MetricsRepository) GetGaugeValue(key string) float64 {
	return m.Gauge[key]
}

func (m *MetricsRepository) GetCounters() map[string]int64 {
	return m.Counter
}

func (m *MetricsRepository) GetCounterValue(key string) int64 {
	return m.Counter[key]
}
