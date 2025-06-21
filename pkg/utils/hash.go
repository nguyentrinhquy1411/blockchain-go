package utils

import "crypto/sha256"

func HashData(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

func CombineHashes(left, right []byte) []byte {
	combined := append(left, right...)
	return HashData(combined)
}
