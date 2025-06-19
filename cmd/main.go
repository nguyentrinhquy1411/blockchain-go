package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"os"
	"strconv"
	"time"

	"github.com/nguyentrinhquy1411/blockchain-go/pkg/blockchain"
	"github.com/nguyentrinhquy1411/blockchain-go/pkg/storage"
	"github.com/nguyentrinhquy1411/blockchain-go/pkg/validator"
	"github.com/nguyentrinhquy1411/blockchain-go/pkg/wallet"
)

type KeyData struct {
	PrivateKey string `json:"private_key"`
	PublicKeyX string `json:"public_key_x"`
	PublicKeyY string `json:"public_key_y"`
}

func saveKey(priv *ecdsa.PrivateKey) error {
	keyData := KeyData{
		PrivateKey: hex.EncodeToString(priv.D.Bytes()),
		PublicKeyX: hex.EncodeToString(priv.PublicKey.X.Bytes()),
		PublicKeyY: hex.EncodeToString(priv.PublicKey.Y.Bytes()),
	}

	f, err := os.Create("user_key.json")
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()

	if err := json.NewEncoder(f).Encode(keyData); err != nil {
		return fmt.Errorf("failed to encode key: %w", err)
	}

	fmt.Println("Saved private key to user_key.json")
	return nil
}

func loadKey() (*ecdsa.PrivateKey, error) {
	f, err := os.Open("user_key.json")
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	var keyData KeyData
	if err := json.NewDecoder(f).Decode(&keyData); err != nil {
		return nil, fmt.Errorf("failed to decode key: %w", err)
	}

	// Decode private key
	privKeyBytes, err := hex.DecodeString(keyData.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decode private key: %w", err)
	}

	// Decode public key coordinates
	pubKeyXBytes, err := hex.DecodeString(keyData.PublicKeyX)
	if err != nil {
		return nil, fmt.Errorf("failed to decode public key X: %w", err)
	}

	pubKeyYBytes, err := hex.DecodeString(keyData.PublicKeyY)
	if err != nil {
		return nil, fmt.Errorf("failed to decode public key Y: %w", err)
	}

	// Reconstruct the private key
	priv := &ecdsa.PrivateKey{
		D: new(big.Int).SetBytes(privKeyBytes),
		PublicKey: ecdsa.PublicKey{
			Curve: elliptic.P256(),
			X:     new(big.Int).SetBytes(pubKeyXBytes),
			Y:     new(big.Int).SetBytes(pubKeyYBytes),
		},
	}

	return priv, nil
}

func saveKeyWithName(priv *ecdsa.PrivateKey, filename string) error {
	keyData := KeyData{
		PrivateKey: hex.EncodeToString(priv.D.Bytes()),
		PublicKeyX: hex.EncodeToString(priv.PublicKey.X.Bytes()),
		PublicKeyY: hex.EncodeToString(priv.PublicKey.Y.Bytes()),
	}

	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filename, err)
	}
	defer f.Close()

	if err := json.NewEncoder(f).Encode(keyData); err != nil {
		return fmt.Errorf("failed to encode key: %w", err)
	}

	fmt.Printf("💾 Saved private key to %s\n", filename)
	return nil
}

func loadKeyFromFile(filename string) (*ecdsa.PrivateKey, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filename, err)
	}
	defer f.Close()

	var keyData KeyData
	if err := json.NewDecoder(f).Decode(&keyData); err != nil {
		return nil, fmt.Errorf("failed to decode key: %w", err)
	}

	// Decode private key
	privKeyBytes, err := hex.DecodeString(keyData.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decode private key: %w", err)
	}

	// Decode public key coordinates
	pubKeyXBytes, err := hex.DecodeString(keyData.PublicKeyX)
	if err != nil {
		return nil, fmt.Errorf("failed to decode public key X: %w", err)
	}

	pubKeyYBytes, err := hex.DecodeString(keyData.PublicKeyY)
	if err != nil {
		return nil, fmt.Errorf("failed to decode public key Y: %w", err)
	}

	// Reconstruct the private key
	priv := &ecdsa.PrivateKey{
		D: new(big.Int).SetBytes(privKeyBytes),
		PublicKey: ecdsa.PublicKey{
			Curve: elliptic.P256(),
			X:     new(big.Int).SetBytes(pubKeyXBytes),
			Y:     new(big.Int).SetBytes(pubKeyYBytes),
		},
	}

	return priv, nil
}

