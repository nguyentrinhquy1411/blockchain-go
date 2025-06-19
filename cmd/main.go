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

func main() {
	args := os.Args
	if len(args) < 2 {
		printUsage()
		return
	}

	switch args[1] {
	case "create":
		createUserKey()
	case "send":
		sendTransaction(args)
	case "demo":
		runAliceBobDemo()
	case "init":
		initBlockchain()
	case "balance":
		checkBalance(args)
	case "blocks":
		listBlocks()
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
	fmt.Println("  send <to> <amount> - Send money to address")
	fmt.Println("  demo            - Run Alice & Bob demo")
	fmt.Println("  init            - Initialize blockchain")
	fmt.Println("  balance <address> - Check balance of address")
	fmt.Println("  blocks          - List all blocks")
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
	fmt.Printf("Alice Address: %x\n", aliceAddr)

	// Create Bob's wallet
	fmt.Println("\n👨 Creating Bob's wallet...")
	bobPriv, err := wallet.GenerateKeyPair()
	if err != nil {
		log.Fatal("Failed to generate Bob's key:", err)
	}
	bobAddr := wallet.PublicKeyToAddress(&bobPriv.PublicKey)
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

func checkBalance(args []string) {
	if len(args) < 3 {
		fmt.Println("Usage: cli balance <address>")
		return
	}

	address := args[2]
	fmt.Printf("Checking balance for address: %s\n", address)
	fmt.Println("⚠️  Balance checking not implemented yet")
	fmt.Println("This would require implementing UTXO model or account-based model")
}

func listBlocks() {
	fmt.Println("📋 Listing all blocks...")

	validator, err := validator.NewValidatorNode("./blockchain_data")
	if err != nil {
		fmt.Printf("Error accessing blockchain: %v\n", err)
		return
	}
	defer validator.Close()

	fmt.Println("⚠️  Block listing not implemented yet")
	fmt.Println("This would require implementing block iteration in storage layer")
}
