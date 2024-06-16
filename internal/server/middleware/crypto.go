package middleware

import "github.com/sotavant/yandex-metrics/internal/utils"

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