func main() {
	args := os.Args
	if len(args) < 2 {
		printUsage()
		return
	}
	switch args[1] {
	case "create":
		createUserKey()
	case "create-alice":
		createAlice()
	case "create-bob":
		createBob()
	case "alice-to-bob":
		aliceToBobTransaction(args)
	case "send":
		sendTransaction(args)
	case "demo":
		runAliceBobDemo()
	case "pool-demo":
		runTransactionPoolDemo()
	case "init":
		initBlockchain()
	case "count":
		checkBlockCount()
	case "help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", args[1])
		printUsage()
	}
}

func printUsage() {
	fmt.Println("Blockchain CLI Usage:")
	fmt.Println("  create          - Create a new wallet key pair")
	fmt.Println("  create-alice    - Create Alice's wallet")
	fmt.Println("  create-bob      - Create Bob's wallet")
	fmt.Println("  alice-to-bob <amount> - Send money from Alice to Bob")
	fmt.Println("  send <to> <amount> - Send money to address")
	fmt.Println("  demo            - Run Alice & Bob demo")
	fmt.Println("  pool-demo       - Demo transaction pool (5 transactions per block)")
	fmt.Println("  init            - Initialize blockchain")
	fmt.Println("  count           - Show blockchain block count")
	fmt.Println("  help            - Show this help message")
}
func createUserKey() {
	priv, err := wallet.GenerateKeyPair()
	if err != nil {
		fmt.Printf("Error generating key pair: %v\n", err)
		return
	}

	if err := saveKey(priv); err != nil {
		fmt.Printf("Error saving key: %v\n", err)
		return
	}

	address := wallet.PublicKeyToAddress(&priv.PublicKey)
	fmt.Printf("✅ Wallet created successfully!\n")
	fmt.Printf("Address: %x\n", address)
}

func createAlice() {
	fmt.Println("👩 Creating Alice's wallet...")

	alicePriv, err := wallet.GenerateKeyPair()
	if err != nil {
		fmt.Printf("Error generating Alice's key pair: %v\n", err)
		return
	}

	if err := saveKeyWithName(alicePriv, "alice_key.json"); err != nil {
		fmt.Printf("Error saving Alice's key: %v\n", err)
		return
	}

	aliceAddr := wallet.PublicKeyToAddress(&alicePriv.PublicKey)
	fmt.Printf("✅ Alice's wallet created successfully!\n")
	fmt.Printf("Alice Address: %x\n", aliceAddr)
	fmt.Printf("💾 Keys saved to: alice_key.json\n")
}

func createBob() {
	fmt.Println("👨 Creating Bob's wallet...")

	bobPriv, err := wallet.GenerateKeyPair()
	if err != nil {
		fmt.Printf("Error generating Bob's key pair: %v\n", err)
		return
	}

	if err := saveKeyWithName(bobPriv, "bob_key.json"); err != nil {
		fmt.Printf("Error saving Bob's key: %v\n", err)
		return
	}

	bobAddr := wallet.PublicKeyToAddress(&bobPriv.PublicKey)
	fmt.Printf("✅ Bob's wallet created successfully!\n")
	fmt.Printf("Bob Address: %x\n", bobAddr)
	fmt.Printf("💾 Keys saved to: bob_key.json\n")
}

