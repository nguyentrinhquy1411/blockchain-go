package validator

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/nguyentrinhquy1411/blockchain-go/pkg/blockchain"
	"github.com/nguyentrinhquy1411/blockchain-go/pkg/p2p"
	"github.com/nguyentrinhquy1411/blockchain-go/pkg/storage"
)

// ValidatorNode updated for P2P network and consensus
type ValidatorNode struct {
	NodeID       string
	IsLeader     bool
	Peers        []string
	Blockchain   *blockchain.Blockchain
	Storage      *storage.LevelDB
	Server       *p2p.BlockchainServer // Keep existing functionality
	blockStorage *storage.BlockStorage
	// transactionPool []*blockchain.Transaction // Pool để tích lũy transactions - commented out due to import issue
}

// NewValidatorNode tạo validator node mới với P2P capabilities
func NewValidatorNode() (*ValidatorNode, error) {
	// Get configuration from environment
	nodeID := os.Getenv("NODE_ID")
	if nodeID == "" {
		nodeID = "node1" // default
	}

	isLeaderStr := os.Getenv("IS_LEADER")
	isLeader, _ := strconv.ParseBool(isLeaderStr)

	peersStr := os.Getenv("PEERS")
	var peers []string
	if peersStr != "" {
		peers = strings.Split(peersStr, ",")
	}

	// Initialize storage
	dbPath := fmt.Sprintf("data/%s", nodeID)
	storage, err := storage.NewLevelDB(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage: %w", err)
	}

	// Initialize blockchain
	blockchain, err := blockchain.NewBlockchain(storage)
	if err != nil {
		return nil, fmt.Errorf("failed to create blockchain: %w", err)
	}
	// Initialize P2P server
	server := p2p.NewBlockchainServer(nodeID, blockchain, storage, peers, isLeader)

	return &ValidatorNode{
		NodeID:     nodeID,
		IsLeader:   isLeader,
		Peers:      peers,
		Blockchain: blockchain,
		Storage:    storage,
		Server:     server,
	}, nil
}

// Legacy constructor for backward compatibility
func NewValidatorNodeLegacy(dbPath string) (*ValidatorNode, error) {
	blockStorage, err := storage.NewBlockStorage(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage: %w", err)
	}
	return &ValidatorNode{
		blockStorage: blockStorage,
	}, nil
}

// Close đóng validator node (legacy)
func (vn *ValidatorNode) CloseLegacy() error {
	if vn.blockStorage != nil {
		return vn.blockStorage.Close()
	}
	return nil
}

// CreateBlock tạo block mới từ transactions (core functionality)
func (vn *ValidatorNode) CreateBlock(transactions []*blockchain.Transaction) (*blockchain.Block, error) {
	// Lấy previous block hash nếu có
	var prevHash []byte
	latestIndex, err := vn.blockStorage.GetLatestIndex()
	if err == nil && latestIndex >= 0 {
		prevBlock, err := vn.blockStorage.GetBlockByIndex(latestIndex)
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
	if err := vn.blockStorage.SaveBlock(newBlock); err != nil {
		return nil, fmt.Errorf("failed to save block: %w", err)
	}

	// Lưu index mapping
	if err := vn.blockStorage.StoreBlockByIndex(newBlock); err != nil {
		return nil, fmt.Errorf("failed to store block index: %w", err)
	}

	return newBlock, nil
}

// AddTransaction thêm transaction vào pool và tự động tạo block khi đủ 5 transactions
// Commented out due to transaction pool issues - will be reimplemented later
/*
func (vn *ValidatorNode) AddTransaction(tx *blockchain.Transaction) (*blockchain.Block, error) {
	// Implementation will be added later when transaction pool is fixed
	return nil, fmt.Errorf("not implemented yet")
}
*/

// GetBlock lấy block từ LevelDB theo hash
func (vn *ValidatorNode) GetBlock(hash []byte) (*blockchain.Block, error) {
	if vn.blockStorage != nil {
		return vn.blockStorage.GetBlock(hash)
	}
	// TODO: Implement for new storage
	return nil, fmt.Errorf("not implemented for new storage")
}

// Start starts the validator node P2P server
func (vn *ValidatorNode) Start() error {
	log.Printf("Starting validator node %s (Leader: %v)", vn.NodeID, vn.IsLeader)
	log.Printf("Peers: %v", vn.Peers)

	// Start P2P server
	port := "50051" // default port
	return vn.Server.StartServer(port)
}

// Stop stops the validator node
func (vn *ValidatorNode) Stop() error {
	if vn.Storage != nil {
		return vn.Storage.Close()
	}
	return nil
}

// Close alias for Stop for backward compatibility
func (vn *ValidatorNode) Close() error {
	return vn.Stop()
}
