package wallet

import (
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	"math/big"

	"github.com/nguyentrinhquy1411/blockchain-go/pkg/blockchain"
)

func SignTransaction(tx *blockchain.Transaction, privKey *ecdsa.PrivateKey) error {
	hash, _ := tx.Hash()
	r, s, err := ecdsa.Sign(rand.Reader, privKey, hash)
	if err != nil {
		return fmt.Errorf("sign error: %w", err)
	}

	tx.Signature = append(r.Bytes(), s.Bytes()...)
	return nil
}

func VerifyTransaction(tx *blockchain.Transaction, pubKey *ecdsa.PublicKey) bool {
	hash, _ := tx.Hash()
	r := new(big.Int).SetBytes(tx.Signature[:len(tx.Signature)/2])
	s := new(big.Int).SetBytes(tx.Signature[len(tx.Signature)/2:])
	return ecdsa.Verify(pubKey, hash, r, s)
}
