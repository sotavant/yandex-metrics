package in_memory

import (
	"context"
	"fmt"
	"github.com/sotavant/yandex-metrics/internal"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMemStorage_AddGaugeValue(t *testing.T) {
	type fields struct {
		Gauge   map[string]float64
		Counter map[string]int64
	}
	type args struct {
		key   string
		value float64
	}

	ctx := context.Background()

	tests := []struct {
		name   string
		fields fields
		args   args
		wants  args
	}{
		{
			name:   `newValue`,
			fields: fields{},
			args: args{
				key:   `newValue`,
				value: 345.555,
			},
			wants: args{
				value: 345.555,
			},
		},
		{
			name:   `updateValue`,
			fields: fields{},
			args: args{
				key:   `updateValue`,
				value: 345.555,
			},
			wants: args{
				value: 345.555,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMetricsRepository()
			err := m.AddGaugeValue(ctx, tt.args.key, tt.args.value)
			assert.NoError(t, err)
			if tt.name == `updateValue` {
				err = m.AddGaugeValue(ctx, tt.args.key, tt.args.value)
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.wants.value, m.Gauge[tt.args.key])
		})
	}
}

func TestMemStorage_AddCounterValue(t *testing.T) {
	type fields struct {
		Gauge   map[string]float64
		Counter map[string]int64
	}

	type args struct {
		key   string
		value int64
	}

	ctx := context.Background()

	tests := []struct {
		name   string
		fields fields
		args   args
		wants  args
	}{
		{
			name:   `newValue`,
			fields: fields{},
			args: args{
				key:   `newValue`,
				value: 3,
			},
			wants: args{
				value: 3,
			},
		},
		{
			name:   `updateValue`,
			fields: fields{},
			args: args{
				key:   `updateValue`,
				value: 5,
			},
			wants: args{
				value: 10,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMetricsRepository()
			err := m.AddCounterValue(ctx, tt.args.key, tt.args.value)
			assert.NoError(t, err)
			if tt.name == `updateValue` {
				err = m.AddCounterValue(ctx, tt.args.key, tt.args.value)
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.wants.value, m.Counter[tt.args.key])
		})
	}
}

func TestMemStorage_AddValue(t *testing.T) {
	m := NewMetricsRepository()

	type args struct {
		metric internal.Metrics
	}

	var delta int64 = 11

	tests := []struct {
		name    string
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "correct type",
			args: struct{ metric internal.Metrics }{metric: internal.Metrics{
				ID:    "aa",
				MType: internal.CounterType,
				Delta: &delta,
				Value: nil,
			}},
			wantErr: assert.NoError,
		},
		{
			name: "bad type",
			args: struct{ metric internal.Metrics }{metric: internal.Metrics{
				ID:    "aa",
				MType: "someBadType",
				Delta: &delta,
				Value: nil,
			}},
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			tt.wantErr(t, m.AddValue(tt.args.metric), fmt.Sprintf("AddValue(%v)", tt.args.metric))
		})
	}
}
