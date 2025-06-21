package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
)

func GenerateKeyPair() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
}

func PublicKeyToAddress(pubKey *ecdsa.PublicKey) []byte {
	pubBytes := append(pubKey.X.Bytes(), pubKey.Y.Bytes()...)
	hash := sha256.Sum256(pubBytes)
	return hash[:20]
}
