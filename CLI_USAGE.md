# Blockchain CLI Usage Guide

Đây là hướng dẫn sử dụng CLI cho blockchain project của bạn thông qua `run.bat`.

## 📚 Tài Liệu Liên Quan

- **[DATA_STRUCTURE.md](DATA_STRUCTURE.md)** - Chi tiết về cấu trúc dữ liệu và LevelDB
- **[BLOCKCHAIN_EXPLAINED.md](BLOCKCHAIN_EXPLAINED.md)** - Giải thích code từng phần
- **[BLOCKCHAIN_CHEATSHEET.md](BLOCKCHAIN_CHEATSHEET.md)** - Tóm tắt các khái niệm
- **[QUICK_START.md](QUICK_START.md)** - Hướng dẫn học blockchain
- **[CODE_REFERENCE.md](CODE_REFERENCE.md)** - Bản đồ tham chiếu code

## 🔧 Cài Đặt và Build

### ⚠️ Quy Tắc Quan Trọng

**CHỈ sử dụng `./run.bat` - KHÔNG gọi trực tiếp `./cli.exe`!**

```bash
# Build project
./run.bat build

# Xem tất cả commands
./run.bat
```

### Kiểm tra cài đặt

```bash
# Xem danh sách commands có sẵn
./run.bat

# Build và test demo
./run.bat demo
```

## 🚀 Quick Start

### 1. Build Project

```bash
./run.bat build
```

### 2. Chạy Demo Alice & Bob

```bash
./run.bat demo
```

### 3. Tạo Wallet Riêng Biệt

```bash
./run.bat create-alice
./run.bat create-bob
```

### 4. Thực hiện giao dịch Alice → Bob

```bash
./run.bat alice-to-bob 100.0
```

### 5. Dọn dẹp workspace

```bash
./run.bat clean
```

## 📝 Các Lệnh run.bat

### 🏗️ Build & Management

```bash
./run.bat build    # Build CLI executable
./run.bat clean    # Xóa tất cả files được tạo
./run.bat help     # Hiển thị CLI help
```

### 👤 Wallet Management

```bash
./run.bat create        # Tạo wallet chung (user_key.json)
./run.bat create-alice  # Tạo Alice wallet (alice_key.json)
./run.bat create-bob    # Tạo Bob wallet (bob_key.json)
```

### 💸 Transactions

```bash
./run.bat alice-to-bob <amount>  # Alice gửi tiền cho Bob
./run.bat send <address> <amount>  # Gửi từ wallet chung

# Ví dụ:
./run.bat alice-to-bob 75.5
./run.bat send 437c6e08e2fc87d08d056b8db9fc174fe003560d 50.0
```

### 🎮 Demo & Testing

```bash
./run.bat demo  # Chạy demo Alice & Bob đầy đủ
./run.bat init  # Khởi tạo blockchain
```

## 📊 Workflows Thực Tế

### 🎯 Workflow 1: Alice & Bob Individual

```bash
# 1. Tạo wallets riêng biệt
./run.bat create-alice
./run.bat create-bob

# 2. Alice gửi tiền cho Bob
./run.bat alice-to-bob 100.0

# 3. Kiểm tra files được tạo
ls *.json          # alice_key.json, bob_key.json
ls *blockchain*/   # blockchain_data/
```

### 🎯 Workflow 2: Demo Nhanh (Recommended)

```bash
# Chạy demo hoàn chỉnh một lần
./run.bat demo

# Files được tạo:
# - alice_key.json (Alice's private key)
# - bob_key.json (Bob's private key)
# - demo_blockchain/ (blockchain database)
# - cli.exe (executable)
```

### 🎯 Workflow 3: Personal Wallet

```bash
# 1. Tạo wallet cá nhân
./run.bat create

# 2. Lấy địa chỉ từ demo Alice/Bob
./run.bat demo  # Copy Bob's address từ output

# 3. Gửi tiền từ wallet chính
./run.bat send 1e410ff495f3f0328e817bb9cae934c93b5dcbb0 50.0
```

