package utils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCipher_Encrypt(t *testing.T) {
	pubKeyPath := os.Getenv("TEST_CRYPT_PUB_KEY")
	privKeyPath := os.Getenv("TEST_CRYPT_PRIV_KEY")

	if pubKeyPath == "" && privKeyPath == "" {
		return
	}

	c, err := NewCipher(privKeyPath, pubKeyPath)
	assert.NoError(t, err)
	assert.NotNil(t, c)

	testMsg := "test message"

	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "success",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got, decoded []byte
			got, err = c.Encrypt([]byte(testMsg))
			if (err != nil) != tt.wantErr {
				t.Errorf("Encrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			decoded, err = c.Decrypt(got)
			if (err != nil) != tt.wantErr {
				t.Errorf("Encrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equal(t, string(decoded), testMsg)
		})
	}
}
