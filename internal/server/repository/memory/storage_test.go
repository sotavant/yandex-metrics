package memory

import (
	"context"
	"fmt"
	"testing"

	"github.com/sotavant/yandex-metrics/internal"
	"github.com/stretchr/testify/assert"
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
		wantErr assert.ErrorAssertionFunc
		args    args
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
			tt.wantErr(t, m.AddValue(context.Background(), tt.args.metric), fmt.Sprintf("AddValue(%v)", tt.args.metric))
		})
	}
}

func BenchmarkMetricsRepository_AddValues(b *testing.B) {
	m := fillMetrics()
	b.ResetTimer()
	ctx := context.Background()

	for n := 0; n < b.N; n++ {
		storage := NewMetricsRepository()
		err := storage.AddValues(ctx, m)
		assert.NoError(b, err)
	}
}

func fillMetrics() []internal.Metrics {
	var counter int64 = 1
	res := make([]internal.Metrics, 0)

	for i := 'a'; i < 'z'; i++ {
		metric := internal.Metrics{
			ID: string(i),
		}

		if counter%5 == 0 {
			metric.MType = internal.CounterType
			metric.Delta = &counter
		} else {
			gVal := float64(counter)
			metric.MType = internal.GaugeType
			metric.Value = &gVal
		}
		counter++

		res = append(res, metric)
	}

	return res
}
