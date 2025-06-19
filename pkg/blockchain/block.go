package blockchain

import (
	"crypto/sha256"
	"encoding/json"
	"time"
)

// Block theo yêu cầu: danh sách giao dịch, Merkle Root, PreviousBlockHash, CurrentBlockHash
type Block struct {
	Index             int            `json:"index"`
	Timestamp         int64          `json:"timestamp"`
	Transactions      []*Transaction `json:"transactions"`
	MerkleRoot        []byte         `json:"merkle_root"`
	PreviousBlockHash []byte         `json:"previous_block_hash"`
	CurrentBlockHash  []byte         `json:"current_block_hash"`
}

// NewBlock tạo block mới khi đủ 5 giao dịch
func NewBlock(index int, transactions []*Transaction, prevHash []byte) *Block {
	// Kiểm tra số lượng transactions tối thiểu
	if len(transactions) < 5 {
		return nil // Không tạo block nếu chưa đủ 5 giao dịch
	}

	block := &Block{
		Index:             index,
		Timestamp:         time.Now().Unix(),
		Transactions:      transactions,
		PreviousBlockHash: prevHash,
	}

	// Tính Merkle Root từ transactions
	block.calculateMerkleRoot()

	// Tính Current Block Hash
	block.calculateHash()

	return block
}

// CalculateMerkleRoot tính Merkle Root từ transactions (exported method)
func (b *Block) CalculateMerkleRoot() {
	if len(b.Transactions) == 0 {
		b.MerkleRoot = []byte{}
		return
	}

	// Lấy hash của tất cả transactions
	var txHashes [][]byte
	for _, tx := range b.Transactions {
		hash, err := tx.Hash()
		if err != nil {
			continue
		}
		txHashes = append(txHashes, hash)
	}

	// Tạo Merkle Tree và lấy root
	merkleTree := NewMerkleTree(txHashes)
	b.MerkleRoot = merkleTree.GetRoot()
}

// calculateMerkleRoot tính Merkle Root từ transactions (unexported method for internal use)
func (b *Block) calculateMerkleRoot() {
	b.CalculateMerkleRoot()
}

// CalculateHash tính Current Block Hash (exported method)
func (b *Block) CalculateHash() {
	// Tạo struct chỉ chứa data cần hash (không bao gồm CurrentBlockHash)
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

// calculateHash tính Current Block Hash (unexported method for internal use)
func (b *Block) calculateHash() {
	b.CalculateHash()
}

// IsValid kiểm tra tính hợp lệ của block theo yêu cầu
func (b *Block) IsValid() bool {
	// Kiểm tra Merkle Root integrity
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

	// So sánh calculated vs stored Merkle Root
	if len(calculatedRoot) != len(b.MerkleRoot) {
		return false
	}
	for i := range calculatedRoot {
		if calculatedRoot[i] != b.MerkleRoot[i] {
			return false
		}
	}

	// Kiểm tra Current Block Hash integrity
	originalHash := make([]byte, len(b.CurrentBlockHash))
	copy(originalHash, b.CurrentBlockHash)

	b.calculateHash()

	for i := range originalHash {
		if originalHash[i] != b.CurrentBlockHash[i] {
			return false
		}
	}

	return true
}
