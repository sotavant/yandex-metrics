package main

import (
	"encoding/json"
	"github.com/sotavant/yandex-metrics/internal"
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
		case gaugeType:
			if m.Value == nil {
				http.Error(res, "value absent", http.StatusBadRequest)
				return
			}

			appInstance.memStorage.AddGaugeValue(m.ID, *m.Value)
		case counterType:
			if m.Delta == nil {
				http.Error(res, "value absent", http.StatusBadRequest)
				return
			}

			appInstance.memStorage.AddCounterValue(m.ID, *m.Delta)
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

		if m.MType != gaugeType && m.MType != counterType {
			http.Error(res, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		if !appInstance.memStorage.KeyExist(m.MType, m.ID) {
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

func getMetricsStruct(storage Storage, before internal.Metrics) internal.Metrics {
	m := before

	switch m.MType {
	case gaugeType:
		gValue := storage.GetGaugeValue(m.ID)
		m.Value = &gValue
	case counterType:
		cValue := storage.GetCounterValue(m.ID)
		m.Delta = &cValue
	}

	return m
}
