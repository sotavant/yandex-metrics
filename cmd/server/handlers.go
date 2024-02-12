package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
	"strings"
)

func updateHandler(storage Storage) func(res http.ResponseWriter, req *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		mType := chi.URLParam(req, "type")
		mName := chi.URLParam(req, "name")
		mVal := chi.URLParam(req, "value")

		switch mType {
		case gaugeType:
			val, err := parseValue[float64](mType, mVal)
			if err != nil {
				http.Error(res, "bad request", http.StatusBadRequest)
			}

			storage.AddGaugeValue(mName, val)
		case counterType:
			val, err := parseValue[int64](mType, mVal)
			if err != nil {
				http.Error(res, "bad request", http.StatusBadRequest)
			}

			storage.AddCounterValue(mName, val)
		default:
			http.Error(res, "bad request", http.StatusBadRequest)
			return
		}

		res.WriteHeader(http.StatusOK)
	}
}

func getValueHandler(storage Storage) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		var strValue string
		mType := chi.URLParam(req, "type")
		mName := chi.URLParam(req, "name")

		if mType != gaugeType && mType != counterType {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		value := storage.GetValue(mType, mName)
		if value == nil {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		if mType == gaugeType {
			strValue = strings.TrimRight(strings.TrimRight(fmt.Sprintf(`%f`, value), "0"), ".")
		} else {
			strValue = strconv.FormatInt(value.(int64), 10)
		}

		_, err := w.Write([]byte(strValue))
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)
			return
		}
	}
}

func getValuesHandler(storage Storage) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		var resp = ""

		if len(storage.GetGauge()) != 0 {
			for k, v := range storage.GetGauge() {
				resp += fmt.Sprintf("<p>%s: %s</p>", k, strings.TrimRight(strings.TrimRight(fmt.Sprintf(`%f`, v), "0"), "."))
			}
		} else {
			resp = "no value"
		}

		w.Header().Set("Content-Type", "text/html; charset=utf8")
		_, err := fmt.Fprint(w, resp)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)
			return
		}
	}
}

func parseValue[T float64 | int64](mType, mValue string) (T, error) {
	switch mType {
	case gaugeType:
		floatVal, err := strconv.ParseFloat(strings.TrimSpace(mValue), 64)
		if err != nil {
			return 0, err
		}

		return T(floatVal), nil
	case counterType:
		intVal, err := strconv.ParseInt(strings.TrimSpace(mValue), 10, 64)
		if err != nil {
			return 0, err
		}

		return T(intVal), nil
	}

	return T(0), nil
}