### 🎯 Workflow 4: Development Cycle

```bash
# Development workflow
./run.bat build    # Build thay đổi
./run.bat demo     # Test functionality
./run.bat clean    # Clean workspace
```

## 🔄 Command Reference

| **Command**              | **Purpose**               | **Files Created**             | **Requirements** |
| ------------------------ | ------------------------- | ----------------------------- | ---------------- |
| `./run.bat build`        | Build CLI executable      | `cli.exe`                     | None             |
| `./run.bat create`       | General wallet            | `user_key.json`               | Build completed  |
| `./run.bat create-alice` | Alice wallet              | `alice_key.json`              | Build completed  |
| `./run.bat create-bob`   | Bob wallet                | `bob_key.json`                | Build completed  |
| `./run.bat alice-to-bob` | Alice→Bob transaction     | `blockchain_data/`            | Alice & Bob keys |
| `./run.bat send`         | Generic transaction       | `blockchain_data/`            | `user_key.json`  |
| `./run.bat demo`         | Full Alice & Bob demo     | All keys + `demo_blockchain/` | None             |
| `./run.bat init`         | Initialize blockchain     | `blockchain_data/`            | Build completed  |
| `./run.bat help`         | Show CLI help             | None                          | Build completed  |
| `./run.bat clean`        | Clean all generated files | None (deletes files)          | None             |

## 📈 Demo Output Example

```bash
$ ./run.bat demo
Building CLI...
✅ Build successful!
Running Alice & Bob Demo...

🚀 Running Alice & Bob Demo...

👩 Creating Alice's wallet...
💾 Saved private key to alice_key.json
Alice Address: 09cc9eb4b3bd7807e795486597af429cd495482c

👨 Creating Bob's wallet...
💾 Saved private key to bob_key.json
Bob Address: 1e410ff495f3f0328e817bb9cae934c93b5dcbb0

💰 Alice sends 50.0 coins to Bob...
✅ Transaction signature verified

💰 Bob sends 20.0 coins back to Alice...
✅ Transaction signature verified

📦 Creating Block 1...
✅ Block 1 created: Hash=d3f695a7b3a9a4d9e9def8ac7699a5509eec90f1...
   Transactions: 1
   Merkle Root: a551803d092d447e5974fd4bf8d70351527db10a4e89b6e8...

📦 Creating Block 2...
✅ Block 2 created: Hash=b8eea640eefdca9d5b2e8737d91320e66e9942b9...
   Transactions: 1
   Merkle Root: fc62b9da2a5dcab5dccbec8381a613ad3cfda10ceb3d7e17...
   Previous Block: d3f695a7b3a9a4d9e9def8ac7699a5509eec90f1...

🎉 Demo completed successfully!
Summary:
- Alice sent 50.0 coins to Bob
- Bob sent 20.0 coins back to Alice
- 2 blocks created with valid signatures and Merkle Trees

📁 Files created:
- alice_key.json (Alice's private key)
- bob_key.json (Bob's private key)
- demo_blockchain/ (blockchain database)

💡 Next steps:
- Use './run.bat alice-to-bob 75.5' for more transactions
- Use './run.bat clean' to reset everything
```

## 🔐 Bảo Mật & Files

### 📁 Key Files (🔒 Bảo Mật Cao!)

```
user_key.json    # Wallet chung - PRIVATE KEY!
alice_key.json   # Alice's private key - BẢO MẬT!
bob_key.json     # Bob's private key - BẢO MẬT!
```

**⚠️ QUAN TRỌNG:** Các file này chứa private keys - không chia sẻ!

### 📂 Database Folders

```
blockchain_data/    # Blockchain chính (init, send)
demo_blockchain/    # Demo blockchain (demo)
├── LOCK           # Database lock
├── CURRENT        # Manifest pointer
├── MANIFEST-*     # Database metadata
├── LOG           # Operation logs
└── *.log         # Block data
```

