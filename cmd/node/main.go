package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/nguyentrinhquy1411/blockchain-go/pkg/blockchain"
	"github.com/nguyentrinhquy1411/blockchain-go/pkg/p2p"
	"github.com/nguyentrinhquy1411/blockchain-go/pkg/storage"
)

func main() {
	// Get configuration from environment
	nodeID := os.Getenv("NODE_ID")
	if nodeID == "" {
		nodeID = "node1"
	}

	log.Printf("Starting blockchain node %s...", nodeID)

	// Create storage
	storage, err := storage.NewLevelDB("data/" + nodeID)
	if err != nil {
		log.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	// Create blockchain
	bc, err := blockchain.NewBlockchain(storage)
	if err != nil {
		log.Fatalf("Failed to create blockchain: %v", err)
	}

	// Create P2P server
	isLeader := nodeID == "node1"
	var peers []string
	if nodeID != "node1" {
		peers = append(peers, "node1:50051")
	}

	server := p2p.NewBlockchainServer(nodeID, bc, storage, peers, isLeader)

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		port := "50051"
		if nodeID == "node2" {
			port = "50052"
		} else if nodeID == "node3" {
			port = "50053"
		}

		log.Printf("Starting gRPC server on port %s", port)
		if err := server.StartServer(port); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	log.Printf("Node %s is running. Press Ctrl+C to exit.", nodeID)
	<-sigChan
	log.Println("Shutting down...")
}
