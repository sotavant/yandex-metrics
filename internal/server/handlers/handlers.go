package handlers

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/sotavant/yandex-metrics/internal"
	"github.com/sotavant/yandex-metrics/internal/server"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

func UpdateHandler(appInstance *server.App) func(res http.ResponseWriter, req *http.Request) {
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

			err = appInstance.Storage.AddGaugeValue(req.Context(), mName, val)
			if err != nil {
				http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		case internal.CounterType:
			val, err := parseValue[int64](mType, mVal)
			if err != nil {
				http.Error(res, "bad request", http.StatusBadRequest)
			}

			err = appInstance.Storage.AddCounterValue(req.Context(), mName, val)
			if err != nil {
				http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		default:
			http.Error(res, "bad request", http.StatusBadRequest)
			return
		}

		if appInstance.Fs != nil && appInstance.Fs.StoreInterval == 0 {
			if err := appInstance.Fs.Sync(req.Context(), appInstance.Storage); err != nil {
				internal.Logger.Infow("error in sync")
				http.Error(res, "internal server error", http.StatusInternalServerError)
				return
			}
		}

		res.WriteHeader(http.StatusOK)
	}
}

func GetValueHandler(appInstance *server.App) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		var strValue string
		mType := chi.URLParam(req, "type")
		mName := chi.URLParam(req, "name")

		if mType != internal.GaugeType && mType != internal.CounterType {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		value, err := appInstance.Storage.GetValue(req.Context(), mType, mName)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		if value == nil {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		if mType == internal.GaugeType {
			strValue = strings.TrimRight(strings.TrimRight(fmt.Sprintf(`%f`, value), "0"), ".")
		} else {
			strValue = strconv.FormatInt(value.(int64), 10)
		}

		_, err = w.Write([]byte(strValue))
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func GetValuesHandler(appInstance *server.App) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		var resp = ""

		gaugeValues, err := appInstance.Storage.GetGauge(req.Context())
		if err != nil {
			internal.Logger.Infow("get gauge values error", "err", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		resp = getHTMLResponseForGaugeList(gaugeValues)

		w.Header().Set("Content-Type", "text/html; charset=utf8")
		_, err = fmt.Fprint(w, resp)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)
			return
		}
	}
}

func PingDBHandler(dbConn *pgx.Conn) func(w http.ResponseWriter, req *http.Request) {
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

func getHTMLResponseForGaugeList(gaugeValues map[string]float64) (resp string) {
	if len(gaugeValues) != 0 {
		keys := make([]string, 0, len(gaugeValues))
		for k := range gaugeValues {
			keys = append(keys, k)
		}

		sort.Strings(keys)

		for _, k := range keys {
			resp += fmt.Sprintf("<p>%s: %s</p>", k, strings.TrimRight(strings.TrimRight(fmt.Sprintf(`%f`, gaugeValues[k]), "0"), "."))
		}
	} else {
		resp = "no value"
	}

	return
}
