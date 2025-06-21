// Package consensus implements node recovery and synchronization mechanisms
// This file handles automatic node recovery, block synchronization, and fault tolerance
package consensus

import (
	"context"
	"log"
	"time"

	"github.com/nguyentrinhquy1411/blockchain-go/pkg/blockchain"
	"github.com/nguyentrinhquy1411/blockchain-go/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// RecoveryEngine handles node recovery and synchronization
// It ensures nodes can recover from failures and stay synchronized with the network
type RecoveryEngine struct {
	nodeID     string                 // Unique identifier for this node
	blockchain *blockchain.Blockchain // Reference to local blockchain
	peers      []string               // List of peer node addresses
	isLeader   bool                   // Whether this node is the leader

	// Recovery parameters
	syncInterval   time.Duration // How often to check for sync (default: 30s)
	syncTimeout    time.Duration // Timeout for sync operations (default: 10s)
	maxRetries     int           // Maximum retry attempts for failed operations
	recoveryActive bool          // Whether recovery process is currently active
}

// NewRecoveryEngine creates a new recovery engine
// Parameters:
//   - nodeID: unique identifier for this node
//   - blockchain: reference to the blockchain instance
//   - peers: list of peer node addresses
//   - isLeader: whether this node is the leader
func NewRecoveryEngine(nodeID string, blockchain *blockchain.Blockchain, peers []string, isLeader bool) *RecoveryEngine {
	return &RecoveryEngine{
		nodeID:         nodeID,
		blockchain:     blockchain,
		peers:          peers,
		isLeader:       isLeader,
		syncInterval:   30 * time.Second, // Sync every 30 seconds
		syncTimeout:    10 * time.Second, // 10 second timeout for sync operations
		maxRetries:     3,                // Try up to 3 times before giving up
		recoveryActive: false,
	}
}

// StartRecovery starts the automatic recovery process
// This should be called when the node starts up or detects it needs to recover
func (re *RecoveryEngine) StartRecovery() {
	if len(re.peers) == 0 {
		log.Printf("[%s] RECOVERY: No peers configured, skipping recovery", re.nodeID)
		return
	}

	log.Printf("[%s] RECOVERY: Starting node recovery process...", re.nodeID)
	re.recoveryActive = true

	// Step 1: Immediate recovery attempt
	re.performRecoverySync()

	// Step 2: Start periodic sync for ongoing recovery
	// Only followers need continuous sync, leaders are authoritative
	if !re.isLeader {
		go re.startPeriodicSync()
	}

	log.Printf("[%s] RECOVERY: Recovery process initialized", re.nodeID)
}

// StopRecovery stops the recovery process
func (re *RecoveryEngine) StopRecovery() {
	log.Printf("[%s] RECOVERY: Stopping recovery process", re.nodeID)
	re.recoveryActive = false
}

// startPeriodicSync starts the periodic synchronization process
// This runs in the background and continuously keeps the node synchronized
func (re *RecoveryEngine) startPeriodicSync() {
	log.Printf("[%s] RECOVERY: Starting periodic sync (interval: %v)", re.nodeID, re.syncInterval)

	// Create ticker for periodic sync
	ticker := time.NewTicker(re.syncInterval)
	defer ticker.Stop()

	// Wait a bit before starting to allow the node to fully initialize
	time.Sleep(3 * time.Second)

	for range ticker.C {
		if !re.recoveryActive {
			log.Printf("[%s] RECOVERY: Periodic sync stopped", re.nodeID)
			return
		}

		// Perform sync check
		re.performRecoverySync()
	}
}

// performRecoverySync performs a single recovery synchronization attempt
// This tries to sync with all available peers and updates the local blockchain
func (re *RecoveryEngine) performRecoverySync() {
	log.Printf("[%s] RECOVERY: Performing recovery sync with %d peers", re.nodeID, len(re.peers))

	// Try to sync with each peer
	syncSuccess := false
	for _, peerAddr := range re.peers {
		log.Printf("[%s] RECOVERY: Attempting sync with peer %s", re.nodeID, peerAddr)

		success := re.syncWithPeer(peerAddr)
		if success {
			log.Printf("[%s] RECOVERY: Successfully synced with peer %s", re.nodeID, peerAddr)
			syncSuccess = true
			break // Success with one peer is sufficient
		} else {
			log.Printf("[%s] RECOVERY: Failed to sync with peer %s", re.nodeID, peerAddr)
		}
	}

	if !syncSuccess {
		log.Printf("[%s] RECOVERY: Failed to sync with any peer", re.nodeID)
	} else {
		log.Printf("[%s] RECOVERY: Sync completed successfully", re.nodeID)
	}
}

// syncWithPeer attempts to synchronize blockchain state with a specific peer
// Returns true if sync was successful, false otherwise
func (re *RecoveryEngine) syncWithPeer(peerAddr string) bool {
	// Step 1: Establish connection to peer with retries
	conn, err := re.connectToPeerWithRetry(peerAddr)
	if err != nil {
		log.Printf("[%s] RECOVERY: Failed to connect to peer %s after retries: %v",
			re.nodeID, peerAddr, err)
		return false
	}
	defer conn.Close()

	// Step 2: Create gRPC client
	client := proto.NewBlockchainServiceClient(conn)

	// Step 3: Get peer's latest block information
	peerLatestHeight, err := re.getPeerLatestHeight(client)
	if err != nil {
		log.Printf("[%s] RECOVERY: Failed to get latest height from %s: %v",
			re.nodeID, peerAddr, err)
		return false
	}

	// Step 4: Compare with local blockchain height
	localLatestBlock := re.blockchain.GetLatestBlock()
	localHeight := int32(localLatestBlock.Index)

	log.Printf("[%s] RECOVERY: Blockchain height comparison - Local: %d, Peer %s: %d",
		re.nodeID, localHeight, peerAddr, peerLatestHeight)

	// Step 5: Determine if sync is needed
	if peerLatestHeight <= localHeight {
		log.Printf("[%s] RECOVERY: Local blockchain is up to date or ahead", re.nodeID)
		return true
	}

	// Step 6: Sync missing blocks
	return re.syncMissingBlocks(client, localHeight+1, peerLatestHeight, peerAddr)
}

// connectToPeerWithRetry attempts to connect to a peer with retries
func (re *RecoveryEngine) connectToPeerWithRetry(peerAddr string) (*grpc.ClientConn, error) {
	var conn *grpc.ClientConn
	var err error

	for attempt := 1; attempt <= re.maxRetries; attempt++ {
		log.Printf("[%s] RECOVERY: Connection attempt %d/%d to %s",
			re.nodeID, attempt, re.maxRetries, peerAddr)

		conn, err = grpc.NewClient(peerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err == nil {
			log.Printf("[%s] RECOVERY: Successfully connected to %s", re.nodeID, peerAddr)
			return conn, nil
		}

		log.Printf("[%s] RECOVERY: Connection attempt %d failed: %v", re.nodeID, attempt, err)

		// Wait before retrying (exponential backoff)
		if attempt < re.maxRetries {
			waitTime := time.Duration(attempt) * time.Second
			log.Printf("[%s] RECOVERY: Waiting %v before retry", re.nodeID, waitTime)
			time.Sleep(waitTime)
		}
	}

	return nil, err
}

// getPeerLatestHeight gets the latest block height from a peer
func (re *RecoveryEngine) getPeerLatestHeight(client proto.BlockchainServiceClient) (int32, error) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), re.syncTimeout)
	defer cancel()

	// Request latest block information
	resp, err := client.GetLatestBlock(ctx, &proto.GetLatestBlockRequest{})
	if err != nil {
		return 0, err
	}

	return resp.Height, nil
}

