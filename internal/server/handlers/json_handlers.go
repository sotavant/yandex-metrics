package handlers

import (
	"context"
	"net/http"

	jsoniter "github.com/json-iterator/go"
	"github.com/sotavant/yandex-metrics/internal"
	"github.com/sotavant/yandex-metrics/internal/server"
	"github.com/sotavant/yandex-metrics/internal/server/repository"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func UpdateJSONHandler(appInstance *server.App) func(res http.ResponseWriter, req *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		var m internal.Metrics

		if err := json.NewDecoder(req.Body).Decode(&m); err != nil {
			internal.Logger.Infow("decode error", "err", err)
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}

		if m.ID == "" {
			http.Error(res, "id absent", http.StatusBadRequest)
			return
		}

		switch m.MType {
		case internal.GaugeType:
			if m.Value == nil {
				http.Error(res, "value absent", http.StatusBadRequest)
				return
			}

			err := appInstance.Storage.AddGaugeValue(req.Context(), m.ID, *m.Value)
			if err != nil {
				internal.Logger.Infow("error in add value", "err", err)
				http.Error(res, "internal server error", http.StatusInternalServerError)
				return
			}
		case internal.CounterType:
			if m.Delta == nil {
				http.Error(res, "value absent", http.StatusBadRequest)
				return
			}

			err := appInstance.Storage.AddCounterValue(req.Context(), m.ID, *m.Delta)
			if err != nil {
				internal.Logger.Infow("error in add counter value", "err", err)
				http.Error(res, "internal server error", http.StatusInternalServerError)
				return
			}
		default:
			http.Error(res, "bad request", http.StatusBadRequest)
			return
		}

		respStruct, err := getMetricsStruct(req.Context(), appInstance.Storage, m)
		if err != nil {
			internal.Logger.Infow("error in get metric struct", "err", err)
			http.Error(res, "internal server error", http.StatusInternalServerError)
			return
		}

		res.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(res)
		if err = enc.Encode(respStruct); err != nil {
			internal.Logger.Infow("error in encode")
			http.Error(res, "internal server error", http.StatusInternalServerError)
			return
		}

		if appInstance.Fs != nil && appInstance.Fs.StoreInterval == 0 {
			if err = appInstance.Fs.Sync(req.Context(), appInstance.Storage); err != nil {
				internal.Logger.Infow("error in sync", "err", err)
				http.Error(res, "internal server error", http.StatusInternalServerError)
				return
			}
		}

		res.WriteHeader(http.StatusOK)
	}
}

func UpdateBatchJSONHandler(appInstance *server.App) func(res http.ResponseWriter, req *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		var m []internal.Metrics

		if err := json.NewDecoder(req.Body).Decode(&m); err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}

		if len(m) == 0 {
			http.Error(res, "data absent", http.StatusBadRequest)
			return
		}

		err := appInstance.Storage.AddValues(req.Context(), m)
		if err != nil {
			internal.Logger.Infow("error in addValues", "err", err)
			http.Error(res, "internal server error", http.StatusInternalServerError)
			return
		}

		metrics, err := appInstance.Storage.GetValues(req.Context())
		if err != nil {
			internal.Logger.Infow("error in get metrics", "err", err)
			http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		res.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(res)
		if err := enc.Encode(metrics); err != nil {
			internal.Logger.Infow("error in encode")
			http.Error(res, "internal server error", http.StatusInternalServerError)
			return
		}

		if appInstance.Fs != nil && appInstance.Fs.StoreInterval == 0 {
			if err = appInstance.Fs.Sync(req.Context(), appInstance.Storage); err != nil {
				internal.Logger.Infow("error in sync")
				http.Error(res, "internal server error", http.StatusInternalServerError)
				return
			}
		}

		res.WriteHeader(http.StatusOK)
	}
}

func GetValueJSONHandler(appInstance *server.App) func(res http.ResponseWriter, req *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		var m internal.Metrics

		if err := json.NewDecoder(req.Body).Decode(&m); err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
		}

		if m.ID == "" {
			http.Error(res, "id absent", http.StatusBadRequest)
		}

		if m.MType != internal.GaugeType && m.MType != internal.CounterType {
			http.Error(res, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		exist, err := appInstance.Storage.KeyExist(req.Context(), m.MType, m.ID)
		if err != nil {
			internal.Logger.Infow("error in encode")
			http.Error(res, "internal server error", http.StatusInternalServerError)
			return
		}

		if !exist {
			http.Error(res, "not found", http.StatusNotFound)
			return
		}

		respStruct, err := getMetricsStruct(req.Context(), appInstance.Storage, m)
		if err != nil {
			internal.Logger.Infow("error in getMetricsStruct", "err", err)
			http.Error(res, "internal server error", http.StatusInternalServerError)
			return
		}

		res.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(res)
		if err := enc.Encode(respStruct); err != nil {
			internal.Logger.Infow("error in encode")
			http.Error(res, "internal server error", http.StatusInternalServerError)
			return
		}

		res.WriteHeader(http.StatusOK)
	}
}

func getMetricsStruct(ctx context.Context, storage repository.Storage, before internal.Metrics) (internal.Metrics, error) {
	var err error
	m := before

	switch m.MType {
	case internal.GaugeType:
		gValue, err := storage.GetGaugeValue(ctx, m.ID)
		if err != nil {
			return m, err
		}
		m.Value = &gValue
	case internal.CounterType:
		cValue, err := storage.GetCounterValue(ctx, m.ID)
		if err != nil {
			return m, err
		}
		m.Delta = &cValue
	}

	return m, err
}
