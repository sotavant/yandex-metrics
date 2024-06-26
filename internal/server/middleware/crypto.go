package middleware

import (
	"bytes"
	"io"
	"net/http"

	"github.com/sotavant/yandex-metrics/internal"
	"github.com/sotavant/yandex-metrics/internal/utils"
)

// Crypto structure with initialized Cipher
type Crypto struct {
	Cipher *utils.Cipher
}

// NewCrypto initialize struct
func NewCrypto(pathToPrivateKey string) (*Crypto, error) {
	cipher, err := utils.NewCipher(pathToPrivateKey, "")
	if err != nil {
		return nil, err
	}

	return &Crypto{
		Cipher: cipher,
	}, nil
}

// Handler Данный middleware служит для расшифровки тела запроса.
// Если приватный ключ неустановлен, то тело запроса считается не зашифрованным.
func (h *Crypto) Handler(next http.Handler) http.Handler {
	f := func(w http.ResponseWriter, r *http.Request) {
		ow := w
		if h.Cipher.IsPrivateKeyExist() {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				internal.Logger.Infow("read body error", "error", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			encryptedText, err := h.Cipher.Decrypt(body)
			if err != nil {
				internal.Logger.Infow("bad attempt to decrypt")
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			r.Body = io.NopCloser(bytes.NewReader(encryptedText))
		}

		next.ServeHTTP(ow, r)
	}

	return http.HandlerFunc(f)
}
