package main

import (
	"encoding/json"
	"github.com/sotavant/yandex-metrics/internal"
	"github.com/sotavant/yandex-metrics/internal/server/repository"
	"net/http"
)

func updateJSONHandler(appInstance *app) func(res http.ResponseWriter, req *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		var m internal.Metrics

		if err := json.NewDecoder(req.Body).Decode(&m); err != nil {
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

			err := appInstance.memStorage.AddGaugeValue(req.Context(), m.ID, *m.Value)
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

			err := appInstance.memStorage.AddCounterValue(req.Context(), m.ID, *m.Delta)
			if err != nil {
				internal.Logger.Infow("error in add counter value", "err", err)
				http.Error(res, "internal server error", http.StatusInternalServerError)
				return
			}
		default:
			http.Error(res, "bad request", http.StatusBadRequest)
			return
		}

		respStruct := getMetricsStruct(appInstance.memStorage, m)
		res.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(res)
		if err := enc.Encode(respStruct); err != nil {
			internal.Logger.Infow("error in encode")
			http.Error(res, "internal server error", http.StatusInternalServerError)
			return
		}

		if appInstance.fs.storeInterval == 0 {
			if err := appInstance.fs.Sync(appInstance.memStorage); err != nil {
				internal.Logger.Infow("error in sync")
				http.Error(res, "internal server error", http.StatusInternalServerError)
				return
			}
		}

		res.WriteHeader(http.StatusOK)
	}
}

func getValueJSONHandler(appInstance *app) func(res http.ResponseWriter, req *http.Request) {
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

		exist, err := appInstance.memStorage.KeyExist(req.Context(), m.MType, m.ID)
		if err != nil {
			internal.Logger.Infow("error in encode")
			http.Error(res, "internal server error", http.StatusInternalServerError)
			return
		}

		if !exist {
			http.Error(res, "not found", http.StatusNotFound)
			return
		}

		respStruct := getMetricsStruct(appInstance.memStorage, m)
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

func getMetricsStruct(storage repository.Storage, before internal.Metrics) internal.Metrics {
	m := before

	switch m.MType {
	case internal.GaugeType:
		gValue := storage.GetGaugeValue(m.ID)
		m.Value = &gValue
	case internal.CounterType:
		cValue := storage.GetCounterValue(m.ID)
		m.Delta = &cValue
	}

	return m
}
