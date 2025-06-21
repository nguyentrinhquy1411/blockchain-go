package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nguyentrinhquy1411/blockchain-go/pkg/blockchain"
	"github.com/nguyentrinhquy1411/blockchain-go/pkg/storage"
)

func main() {
	log.Println("Starting simple blockchain node...")

	// Create storage
	storage, err := storage.NewLevelDB("data/simple")
	if err != nil {
		log.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	// Create blockchain
	bc, err := blockchain.NewBlockchain(storage)
	if err != nil {
		log.Fatalf("Failed to create blockchain: %v", err)
	}

	// Create a simple transaction
	tx := &blockchain.Transaction{
		Sender:    []byte("Alice"),
		Receiver:  []byte("Bob"),
		Amount:    10.0,
		Timestamp: time.Now().Unix(),
	}

	log.Printf("Created transaction: %s -> %s (%.2f)", tx.Sender, tx.Receiver, tx.Amount)

	// Create a block with the transaction
	transactions := []*blockchain.Transaction{tx}
	latestBlock := bc.GetLatestBlock()

	newBlock := blockchain.NewBlock(latestBlock.Index+1, transactions, latestBlock.CurrentBlockHash)
	log.Printf("Created block %d with %d transactions", newBlock.Index, len(newBlock.Transactions))

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Println("Simple blockchain node running. Press Ctrl+C to exit.")
	<-sigChan
	log.Println("Shutting down...")
}
