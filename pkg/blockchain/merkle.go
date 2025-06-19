package blockchain

import "crypto/sha256"

// MerkleTree tối giản cho xác thực tính toàn vẹn
type MerkleTree struct {
	Root []byte
}

// NewMerkleTree tạo Merkle Tree từ transaction hashes
func NewMerkleTree(txHashes [][]byte) *MerkleTree {
	if len(txHashes) == 0 {
		return &MerkleTree{Root: nil}
	}

	nodes := txHashes

	// Build tree bottom-up
	for len(nodes) > 1 {
		var level [][]byte

		for i := 0; i < len(nodes); i += 2 {
			left := nodes[i]
			var right []byte

			if i+1 < len(nodes) {
				right = nodes[i+1]
			} else {
				right = left // Duplicate if odd number
			}

			// Combine and hash
			combined := append(left, right...)
			hash := sha256.Sum256(combined)
			level = append(level, hash[:])
		}
		nodes = level
	}

	return &MerkleTree{Root: nodes[0]}
}

// GetRoot trả về Merkle Root
func (mt *MerkleTree) GetRoot() []byte {
	return mt.Root
}
