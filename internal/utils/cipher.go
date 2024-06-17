package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
)

type Cipher struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

func NewCipher(privateKeyPath, publicKeyPath string) (*Cipher, error) {
	if privateKeyPath == "" && publicKeyPath == "" {
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
		privateKey: privKey,
		publicKey:  pubKey,
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
	return c.privateKey != nil
}

func (c *Cipher) IsPublicKeyExist() bool {
	return c.publicKey != nil
}
