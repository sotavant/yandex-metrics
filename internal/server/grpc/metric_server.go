package grpc

import (
	"context"
	"errors"

	"github.com/sotavant/yandex-metrics/internal"
	"github.com/sotavant/yandex-metrics/internal/server/metric"
	pb "github.com/sotavant/yandex-metrics/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MetricServer struct {
	pb.UnimplementedMetricsServer
	MService *metric.MetricService
}

func NewMetricServer(mService *metric.MetricService) *MetricServer {
	return &MetricServer{
		MService: mService,
	}
}

func (m *MetricServer) UpdateMetric(ctx context.Context, req *pb.UpdateMetricRequest) (*pb.UpdateMetricResponse, error) {
	value := req.Metric.Value
	delta := req.Metric.Delta

	reqMetric := internal.Metrics{
		Value: &value,
		Delta: &delta,
		ID:    req.Metric.ID,
		MType: req.Metric.MType,
	}

	respStruct, err := m.MService.Upsert(ctx, reqMetric)
	if err != nil {
		return nil, getError(err)
	}

	return &pb.UpdateMetricResponse{
		Metric: &pb.Metric{
			Value: *respStruct.Value,
			Delta: *respStruct.Delta,
			ID:    respStruct.ID,
			MType: respStruct.MType,
		},
		Error: "",
	}, nil
}

func (m *MetricServer) UpdateMetricTest(ctx context.Context, req *pb.UpdateMetricRequest) (*pb.UpdateMetricResponse, error) {
	return nil, nil
}

func getError(err error) error {
	switch {
	case errors.Is(err, metric.ErrIDAbsent), errors.Is(err, metric.ErrBadType), errors.Is(err, metric.ErrValueAbsent):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, metric.ErrAddGaugeValue), errors.Is(err, metric.ErrAddCounterValue):
		return status.Error(codes.Internal, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}