// syncMissingBlocks synchronizes missing blocks from a peer
func (re *RecoveryEngine) syncMissingBlocks(client proto.BlockchainServiceClient,
	fromHeight, toHeight int32, peerAddr string) bool {

	log.Printf("[%s] RECOVERY: Syncing blocks from height %d to %d from peer %s",
		re.nodeID, fromHeight, toHeight, peerAddr)

	// Step 1: Request missing blocks from peer
	ctx, cancel := context.WithTimeout(context.Background(), re.syncTimeout)
	defer cancel()

	syncResp, err := client.SyncBlocks(ctx, &proto.SyncBlocksRequest{
		FromHeight: fromHeight,
		ToHeight:   toHeight,
	})

	if err != nil {
		log.Printf("[%s] RECOVERY: Failed to request blocks from peer %s: %v",
			re.nodeID, peerAddr, err)
		return false
	}

	// Step 2: Process and validate each received block
	syncedCount := 0
	for _, protoBlock := range syncResp.Blocks {
		success := re.processReceivedBlock(protoBlock)
		if !success {
			log.Printf("[%s] RECOVERY: Failed to process block at height %d",
				re.nodeID, protoBlock.Height)
			return false
		}
		syncedCount++
	}

	log.Printf("[%s] RECOVERY: Successfully synced %d blocks from peer %s",
		re.nodeID, syncedCount, peerAddr)
	return true
}

