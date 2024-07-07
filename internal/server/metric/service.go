package metric

import (
	"context"
	"errors"
	"net/http"

	"github.com/sotavant/yandex-metrics/internal"
	"github.com/sotavant/yandex-metrics/internal/server/repository"
	"golang.org/x/mod/sumdb/storage"
)

var (
	ErrIDAbsent = errors.New("ID is absent")
	ErrBadType  = errors.New("Bad metric type")
)

type MetricService struct {
	storage *repository.Storage
}

func (ms *MetricService) Upsert(ctx context.Context, m internal.Metrics) (internal.Metrics, error) {
	if m.ID == "" {
		return internal.Metrics{}, ErrIDAbsent
	}

	if m.MType != internal.GaugeType && m.MType != internal.CounterType {
		return internal.Metrics{}, ErrBadType
	}

	exist, err := ms.storage.KeyExist(req.Context(), m.MType, m.ID)
	if err != nil {
		internal.Logger.Infow("error in encode")
		http.Error(res, "internal server error", http.StatusInternalServerError)
		return
	}

	if !exist {
		http.Error(res, "not found", http.StatusNotFound)
		return
	}

	respStruct, err := ms.getMetricsStruct(req.Context(), appInstance.Storage, m)
	if err != nil {
		internal.Logger.Infow("error in getMetricsStruct", "err", err)
		http.Error(res, "internal server error", http.StatusInternalServerError)
		return
	}

	return m, nil
}

func (ms MetricService) getMetricsStruct(ctx context.Context, s storage.Storage, metric internal.Metrics) (internal.Metrics, error) {

}
