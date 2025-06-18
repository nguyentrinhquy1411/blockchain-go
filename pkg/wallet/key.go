package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
)

func GenerateKeyPair() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(elliptic.P256(), rand.Reader) // xài đường cong elliptic.P256() để tạo khóa
}

func PublicKeyToAddress(pubKey *ecdsa.PublicKey) []byte {
	pubBytes := append(pubKey.X.Bytes(), pubKey.Y.Bytes()...) // nghiên cứu thêm
	hash := sha256.Sum256(pubBytes)
	return hash[:20]
}
