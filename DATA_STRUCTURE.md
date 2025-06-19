# 💾 Blockchain Data Structure Documentation

## 📖 Mục Lục

1. [Tổng Quan Data Storage](#tổng-quan-data-storage)
2. [LevelDB Database Structure](#leveldb-database-structure)
3. [Key-Value Mapping](#key-value-mapping)
4. [File Structure Chi Tiết](#file-structure-chi-tiết)
5. [Code Implementation](#code-implementation)
6. [Data Flow](#data-flow)

---

## 🎯 Tổng Quan Data Storage

Blockchain của bạn sử dụng **LevelDB** - một key-value database để lưu trữ:

- **Blocks** - Các khối chứa transactions
- **Index mapping** - Mapping từ block index → block hash
- **Metadata** - Thông tin database

### Code Location:

```go
// pkg/storage/leveldb.go - Line 15-25
func NewLevelDB(path string) (*LevelDB, error) {
    db, err := leveldb.OpenFile(path, nil)  // ← Tạo/mở database folder
    if err != nil {
        return nil, fmt.Errorf("failed to open leveldb: %w", err)
    }
    return &LevelDB{db: db}, nil
}
```

---

## 🗂️ LevelDB Database Structure

### Folder Structure:

```
demo_blockchain/          ← Database directory
├── LOCK                 ← Process lock file
├── CURRENT              ← Active manifest pointer
├── MANIFEST-000000      ← Database metadata
├── LOG                  ← Operation logs
└── 000001.log          ← Write-ahead log (actual data)
```

### Code Location:

```go
// cmd/main.go - Line 140
validator, err := validator.NewValidatorNode("./demo_blockchain")

// pkg/validator/node.go - Line 20-25
func NewValidatorNode(dbPath string) (*ValidatorNode, error) {
    storage, err := storage.NewLevelDB(dbPath)  // ← Tạo folder này
    if err != nil {
        return nil, fmt.Errorf("failed to create storage: %w", err)
    }
    // ...
}
```

---

## 🔑 Key-Value Mapping

LevelDB lưu data dưới dạng **key-value pairs**:

### 1. Block Storage

```
Key:   [block_hash]           (32 bytes SHA256)
Value: [serialized_block]     (JSON của Block struct)
```

### 2. Index Mapping

```
Key:   "index_0", "index_1", "index_2"...
Value: [block_hash]           (32 bytes)
```

### Code Implementation:

```go
// pkg/storage/leveldb.go - Line 35-55
func (ldb *LevelDB) SaveBlock(block *blockchain.Block) error {
    // 1. Serialize block thành JSON
    blockBytes, err := json.Marshal(block)
    if err != nil {
        return fmt.Errorf("failed to marshal block: %w", err)
    }

    // 2. Lưu với key = block hash
    key := block.CurrentBlockHash  // ← SHA256 hash làm key
    if err := ldb.db.Put(key, blockBytes, nil); err != nil {
        return fmt.Errorf("failed to save block: %w", err)
    }

    // 3. Lưu index mapping: "index_0" → block_hash
    indexKey := []byte(fmt.Sprintf("index_%d", block.Index))
    return ldb.db.Put(indexKey, key, nil)
}
```

---

## 📄 File Structure Chi Tiết

### 1. **LOCK File**

```
File: demo_blockchain/LOCK
Content: (empty file)
Purpose: Prevent concurrent access
```

**Code tạo ra:**

```go
// Khi gọi leveldb.OpenFile() - internal LevelDB code
// Tự động tạo LOCK file để prevent multiple processes
```

### 2. **CURRENT File**

```
File: demo_blockchain/CURRENT
Content: MANIFEST-000000
Purpose: Points to active manifest file
```

### 3. **MANIFEST-000000 File**

```
File: demo_blockchain/MANIFEST-000000
Content: Binary metadata
Contains:
- Comparator: leveldb.BytewiseComparator
- Database version info
- Level structure metadata
```

### 4. **LOG File**

```
File: demo_blockchain/LOG
Content: Text logs
Example:
=============== Jun 19, 2025 (+07) ===============
15:49:25.268762 db@open opening
15:49:25.269796 version@stat F·[] S·0B[] Sc·[]
15:49:25.271868 db@open done T·3.1059ms
```

**Giải thích log entries:**

- `db@open opening` - Database being opened
- `version@stat F·[] S·0B[]` - No files, 0 bytes
- `db@open done T·3.1059ms` - Opened in 3.1ms

### 5. **000001.log File (Write-Ahead Log)**

```
File: demo_blockchain/000001.log
Content: Binary data
Contains: Actual block data before compaction
```

---

## 💻 Code Implementation Details

### Block Data Structure:

```go
// pkg/blockchain/block.go - Line 8-15
type Block struct {
    Index             int            `json:"index"`              // Block number
    Timestamp         int64          `json:"timestamp"`          // Unix timestamp
    Transactions      []*Transaction `json:"transactions"`       // List of txs
    MerkleRoot        []byte         `json:"merkle_root"`        // Merkle tree root
    PreviousBlockHash []byte         `json:"previous_block_hash"` // Previous block hash
    CurrentBlockHash  []byte         `json:"current_block_hash"`  // This block hash
}
```

### Serialized Block Example:

```json
{
  "index": 0,
  "timestamp": 1640995200,
  "transactions": [
    {
      "sender": "4d47ace7bbcdde1ec0dda61bf0600f3c22221dbc",
      "receiver": "437c6e08e2fc87d08d056b8db9fc174fe003560d",
      "amount": 50.0,
      "timestamp": 1640995200,
      "signature": "1a2b3c4d5e6f7890abcdef..."
    }
  ],
  "merkle_root": "9250bad8341649f09e8bdf0b48135750b6ce51dcb6ccbc446dbaae035053e66c",
  "previous_block_hash": null,
  "current_block_hash": "d9050ddddd56fb958e1e3a7e7f3386ef90d62b36726896ec561e808072664d94"
}
```

### Save Process Code:

```go
// pkg/validator/node.go - Line 65-85
func (vn *ValidatorNode) CreateBlock(transactions []*blockchain.Transaction) (*blockchain.Block, error) {
    // ... validation code ...

    // Tạo block mới
    newBlock := &blockchain.Block{
        Index:             vn.currentIndex,
        Timestamp:         time.Now().Unix(),
        Transactions:      transactions,
        MerkleRoot:        merkleTree.Root.Data,      // ← Merkle root
        PreviousBlockHash: vn.getLastBlockHash(),     // ← Chain linking
    }

    // Tính hash cho block
    newBlock.CurrentBlockHash = newBlock.Hash()       // ← Block hash

    // Lưu vào database
    if err := vn.storage.SaveBlock(newBlock); err != nil {  // ← Save to LevelDB
        return nil, fmt.Errorf("failed to save block: %w", err)
    }

    return newBlock, nil
}
```

---

## 🔄 Data Flow

### 1. Demo Workflow:

```
cli.exe demo
    ↓
runAliceBobDemo() [cmd/main.go:140]
    ↓
validator.NewValidatorNode("./demo_blockchain") [pkg/validator/node.go:20]
    ↓
storage.NewLevelDB(dbPath) [pkg/storage/leveldb.go:15]
    ↓
leveldb.OpenFile(path, nil) ← Creates demo_blockchain/ folder
```

### 2. Block Creation Flow:

```
Alice Transaction Created [cmd/main.go:165]
    ↓
validator.CreateBlock([tx1]) [pkg/validator/node.go:65]
    ↓
storage.SaveBlock(newBlock) [pkg/storage/leveldb.go:35]
    ↓
db.Put(blockHash, blockJSON, nil) ← Writes to 000001.log
    ↓
db.Put("index_0", blockHash, nil) ← Index mapping
```

### 3. File Creation Timeline:

```
Step 1: leveldb.OpenFile()
├── Creates demo_blockchain/ folder
├── Creates LOCK file (process lock)
├── Creates CURRENT file (manifest pointer)
├── Creates MANIFEST-000000 (metadata)
└── Creates LOG file (operation log)

Step 2: First SaveBlock()
└── Creates 000001.log (write-ahead log with actual data)
```

---

## 🔍 Debugging Data

### View Raw Data:

```go
// Add this function to cmd/main.go for debugging
func debugDatabase() {
    db, err := leveldb.OpenFile("./demo_blockchain", nil)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    fmt.Println("=== DATABASE CONTENTS ===")
    iter := db.NewIterator(nil, nil)
    for iter.Next() {
        key := iter.Key()
        value := iter.Value()

        if strings.HasPrefix(string(key), "index_") {
            fmt.Printf("Index Key: %s → Hash: %x\n", key, value)
        } else {
            fmt.Printf("Block Hash: %x\n", key)
            fmt.Printf("Block Data: %s\n\n", value)
        }
    }
    iter.Release()
}
```

### Read Specific Block:

```go
// pkg/storage/leveldb.go - Line 75-85
func (ldb *LevelDB) GetBlock(hash []byte) (*blockchain.Block, error) {
    data, err := ldb.db.Get(hash, nil)     // ← Read from 000001.log
    if err != nil {
        return nil, fmt.Errorf("failed to get block: %w", err)
    }

    var block blockchain.Block
    if err := json.Unmarshal(data, &block); err != nil {  // ← Deserialize JSON
        return nil, fmt.Errorf("failed to unmarshal block: %w", err)
    }

    return &block, nil
}
```

---

## 📊 Storage Statistics

### After Demo Completion:

```
Blocks Created: 2
Total Keys: 4
├── Block 0 hash → Block 0 data
├── "index_0" → Block 0 hash
├── Block 1 hash → Block 1 data
└── "index_1" → Block 1 hash

File Sizes (approx):
├── LOCK: 0 bytes
├── CURRENT: 15 bytes
├── MANIFEST-000000: ~100 bytes
├── LOG: ~500 bytes
└── 000001.log: ~2KB (contains actual blocks)
```

### Code to Check Stats:

```go
// Add to CLI for monitoring
func showStats() {
    db, _ := leveldb.OpenFile("./demo_blockchain", nil)
    defer db.Close()

    count := 0
    iter := db.NewIterator(nil, nil)
    for iter.Next() {
        count++
    }
    iter.Release()

    fmt.Printf("Total keys in database: %d\n", count)
}
```

---

## ⚡ Performance Notes

### Write Performance:

```go
// pkg/storage/leveldb.go - Write operations
db.Put(key, value, nil)  // O(log N) - uses LSM trees
```

### Read Performance:

```go
// pkg/storage/leveldb.go - Read operations
db.Get(key, nil)         // O(log N) - indexed lookup
```

### Batch Operations:

```go
// For multiple blocks (not implemented yet)
batch := new(leveldb.Batch)
batch.Put(key1, value1)
batch.Put(key2, value2)
db.Write(batch, nil)     // Atomic batch write
```

---

## 🎯 Key Takeaways

1. **LevelDB** tự động quản lý file structure
2. **Blocks** được lưu dưới dạng JSON serialization
3. **Dual indexing**: Hash-based và Index-based lookup
4. **WAL** đảm bảo durability và crash recovery
5. **ACID** properties được LevelDB đảm bảo

### Main Code Files:

- **Storage Logic**: `pkg/storage/leveldb.go`
- **Block Definition**: `pkg/blockchain/block.go`
- **Save Process**: `pkg/validator/node.go`
- **Demo Creation**: `cmd/main.go`

**🎉 Đây là cách blockchain data được lưu trữ và tổ chức trong project của bạn!**
