package grpc

import (
	"context"
	"testing"

	"github.com/sotavant/yandex-metrics/internal/server/metric"
	"github.com/sotavant/yandex-metrics/internal/server/repository/memory"
	pb "github.com/sotavant/yandex-metrics/proto"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestMetricServer_UpdateMetric(t *testing.T) {
	st := memory.NewMetricsRepository()

	server := NewMetricServer(metric.NewMetricService(st))

	tests := []struct {
		req       *pb.Metric
		wantError error
		name      string
		wantValue float64
	}{
		{
			name: "success",
			req: &pb.Metric{
				Value: 1,
				Delta: 0,
				ID:    "ss",
				MType: "gauge",
			},
			wantError: nil,
			wantValue: 1,
		},
		{
			name: "id absent",
			req: &pb.Metric{
				Value: 1,
				Delta: 0,
				ID:    "",
				MType: "gauge",
			},
			wantError: status.Error(codes.InvalidArgument, metric.ErrIDAbsent.Error()),
			wantValue: 1,
		},
		{
			name: "bad type",
			req: &pb.Metric{
				Value: 1,
				Delta: 0,
				ID:    "sss",
				MType: "ssss",
			},
			wantError: status.Error(codes.InvalidArgument, metric.ErrBadType.Error()),
			wantValue: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := pb.UpdateMetricRequest{Metric: tt.req}
			res, err := server.UpdateMetric(context.Background(), &req)
			if tt.wantError != nil {
				assert.ErrorIs(t, err, tt.wantError)
			} else {
				assert.Equal(t, tt.wantValue, res.Metric.Value)
			}
		})
	}
}
