# Blockchain CLI Usage Guide

Đây là hướng dẫn sử dụng CLI cho blockchain project của bạn.

## 🔧 Cài Đặt và Build

```bash
# Clone project và build
go build -o cli.exe ./cmd/main.go
```

## 📝 Các Lệnh Cơ Bản

### 1. Xem trợ giúp

```bash
./cli.exe help
```

### 2. Tạo ví (wallet) mới

```bash
./cli.exe create
```

- Tạo cặp khóa ECDSA mới
- Lưu private key vào `user_key.json`
- Hiển thị địa chỉ ví

### 3. Gửi giao dịch

```bash
./cli.exe send <địa_chỉ_nhận> <số_tiền>
```

Ví dụ:

```bash
./cli.exe send 437c6e08e2fc87d08d056b8db9fc174fe003560d 25.5
```

### 4. Demo Alice & Bob 🚀

```bash
./cli.exe demo
```

Chạy demo hoàn chỉnh:

- Tạo ví cho Alice và Bob
- Alice gửi 50 coins cho Bob
- Bob gửi 20 coins lại cho Alice
- Tạo 2 blocks với chữ ký và Merkle Tree

### 5. Khởi tạo blockchain

```bash
./cli.exe init
```

### 6. Kiểm tra số dư (đang phát triển)

```bash
./cli.exe balance <địa_chỉ>
```

### 7. Liệt kê blocks (đang phát triển)

```bash
./cli.exe blocks
```

## 📊 Ví Dụ Workflow

### Tạo ví và gửi tiền:

```bash
# 1. Tạo ví của bạn
./cli.exe create

# 2. Tạo ví khác (hoặc dùng address từ demo)
./cli.exe demo  # Lấy Bob's address từ đây

# 3. Gửi tiền
./cli.exe send 437c6e08e2fc87d08d056b8db9fc174fe003560d 100.0
```

### Demo đầy đủ:

```bash
# Chạy demo Alice & Bob
./cli.exe demo
```

## 🔐 Bảo Mật

- Private key được lưu trong `user_key.json` - **BẢO MẬT FILE NÀY!**
- Mỗi giao dịch được ký bằng ECDSA
- Chữ ký được xác thực trước khi tạo block
- Merkle Tree đảm bảo tính toàn vẹn của block

## 📁 Cấu Trúc Dữ Liệu

```
blockchain-go/
├── cli.exe                 # CLI executable
├── user_key.json          # Private key của bạn (GIỮ BÍ MẬT!)
├── blockchain_data/       # Blockchain chính
├── demo_blockchain/       # Blockchain cho demo
└── ...
```

## 🎯 Tính Năng Đã Thực Hiện

✅ **ECDSA Digital Signatures**

- Tạo và quản lý cặp khóa
- Ký và xác thực giao dịch
- Address generation

✅ **LevelDB Storage**

- Lưu trữ blocks persistently
- Key-value database
- Error handling

✅ **Merkle Tree**

- Xây dựng Merkle Tree từ transactions
- Merkle Root cho mỗi block
- Integrity verification

✅ **Blockchain Structure**

- Block linking với previous hash
- Transaction validation
- Block creation với validator

✅ **CLI Interface**

- User-friendly commands
- Alice & Bob demo
- Error handling và help

## 🔄 Demo Output Mẫu

```bash
$ ./cli.exe demo
🚀 Running Alice & Bob Demo...

👩 Creating Alice's wallet...
Alice Address: 4d47ace7bbcdde1ec0dda61bf0600f3c22221dbc

👨 Creating Bob's wallet...
Bob Address: 437c6e08e2fc87d08d056b8db9fc174fe003560d

💰 Alice sends 50.0 coins to Bob...
✅ Transaction signature verified

💰 Bob sends 20.0 coins back to Alice...
✅ Transaction signature verified

📦 Creating Block 1...
✅ Block 1 created: Hash=d9050ddddd56fb958e1e3a7e7f3386ef90d62b36726896ec561e808072664d94
   Transactions: 1
   Merkle Root: 9250bad8341649f09e8bdf0b48135750b6ce51dcb6ccbc446dbaae035053e66c

📦 Creating Block 2...
✅ Block 2 created: Hash=ce3a142a4d34cad68bd82d8abe31e8d575e386e9c326ff34319d65ee5920fd89
   Transactions: 1
   Merkle Root: 04aecb25e8117cb1b2b87f7e820235783814b7208c893832d692e3aa39d4137b
   Previous Block: d9050ddddd56fb958e1e3a7e7f3386ef90d62b36726896ec561e808072664d94

🎉 Demo completed successfully!
Summary:
- Alice sent 50.0 coins to Bob
- Bob sent 20.0 coins back to Alice
- 2 blocks created with valid signatures and Merkle Trees
```

## 🚀 Tính Năng Có Thể Mở Rộng

- [ ] Balance tracking (UTXO hoặc account-based)
- [ ] Block explorer UI
- [ ] Network/P2P communication
- [ ] Consensus mechanism
- [ ] Smart contracts
- [ ] Mining/staking rewards
