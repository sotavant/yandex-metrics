// Package utils Вспомогательные утилиты
package utils

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"strings"

	"github.com/sotavant/yandex-metrics/internal"
)

const HasherHeaderKey = "HashSHA256"

// GetHash Получения зашифрованного сообщения, с помощью ключа key. Используется алгоритм sha256
func GetHash(data []byte, key string) (hash string, err error) {
	cleanData := strings.TrimSuffix(string(data), "\n")
	h := sha256.New()
	_, err = h.Write([]byte(cleanData))
	if err != nil {
		return
	}

	_, err = h.Write([]byte(key))
	if err != nil {
		return
	}

	return fmt.Sprintf("%x", h.Sum(nil)), err
}

func GetMetricHash(m internal.Metrics, key string) (hash string, err error) {
	var metricsBuf bytes.Buffer
	enc := gob.NewEncoder(&metricsBuf)
	err = enc.Encode(m)
	if err != nil {
		return
	}

	return GetHash(metricsBuf.Bytes(), key)
}
