# 📚 Hướng Dẫn Hiểu Code Blockchain - Dành Cho Người Mới Bắt Đầu

## 📖 Mục Lục

1. [Tổng Quan Dự Án](#tổng-quan-dự-án)
2. [Cấu Trúc Thư Mục](#cấu-trúc-thư-mục)
3. [Giải Thích Chi Tiết Từng File](#giải-thích-chi-tiết-từng-file)
4. [Khái Niệm Blockchain Cơ Bản](#khái-niệm-blockchain-cơ-bản)
5. [Luồng Hoạt Động](#luồng-hoạt-động)

---

## 🎯 Tổng Quan Dự Án

Dự án này là một **blockchain đơn giản** được viết bằng Go, mô phỏng việc chuyển tiền giữa Alice và Bob. Các tính năng chính:

- **ECDSA Digital Signatures** - Ký số giao dịch
- **Merkle Tree** - Xác thực tính toàn vẹn
- **LevelDB** - Lưu trữ dữ liệu
- **CLI Interface** - Giao diện dòng lệnh

---

## 📁 Cấu Trúc Thư Mục

```
blockchain-go/
├── cmd/main.go           # CLI chính - điểm vào chương trình
├── pkg/
│   ├── blockchain/       # Logic blockchain cốt lõi
│   │   ├── block.go     # Định nghĩa Block
│   │   ├── transaction.go # Định nghĩa Transaction
│   │   └── merkle.go    # Merkle Tree
│   ├── wallet/          # Quản lý ví và chữ ký
│   │   ├── key.go       # Tạo và quản lý khóa
│   │   └── sign.go      # Ký và xác thực
│   ├── storage/         # Lưu trữ dữ liệu
│   │   └── leveldb.go   # Tương tác LevelDB
│   ├── validator/       # Node xác thực
│   │   └── node.go      # Logic validator
│   └── utils/           # Tiện ích
│       └── hash.go      # Hàm hash
└── blockchain_data/     # Dữ liệu blockchain
```

---

## 🔍 Giải Thích Chi Tiết Từng File

### 1. 🔑 pkg/wallet/key.go - Quản Lý Khóa Mật Mã

```go
package wallet

import (
    "crypto/ecdsa"
    "crypto/elliptic"
    "crypto/rand"
    "crypto/sha256"
    "fmt"
)
```

**Giải thích import:**

- `crypto/ecdsa`: Thuật toán ký số ECDSA (Elliptic Curve Digital Signature Algorithm)
- `crypto/elliptic`: Đường cong elliptic cho mật mã học
- `crypto/rand`: Tạo số ngẫu nhiên an toàn
- `crypto/sha256`: Hàm hash SHA-256

```go
func GenerateKeyPair() (*ecdsa.PrivateKey, error) {
    // Tạo private key sử dụng đường cong P256
    privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
    if err != nil {
        return nil, fmt.Errorf("failed to generate key pair: %w", err)
    }
    return privKey, nil
}
```

**Giải thích:**

- **Private Key**: Khóa riêng tư, dùng để ký giao dịch
- **Public Key**: Khóa công khai, được tính từ private key
- **P256**: Đường cong elliptic tiêu chuẩn, an toàn và hiệu quả
- **rand.Reader**: Nguồn entropy an toàn cho việc tạo khóa

```go
func PublicKeyToAddress(pubKey *ecdsa.PublicKey) []byte {
    // Chuyển public key thành bytes
    pubKeyBytes := append(pubKey.X.Bytes(), pubKey.Y.Bytes()...)

    // Hash để tạo address
    hash := sha256.Sum256(pubKeyBytes)
    return hash[:20] // Lấy 20 bytes đầu làm address
}
```

**Giải thích:**

- **Address**: Địa chỉ ví, tính từ public key
- **X, Y**: Tọa độ của điểm trên đường cong elliptic
- **SHA-256**: Hash công khai để tạo address duy nhất
- **[:20]**: Lấy 20 bytes = 160 bits cho address

---

### 2. ✍️ pkg/wallet/sign.go - Ký Số và Xác Thực

```go
func SignTransaction(tx *Transaction, privKey *ecdsa.PrivateKey) error {
    // Tạo hash của transaction (không bao gồm signature)
    txHash := tx.Hash()

    // Ký hash bằng private key
    r, s, err := ecdsa.Sign(rand.Reader, privKey, txHash)
    if err != nil {
        return fmt.Errorf("failed to sign transaction: %w", err)
    }

    // Lưu chữ ký (r và s ghép lại)
    tx.Signature = append(r.Bytes(), s.Bytes()...)
    return nil
}
```

**Giải thích ECDSA Signing:**

- **Hash trước khi ký**: Chỉ ký hash, không ký toàn bộ dữ liệu
- **r, s**: Hai thành phần của chữ ký ECDSA
- **Deterministic**: Cùng private key + hash → cùng chữ ký
- **Non-repudiation**: Chỉ chủ private key mới tạo được chữ ký

```go
func VerifyTransaction(tx *Transaction, pubKey *ecdsa.PublicKey) bool {
    // Tạo lại hash của transaction
    txHash := tx.Hash()

    // Tách r và s từ signature
    sigLen := len(tx.Signature)
    r := new(big.Int).SetBytes(tx.Signature[:sigLen/2])
    s := new(big.Int).SetBytes(tx.Signature[sigLen/2:])

    // Xác thực chữ ký
    return ecdsa.Verify(pubKey, txHash, r, s)
}
```

**Giải thích Verification:**

- **Public Key Verification**: Dùng public key để xác thực
- **Hash Matching**: Hash phải giống khi ký
- **Mathematical Proof**: Chứng minh toán học chủ private key đã ký

---

### 3. 💰 pkg/blockchain/transaction.go - Giao Dịch

```go
type Transaction struct {
    Sender    []byte  `json:"sender"`     // Địa chỉ người gửi
    Receiver  []byte  `json:"receiver"`   // Địa chỉ người nhận
    Amount    float64 `json:"amount"`     // Số tiền
    Timestamp int64   `json:"timestamp"`  // Thời gian
    Signature []byte  `json:"signature"`  // Chữ ký
}
```

**Giải thích các trường:**

- **Sender/Receiver**: Địa chỉ ví (20 bytes từ public key hash)
- **Amount**: Số tiền chuyển (trong thực tế dùng integer cho độ chính xác)
- **Timestamp**: Unix timestamp, chống replay attack
- **Signature**: Chữ ký ECDSA chứng minh quyền sở hữu

```go
func (t *Transaction) Hash() []byte {
    // Tạo bản sao transaction không có signature
    txCopy := *t
    txCopy.Signature = nil

    // Serialize thành JSON
    data, _ := json.Marshal(txCopy)

    // Hash bằng SHA-256
    hash := sha256.Sum256(data)
    return hash[:]
}
```

**Tại sao loại bỏ Signature khi hash?**

- **Circular Dependency**: Signature được tính từ hash, nếu hash bao gồm signature sẽ bị lặp vô hạn
- **Immutable Content**: Hash chỉ phản ánh nội dung giao dịch, không phải chữ ký

---

### 4. 🌳 pkg/blockchain/merkle.go - Cây Merkle

```go
type MerkleTree struct {
    Root *MerkleNode `json:"root"`
}

type MerkleNode struct {
    Left  *MerkleNode `json:"left"`
    Right *MerkleNode `json:"right"`
    Data  []byte      `json:"data"`
}
```

**Cây Merkle là gì?**

- **Binary Tree**: Cây nhị phân, mỗi node có tối đa 2 con
- **Hash Tree**: Mỗi node chứa hash
- **Leaf Nodes**: Lá cây là hash của từng transaction
- **Root**: Gốc cây là hash tổng hợp của tất cả transactions

```go
func NewMerkleTree(data [][]byte) *MerkleTree {
    if len(data) == 0 {
        return &MerkleTree{Root: nil}
    }

    // Tạo leaf nodes từ data
    var nodes []*MerkleNode
    for _, datum := range data {
        node := &MerkleNode{Data: datum}
        nodes = append(nodes, node)
    }

    // Nếu số lượng lẻ, duplicate node cuối
    if len(nodes)%2 != 0 {
        nodes = append(nodes, nodes[len(nodes)-1])
    }

    // Xây dựng cây từ dưới lên
    for len(nodes) > 1 {
        var level []*MerkleNode

        for i := 0; i < len(nodes); i += 2 {
            // Ghép cặp nodes và hash
            left := nodes[i]
            right := nodes[i+1]

            // Tạo parent node
            parent := &MerkleNode{
                Left:  left,
                Right: right,
                Data:  hashPair(left.Data, right.Data),
            }
            level = append(level, parent)
        }
        nodes = level
    }

    return &MerkleTree{Root: nodes[0]}
}
```

**Tại sao cần Merkle Tree?**

- **Integrity**: Đảm bảo tính toàn vẹn - nếu 1 transaction thay đổi, root hash sẽ khác
- **Efficiency**: Chỉ cần lưu root hash thay vì tất cả transaction hashes
- **Proof**: Có thể chứng minh 1 transaction có trong block mà không cần toàn bộ block

---

### 5. 🧱 pkg/blockchain/block.go - Khối

```go
type Block struct {
    Index             int            `json:"index"`
    Timestamp         int64          `json:"timestamp"`
    Transactions      []*Transaction `json:"transactions"`
    MerkleRoot        []byte         `json:"merkle_root"`
    PreviousBlockHash []byte         `json:"previous_block_hash"`
    CurrentBlockHash  []byte         `json:"current_block_hash"`
}
```

**Cấu trúc Block:**

- **Index**: Số thứ tự block trong chain
- **Timestamp**: Thời gian tạo block
- **Transactions**: Danh sách giao dịch trong block
- **MerkleRoot**: Hash gốc của Merkle Tree
- **PreviousBlockHash**: Hash của block trước (tạo chain)
- **CurrentBlockHash**: Hash của block hiện tại

```go
func (b *Block) Hash() []byte {
    // Tạo bản sao block không có current hash
    blockCopy := *b
    blockCopy.CurrentBlockHash = nil

    // Serialize và hash
    data, _ := json.Marshal(blockCopy)
    hash := sha256.Sum256(data)
    return hash[:]
}
```

**Blockchain Chain:**

```
Genesis Block → Block 1 → Block 2 → Block 3 → ...
     ↑            ↑         ↑         ↑
   Hash A      Hash B    Hash C    Hash D
              (chứa A)  (chứa B)  (chứa C)
```

---

### 6. 💾 pkg/storage/leveldb.go - Lưu Trữ

```go
type LevelDB struct {
    db *leveldb.DB
}

func NewLevelDB(path string) (*LevelDB, error) {
    db, err := leveldb.OpenFile(path, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to open leveldb: %w", err)
    }
    return &LevelDB{db: db}, nil
}
```

**LevelDB là gì?**

- **Key-Value Store**: Lưu trữ dạng key-value
- **Persistent**: Dữ liệu lưu trên disk, không mất khi tắt máy
- **Ordered**: Keys được sắp xếp tự động
- **Fast**: Tối ưu cho read/write operations

```go
func (ldb *LevelDB) SaveBlock(block *blockchain.Block) error {
    // Serialize block thành JSON
    blockBytes, err := json.Marshal(block)
    if err != nil {
        return fmt.Errorf("failed to marshal block: %w", err)
    }

    // Lưu với key là hash của block
    key := block.CurrentBlockHash
    if err := ldb.db.Put(key, blockBytes, nil); err != nil {
        return fmt.Errorf("failed to save block: %w", err)
    }

    // Lưu mapping index → hash để dễ truy xuất
    indexKey := []byte(fmt.Sprintf("index_%d", block.Index))
    return ldb.db.Put(indexKey, key, nil)
}
```

**Storage Strategy:**

- **Primary Key**: Hash của block → Full block data
- **Index Key**: Index number → Hash (để tìm block theo số thứ tự)
- **Redundant but Fast**: Lưu 2 lần nhưng truy xuất nhanh

---

### 7. ⚡ pkg/validator/node.go - Node Xác Thực

```go
type ValidatorNode struct {
    storage     *storage.LevelDB
    blockchain  []*blockchain.Block
    currentIndex int
}
```

**Validator Node làm gì?**

- **Validate Transactions**: Kiểm tra chữ ký và tính hợp lệ
- **Create Blocks**: Tạo block mới từ transactions
- **Store Blocks**: Lưu blocks vào database
- **Maintain Chain**: Duy trì tính liên tục của blockchain

```go
func (vn *ValidatorNode) CreateBlock(transactions []*blockchain.Transaction) (*blockchain.Block, error) {
    // Xác thực tất cả transactions
    for _, tx := range transactions {
        if !vn.isValidTransaction(tx) {
            return nil, fmt.Errorf("invalid transaction")
        }
    }

    // Tạo Merkle Tree từ transaction hashes
    var txHashes [][]byte
    for _, tx := range transactions {
        txHashes = append(txHashes, tx.Hash())
    }
    merkleTree := blockchain.NewMerkleTree(txHashes)

    // Tạo block mới
    newBlock := &blockchain.Block{
        Index:             vn.currentIndex,
        Timestamp:         time.Now().Unix(),
        Transactions:      transactions,
        MerkleRoot:        merkleTree.Root.Data,
        PreviousBlockHash: vn.getLastBlockHash(),
    }

    // Tính hash cho block
    newBlock.CurrentBlockHash = newBlock.Hash()

    // Lưu block
    if err := vn.storage.SaveBlock(newBlock); err != nil {
        return nil, err
    }

    // Cập nhật local blockchain
    vn.blockchain = append(vn.blockchain, newBlock)
    vn.currentIndex++

    return newBlock, nil
}
```

**Block Creation Process:**

1. **Validate**: Kiểm tra từng transaction
2. **Merkle**: Tạo Merkle Tree từ transaction hashes
3. **Link**: Liên kết với block trước bằng hash
4. **Hash**: Tính hash cho block hiện tại
5. **Store**: Lưu vào database
6. **Update**: Cập nhật local chain

---

### 8. 🖥️ cmd/main.go - CLI Interface

```go
func runAliceBobDemo() {
    // Tạo validator
    validator, err := validator.NewValidatorNode("./demo_blockchain")

    // Tạo Alice's wallet
    alicePriv, _ := wallet.GenerateKeyPair()
    aliceAddr := wallet.PublicKeyToAddress(&alicePriv.PublicKey)

    // Tạo Bob's wallet
    bobPriv, _ := wallet.GenerateKeyPair()
    bobAddr := wallet.PublicKeyToAddress(&bobPriv.PublicKey)

    // Alice gửi tiền cho Bob
    tx1 := &blockchain.Transaction{
        Sender:    aliceAddr,
        Receiver:  bobAddr,
        Amount:    50.0,
        Timestamp: time.Now().Unix(),
    }
    wallet.SignTransaction(tx1, alicePriv)

    // Tạo block
    block1, _ := validator.CreateBlock([]*blockchain.Transaction{tx1})
}
```

**Demo Flow:**

1. **Setup**: Tạo validator node và wallets
2. **Create Transaction**: Alice → Bob
3. **Sign**: Alice ký transaction bằng private key
4. **Verify**: Validator xác thực chữ ký
5. **Block**: Tạo block chứa transaction
6. **Store**: Lưu block vào LevelDB

---

## 🔗 Khái Niệm Blockchain Cơ Bản

### 1. **Digital Signature (Chữ Ký Số)**

```
Private Key → Sign Transaction → Signature
Public Key + Signature + Transaction → Verify → True/False
```

### 2. **Hash Function (Hàm Hash)**

```
Input: "Hello World"
SHA-256: a591a6d40bf420404a011733cfb7b190d62c65bf0bcda32b57b277d9ad9f146e
```

**Tính chất:**

- **Deterministic**: Cùng input → cùng output
- **Fixed Size**: Output luôn 256 bits
- **Avalanche Effect**: Thay đổi 1 bit input → thay đổi hoàn toàn output

### 3. **Merkle Tree**

```
      Root Hash
     /         \
  Hash12      Hash34
  /    \      /    \
Hash1 Hash2 Hash3 Hash4
  |     |     |     |
 Tx1   Tx2   Tx3   Tx4
```

### 4. **Blockchain Structure**

```
Block 0 (Genesis)
├── Index: 0
├── Transactions: []
├── PrevHash: null
└── Hash: ABC123

Block 1
├── Index: 1
├── Transactions: [Tx1, Tx2]
├── PrevHash: ABC123 ← Links to Block 0
└── Hash: DEF456

Block 2
├── Index: 2
├── Transactions: [Tx3]
├── PrevHash: DEF456 ← Links to Block 1
└── Hash: GHI789
```

---

## 🔄 Luồng Hoạt Động

### 1. **Tạo Wallet**

```go
// Tạo private key
privKey := GenerateKeyPair()

// Tính public key (tự động từ private key)
pubKey := privKey.PublicKey

// Tạo address từ public key
address := SHA256(pubKey.X + pubKey.Y)[:20]
```

### 2. **Tạo Transaction**

```go
// Tạo transaction
tx := Transaction{
    Sender: aliceAddress,
    Receiver: bobAddress,
    Amount: 50.0,
    Timestamp: now(),
}

// Alice ký transaction
signature := ECDSA_Sign(alicePrivateKey, SHA256(tx))
tx.Signature = signature
```

### 3. **Xác Thực Transaction**

```go
// Validator kiểm tra
txHash := SHA256(tx without signature)
isValid := ECDSA_Verify(alicePublicKey, txHash, tx.Signature)
```

### 4. **Tạo Block**

```go
// Thu thập transactions
transactions := [tx1, tx2, tx3]

// Tạo Merkle Tree
merkleRoot := BuildMerkleTree(transactions)

// Tạo block
block := Block{
    Index: currentIndex,
    Transactions: transactions,
    MerkleRoot: merkleRoot,
    PreviousHash: lastBlock.Hash,
    Timestamp: now(),
}

// Tính hash cho block
block.Hash = SHA256(block)
```

### 5. **Lưu Trữ**

```go
// Lưu vào LevelDB
levelDB.Put(block.Hash, serialize(block))
levelDB.Put("index_" + block.Index, block.Hash)
```

---

## 🎓 Bài Tập Thực Hành

### Bài 1: **Hash Experiment**

```bash
# Thử thay đổi 1 ký tự và xem hash thay đổi như thế nào
echo "Hello World" | sha256sum
echo "Hello World!" | sha256sum
```

### Bài 2: **Signature Verification**

```go
// Tạo transaction và thử thay đổi amount
// Xem signature còn valid không?
tx.Amount = 100.0 // Thay đổi sau khi ký
isValid := VerifyTransaction(tx, publicKey) // Should be false
```

### Bài 3: **Merkle Tree Analysis**

```go
// So sánh Merkle Root trước và sau khi thay đổi 1 transaction
```

---

## 🚀 Mở Rộng Kiến Thức

### 1. **Consensus Mechanisms**

- **Proof of Work**: Mining, difficulty adjustment
- **Proof of Stake**: Validators, staking rewards
- **PBFT**: Byzantine fault tolerance

### 2. **Advanced Features**

- **Smart Contracts**: Programmable transactions
- **UTXO Model**: Unspent transaction outputs
- **Lightning Network**: Layer 2 scaling

### 3. **Security Considerations**

- **51% Attack**: Majority control
- **Double Spending**: Prevent same coin spending twice
- **Replay Attacks**: Timestamp protection

---

## 📚 Tài Liệu Tham Khảo

1. **Bitcoin Whitepaper** - Satoshi Nakamoto
2. **Mastering Bitcoin** - Andreas Antonopoulos
3. **Ethereum Yellowpaper** - Gavin Wood
4. **Cryptography Engineering** - Ferguson, Schneier, Kohno

---

_📝 File này được tạo để giúp người mới bắt đầu hiểu rõ cách blockchain hoạt động thông qua code thực tế. Hãy đọc từng phần một cách cẩn thận và thực hành các ví dụ!_