// processReceivedBlock processes a single block received during sync
func (re *RecoveryEngine) processReceivedBlock(protoBlock *proto.Block) bool {
	log.Printf("[%s] RECOVERY: Processing received block %d (hash: %s)",
		re.nodeID, protoBlock.Height, protoBlock.Hash[:8])

	// Step 1: Convert protobuf block to internal format
	block := re.protoToBlock(protoBlock)

	// Step 2: Validate the block before adding
	if !re.validateReceivedBlock(block) {
		log.Printf("[%s] RECOVERY: Block validation failed for block %d",
			re.nodeID, block.Index)
		return false
	}

	// Step 3: Add block to local blockchain
	if err := re.blockchain.AddBlock(block); err != nil {
		log.Printf("[%s] RECOVERY: Failed to add block %d to blockchain: %v",
			re.nodeID, block.Index, err)
		return false
	}

	log.Printf("[%s] RECOVERY: Successfully added block %d to blockchain",
		re.nodeID, block.Index)
	return true
}

// validateReceivedBlock validates a block received during recovery
func (re *RecoveryEngine) validateReceivedBlock(block *blockchain.Block) bool {
	// Step 1: Basic structure validation
	if block == nil {
		log.Printf("[%s] RECOVERY: Block is nil", re.nodeID)
		return false
	}

	if block.Index <= 0 {
		log.Printf("[%s] RECOVERY: Invalid block index: %d", re.nodeID, block.Index)
		return false
	}

	// Step 2: Validate transactions
	if len(block.Transactions) == 0 {
		log.Printf("[%s] RECOVERY: Block has no transactions", re.nodeID)
		return false
	}

	for i, tx := range block.Transactions {
		if tx.Amount <= 0 {
			log.Printf("[%s] RECOVERY: Invalid transaction amount at index %d: %f",
				re.nodeID, i, tx.Amount)
			return false
		}
	}

	// Step 3: Use existing blockchain validation
	if !block.IsValid() {
		log.Printf("[%s] RECOVERY: Block failed blockchain validation", re.nodeID)
		return false
	}

	log.Printf("[%s] RECOVERY: Block %d validation successful", re.nodeID, block.Index)
	return true
}

// PerformHealthCheck performs a health check to determine if recovery is needed
func (re *RecoveryEngine) PerformHealthCheck() bool {
	log.Printf("[%s] RECOVERY: Performing health check", re.nodeID)

	if len(re.peers) == 0 {
		log.Printf("[%s] RECOVERY: No peers configured for health check", re.nodeID)
		return true // Consider healthy if no peers to compare with
	}

	// Check if we can connect to at least one peer
	for _, peerAddr := range re.peers {
		if re.canConnectToPeer(peerAddr) {
			log.Printf("[%s] RECOVERY: Health check passed - can connect to %s",
				re.nodeID, peerAddr)
			return true
		}
	}

	log.Printf("[%s] RECOVERY: Health check failed - cannot connect to any peer", re.nodeID)
	return false
}