func aliceToBobTransaction(args []string) {
	if len(args) < 3 {
		fmt.Println("Usage: cli alice-to-bob <amount>")
		fmt.Println("Example: cli alice-to-bob 25.5")
		return
	}

	// Parse amount
	amount, err := strconv.ParseFloat(args[2], 64)
	if err != nil {
		fmt.Printf("Invalid amount: %v\n", err)
		return
	}

	// Load Alice's key
	fmt.Println("🔑 Loading Alice's key...")
	alicePriv, err := loadKeyFromFile("alice_key.json")
	if err != nil {
		fmt.Printf("Error loading Alice's key: %v\n", err)
		fmt.Println("💡 Please run 'cli create-alice' first to create Alice's wallet")
		return
	}

	// Load Bob's key to get his address
	fmt.Println("🔑 Loading Bob's address...")
	bobPriv, err := loadKeyFromFile("bob_key.json")
	if err != nil {
		fmt.Printf("Error loading Bob's key: %v\n", err)
		fmt.Println("💡 Please run 'cli create-bob' first to create Bob's wallet")
		return
	}

	aliceAddr := wallet.PublicKeyToAddress(&alicePriv.PublicKey)
	bobAddr := wallet.PublicKeyToAddress(&bobPriv.PublicKey)

	fmt.Printf("💸 Alice (%x) sending %.2f coins to Bob (%x)...\n",
		aliceAddr[:8], amount, bobAddr[:8])

	// Create transaction
	tx := &blockchain.Transaction{
		Sender:    aliceAddr,
		Receiver:  bobAddr,
		Amount:    amount,
		Timestamp: time.Now().Unix(),
	}

	// Alice signs the transaction
	fmt.Println("🔏 Alice signing transaction...")
	if err := wallet.SignTransaction(tx, alicePriv); err != nil {
		fmt.Printf("Error signing transaction: %v\n", err)
		return
	}

	// Verify signature
	if !wallet.VerifyTransaction(tx, &alicePriv.PublicKey) {
		fmt.Println("❌ Transaction signature invalid")
		return
	}
	fmt.Println("✅ Transaction signature verified")

	// Create validator and process transaction
	fmt.Println("📦 Creating block...")
	validator, err := validator.NewValidatorNode("./blockchain_data")
	if err != nil {
		fmt.Printf("Error creating validator: %v\n", err)
		return
	}
	defer validator.Close()

	block, err := validator.CreateBlock([]*blockchain.Transaction{tx})
	if err != nil {
		fmt.Printf("Error creating block: %v\n", err)
		return
	}

	fmt.Println("\n🎉 Transaction completed successfully!")
	fmt.Printf("📋 Transaction Details:\n")
	fmt.Printf("   From: Alice (%x)\n", aliceAddr)
	fmt.Printf("   To: Bob (%x)\n", bobAddr)
	fmt.Printf("   Amount: %.2f coins\n", amount)
	fmt.Printf("   Block: %d\n", block.Index)
	fmt.Printf("   Block Hash: %x\n", block.CurrentBlockHash)
	fmt.Printf("   Merkle Root: %x\n", block.MerkleRoot)
}

func sendTransaction(args []string) {
	if len(args) < 4 {
		fmt.Println("Usage: cli send <to_address> <amount>")
		return
	}

	priv, err := loadKey()
	if err != nil {
		fmt.Printf("Error loading key: %v\n", err)
		fmt.Println("Please run 'cli create' first to create a wallet")
		return
	}

	// Parse receiver address
	receiverHex := args[2]
	receiver, err := hex.DecodeString(receiverHex)
	if err != nil {
		fmt.Printf("Invalid receiver address: %v\n", err)
		return
	}

	// Parse amount
	amount, err := strconv.ParseFloat(args[3], 64)
	if err != nil {
		fmt.Printf("Invalid amount: %v\n", err)
		return
	}

	sender := wallet.PublicKeyToAddress(&priv.PublicKey)

	tx := &blockchain.Transaction{
		Sender:    sender,
		Receiver:  receiver,
		Amount:    amount,
		Timestamp: time.Now().Unix(),
	}

	if err := wallet.SignTransaction(tx, priv); err != nil {
		fmt.Printf("Error signing transaction: %v\n", err)
		return
	}

	// Create validator and process transaction
	validator, err := validator.NewValidatorNode("./blockchain_data")
	if err != nil {
		fmt.Printf("Error creating validator: %v\n", err)
		return
	}
	defer validator.Close()

	block, err := validator.CreateBlock([]*blockchain.Transaction{tx})
	if err != nil {
		fmt.Printf("Error creating block: %v\n", err)
		return
	}

	fmt.Printf("✅ Transaction sent successfully!\n")
	fmt.Printf("From: %x\n", sender)
	fmt.Printf("To: %x\n", receiver)
	fmt.Printf("Amount: %.2f\n", amount)
	fmt.Printf("Block: %d (Hash: %x)\n", block.Index, block.CurrentBlockHash)
}

