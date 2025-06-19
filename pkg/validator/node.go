package validator

import (
	"fmt"

	"github.com/nguyentrinhquy1411/blockchain-go/pkg/blockchain"
	"github.com/nguyentrinhquy1411/blockchain-go/pkg/storage"
)

// ValidatorNode đơn giản - chỉ tập trung vào yêu cầu đề bài:
// - Lưu trữ blocks trong LevelDB
// - Xác thực bằng Merkle Tree
// - Transaction pool để tích lũy transactions
type ValidatorNode struct {
	storage         *storage.BlockStorage
	transactionPool []*blockchain.Transaction // Pool để tích lũy transactions
}

// NewValidatorNode tạo validator node mới
func NewValidatorNode(dbPath string) (*ValidatorNode, error) {
	storage, err := storage.NewBlockStorage(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage: %w", err)
	}

	return &ValidatorNode{
		storage:         storage,
		transactionPool: make([]*blockchain.Transaction, 0),
	}, nil
}

// Close đóng validator node
func (vn *ValidatorNode) Close() error {
	return vn.storage.Close()
}

// CreateBlock tạo block mới từ transactions (core functionality)
func (vn *ValidatorNode) CreateBlock(transactions []*blockchain.Transaction) (*blockchain.Block, error) {
	// Lấy previous block hash nếu có
	var prevHash []byte
	latestIndex, err := vn.storage.GetLatestIndex()
	if err == nil && latestIndex >= 0 {
		prevBlock, err := vn.storage.GetBlockByIndex(latestIndex)
		if err == nil {
			prevHash = prevBlock.CurrentBlockHash
		}
	}
	// Tạo block mới
	newBlock := blockchain.NewBlock(latestIndex+1, transactions, prevHash)

	// Xác thực bằng Merkle Tree (yêu cầu đề bài)
	if !newBlock.IsValid() {
		return nil, fmt.Errorf("block invalid - Merkle Tree verification failed")
	}

	// Lưu trữ vào LevelDB (yêu cầu đề bài)
	if err := vn.storage.SaveBlock(newBlock); err != nil {
		return nil, fmt.Errorf("failed to save block: %w", err)
	}

	// Lưu index mapping
	if err := vn.storage.StoreBlockByIndex(newBlock); err != nil {
		return nil, fmt.Errorf("failed to store block index: %w", err)
	}

	return newBlock, nil
}

// AddTransaction thêm transaction vào pool và tự động tạo block khi đủ 5 transactions
func (vn *ValidatorNode) AddTransaction(tx *blockchain.Transaction) (*blockchain.Block, error) {
	// Thêm transaction vào pool
	vn.transactionPool = append(vn.transactionPool, tx)

	// Kiểm tra xem đã đủ 5 transactions chưa
	if len(vn.transactionPool) >= 5 {
		// Lấy 5 transactions đầu tiên
		transactionsToProcess := vn.transactionPool[:5]

		// Tạo block
		block, err := vn.CreateBlock(transactionsToProcess)
		if err != nil {
			return nil, err
		}

		// Xóa 5 transactions đã được xử lý khỏi pool
		vn.transactionPool = vn.transactionPool[5:]

		return block, nil
	}

	// Chưa đủ 5 transactions, trả về nil
	return nil, nil
}

// GetBlock lấy block từ LevelDB theo hash
func (vn *ValidatorNode) GetBlock(hash []byte) (*blockchain.Block, error) {
	return vn.storage.GetBlock(hash)
}
