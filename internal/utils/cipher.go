package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"

	"github.com/sotavant/yandex-metrics/internal"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type Cipher struct {
	privateKey     *rsa.PrivateKey
	publicKey      *rsa.PublicKey
	certPath       string
	privateKeyPath string
}

func NewCipher(privateKeyPath, publicKeyPath, certPath string) (*Cipher, error) {
	if privateKeyPath == "" && publicKeyPath == "" && certPath == "" {
		return nil, nil
	}

	privKey, err := getPrivKey(privateKeyPath)
	if err != nil {
		return nil, err
	}

	pubKey, err := getPubKey(publicKeyPath)
	if err != nil {
		return nil, err
	}

	return &Cipher{
		privateKey:     privKey,
		publicKey:      pubKey,
		certPath:       certPath,
		privateKeyPath: privateKeyPath,
	}, nil
}

func getPrivKey(privateKeyPath string) (*rsa.PrivateKey, error) {
	if privateKeyPath == "" {
		return nil, nil
	}

	privateKeyPEM, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, err
	}

	privateKeyBlock, _ := pem.Decode(privateKeyPEM)
	privateKey, err := x509.ParsePKCS1PrivateKey(privateKeyBlock.Bytes)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

func getPubKey(publicKeyPath string) (*rsa.PublicKey, error) {
	if publicKeyPath == "" {
		return nil, nil
	}

	publicKeyPEM, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return nil, err
	}
	publicKeyBlock, _ := pem.Decode(publicKeyPEM)
	publicKey, err := x509.ParsePKCS1PublicKey(publicKeyBlock.Bytes)
	if err != nil {
		return nil, err
	}

	return publicKey, nil
}

func (c *Cipher) Encrypt(plaintext []byte) ([]byte, error) {
	if c.publicKey == nil {
		return plaintext, nil
	}

	ciphertext, err := rsa.EncryptPKCS1v15(rand.Reader, c.publicKey, plaintext)
	if err != nil {
		return nil, err
	}

	return ciphertext, nil
}

func (c *Cipher) Decrypt(encrypted []byte) ([]byte, error) {
	if c.privateKey == nil {
		return encrypted, nil
	}

	plaintext, err := rsa.DecryptPKCS1v15(rand.Reader, c.privateKey, encrypted)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

func (c *Cipher) IsPrivateKeyExist() bool {
	if c == nil || c.privateKey == nil {
		return false
	}

	return true
}

func (c *Cipher) IsPublicKeyExist() bool {
	if c == nil || c.publicKey == nil {
		return false
	}

	return true
}

func (c *Cipher) GetClientGRPCTransportCreds() credentials.TransportCredentials {
	if c == nil || c.certPath == "" {
		return insecure.NewCredentials()
	}

	creds, err := credentials.NewClientTLSFromFile(c.certPath, "")
	if err != nil {
		internal.Logger.Fatalw("Failed to create TLS credentials", "err", err)
	}

	return creds
}

func (c *Cipher) GetServerGRPCTransportCreds() credentials.TransportCredentials {
	if c == nil || c.certPath == "" || c.privateKeyPath == "" {
		return insecure.NewCredentials()
	}

	creds, err := credentials.NewServerTLSFromFile(c.certPath, c.privateKeyPath)
	if err != nil {
		internal.Logger.Fatalw("Failed to create server TLS credentials", "err", err)
	}

	return creds
}
