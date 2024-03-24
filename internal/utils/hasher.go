package utils

import (
	"crypto/sha256"
	"fmt"
	"strings"
)

const HasherHeaderKey = "HashSHA256"

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