func runAliceBobDemo() {
	fmt.Println("🚀 Running Alice & Bob Demo...")

	// Create validator
	validator, err := validator.NewValidatorNode("./demo_blockchain")
	if err != nil {
		log.Fatal("Failed to create validator:", err)
	}
	defer validator.Close()
	// Create Alice's wallet
	fmt.Println("\n👩 Creating Alice's wallet...")
	alicePriv, err := wallet.GenerateKeyPair()
	if err != nil {
		log.Fatal("Failed to generate Alice's key:", err)
	}
	aliceAddr := wallet.PublicKeyToAddress(&alicePriv.PublicKey)

	// Save Alice's key to file
	if err := saveKeyWithName(alicePriv, "alice_key.json"); err != nil {
		log.Fatal("Failed to save Alice's key:", err)
	}
	fmt.Printf("Alice Address: %x\n", aliceAddr)

	// Create Bob's wallet
	fmt.Println("\n👨 Creating Bob's wallet...")
	bobPriv, err := wallet.GenerateKeyPair()
	if err != nil {
		log.Fatal("Failed to generate Bob's key:", err)
	}
	bobAddr := wallet.PublicKeyToAddress(&bobPriv.PublicKey)

	// Save Bob's key to file
	if err := saveKeyWithName(bobPriv, "bob_key.json"); err != nil {
		log.Fatal("Failed to save Bob's key:", err)
	}
	fmt.Printf("Bob Address: %x\n", bobAddr)

	// Alice sends money to Bob
	fmt.Println("\n💰 Alice sends 50.0 coins to Bob...")
	tx1 := &blockchain.Transaction{
		Sender:    aliceAddr,
		Receiver:  bobAddr,
		Amount:    50.0,
		Timestamp: time.Now().Unix(),
	}

	// Alice signs the transaction
	if err := wallet.SignTransaction(tx1, alicePriv); err != nil {
		log.Fatal("Failed to sign transaction:", err)
	}

	// Verify signature
	if !wallet.VerifyTransaction(tx1, &alicePriv.PublicKey) {
		log.Fatal("Transaction signature invalid")
	}
	fmt.Println("✅ Transaction signature verified")

	// Bob sends money back to Alice
	fmt.Println("\n💰 Bob sends 20.0 coins back to Alice...")
	tx2 := &blockchain.Transaction{
		Sender:    bobAddr,
		Receiver:  aliceAddr,
		Amount:    20.0,
		Timestamp: time.Now().Unix() + 1,
	}

	// Bob signs the transaction
	if err := wallet.SignTransaction(tx2, bobPriv); err != nil {
		log.Fatal("Failed to sign transaction:", err)
	}

	// Verify signature
	if !wallet.VerifyTransaction(tx2, &bobPriv.PublicKey) {
		log.Fatal("Transaction signature invalid")
	}
	fmt.Println("✅ Transaction signature verified")

	// Create first block with Alice's transaction
	fmt.Println("\n📦 Creating Block 1...")
	block1, err := validator.CreateBlock([]*blockchain.Transaction{tx1})
	if err != nil {
		log.Fatal("Failed to create block 1:", err)
	}
	fmt.Printf("✅ Block 1 created: Hash=%x\n", block1.CurrentBlockHash)
	fmt.Printf("   Transactions: %d\n", len(block1.Transactions))
	fmt.Printf("   Merkle Root: %x\n", block1.MerkleRoot)

	// Create second block with Bob's transaction
	fmt.Println("\n📦 Creating Block 2...")
	block2, err := validator.CreateBlock([]*blockchain.Transaction{tx2})
	if err != nil {
		log.Fatal("Failed to create block 2:", err)
	}
	fmt.Printf("✅ Block 2 created: Hash=%x\n", block2.CurrentBlockHash)
	fmt.Printf("   Transactions: %d\n", len(block2.Transactions))
	fmt.Printf("   Merkle Root: %x\n", block2.MerkleRoot)
	fmt.Printf("   Previous Block: %x\n", block2.PreviousBlockHash)
	fmt.Println("\n🎉 Demo completed successfully!")
	fmt.Println("Summary:")
	fmt.Printf("- Alice sent 50.0 coins to Bob\n")
	fmt.Printf("- Bob sent 20.0 coins back to Alice\n")
	fmt.Printf("- 2 blocks created with valid signatures and Merkle Trees\n")
	fmt.Println("\n📁 Wallet files created:")
	fmt.Printf("- alice_key.json (Alice's private key)\n")
	fmt.Printf("- bob_key.json (Bob's private key)\n")
	fmt.Printf("- demo_blockchain/ (blockchain database)\n")
	fmt.Println("\n💡 You can now use these wallets:")
	fmt.Printf("- Load Alice's key: loadKeyFromFile(\"alice_key.json\")\n")
	fmt.Printf("- Load Bob's key: loadKeyFromFile(\"bob_key.json\")\n")
}

