# 🧪 Blockchain Testing Guide

Complete testing guide for all blockchain requirements validation.

## 🚀 Quick Start Flow

```bash
# 1. Setup project
.\setup.bat

# 2. Run complete system test
.\blockchain.exe test

# 3. Test individual features
.\blockchain.exe demo

# 4. Test 3-node consensus
.\test-consensus.bat

# 5. Clean up after testing
.\cleanup.bat
```

## Testing All Requirements

### ✅ 1. ECDSA Digital Signatures

```bash
# Test individual wallet creation
.\blockchain.exe create-alice
.\blockchain.exe create-bob

# Test transaction signing
.\blockchain.exe alice-to-bob 25.5

# Full demo with verification
.\blockchain.exe demo
```

**Verification**:

- Alice & Bob ECDSA key pairs generated ✅
- Transactions signed with private keys ✅
- Signatures verified with public keys ✅

### ✅ 2. LevelDB Storage & Merkle Tree

```bash
# Demo includes Merkle Tree validation
.\blockchain.exe demo
```

**Verification**:

- Blocks stored in LevelDB ✅
- Merkle Tree calculated from transactions ✅
- Block integrity validated ✅
- Previous block hash chaining ✅

### ✅ 3. 3-Node Consensus (Docker)

```bash
# Start 3-node network
docker-compose up -d

# View consensus logs
docker-compose logs -f

# Stop network
docker-compose down
```

**Verification**:

- Node1 (Leader) on port 50051 ✅
- Node2 (Follower) on port 50052 ✅
- Node3 (Follower) on port 50053 ✅
- Leader-Follower consensus mechanism ✅

### ✅ 4. Node Recovery

```bash
# Test node recovery
.\test-consensus.bat

# Stop specific node
docker stop blockchain-node2

# Restart node (auto-reconnects)
docker start blockchain-node2
```

**Verification**:

- Node disconnection handling ✅
- Automatic reconnection ✅
- Block synchronization ✅
- Consensus participation resume ✅

## Test Output Validation

### Core Features Test Results

- ✅ ECDSA digital signatures - PASSED
- ✅ LevelDB persistent storage - PASSED
- ✅ Merkle Tree validation - PASSED
- ✅ Block creation & chaining - PASSED

### Demo Output Validation

- Alice & Bob wallets created with ECDSA keys
- Transactions signed and verified
- 2 blocks created with valid Merkle Trees
- Data persisted in LevelDB
- Block chaining with previous hash links

### Files Created During Test

- `alice_key.json` - Alice's ECDSA private key
- `bob_key.json` - Bob's ECDSA private key
- `demo_blockchain/` - LevelDB database directory

## Troubleshooting

### Build Issues

```bash
go mod tidy
go build -o blockchain.exe .\cmd\main.go
```

### Docker Issues

```bash
docker-compose down --remove-orphans -v
docker-compose build --no-cache
docker-compose up -d
```

### Clean Start

```bash
.\cleanup.bat
.\setup.bat
```
