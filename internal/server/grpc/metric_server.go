package grpc

import (
	"context"
	"fmt"

	pb "github.com/sotavant/yandex-metrics/proto"
)

type MetricServer struct {
	pb.UnimplementedMetricsServer
}

func (m *MetricServer) UpdateMetric(ctx context.Context, req *pb.UpdateMetricRequest) (*pb.UpdateMetricResponse, error) {
	fmt.Println(req.Metric)
	return nil, nil
}
