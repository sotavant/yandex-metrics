package main

import (
	"encoding/json"
	"github.com/sotavant/yandex-metrics/internal"
	"net/http"
)

func updateJsonHandler(storage Storage) func(res http.ResponseWriter, req *http.Request) {
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

			storage.AddGaugeValue(m.ID, *m.Value)
		case counterType:
			if m.Delta == nil {
				http.Error(res, "value absent", http.StatusBadRequest)
				return
			}

			storage.AddCounterValue(m.ID, *m.Delta)
		default:
			http.Error(res, "bad request", http.StatusBadRequest)
			return
		}

		respStruct := getMetricsStruct(storage, m)
		res.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(res)
		if err := enc.Encode(respStruct); err != nil {
			logger.Infow("error in encode")
			http.Error(res, "internal server error", http.StatusInternalServerError)
			return
		}

		res.WriteHeader(http.StatusOK)
	}
}

func getValueJsonHandler(storage Storage) func(res http.ResponseWriter, req *http.Request) {
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

		if !storage.KeyExist(m.MType, m.ID) {
			http.Error(res, "not found", http.StatusNotFound)
			return
		}

		respStruct := getMetricsStruct(storage, m)
		res.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(res)
		if err := enc.Encode(respStruct); err != nil {
			logger.Infow("error in encode")
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