### 🧹 Clean Commands

```bash
# Xóa tất cả files generated
./run.bat clean

# Xóa từng loại file (manual)
del *.json                                    # Xóa keys
rmdir /s blockchain_data demo_blockchain      # Xóa databases
del cli.exe                                   # Xóa executable
```

## 💾 Data Structure Deep Dive

### Block Structure

```json
{
  "index": 0,
  "timestamp": 1640995200,
  "transactions": [
    {
      "sender": "09cc9eb4b3bd7807e795486597af429cd495482c",
      "receiver": "1e410ff495f3f0328e817bb9cae934c93b5dcbb0",
      "amount": 50.0,
      "timestamp": 1640995200,
      "signature": "a551803d092d447e5974fd4bf8d70351527db10a..."
    }
  ],
  "merkle_root": "9250bad8341649f09e8bdf0b48135750b6ce51dcb6c...",
  "previous_block_hash": null,
  "current_block_hash": "d3f695a7b3a9a4d9e9def8ac7699a5509eec90f1..."
}
```

### Key File Structure

```json
{
  "private_key": "ea1248ae04595cdb89ac570119b75f86a8ac7341666167e31ad667d6e7b83cb2",
  "public_key_x": "64f9d6a1d1d819b2261dc4b3890447a9a3a141f8f02b2cc926bd05d8d6e6582c",
  "public_key_y": "badd58fe34c6dc5d7b300941fa00ae0083a908cce3525a8c4294ac87f936cd4a"
}
```

**📖 Chi tiết: Xem [DATA_STRUCTURE.md](DATA_STRUCTURE.md)**

## 🚀 Advanced Usage

### Batch Operations

```bash
# Setup complete environment
./run.bat clean && ./run.bat build
./run.bat create-alice
./run.bat create-bob
./run.bat alice-to-bob 100.0
./run.bat alice-to-bob 50.0
```

### Multiple Transactions

```bash
# Multiple Alice → Bob transactions
./run.bat alice-to-bob 25.0
./run.bat alice-to-bob 30.0
./run.bat alice-to-bob 45.0
```

### Mixed Wallet Usage

```bash
# Create personal wallet và Alice/Bob
./run.bat create
./run.bat create-alice
./run.bat create-bob

# Get Bob's address từ demo
./run.bat demo | grep "Bob Address"

# Send từ personal wallet to Bob
./run.bat send 1e410ff495f3f0328e817bb9cae934c93b5dcbb0 75.0
```

## 🐛 Troubleshooting

### Common Issues & Solutions

**❌ "Build failed"**

```bash
# Check Go installation
go version
go mod tidy
./run.bat build
```

**❌ "Key file not found"**

```bash
# Recreate missing keys
./run.bat create-alice
./run.bat create-bob
# Hoặc recreate all
./run.bat clean
./run.bat demo
```

**❌ "Database locked"**

```bash
# Close other CLI instances
# Check task manager for cli.exe processes
# Clean và rebuild
./run.bat clean
./run.bat build
```

**❌ "Command not found"**

```bash
# Ensure CLI is built
./run.bat build
# Check if cli.exe exists
ls cli.exe
```

**❌ "Invalid signature"**

```bash
# Key file may be corrupted
del alice_key.json bob_key.json
./run.bat create-alice
./run.bat create-bob
```

### Debug Commands

```bash
# Check generated files
ls *.json *.exe
ls -la *blockchain*/

# Check run.bat help
./run.bat

# Verbose demo
./run.bat demo | more
```

## 🎯 Tính Năng Đã Thực Hiện

### ✅ Core Blockchain Features

- **ECDSA Signatures** - Digital signatures cho transactions
- **LevelDB Storage** - Persistent blockchain database
- **Merkle Tree** - Block integrity verification
- **Block Chain** - Linked blocks với previous hashes
- **Transaction Validation** - Signature verification

### ✅ Wallet Management

