package main

import (
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
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
	isLeaderStr := os.Getenv("IS_LEADER")
	isLeader, _ := strconv.ParseBool(isLeaderStr)

	peersStr := os.Getenv("PEERS")
	var peers []string
	if peersStr != "" {
		peers = strings.Split(peersStr, ",")
	}

	server := p2p.NewBlockchainServer(nodeID, bc, storage, peers, isLeader)
	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		// Get port from environment or use default
		port := os.Getenv("PORT")
		if port == "" {
			port = "50051" // Default port
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
