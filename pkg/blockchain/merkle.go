package blockchain

import "crypto/sha256"

type MerkleTree struct {
	Root []byte
}

func NewMerkleTree(txHashes [][]byte) *MerkleTree {
	if len(txHashes) == 0 {
		return &MerkleTree{Root: nil}
	}

	nodes := txHashes

	for len(nodes) > 1 {
		var level [][]byte

		for i := 0; i < len(nodes); i += 2 {
			left := nodes[i]
			var right []byte

			if i+1 < len(nodes) {
				right = nodes[i+1]
			} else {
				right = left
			}

			combined := append(left, right...)
			hash := sha256.Sum256(combined)
			level = append(level, hash[:])
		}
		nodes = level
	}

	return &MerkleTree{Root: nodes[0]}
}

func (mt *MerkleTree) GetRoot() []byte {
	return mt.Root
}
