package postgres

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/sotavant/yandex-metrics/internal"
	"github.com/sotavant/yandex-metrics/internal/server/repository/postgres/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMetricsRepository_AddGaugeValue(t *testing.T) {
	ctx := context.Background()
	conn, tableName, _, err := test.InitConnection(ctx, t)
	assert.NoError(t, err)
	if conn == nil {
		return
	}
	defer func(ctx context.Context, conn pgx.Conn, tableName string) {
		err = test.DropTable(ctx, conn, tableName)
		assert.NoError(t, err)

		err = conn.Close(ctx)
		assert.NoError(t, err)
	}(ctx, *conn, tableName)

	type args struct {
		ctx   context.Context
		key   string
		value float64
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{
			name: "insert value",
			args: args{
				ctx:   ctx,
				key:   "ss",
				value: -123.123,
			},
			want: -123.123,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MetricsRepository{
				conn:      conn,
				tableName: tableName,
			}

			err := m.AddGaugeValue(tt.args.ctx, tt.args.key, tt.args.value)
			assert.NoError(t, err)

			val, err := m.GetValue(ctx, internal.GaugeType, tt.args.key)
			assert.NoError(t, err)
			assert.Equal(t, val, tt.want)
		})
	}
}

func TestMetricsRepository_AddCounterValue(t *testing.T) {
	ctx := context.Background()
	conn, tableName, _, err := test.InitConnection(ctx, t)
	assert.NoError(t, err)
	if conn == nil {
		return
	}
	defer func(ctx context.Context, conn pgx.Conn, tableName string) {
		err := test.DropTable(ctx, conn, tableName)
		assert.NoError(t, err)
	}(ctx, *conn, tableName)

	type args struct {
		ctx   context.Context
		key   string
		value int64
	}

	tests := []struct {
		name string
		args args
		want int64
	}{
		{
			name: "insert value",
			args: args{
				ctx:   ctx,
				key:   "ss",
				value: 3,
			},
			want: 3,
		},
		{
			name: "insert value",
			args: args{
				ctx:   ctx,
				key:   "ss",
				value: 3,
			},
			want: 6,
		},
		{
			name: "insert big value",
			args: args{
				ctx:   ctx,
				key:   "ss",
				value: 2544985532,
			},
			want: 2544985538,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MetricsRepository{
				conn:      conn,
				tableName: tableName,
			}

			err := m.AddCounterValue(tt.args.ctx, tt.args.key, tt.args.value)
			assert.NoError(t, err)

			val, err := m.GetValue(ctx, internal.CounterType, tt.args.key)
			assert.NoError(t, err)
			assert.Equal(t, val, tt.want)
		})
	}
}

func TestMetricsRepository_AddValue(t *testing.T) {
	ctx := context.Background()
	conn, tableName, _, err := test.InitConnection(ctx, t)
	assert.NoError(t, err)
	if conn == nil {
		return
	}
	defer func(ctx context.Context, conn pgx.Conn, tableName string) {
		err := test.DropTable(ctx, conn, tableName)
		assert.NoError(t, err)
	}(ctx, *conn, tableName)

	type args struct {
		ctx   context.Context
		key   string
		value float64
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{
			name: "insert value",
			args: args{
				ctx:   ctx,
				key:   "ss",
				value: -123.123,
			},
			want: -123.123,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MetricsRepository{
				conn:      conn,
				tableName: tableName,
			}

			metr := &internal.Metrics{
				ID:    tt.args.key,
				MType: internal.GaugeType,
				Delta: nil,
				Value: &tt.args.value,
			}
			err := m.AddValue(tt.args.ctx, *metr)
			assert.NoError(t, err)

			val, err := m.GetValue(ctx, internal.GaugeType, tt.args.key)
			assert.NoError(t, err)
			assert.Equal(t, val, tt.want)
		})
	}
}

func TestMetricsRepository_KeyExist(t *testing.T) {
	ctx := context.Background()
	conn, tableName, _, err := test.InitConnection(ctx, t)
	assert.NoError(t, err)
	if conn == nil {
		return
	}
	defer func(ctx context.Context, conn pgx.Conn, tableName string) {
		err := test.DropTable(ctx, conn, tableName)
		assert.NoError(t, err)
	}(ctx, *conn, tableName)

	type args struct {
		ctx   context.Context
		key   string
		value float64
	}

	type wantArgs struct {
		key string
		res bool
	}

	tests := []struct {
		name    string
		args    args
		want    wantArgs
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "insert value",
			args: args{
				ctx:   ctx,
				key:   "ss",
				value: -123.123,
			},
			want: wantArgs{
				key: "ss",
				res: true,
			},
			wantErr: assert.NoError,
		},
		{
			name: "insert value",
			args: args{
				ctx:   ctx,
				key:   "ss",
				value: -123.123,
			},
			want: wantArgs{
				key: "sss",
				res: false,
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MetricsRepository{
				conn:      conn,
				tableName: tableName,
			}
			err := m.AddGaugeValue(tt.args.ctx, tt.args.key, tt.args.value)
			assert.NoError(t, err)

			got, err := m.KeyExist(tt.args.ctx, internal.GaugeType, tt.want.key)
			if !tt.wantErr(t, err, fmt.Sprintf("KeyExist(%v, %v, %v)", tt.args.ctx, internal.GaugeType, tt.args.key)) {
				return
			}
			assert.Equal(t, tt.want.res, got)
		})
	}
}

