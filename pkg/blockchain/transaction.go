package blockchain

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
)

type Transaction struct {
	Sender    []byte // Public Key or Address
	Receiver  []byte // Public Key or Address
	Amount    float64
	Timestamp int64
	Signature []byte // R and S concatenated
}

func (t *Transaction) Hash() ([]byte, error) {
	txCopy := *t           //copy ra để không thay đổi transaction gốc
	txCopy.Signature = nil //Loại bỏ signature
	data, err := json.Marshal(txCopy)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal transaction: %w", err)
	}
	hash := sha256.Sum256(data)
	// Đặc điểm SHA-256:
	// Deterministic: cùng input → cùng output
	// Irreversible: không thể reverse từ hash về data
	// Collision resistant: rất khó tìm 2 input có cùng hash
	return hash[:], nil
}
