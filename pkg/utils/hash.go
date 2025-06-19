package utils

import "crypto/sha256"

// HashData tạo SHA-256 hash từ data
func HashData(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

// CombineHashes kết hợp 2 hash lại với nhau
func CombineHashes(left, right []byte) []byte {
	combined := append(left, right...)
	return HashData(combined)
}
