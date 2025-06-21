package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/nguyentrinhquy1411/blockchain-go/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	var (
		serverAddr = flag.String("server", "localhost:50051", "Server address")
		command    = flag.String("cmd", "latest", "Command to execute: latest, send, get")
		sender     = flag.String("sender", "Alice", "Transaction sender")
		receiver   = flag.String("receiver", "Bob", "Transaction receiver")
		amount     = flag.Float64("amount", 10.0, "Transaction amount")
	)
	flag.Parse()

	// Connect to server
	conn, err := grpc.NewClient(*serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	client := proto.NewBlockchainServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	switch *command {
	case "latest":
		resp, err := client.GetLatestBlock(ctx, &proto.GetLatestBlockRequest{})
		if err != nil {
			log.Fatalf("Failed to get latest block: %v", err)
		}
		fmt.Printf("Latest Block:\n")
		fmt.Printf("  Height: %d\n", resp.Height)
		fmt.Printf("  Hash: %s\n", resp.Block.Hash[:16]+"...")
		fmt.Printf("  Transactions: %d\n", len(resp.Block.Transactions))

	case "send":
		tx := &proto.Transaction{
			Sender:    *sender,
			Receiver:  *receiver,
			Amount:    *amount,
			Timestamp: time.Now().Unix(),
		}

		resp, err := client.SendTransaction(ctx, &proto.SendTransactionRequest{
			Transaction: tx,
		})
		if err != nil {
			log.Fatalf("Failed to send transaction: %v", err)
		}

		fmt.Printf("Transaction sent: %s\n", resp.Message)
		fmt.Printf("  %s -> %s: %.2f\n", *sender, *receiver, *amount)

	default:
		fmt.Printf("Unknown command: %s\n", *command)
		fmt.Println("Available commands: latest, send")
	}
}
