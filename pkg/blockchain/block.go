package blockchain

import (
	"crypto/sha256"
	"encoding/json"
	"time"
)

type Block struct {
	Index             int            `json:"index"`
	Timestamp         int64          `json:"timestamp"`
	Transactions      []*Transaction `json:"transactions"`
	MerkleRoot        []byte         `json:"merkle_root"`
	PreviousBlockHash []byte         `json:"previous_block_hash"`
	CurrentBlockHash  []byte         `json:"current_block_hash"`
}

func NewBlock(index int, transactions []*Transaction, prevHash []byte) *Block {
	block := &Block{
		Index:             index,
		Timestamp:         time.Now().Unix(),
		Transactions:      transactions,
		PreviousBlockHash: prevHash,
	}

	block.CalculateMerkleRoot()
	block.CalculateHash()

	return block
}

func (b *Block) CalculateMerkleRoot() {
	if len(b.Transactions) == 0 {
		b.MerkleRoot = []byte{}
		return
	}

	var txHashes [][]byte
	for _, tx := range b.Transactions {
		hash, err := tx.Hash()
		if err != nil {
			continue
		}
		txHashes = append(txHashes, hash)
	}

	merkleTree := NewMerkleTree(txHashes)
	b.MerkleRoot = merkleTree.GetRoot()
}

func (b *Block) CalculateHash() {
	blockData := struct {
		Index             int            `json:"index"`
		Timestamp         int64          `json:"timestamp"`
		Transactions      []*Transaction `json:"transactions"`
		MerkleRoot        []byte         `json:"merkle_root"`
		PreviousBlockHash []byte         `json:"previous_block_hash"`
	}{
		Index:             b.Index,
		Timestamp:         b.Timestamp,
		Transactions:      b.Transactions,
		MerkleRoot:        b.MerkleRoot,
		PreviousBlockHash: b.PreviousBlockHash,
	}

	data, err := json.Marshal(blockData)
	if err != nil {
		return
	}

	hash := sha256.Sum256(data)
	b.CurrentBlockHash = hash[:]
}

func (b *Block) IsValid() bool {
	var txHashes [][]byte
	for _, tx := range b.Transactions {
		hash, err := tx.Hash()
		if err != nil {
			return false
		}
		txHashes = append(txHashes, hash)
	}

	merkleTree := NewMerkleTree(txHashes)
	calculatedRoot := merkleTree.GetRoot()

	if len(calculatedRoot) != len(b.MerkleRoot) {
		return false
	}
	for i := range calculatedRoot {
		if calculatedRoot[i] != b.MerkleRoot[i] {
			return false
		}
	}

	originalHash := make([]byte, len(b.CurrentBlockHash))
	copy(originalHash, b.CurrentBlockHash)

	b.CalculateHash()

	for i := range originalHash {
		if originalHash[i] != b.CurrentBlockHash[i] {
			return false
		}
	}

	return true
}