// canConnectToPeer checks if we can establish a connection to a peer
func (re *RecoveryEngine) canConnectToPeer(peerAddr string) bool {
	conn, err := grpc.NewClient(peerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return false
	}
	defer conn.Close()

	// Try a simple call to verify the connection works
	client := proto.NewBlockchainServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err = client.GetLatestBlock(ctx, &proto.GetLatestBlockRequest{})
	return err == nil
}

// GetRecoveryStatus returns the current status of the recovery engine
func (re *RecoveryEngine) GetRecoveryStatus() map[string]interface{} {
	localLatestBlock := re.blockchain.GetLatestBlock()

	return map[string]interface{}{
		"node_id":         re.nodeID,
		"is_leader":       re.isLeader,
		"recovery_active": re.recoveryActive,
		"peers_count":     len(re.peers),
		"local_height":    localLatestBlock.Index,
		"sync_interval":   re.syncInterval.String(),
		"sync_timeout":    re.syncTimeout.String(),
		"max_retries":     re.maxRetries,
	}
}

// Helper function to convert protobuf block to internal block format
func (re *RecoveryEngine) protoToBlock(pb *proto.Block) *blockchain.Block {
	var transactions []*blockchain.Transaction
	for _, tx := range pb.Transactions {
		// Convert hex strings back to bytes
		sender := make([]byte, len(tx.Sender)/2)
		receiver := make([]byte, len(tx.Receiver)/2)

		// Handle potential errors in hex decoding
		if len(tx.Sender)%2 == 0 && len(tx.Receiver)%2 == 0 {
			for i := 0; i < len(tx.Sender); i += 2 {
				if i/2 < len(sender) {
					sender[i/2] = hexToByte(tx.Sender[i : i+2])
				}
			}
			for i := 0; i < len(tx.Receiver); i += 2 {
				if i/2 < len(receiver) {
					receiver[i/2] = hexToByte(tx.Receiver[i : i+2])
				}
			}
		} else {
			// Fallback: use string bytes if hex decoding fails
			sender = []byte(tx.Sender)
			receiver = []byte(tx.Receiver)
		}

		transactions = append(transactions, &blockchain.Transaction{
			Sender:    sender,
			Receiver:  receiver,
			Amount:    tx.Amount,
			Timestamp: tx.Timestamp,
			Signature: tx.Signature,
		})
	}

	// Convert hex strings back to bytes for block hashes
	previousHash := hexToBytes(pb.PreviousHash)
	merkleRoot := hexToBytes(pb.MerkleRoot)
	currentHash := hexToBytes(pb.Hash)

	return &blockchain.Block{
		Index:             int(pb.Height),
		PreviousBlockHash: previousHash,
		MerkleRoot:        merkleRoot,
		Timestamp:         pb.Timestamp,
		Transactions:      transactions,
		CurrentBlockHash:  currentHash,
	}
}

// Helper function to convert hex string to bytes
func hexToBytes(hexStr string) []byte {
	if len(hexStr)%2 != 0 {
		return []byte(hexStr) // Fallback if not valid hex
	}

	result := make([]byte, len(hexStr)/2)
	for i := 0; i < len(hexStr); i += 2 {
		result[i/2] = hexToByte(hexStr[i : i+2])
	}
	return result
}

// Helper function to convert 2-character hex string to byte
func hexToByte(hexStr string) byte {
	if len(hexStr) != 2 {
		return 0
	}

	var result byte
	for i, c := range hexStr {
		var value byte
		if c >= '0' && c <= '9' {
			value = byte(c - '0')
		} else if c >= 'a' && c <= 'f' {
			value = byte(c - 'a' + 10)
		} else if c >= 'A' && c <= 'F' {
			value = byte(c - 'A' + 10)
		}

		if i == 0 {
			result = value << 4
		} else {
			result |= value
		}
	}

	return result
}
