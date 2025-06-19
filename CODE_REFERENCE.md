# 🗺️ Code Reference Map - Where to Find What

## 📋 Quick Navigation

| **Tính Năng**             | **File Location**               | **Function/Type**        | **Line** |
| ------------------------- | ------------------------------- | ------------------------ | -------- |
| **LevelDB Setup**         | `pkg/storage/leveldb.go`        | `NewLevelDB()`           | 15-25    |
| **Save Block**            | `pkg/storage/leveldb.go`        | `SaveBlock()`            | 35-55    |
| **Get Block**             | `pkg/storage/leveldb.go`        | `GetBlock()`             | 75-85    |
| **Block Structure**       | `pkg/blockchain/block.go`       | `type Block`             | 8-15     |
| **Transaction Structure** | `pkg/blockchain/transaction.go` | `type Transaction`       | 8-15     |
| **Merkle Tree**           | `pkg/blockchain/merkle.go`      | `NewMerkleTree()`        | 20-60    |
| **ECDSA Key Generation**  | `pkg/wallet/key.go`             | `GenerateKeyPair()`      | 12-20    |
| **Sign Transaction**      | `pkg/wallet/sign.go`            | `SignTransaction()`      | 12-25    |
| **Verify Transaction**    | `pkg/wallet/sign.go`            | `VerifyTransaction()`    | 35-50    |
| **Validator Node**        | `pkg/validator/node.go`         | `NewValidatorNode()`     | 20-30    |
| **Create Block**          | `pkg/validator/node.go`         | `CreateBlock()`          | 65-85    |
| **CLI Main**              | `cmd/main.go`                   | `main()`                 | 85-105   |
| **Alice Bob Demo**        | `cmd/main.go`                   | `runAliceBobDemo()`      | 140-220  |
| **Key Serialization**     | `cmd/main.go`                   | `saveKey()`, `loadKey()` | 25-75    |

---

## 🔍 Data Flow Code Locations

### 1. Demo Creation Flow:

```
cmd/main.go:140          runAliceBobDemo()
    ↓
pkg/validator/node.go:20  NewValidatorNode("./demo_blockchain")
    ↓
pkg/storage/leveldb.go:15 NewLevelDB(dbPath)
    ↓
leveldb.OpenFile()       Creates demo_blockchain/ folder
```

### 2. Transaction Flow:

```
cmd/main.go:165          Create Transaction struct
    ↓
pkg/wallet/sign.go:12    SignTransaction(tx, alicePriv)
    ↓
pkg/wallet/sign.go:35    VerifyTransaction(tx, &alicePriv.PublicKey)
    ↓
pkg/validator/node.go:65 CreateBlock([]*Transaction{tx})
    ↓
pkg/storage/leveldb.go:35 SaveBlock(newBlock)
```

### 3. Block Creation Flow:

```
pkg/blockchain/merkle.go:20   NewMerkleTree(txHashes)
    ↓
pkg/blockchain/block.go:45    newBlock.Hash()
    ↓
pkg/storage/leveldb.go:40     db.Put(blockHash, blockJSON, nil)
    ↓
pkg/storage/leveldb.go:50     db.Put("index_0", blockHash, nil)
```

---

## 🔧 Key Code Snippets

### Database Initialization:

```go
// pkg/storage/leveldb.go:15
func NewLevelDB(path string) (*LevelDB, error) {
    db, err := leveldb.OpenFile(path, nil)  // ← Creates folder structure
    if err != nil {
        return nil, fmt.Errorf("failed to open leveldb: %w", err)
    }
    return &LevelDB{db: db}, nil
}
```

### Block Storage:

```go
// pkg/storage/leveldb.go:35
func (ldb *LevelDB) SaveBlock(block *blockchain.Block) error {
    blockBytes, err := json.Marshal(block)  // ← Serialize to JSON
    if err != nil {
        return fmt.Errorf("failed to marshal block: %w", err)
    }

    key := block.CurrentBlockHash              // ← Use hash as key
    if err := ldb.db.Put(key, blockBytes, nil); err != nil {
        return fmt.Errorf("failed to save block: %w", err)
    }

    indexKey := []byte(fmt.Sprintf("index_%d", block.Index))  // ← Index mapping
    return ldb.db.Put(indexKey, key, nil)
}
```

