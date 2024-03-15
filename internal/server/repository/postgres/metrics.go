package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/sotavant/yandex-metrics/internal"
)

type MetricsRepository struct {
	conn      *pgx.Conn
	tableName string
}

func NewMemStorage(ctx context.Context, conn *pgx.Conn, tableName string) (*MetricsRepository, error) {
	err := conn.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("error in ping: %s", err)
	}

	err = createTable(ctx, *conn, tableName)
	if err != nil {
		return nil, fmt.Errorf("error in creating table: %s", err)
	}

	return &MetricsRepository{conn, tableName}, nil
}

func createTable(ctx context.Context, conn pgx.Conn, tableName string) error {
	query := `create table if not exists $1
		(
			id    varchar not null,
			type  varchar not null,
			delta integer,
			value double precision,
			constraint $2
				unique (id, type)
		);`

	_, err := conn.Exec(ctx, query, tableName, tableName+"_pk")

	return err
}

func (m *MetricsRepository) AddGaugeValue(ctx context.Context, key string, value float64) error {
	query := `insert into $4 (id, type, value)
		values ($1, $2, $3)
		on conflict do update set id = $1, type = $2, value = $3;`

	_, err := m.conn.Exec(ctx, query, key, internal.GaugeType, value, m.tableName)

	return err
}

func (m *MetricsRepository) AddCounterValue(ctx context.Context, key string, value int64) error {
	var delta int64
	selectQuery := `select delta from $3 where type = $1 and id = $2`
	insertQuery := `insert into $1 (id, type, delta) values ($2, $3, $4)`
	updateQuery := `update $1 set delta = $2 where key = $3 and type = $4`

	err := m.conn.QueryRow(ctx, selectQuery, internal.CounterType, key).Scan(&delta)
	switch err {
	case nil:
		_, err = m.conn.Exec(ctx, updateQuery, m.tableName, value+delta, key, internal.CounterType)
		if err != nil {
			internal.Logger.Infow("error in update", "err", err)
			return err
		}
	case pgx.ErrNoRows:
		_, err = m.conn.Exec(ctx, insertQuery, m.tableName, key, internal.CounterType, value)
		if err != nil {
			internal.Logger.Infow("error in insert", "err", err)
			return err
		}
	default:
		internal.Logger.Infow("error in select", "err", err)
		return err
	}

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
	var delta int64
	var value float64
	var err error

	query := `select $1 from $2 where type = $3 and key = $4`

	switch mType {
	case internal.CounterType:
		err = m.conn.QueryRow(ctx, query, "delta", m.tableName, internal.CounterType, key).Scan(&delta)
	case internal.GaugeType:
		err = m.conn.QueryRow(ctx, query, "value", m.tableName, internal.GaugeType, key).Scan(&value)
	default:
		return nil, nil
	}

	switch {
	case err == nil:
		switch mType {
		case internal.GaugeType:
			return value, nil
		case internal.CounterType:
			return delta, nil
		}
	case errors.Is(err, pgx.ErrNoRows):
		return nil, nil
	default:
		internal.Logger.Infow("error in getValue", "err", err)
		return nil, err
	}

	return nil, nil
}

func (m *MetricsRepository) KeyExist(ctx context.Context, mType, key string) (bool, error) {
	var count int

	query := `select count(*) from $1 where type = $2 and key = $3 limit 1`

	err := m.conn.QueryRow(ctx, query, m.tableName, mType, key).Scan(&count)

	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		internal.Logger.Infow("error in select count", "err", err)
		return false, err
	}

	return count > 0, nil
}

func (m *MetricsRepository) GetGauge(ctx context.Context) (map[string]float64, error) {
	res := make(map[string]float64)
	query := `select key, value from $1 where type = $2`

	rows, err := m.conn.Query(ctx, query, m.tableName, internal.GaugeType)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		internal.Logger.Infow("error in select rows", "err", err)
		return nil, err
	}

	for rows.Next() {
		var key string
		var val float64
		err := rows.Scan(&key, &val)
		if err != nil {
			internal.Logger.Infow("error in scan gauge row", "err", err)
			return nil, err
		}

		res[key] = val
	}

	return res, nil
}

func (m *MetricsRepository) GetGaugeValue(ctx context.Context, key string) (float64, error) {
	val, err := m.GetValue(ctx, internal.GaugeType, key)

	switch i := val.(type) {
	case float64:
		return i, err
	default:
		return 0, errors.New("unknown type of result")
	}
}

func (m *MetricsRepository) GetCounters(ctx context.Context) (map[string]int64, error) {
	res := make(map[string]int64)
	query := `select key, delta from $1 where type = $2`

	rows, err := m.conn.Query(ctx, query, m.tableName, internal.CounterType)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		internal.Logger.Infow("error in select rows", "err", err)
		return nil, err
	}

	for rows.Next() {
		var key string
		var val int64
		err := rows.Scan(&key, &val)
		if err != nil {
			internal.Logger.Infow("error in scan counter row", "err", err)
			return nil, err
		}

		res[key] = val
	}

	return res, nil
}

func (m *MetricsRepository) GetCounterValue(ctx context.Context, key string) (int64, error) {
	val, err := m.GetValue(ctx, internal.CounterType, key)

	switch i := val.(type) {
	case int64:
		return i, err
	default:
		return 0, errors.New("unknown type of result")
	}
}
