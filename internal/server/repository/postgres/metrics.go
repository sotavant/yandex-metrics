package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/sotavant/yandex-metrics/internal"
	"strings"
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

	err = CreateTable(ctx, *conn, tableName)
	if err != nil {
		return nil, fmt.Errorf("error in creating table: %s", err)
	}

	return &MetricsRepository{conn, tableName}, nil
}

func CreateTable(ctx context.Context, conn pgx.Conn, tableName string) error {
	query := strings.ReplaceAll(`create table if not exists #T
		(
			id    varchar not null,
			type  varchar not null,
			delta int8,
			value double precision,
			constraint #T_pk
				unique (id, type)
		);`, "#T", tableName)

	_, err := conn.Exec(ctx, query)

	return err
}

func (m *MetricsRepository) AddGaugeValue(ctx context.Context, key string, value float64) error {
	query := m.setTableName(`insert into #T# (id, type, value)
		values ($1, $2, $3)
		on conflict on constraint #T#_pk do update set id = $1, type = $2, value = $3;`)

	_, err := m.conn.Exec(ctx, query, key, internal.GaugeType, value)

	return err
}

func (m *MetricsRepository) AddCounterValue(ctx context.Context, key string, value int64) error {
	var delta int64
	selectQuery := m.setTableName(`select delta from #T# where type = $1 and id = $2`)
	insertQuery := m.setTableName(`insert into #T# (id, type, delta) values ($1, $2, $3)`)
	updateQuery := m.setTableName(`update #T# set delta = $1 where id = $2 and type = $3`)

	err := m.conn.QueryRow(ctx, selectQuery, internal.CounterType, key).Scan(&delta)
	switch {
	case err == nil:
		_, err = m.conn.Exec(ctx, updateQuery, value+delta, key, internal.CounterType)
		if err != nil {
			internal.Logger.Infow("error in update", "err", err)
			return err
		}
	case errors.Is(err, pgx.ErrNoRows):
		_, err = m.conn.Exec(ctx, insertQuery, key, internal.CounterType, value)
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

func (m *MetricsRepository) AddValues(ctx context.Context, metrics []internal.Metrics) error {
	var err error

	tx, err := m.conn.Begin(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	for _, metric := range metrics {
		switch metric.MType {
		case internal.GaugeType:
			err = m.AddGaugeValue(ctx, metric.ID, *metric.Value)
		case internal.CounterType:
			err = m.AddCounterValue(ctx, metric.ID, *metric.Delta)
		default:
			return errors.New("undefined metric type")
		}

		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (m *MetricsRepository) GetValue(ctx context.Context, mType, key string) (interface{}, error) {
	var delta int64
	var value float64
	var err error

	query := m.setTableName(`select #F# from #T# where type = $1 and id = $2`)

	switch mType {
	case internal.CounterType:
		query = strings.ReplaceAll(query, "#F#", "delta")
		err = m.conn.QueryRow(ctx, query, internal.CounterType, key).Scan(&delta)
	case internal.GaugeType:
		query = strings.ReplaceAll(query, "#F#", "value")
		err = m.conn.QueryRow(ctx, query, internal.GaugeType, key).Scan(&value)
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

func (m *MetricsRepository) GetValues(ctx context.Context) ([]internal.Metrics, error) {
	metrics := make([]internal.Metrics, 0)

	query := m.setTableName(`select type, id, value, delta from #T#`)
	rows, err := m.conn.Query(ctx, query)
	switch {
	case err != nil:
		return nil, err
	case errors.Is(err, pgx.ErrNoRows):
		return nil, nil
	}

	for rows.Next() {
		var metric internal.Metrics
		err = rows.Scan(&metric.MType, &metric.ID, &metric.Value, &metric.Delta)
		if err != nil {
			return nil, err
		}

		metrics = append(metrics, metric)
	}

	return metrics, nil
}

func (m *MetricsRepository) KeyExist(ctx context.Context, mType, key string) (bool, error) {
	var count int

	query := m.setTableName(`select count(*) from #T# where type = $1 and id = $2 limit 1`)

	err := m.conn.QueryRow(ctx, query, mType, key).Scan(&count)

	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		internal.Logger.Infow("error in select count", "err", err)
		return false, err
	}

	return count > 0, nil
}

func (m *MetricsRepository) GetGauge(ctx context.Context) (map[string]float64, error) {
	res := make(map[string]float64)
	query := m.setTableName(`select id, value from #T# where type = $1`)

	rows, err := m.conn.Query(ctx, query, internal.GaugeType)
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
	query := m.setTableName(`select id, delta from #T# where type = $1`)

	rows, err := m.conn.Query(ctx, query, internal.CounterType)
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

func (m *MetricsRepository) setTableName(query string) string {
	return strings.ReplaceAll(query, "#T#", m.tableName)
}
