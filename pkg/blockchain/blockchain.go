package blockchain

import (
	"encoding/json"
	"fmt"
)

// Storage interface để tránh import cycle
type Storage interface {
	Get(key string) ([]byte, error)
	Put(key string, value []byte) error
}

type Blockchain struct {
	storage Storage
	genesis *Block
}

func NewBlockchain(storage Storage) (*Blockchain, error) {
	bc := &Blockchain{
		storage: storage,
	}

	// Try to load existing blockchain or create genesis
	if err := bc.loadOrCreateGenesis(); err != nil {
		return nil, err
	}

	return bc, nil
}

func (bc *Blockchain) loadOrCreateGenesis() error {
	// Try to load genesis block
	genesisData, err := bc.storage.Get("genesis")
	if err != nil {
		// Create genesis block
		genesisTransactions := []*Transaction{
			{
				Sender:    []byte("genesis"),
				Receiver:  []byte("alice"),
				Amount:    100.0,
				Timestamp: 0,
			},
		}

		bc.genesis = NewBlock(0, genesisTransactions, []byte(""))

		// Save genesis block
		return bc.saveBlock(bc.genesis, "genesis")
	}

	// Load existing genesis
	if err := json.Unmarshal(genesisData, &bc.genesis); err != nil {
		return fmt.Errorf("failed to unmarshal genesis block: %w", err)
	}

	return nil
}

func (bc *Blockchain) GetLatestBlock() *Block {
	// For now, return genesis. In a full implementation,
	// you'd track the latest block height
	return bc.genesis
}

func (bc *Blockchain) GetBlockByHeight(height int) (*Block, error) {
	if height == 0 {
		return bc.genesis, nil
	}

	// For now, only genesis is implemented
	return nil, fmt.Errorf("block at height %d not found", height)
}

func (bc *Blockchain) GetBlockByHash(hash string) (*Block, error) {
	// Simple implementation - check if it's genesis
	if string(bc.genesis.CurrentBlockHash) == hash {
		return bc.genesis, nil
	}

	return nil, fmt.Errorf("block with hash %s not found", hash)
}

func (bc *Blockchain) AddBlock(block *Block) error {
	// Validate block
	if !block.IsValid() {
		return fmt.Errorf("invalid block")
	}

	// Save block
	key := fmt.Sprintf("block_%d", block.Index)
	return bc.saveBlock(block, key)
}

func (bc *Blockchain) CalculateMerkleRoot(transactions []*Transaction) string {
	if len(transactions) == 0 {
		return ""
	}

	// Get hashes of all transactions
	var txHashes [][]byte
	for _, tx := range transactions {
		hash, err := tx.Hash()
		if err != nil {
			continue
		}
		txHashes = append(txHashes, hash)
	}

	// Create Merkle Tree and get root
	merkleTree := NewMerkleTree(txHashes)
	return string(merkleTree.GetRoot())
}

func (bc *Blockchain) saveBlock(block *Block, key string) error {
	blockData, err := json.Marshal(block)
	if err != nil {
		return fmt.Errorf("failed to marshal block: %w", err)
	}

	return bc.storage.Put(key, blockData)
}