func initBlockchain() {
	fmt.Println("🔧 Initializing blockchain...")

	validator, err := validator.NewValidatorNode("./blockchain_data")
	if err != nil {
		fmt.Printf("Error initializing blockchain: %v\n", err)
		return
	}
	defer validator.Close()
	fmt.Println("✅ Blockchain initialized successfully!")
	fmt.Println("Data directory: ./blockchain_data")
}

func checkBlockCount() {
	fmt.Println("📊 Checking blockchain statistics...")

	// Check main blockchain
	fmt.Println("\n🔗 Main Blockchain (blockchain_data):")
	checkStorageStats("./blockchain_data")

	// Check demo blockchain if exists
	fmt.Println("\n🎯 Demo Blockchain (demo_blockchain):")
	checkStorageStats("./demo_blockchain")

	// Check pool blockchain if exists
	fmt.Println("\n🔄 Pool Blockchain (pool_blockchain):")
	checkStorageStats("./pool_blockchain")
}

func checkStorageStats(dbPath string) {
	blockStorage, err := storage.NewBlockStorage(dbPath)
	if err != nil {
		fmt.Printf("❌ Cannot open storage at %s: %v\n", dbPath, err)
		return
	}
	defer blockStorage.Close()

	latestIndex, err := blockStorage.GetLatestIndex()
	if err != nil {
		fmt.Printf("❌ Cannot get latest index: %v\n", err)
		return
	}

	if latestIndex == -1 {
		fmt.Printf("📭 No blocks found\n")
		return
	}

	blockCount := latestIndex + 1
	fmt.Printf("📦 Total blocks: %d\n", blockCount)
	fmt.Printf("🏷️  Latest block index: %d\n", latestIndex)

	// Show some block details
	fmt.Printf("📋 Block details:\n")
	start := 0
	if latestIndex > 4 {
		start = latestIndex - 4
		fmt.Printf("   ... (showing last 5 blocks)\n")
	}

	for i := start; i <= latestIndex; i++ {
		block, err := blockStorage.GetBlockByIndex(i)
		if err != nil {
			fmt.Printf("   Block %d: ❌ Error: %v\n", i, err)
			continue
		}
		fmt.Printf("   Block %d: %d transactions, hash: %x\n",
			i, len(block.Transactions), block.CurrentBlockHash[:8])
	}
}

