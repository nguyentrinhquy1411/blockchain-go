package main

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/nguyentrinhquy1411/blockchain-go/pkg/blockchain"
	"github.com/nguyentrinhquy1411/blockchain-go/pkg/wallet"
)

func saveKey(priv *ecdsa.PrivateKey) error {
	f, err := os.Create("user_key.json")
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()

	if err := json.NewEncoder(f).Encode(priv); err != nil {
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

	var priv ecdsa.PrivateKey
	if err := json.NewDecoder(f).Decode(&priv); err != nil {
		return nil, fmt.Errorf("failed to decode key: %w", err)
	}

	return &priv, nil
}

func main() {
	args := os.Args
	if len(args) < 2 {
		fmt.Println("Usage: cli create|send")
		return
	}

	switch args[1] {
	case "create":
		priv, err := wallet.GenerateKeyPair()
		if err != nil {
			fmt.Printf("Error generating key pair: %v\n", err)
			return
		}

		if err := saveKey(priv); err != nil {
			fmt.Printf("Error saving key: %v\n", err)
			return
		}

	case "send":
		priv, err := loadKey()
		if err != nil {
			fmt.Printf("Error loading key: %v\n", err)
			return
		}

		pub := &priv.PublicKey
		sender := wallet.PublicKeyToAddress(pub)

		// Fix: Convert string thành []byte
		receiver := []byte("abcd1234")

		tx := blockchain.Transaction{
			Sender:    sender,
			Receiver:  receiver, // Fix: []byte thay vì string
			Amount:    10.0,
			Timestamp: time.Now().Unix(),
		}

		if err := wallet.SignTransaction(&tx, priv); err != nil {
			fmt.Printf("Error signing transaction: %v\n", err)
			return
		}
		fmt.Printf("Signed TX: %+v\n", tx)

		// Xác minh chữ ký (test)
		valid := wallet.VerifyTransaction(&tx, pub)
		fmt.Println("Valid signature:", valid)

	default:
		fmt.Println("Unknown command. Usage: cli create|send")
	}
}