### Merkle Tree Construction:

```go
// pkg/blockchain/merkle.go:25
func NewMerkleTree(data [][]byte) *MerkleTree {
    var nodes []*MerkleNode
    for _, datum := range data {
        node := &MerkleNode{Data: datum}     // ← Create leaf nodes
        nodes = append(nodes, node)
    }

    for len(nodes) > 1 {
        var level []*MerkleNode
        for i := 0; i < len(nodes); i += 2 {
            left := nodes[i]
            right := nodes[i+1]

            parent := &MerkleNode{              // ← Build tree bottom-up
                Data:  hashPair(left.Data, right.Data),
                Left:  left,
                Right: right,
            }
            level = append(level, parent)
        }
        nodes = level
    }

    return &MerkleTree{Root: nodes[0]}
}
```

### ECDSA Signing:

```go
// pkg/wallet/sign.go:12
func SignTransaction(tx *Transaction, privKey *ecdsa.PrivateKey) error {
    txHash := tx.Hash()                    // ← Hash without signature

    r, s, err := ecdsa.Sign(rand.Reader, privKey, txHash)  // ← ECDSA sign
    if err != nil {
        return fmt.Errorf("failed to sign transaction: %w", err)
    }

    tx.Signature = append(r.Bytes(), s.Bytes()...)        // ← Store signature
    return nil
}
```

---

## 📁 File Structure Reference

```
cmd/
├── main.go              ← CLI interface, demo functions
pkg/
├── blockchain/
│   ├── block.go         ← Block struct and methods
│   ├── transaction.go   ← Transaction struct and hashing
│   └── merkle.go        ← Merkle tree implementation
├── wallet/
│   ├── key.go           ← ECDSA key generation
│   └── sign.go          ← Transaction signing/verification
├── storage/
│   └── leveldb.go       ← Database operations
├── validator/
│   └── node.go          ← Block creation and validation
└── utils/
    └── hash.go          ← Utility hashing functions
```

---

## 🎯 Common Debug Points

### Check Database Contents:

```go
// Add to cmd/main.go for debugging
func debugDB() {
    db, _ := leveldb.OpenFile("./demo_blockchain", nil)
    defer db.Close()

    iter := db.NewIterator(nil, nil)
    for iter.Next() {
        fmt.Printf("Key: %x\nValue: %s\n\n", iter.Key(), iter.Value())
    }
}
```

### Verify Block Hash:

```go
// pkg/blockchain/block.go:45
func (b *Block) Hash() []byte {
    blockCopy := *b
    blockCopy.CurrentBlockHash = nil  // ← Exclude self from hash

    data, _ := json.Marshal(blockCopy)
    hash := sha256.Sum256(data)
    return hash[:]
}
```

### Check Signature Validation:

```go
// pkg/wallet/sign.go:35
func VerifyTransaction(tx *Transaction, pubKey *ecdsa.PublicKey) bool {
    txHash := tx.Hash()  // ← Must be same hash as when signing

    sigLen := len(tx.Signature)
    r := new(big.Int).SetBytes(tx.Signature[:sigLen/2])
    s := new(big.Int).SetBytes(tx.Signature[sigLen/2:])

    return ecdsa.Verify(pubKey, txHash, r, s)
}
```

---

## 🚀 Extension Points

### Add New CLI Command:

```go
// cmd/main.go:95 - Add to switch statement
case "newcommand":
    newCommandFunction()
```

### Add Block Validation:

```go
// pkg/validator/node.go:70 - Add to CreateBlock()
if !vn.validateBlock(newBlock) {
    return nil, fmt.Errorf("invalid block")
}
```

### Add Database Query:

```go
// pkg/storage/leveldb.go - Add new method
func (ldb *LevelDB) GetBlockByIndex(index int) (*blockchain.Block, error) {
    indexKey := []byte(fmt.Sprintf("index_%d", index))
    hash, err := ldb.db.Get(indexKey, nil)
    if err != nil {
        return nil, err
    }
    return ldb.GetBlock(hash)
}
```

---

**🎯 Use this reference to quickly navigate and understand the codebase!**
