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

type ValidatorNode struct {
	NodeID       string
	IsLeader     bool
	Peers        []string
	Blockchain   *blockchain.Blockchain
	Storage      *storage.LevelDB
	Server       *p2p.BlockchainServer
	blockStorage *storage.BlockStorage
}

func NewValidatorNode() (*ValidatorNode, error) {
	nodeID := os.Getenv("NODE_ID")
	if nodeID == "" {
		nodeID = "node1"
	}

	isLeaderStr := os.Getenv("IS_LEADER")
	isLeader, _ := strconv.ParseBool(isLeaderStr)

	peersStr := os.Getenv("PEERS")
	var peers []string
	if peersStr != "" {
		peers = strings.Split(peersStr, ",")
	}

	dbPath := fmt.Sprintf("data/%s", nodeID)
	storage, err := storage.NewLevelDB(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage: %w", err)
	}

	blockchain, err := blockchain.NewBlockchain(storage)
	if err != nil {
		return nil, fmt.Errorf("failed to create blockchain: %w", err)
	}

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

func NewValidatorNodeLegacy(dbPath string) (*ValidatorNode, error) {
	blockStorage, err := storage.NewBlockStorage(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage: %w", err)
	}
	return &ValidatorNode{
		blockStorage: blockStorage,
	}, nil
}

func (vn *ValidatorNode) CloseLegacy() error {
	if vn.blockStorage != nil {
		return vn.blockStorage.Close()
	}
	return nil
}

func (vn *ValidatorNode) CreateBlock(transactions []*blockchain.Transaction) (*blockchain.Block, error) {
	var prevHash []byte
	latestIndex, err := vn.blockStorage.GetLatestIndex()
	if err == nil && latestIndex >= 0 {
		prevBlock, err := vn.blockStorage.GetBlockByIndex(latestIndex)
		if err == nil {
			prevHash = prevBlock.CurrentBlockHash
		}
	}

	newBlock := blockchain.NewBlock(latestIndex+1, transactions, prevHash)

	if !newBlock.IsValid() {
		return nil, fmt.Errorf("block invalid - Merkle Tree verification failed")
	}

	if err := vn.blockStorage.SaveBlock(newBlock); err != nil {
		return nil, fmt.Errorf("failed to save block: %w", err)
	}

	if err := vn.blockStorage.StoreBlockByIndex(newBlock); err != nil {
		return nil, fmt.Errorf("failed to store block index: %w", err)
	}

	return newBlock, nil
}

func (vn *ValidatorNode) GetBlock(hash []byte) (*blockchain.Block, error) {
	if vn.blockStorage != nil {
		return vn.blockStorage.GetBlock(hash)
	}
	return nil, fmt.Errorf("not implemented for new storage")
}

func (vn *ValidatorNode) Start() error {
	log.Printf("Starting validator node %s (Leader: %v)", vn.NodeID, vn.IsLeader)
	log.Printf("Peers: %v", vn.Peers)

	port := "50051"
	return vn.Server.StartServer(port)
}

func (vn *ValidatorNode) Stop() error {
	if vn.Storage != nil {
		return vn.Storage.Close()
	}
	return nil
}

func (vn *ValidatorNode) Close() error {
	return vn.Stop()
}