func TestMetricsRepository_GetGauge(t *testing.T) {
	ctx := context.Background()
	conn, tableName, _, err := test.InitConnection(ctx, t)
	assert.NoError(t, err)
	if conn == nil {
		return
	}
	defer func(ctx context.Context, conn pgx.Conn, tableName string) {
		err := test.DropTable(ctx, conn, tableName)
		assert.NoError(t, err)
	}(ctx, *conn, tableName)

	tests := []struct {
		name    string
		fields  map[string]float64
		want    map[string]float64
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "add values",
			fields: map[string]float64{
				"s":   -1333,
				"344": 34.3455,
			},
			want: map[string]float64{
				"s":   -1333,
				"344": 34.3455,
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MetricsRepository{
				conn:      conn,
				tableName: tableName,
			}

			for k, v := range tt.fields {
				err := m.AddGaugeValue(ctx, k, v)
				assert.NoError(t, err)
			}

			got, err := m.GetGauge(ctx)
			if !tt.wantErr(t, err, fmt.Sprintf("GetGauge(%v)", ctx)) {
				return
			}
			assert.Equalf(t, tt.want, got, "GetGauge(%v)", ctx)
		})
	}
}

func TestMetricsRepository_GetGaugeValue(t *testing.T) {
	ctx := context.Background()
	conn, tableName, _, err := test.InitConnection(ctx, t)
	assert.NoError(t, err)
	if conn == nil {
		return
	}
	defer func(ctx context.Context, conn pgx.Conn, tableName string) {
		err := test.DropTable(ctx, conn, tableName)
		assert.NoError(t, err)
	}(ctx, *conn, tableName)

	tests := []struct {
		name    string
		fields  map[string]float64
		want    map[string]float64
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "add values",
			fields: map[string]float64{
				"s": -1333,
			},
			want: map[string]float64{
				"s": -1333,
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MetricsRepository{
				conn:      conn,
				tableName: tableName,
			}

			for k, v := range tt.fields {
				err := m.AddGaugeValue(ctx, k, v)
				assert.NoError(t, err)

				got, err := m.GetGaugeValue(ctx, k)
				if !tt.wantErr(t, err, fmt.Sprintf("GetGauge(%v)", ctx)) {
					return
				}
				assert.Equal(t, got, v)
			}

			if !tt.wantErr(t, err, fmt.Sprintf("GetGauge(%v)", ctx)) {
				return
			}
		})
	}
}

func TestMetricsRepository_GetCounters(t *testing.T) {
	ctx := context.Background()
	conn, tableName, _, err := test.InitConnection(ctx, t)
	assert.NoError(t, err)
	if conn == nil {
		return
	}
	defer func(ctx context.Context, conn pgx.Conn, tableName string) {
		err := test.DropTable(ctx, conn, tableName)
		assert.NoError(t, err)
	}(ctx, *conn, tableName)

	tests := []struct {
		name    string
		fields  map[string]int64
		want    map[string]int64
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "add values",
			fields: map[string]int64{
				"s":   -1333,
				"344": 34,
			},
			want: map[string]int64{
				"s":   -1333,
				"344": 34,
			},
			wantErr: assert.NoError,
		},
		{
			name: "add values",
			fields: map[string]int64{
				"s":   1333,
				"344": 34,
			},
			want: map[string]int64{
				"s":   0,
				"344": 68,
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MetricsRepository{
				conn:      conn,
				tableName: tableName,
			}

			for k, v := range tt.fields {
				err := m.AddCounterValue(ctx, k, v)
				assert.NoError(t, err)
			}

			got, err := m.GetCounters(ctx)
			if !tt.wantErr(t, err, fmt.Sprintf("GetGauge(%v)", ctx)) {
				return
			}
			assert.Equalf(t, tt.want, got, "GetGauge(%v)", ctx)
		})
	}
}

func TestMetricsRepository_GetCounterValue(t *testing.T) {
	ctx := context.Background()
	conn, tableName, _, err := test.InitConnection(ctx, t)
	assert.NoError(t, err)
	if conn == nil {
		return
	}
	defer func(ctx context.Context, conn pgx.Conn, tableName string) {
		err := test.DropTable(ctx, conn, tableName)
		assert.NoError(t, err)
	}(ctx, *conn, tableName)

	type fields struct {
		key   string
		value int64
	}

	tests := []struct {
		name    string
		fields  fields
		want    fields
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "add values",
			fields: fields{
				key:   "s",
				value: -1333,
			},
			want: fields{
				key:   "s",
				value: -1333,
			},
			wantErr: assert.NoError,
		},
		{
			name: "repeat values",
			fields: fields{
				key:   "s",
				value: 1333,
			},
			want: fields{
				key:   "s",
				value: 0,
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MetricsRepository{
				conn:      conn,
				tableName: tableName,
			}

			err := m.AddCounterValue(ctx, tt.fields.key, tt.fields.value)
			assert.NoError(t, err)

			got, err := m.GetCounterValue(ctx, tt.fields.key)
			if !tt.wantErr(t, err, fmt.Sprintf("GetCounterValue(%v)", ctx)) {
				return
			}

			assert.Equal(t, got, tt.want.value)
		})
	}
}
