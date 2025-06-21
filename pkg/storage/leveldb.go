package storage

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/nguyentrinhquy1411/blockchain-go/pkg/blockchain"
	"github.com/syndtr/goleveldb/leveldb"
)

type BlockStorage struct {
	db *leveldb.DB
}

func NewBlockStorage(dbPath string) (*BlockStorage, error) {
	db, err := leveldb.OpenFile(dbPath, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open leveldb: %w", err)
	}

	return &BlockStorage{db: db}, nil
}

func (bs *BlockStorage) Close() error {
	return bs.db.Close()
}

func (bs *BlockStorage) SaveBlock(block *blockchain.Block) error {
	blockBytes, err := json.Marshal(block)
	if err != nil {
		return fmt.Errorf("failed to marshal block: %w", err)
	}

	return bs.db.Put(block.CurrentBlockHash, blockBytes, nil)
}

func (bs *BlockStorage) GetBlock(hash []byte) (*blockchain.Block, error) {
	blockBytes, err := bs.db.Get(hash, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get block: %w", err)
	}

	var block blockchain.Block
	if err := json.Unmarshal(blockBytes, &block); err != nil {
		return nil, fmt.Errorf("failed to unmarshal block: %w", err)
	}

	return &block, nil
}

func (bs *BlockStorage) StoreBlockByIndex(block *blockchain.Block) error {
	key := "height_" + strconv.Itoa(block.Index)
	return bs.db.Put([]byte(key), block.CurrentBlockHash, nil)
}

func (bs *BlockStorage) GetBlockByIndex(index int) (*blockchain.Block, error) {
	key := "height_" + strconv.Itoa(index)
	hash, err := bs.db.Get([]byte(key), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get block hash by index: %w", err)
	}

	return bs.GetBlock(hash)
}

func (bs *BlockStorage) GetLatestIndex() (int, error) {
	iter := bs.db.NewIterator(nil, nil)
	defer iter.Release()

	latestIndex := -1
	for iter.Next() {
		key := string(iter.Key())
		if len(key) > 7 && key[:7] == "height_" {
			index, err := strconv.Atoi(key[7:])
			if err == nil && index > latestIndex {
				latestIndex = index
			}
		}
	}

	if err := iter.Error(); err != nil {
		return -1, fmt.Errorf("iterator error: %w", err)
	}

	return latestIndex, nil
}

type LevelDB struct {
	db *leveldb.DB
}

func NewLevelDB(dbPath string) (*LevelDB, error) {
	db, err := leveldb.OpenFile(dbPath, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open leveldb: %w", err)
	}

	return &LevelDB{db: db}, nil
}

func (ldb *LevelDB) Get(key string) ([]byte, error) {
	return ldb.db.Get([]byte(key), nil)
}

func (ldb *LevelDB) Put(key string, value []byte) error {
	return ldb.db.Put([]byte(key), value, nil)
}

func (ldb *LevelDB) Close() error {
	return ldb.db.Close()
}
