package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/sotavant/yandex-metrics/internal"
	"net/http"
	"strconv"
	"strings"
)

func updateHandler(appInstance *app) func(res http.ResponseWriter, req *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		mType := chi.URLParam(req, "type")
		mName := chi.URLParam(req, "name")
		mVal := chi.URLParam(req, "value")

		switch mType {
		case internal.GaugeType:
			val, err := parseValue[float64](mType, mVal)
			if err != nil {
				http.Error(res, "bad request", http.StatusBadRequest)
				return
			}

			err = appInstance.memStorage.AddGaugeValue(req.Context(), mName, val)
			if err != nil {
				http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		case internal.CounterType:
			val, err := parseValue[int64](mType, mVal)
			if err != nil {
				http.Error(res, "bad request", http.StatusBadRequest)
			}

			err = appInstance.memStorage.AddCounterValue(req.Context(), mName, val)
			if err != nil {
				http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		default:
			http.Error(res, "bad request", http.StatusBadRequest)
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

func getValueHandler(appInstance *app) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		var strValue string
		mType := chi.URLParam(req, "type")
		mName := chi.URLParam(req, "name")

		if mType != internal.GaugeType && mType != internal.CounterType {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		value := appInstance.memStorage.GetValue(mType, mName)
		if value == nil {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		if mType == internal.GaugeType {
			strValue = strings.TrimRight(strings.TrimRight(fmt.Sprintf(`%f`, value), "0"), ".")
		} else {
			strValue = strconv.FormatInt(value.(int64), 10)
		}

		_, err := w.Write([]byte(strValue))
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func getValuesHandler(appInstance *app) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		var resp = ""

		if len(appInstance.memStorage.GetGauge()) != 0 {
			for k, v := range appInstance.memStorage.GetGauge() {
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

func pingDBHandler(dbConn *pgx.Conn) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		if dbConn == nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		err := dbConn.Ping(req.Context())
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func parseValue[T float64 | int64](mType, mValue string) (T, error) {
	switch mType {
	case internal.GaugeType:
		floatVal, err := strconv.ParseFloat(strings.TrimSpace(mValue), 64)
		if err != nil {
			return 0, err
		}

		return T(floatVal), nil
	case internal.CounterType:
		intVal, err := strconv.ParseInt(strings.TrimSpace(mValue), 10, 64)
		if err != nil {
			return 0, err
		}

		return T(intVal), nil
	}

	return T(0), nil
}
