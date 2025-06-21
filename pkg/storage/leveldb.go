package storage

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/nguyentrinhquy1411/blockchain-go/pkg/blockchain"
	"github.com/syndtr/goleveldb/leveldb"
)

// BlockStorage quản lý việc lưu trữ blocks trong LevelDB
type BlockStorage struct {
	db *leveldb.DB
}

// NewBlockStorage tạo BlockStorage mới
func NewBlockStorage(dbPath string) (*BlockStorage, error) {
	db, err := leveldb.OpenFile(dbPath, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open leveldb: %w", err)
	}

	return &BlockStorage{db: db}, nil
}

// Close đóng database connection
func (bs *BlockStorage) Close() error {
	return bs.db.Close()
}

// SaveBlock lưu block vào LevelDB với block hash làm key
func (bs *BlockStorage) SaveBlock(block *blockchain.Block) error {
	blockBytes, err := json.Marshal(block)
	if err != nil {
		return fmt.Errorf("failed to marshal block: %w", err)
	}

	// Sử dụng CurrentBlockHash làm key
	return bs.db.Put(block.CurrentBlockHash, blockBytes, nil)
}

// GetBlock lấy block từ LevelDB bằng hash
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

// StoreBlockByIndex lưu block với index làm key (để tìm theo height)
func (bs *BlockStorage) StoreBlockByIndex(block *blockchain.Block) error {
	key := "height_" + strconv.Itoa(block.Index)

	// Chỉ lưu hash, không lưu toàn bộ block để tiết kiệm space
	return bs.db.Put([]byte(key), block.CurrentBlockHash, nil)
}

// GetBlockByIndex lấy block theo index/height
func (bs *BlockStorage) GetBlockByIndex(index int) (*blockchain.Block, error) {
	key := "height_" + strconv.Itoa(index)
	hash, err := bs.db.Get([]byte(key), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get block hash by index: %w", err)
	}

	// Sau đó lấy block bằng hash
	return bs.GetBlock(hash)
}

// GetLatestIndex lấy index của block mới nhất
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

// HasBlock kiểm tra xem block có tồn tại không (theo hash)
// func (bs *BlockStorage) HasBlock(hash []byte) bool {
// 	exists, err := bs.db.Has(hash, nil)
// 	if err != nil {
// 		// Log error nhưng return false để an toàn
// 		return false
// 	}
// 	return exists
// }

// HasBlockByIndex kiểm tra xem block có tồn tại không (theo index)
// func (bs *BlockStorage) HasBlockByIndex(index int) bool {
// 	key := "height_" + strconv.Itoa(index)
// 	exists, err := bs.db.Has([]byte(key), nil)
// 	if err != nil {
// 		return false
// 	}
// 	return exists
// }

// DeleteBlock xóa block (nếu cần)
// func (bs *BlockStorage) DeleteBlock(hash []byte) error {
// 	return bs.db.Delete(hash, nil)
// }

// LevelDB simple wrapper that implements blockchain.Storage interface
type LevelDB struct {
	db *leveldb.DB
}

// NewLevelDB creates a new LevelDB instance
func NewLevelDB(dbPath string) (*LevelDB, error) {
	db, err := leveldb.OpenFile(dbPath, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open leveldb: %w", err)
	}

	return &LevelDB{db: db}, nil
}

// Get implements blockchain.Storage interface
func (ldb *LevelDB) Get(key string) ([]byte, error) {
	return ldb.db.Get([]byte(key), nil)
}

// Put implements blockchain.Storage interface
func (ldb *LevelDB) Put(key string, value []byte) error {
	return ldb.db.Put([]byte(key), value, nil)
}

// Close closes the database
func (ldb *LevelDB) Close() error {
	return ldb.db.Close()
}
