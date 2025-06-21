package blockchain

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
)

type Transaction struct {
	Sender    []byte
	Receiver  []byte
	Amount    float64
	Timestamp int64
	Signature []byte
}

func (t *Transaction) Hash() ([]byte, error) {
	txCopy := *t
	txCopy.Signature = nil
	data, err := json.Marshal(txCopy)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal transaction: %w", err)
	}
	hash := sha256.Sum256(data)
	return hash[:], nil
}