- **Individual Wallets** - Alice, Bob, và general wallets
- **Key Persistence** - Save/load keys từ JSON files
- **Address Generation** - ECDSA public key → address
- **Key Security** - Private key protection

### ✅ CLI Interface

- **run.bat Integration** - Wrapper script cho tất cả operations
- **User-friendly Commands** - Intuitive command structure
- **Comprehensive Help** - Built-in documentation
- **Error Handling** - Clear error messages
- **Demo System** - Complete Alice & Bob workflow

### ✅ Development Tools

- **Build System** - Integrated Go build
- **Clean System** - Remove generated files
- **Demo Mode** - End-to-end testing
- **Batch Scripts** - Windows-friendly automation

## 🎓 Learning Path

### 🔰 Beginner (30 phút)

1. `./run.bat demo` - Hiểu basic workflow
2. `./run.bat help` - Tìm hiểu commands
3. Đọc [QUICK_START.md](QUICK_START.md)
4. Test create-alice, create-bob

### 🎯 Intermediate (1 giờ)

1. Tạo wallets riêng biệt
2. Thực hiện transactions manual
3. Đọc [BLOCKCHAIN_EXPLAINED.md](BLOCKCHAIN_EXPLAINED.md)
4. Hiểu cấu trúc database

### 🚀 Advanced (2+ giờ)

1. Study [DATA_STRUCTURE.md](DATA_STRUCTURE.md)
2. Examine code trong [CODE_REFERENCE.md](CODE_REFERENCE.md)
3. Customize và extend features
4. Implement new commands

## 🔜 Tính Năng Có Thể Mở Rộng

### Planned Features

- [ ] **Balance Tracking** - UTXO hoặc account-based model
- [ ] **Multi-signature** - Multiple signatures per transaction
- [ ] **Block Explorer** - Web UI để view blockchain
- [ ] **P2P Network** - Distributed blockchain network
- [ ] **Smart Contracts** - Programmable transactions
- [ ] **Mining/Consensus** - Proof of Work hoặc Proof of Stake

### Extension Points

- [ ] **REST API** - HTTP endpoints cho blockchain operations
- [ ] **GraphQL Interface** - Query language cho blockchain data
- [ ] **Mobile App** - React Native hoặc Flutter frontend
- [ ] **Hardware Wallets** - Ledger/Trezor integration
- [ ] **Cross-chain** - Bridge với other blockchains

## 📊 Ví Dụ Use Cases

### Personal Learning

```bash
# Học blockchain basics
./run.bat demo
# Study output và files created
```

### Development Testing

```bash
# Test new features
./run.bat clean
./run.bat build
./run.bat demo
```

### Educational Demo

```bash
# Classroom demonstration
./run.bat clean
./run.bat demo | tee demo_output.txt
# Show demo_output.txt to students
```

### Production Setup

```bash
# Prepare for deployment
./run.bat build
# Copy cli.exe (DO NOT copy *.json files!)
```

## 🎉 Conclusion

Bạn đã có một complete blockchain system với:

- ✅ **ECDSA Digital Signatures**
- ✅ **LevelDB Persistent Storage**
- ✅ **Merkle Tree Verification**
- ✅ **Alice & Bob Workflow**
- ✅ **User-friendly CLI via run.bat**
- ✅ **Comprehensive Documentation**

**💡 Next Steps:**

1. Explore các tài liệu khác trong project
2. Customize code theo nhu cầu
3. Implement additional features
4. Share và contribute back!

**🔗 Quick Links:**

- [DATA_STRUCTURE.md](DATA_STRUCTURE.md) - Technical details
- [BLOCKCHAIN_EXPLAINED.md](BLOCKCHAIN_EXPLAINED.md) - Code explanation
- [QUICK_START.md](QUICK_START.md) - Learning guide
- [CODE_REFERENCE.md](CODE_REFERENCE.md) - Code map

**🌟 Remember: Always use `./run.bat` - never call `./cli.exe` directly!**