func runTransactionPoolDemo() {
	fmt.Println("🔄 Running Transaction Pool Demo (5 transactions per block)...")

	// Create validator với transaction pool
	validator, err := validator.NewValidatorNode("./pool_blockchain")
	if err != nil {
		log.Fatal("Failed to create validator:", err)
	}
	defer validator.Close()

	// Tạo nhiều wallets để demo
	fmt.Println("\n👥 Creating multiple wallets...")
	wallets := make([]*ecdsa.PrivateKey, 8)
	addresses := make([][]byte, 8)

	for i := 0; i < 8; i++ {
		priv, err := wallet.GenerateKeyPair()
		if err != nil {
			log.Fatal(fmt.Sprintf("Failed to generate wallet %d:", i), err)
		}
		wallets[i] = priv
		addresses[i] = wallet.PublicKeyToAddress(&priv.PublicKey)
		fmt.Printf("   Wallet %d: %x\n", i+1, addresses[i][:8])
	}

	fmt.Println("\n💰 Adding transactions to pool...")

	// Thêm 12 transactions vào pool (sẽ tạo 2 blocks + 2 transactions còn lại)
	transactions := []struct {
		from   int
		to     int
		amount float64
	}{
		{0, 1, 10.0}, {1, 2, 15.0}, {2, 3, 20.0}, {3, 4, 25.0}, {4, 5, 30.0}, // Block 1 (5 transactions)
		{5, 6, 35.0}, {6, 7, 40.0}, {7, 0, 45.0}, {0, 3, 50.0}, {1, 4, 55.0}, // Block 2 (5 transactions)
		{2, 5, 60.0}, {3, 6, 65.0}, // Pending (2 transactions)
	}

	blockCount := 0
	for i, txData := range transactions {
		// Tạo transaction
		tx := &blockchain.Transaction{
			Sender:    addresses[txData.from],
			Receiver:  addresses[txData.to],
			Amount:    txData.amount,
			Timestamp: time.Now().Unix() + int64(i), // Unique timestamp
		}

		// Ký transaction
		if err := wallet.SignTransaction(tx, wallets[txData.from]); err != nil {
			log.Fatal("Failed to sign transaction:", err)
		}

		// Verify signature
		if !wallet.VerifyTransaction(tx, &wallets[txData.from].PublicKey) {
			log.Fatal("Transaction signature invalid")
		}

		// Thêm vào pool
		fmt.Printf("   Transaction %d: Wallet%d → Wallet%d (%.1f coins)\n",
			i+1, txData.from+1, txData.to+1, txData.amount)

		block, err := validator.AddTransaction(tx)
		if err != nil {
			log.Fatal("Failed to add transaction:", err)
		}

		// Kiểm tra xem có block mới được tạo không
		if block != nil {
			blockCount++
			fmt.Printf("\n🎉 Block %d created automatically!\n", blockCount)
			fmt.Printf("   Block Index: %d\n", block.Index)
			fmt.Printf("   Transactions: %d\n", len(block.Transactions))
			fmt.Printf("   Block Hash: %x\n", block.CurrentBlockHash[:8])
			fmt.Printf("   Merkle Root: %x\n", block.MerkleRoot[:8])
			if block.PreviousBlockHash != nil {
				fmt.Printf("   Previous Hash: %x\n", block.PreviousBlockHash[:8])
			}
		}

		// Hiển thị pool status
		poolSize := validator.GetPoolSize()
		fmt.Printf("   Pool size: %d transactions\n", poolSize)
	}

	// Hiển thị transactions còn lại trong pool
	fmt.Printf("\n📊 Final Status:\n")
	fmt.Printf("   Blocks created: %d\n", blockCount)
	fmt.Printf("   Pending transactions: %d\n", validator.GetPoolSize())

	if validator.GetPoolSize() > 0 {
		fmt.Printf("\n💡 Remaining transactions in pool:\n")
		pending := validator.GetPendingTransactions()
		for i, tx := range pending {
			fmt.Printf("   Pending %d: %x → %x (%.1f coins)\n",
				i+1, tx.Sender[:4], tx.Receiver[:4], tx.Amount)
		}

		// Option để force create block với remaining transactions
		fmt.Printf("\n🔧 Force creating block with remaining transactions...\n")
		finalBlock, err := validator.ForceCreateBlock()
		if err != nil {
			fmt.Printf("   Error: %v\n", err)
		} else {
			blockCount++
			fmt.Printf("✅ Final block created!\n")
			fmt.Printf("   Block Index: %d\n", finalBlock.Index)
			fmt.Printf("   Transactions: %d\n", len(finalBlock.Transactions))
			fmt.Printf("   Block Hash: %x\n", finalBlock.CurrentBlockHash[:8])
		}
	}

	fmt.Printf("\n🎉 Transaction Pool Demo completed!\n")
	fmt.Printf("📈 Summary:\n")
	fmt.Printf("   Total blocks created: %d\n", blockCount)
	fmt.Printf("   Total transactions processed: %d\n", len(transactions))
	fmt.Printf("   Database location: ./pool_blockchain/\n")
}
