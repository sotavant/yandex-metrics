// Package middleware содержит middleware
package middleware

import (
	"bytes"
	"io"
	"net/http"
	"strings"

	"github.com/sotavant/yandex-metrics/internal"
	"github.com/sotavant/yandex-metrics/internal/utils"
)

// Hasher содержит ключ шифрования
type Hasher struct {
	key string
}

func NewHasher(key string) *Hasher {
	return &Hasher{key: key}
}

// Handler Данный middleware служит проверки зашифрованного текста запроса.
// Если ключа нет в Header (HashSHA256) запроса, то тело запроса считается не зашифрованным.
// Иначе проверяется на корректность.
// В ответ на запрос также возвращается зашифрованное сообщение, если был передан ключ.
func (h *Hasher) Handler(next http.Handler) http.Handler {
	f := func(w http.ResponseWriter, r *http.Request) {
		ow := w
		hash := r.Header.Get(utils.HasherHeaderKey)
		if h.key != "" && hash != "" {
			check, err := h.checkHash(hash, r)
			if err != nil {
				internal.Logger.Infow("error in check hash", "err", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			if !check {
				internal.Logger.Infow("bad hash")
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			hw := &hasherResponseWriter{
				ResponseWriter: w,
				hashKey:        h.key,
			}

			ow = hw
		}

		next.ServeHTTP(ow, r)
	}

	return http.HandlerFunc(f)
}

func (h *Hasher) checkHash(reqHash string, r *http.Request) (bool, error) {
	var body []byte

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return false, err
	}

	err = r.Body.Close()
	if err != nil {
		return false, err
	}

	r.Body = io.NopCloser(bytes.NewBuffer(body))

	bodyHash, err := utils.GetHash(body, h.key)
	if err != nil {
		return false, err
	}

	return strings.TrimSpace(bodyHash) == strings.TrimSpace(reqHash), nil
}

type hasherResponseWriter struct {
	http.ResponseWriter
	data        []byte
	hashKey     string
	wroteHeader bool
}

func (hr *hasherResponseWriter) Write(p []byte) (int, error) {
	hr.data = p
	if !hr.wroteHeader {
		hr.WriteHeader(http.StatusOK)
	}
	return hr.ResponseWriter.Write(p)
}

func (hr *hasherResponseWriter) WriteHeader(code int) {
	if hr.wroteHeader {
		hr.ResponseWriter.WriteHeader(code)
	}
	hr.wroteHeader = true
	defer hr.ResponseWriter.WriteHeader(code)

	if len(hr.data) != 0 {
		hash, err := utils.GetHash(hr.data, hr.hashKey)
		if err != nil {
			internal.Logger.Infow(
				"error in get hash of data",
				"error", err,
			)
			return
		}

		hr.Header().Set(utils.HasherHeaderKey, hash)
	}
}

func (hr *hasherResponseWriter) Header() http.Header {
	return hr.ResponseWriter.Header()
}
