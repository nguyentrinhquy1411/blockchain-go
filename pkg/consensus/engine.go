// Package consensus implements the blockchain consensus mechanism
// This package handles leader election, voting, and block consensus for the blockchain network
package consensus

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/nguyentrinhquy1411/blockchain-go/pkg/blockchain"
	"github.com/nguyentrinhquy1411/blockchain-go/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// VoteType represents the type of vote in consensus
type VoteType int

const (
	VoteApprove VoteType = iota // Vote to approve a block
	VoteReject                  // Vote to reject a block
)

// ConsensusEngine manages the consensus mechanism for the blockchain
// It handles leader election, voting, and block consensus
type ConsensusEngine struct {
	nodeID   string         // Unique identifier for this node
	isLeader bool           // Whether this node is the leader
	peers    []string       // List of peer node addresses
	votes    map[string]int // Maps block hash to vote count
	mutex    sync.RWMutex   // Protects the votes map from race conditions

	// Blockchain components
	blockchain *blockchain.Blockchain // Reference to the blockchain

	// Consensus parameters
	majorityThreshold int           // Minimum votes needed for consensus (2/3 majority)
	blockProposalTime time.Duration // Time interval between block proposals
	voteTimeout       time.Duration // Maximum time to wait for votes
}

// NewConsensusEngine creates a new consensus engine
// Parameters:
//   - nodeID: unique identifier for this node
//   - blockchain: reference to the blockchain instance
//   - peers: list of peer node addresses
//   - isLeader: whether this node should act as leader
func NewConsensusEngine(nodeID string, blockchain *blockchain.Blockchain, peers []string, isLeader bool) *ConsensusEngine {
	return &ConsensusEngine{
		nodeID:            nodeID,
		isLeader:          isLeader,
		peers:             peers,
		votes:             make(map[string]int),
		blockchain:        blockchain,
		majorityThreshold: calculateMajority(len(peers) + 1), // +1 for this node
		blockProposalTime: 10 * time.Second,
		voteTimeout:       5 * time.Second,
	}
}

// calculateMajority calculates the minimum votes needed for majority consensus
// For Byzantine fault tolerance, we need at least 2/3 of nodes to agree
func calculateMajority(totalNodes int) int {
	return (totalNodes*2)/3 + 1
}

// Vote represents a vote for a specific block
type Vote struct {
	BlockHash string    // Hash of the block being voted on
	VoterID   string    // ID of the node casting the vote
	VoteType  VoteType  // Type of vote (approve/reject)
	Timestamp time.Time // When the vote was cast
}

// BlockProposal represents a proposed block waiting for consensus
type BlockProposal struct {
	Block     *blockchain.Block // The proposed block
	Hash      string            // Hash of the block
	Proposer  string            // ID of the node that proposed the block
	Votes     []Vote            // List of votes received for this block
	Timestamp time.Time         // When the block was proposed
}

// StartConsensus starts the consensus mechanism
// Leaders will start proposing blocks, followers will participate in voting
func (ce *ConsensusEngine) StartConsensus() {
	log.Printf("[%s] Starting consensus engine (Leader: %v)", ce.nodeID, ce.isLeader)

	if ce.isLeader {
		// Leaders propose new blocks periodically
		go ce.leaderConsensusLoop()
	}

	// All nodes can receive and process votes
	log.Printf("[%s] Consensus engine started successfully", ce.nodeID)
}

// leaderConsensusLoop is the main loop for leader nodes
// It periodically proposes new blocks and manages the consensus process
func (ce *ConsensusEngine) leaderConsensusLoop() {
	log.Printf("[%s] Starting leader consensus loop", ce.nodeID)

	// Create a ticker to propose blocks at regular intervals
	ticker := time.NewTicker(ce.blockProposalTime)
	defer ticker.Stop()
	for range ticker.C {
		// Time to propose a new block
		ce.proposeNewBlock()
	}
}

// proposeNewBlock creates and proposes a new block to the network
// This is only called by the leader node
func (ce *ConsensusEngine) proposeNewBlock() {
	if !ce.isLeader {
		log.Printf("[%s] Error: Non-leader node attempting to propose block", ce.nodeID)
		return
	}

	log.Printf("[%s] CONSENSUS: Proposing new block...", ce.nodeID)

	// Step 1: Create a new block with consensus transaction
	// In a real blockchain, this would include pending transactions from mempool
	transactions := []*blockchain.Transaction{
		{
			Sender:    []byte("consensus"), // System transaction
			Receiver:  []byte("reward"),    // Block reward
			Amount:    1.0,                 // Fixed reward amount
			Timestamp: time.Now().Unix(),   // Current timestamp
		},
	}

	// Step 2: Get the latest block to build upon
	latestBlock := ce.blockchain.GetLatestBlock()

	// Step 3: Create new block using blockchain constructor
	// This automatically calculates merkle root, hash, etc.
	newBlock := blockchain.NewBlock(
		latestBlock.Index+1,          // Next block index
		transactions,                 // Block transactions
		latestBlock.CurrentBlockHash, // Previous block hash
	)

	// Step 4: Calculate block hash for voting
	blockHash := fmt.Sprintf("%x", newBlock.CurrentBlockHash)

	log.Printf("[%s] CONSENSUS: Created block %d with hash %s",
		ce.nodeID, newBlock.Index, blockHash[:8])

	// Step 5: Initialize voting for this block
	// Leader automatically votes for their own proposal
	ce.mutex.Lock()
	ce.votes[blockHash] = 1 // Leader's automatic vote
	ce.mutex.Unlock()

	log.Printf("[%s] CONSENSUS: Leader vote recorded for block %s",
		ce.nodeID, blockHash[:8])

	// Step 6: Send block proposal to all peer nodes
	ce.broadcastBlockProposal(newBlock)

	// Step 7: Wait for votes and check consensus
	go ce.waitForConsensus(blockHash, newBlock)
}

// broadcastBlockProposal sends a block proposal to all peer nodes
func (ce *ConsensusEngine) broadcastBlockProposal(block *blockchain.Block) {
	blockHash := fmt.Sprintf("%x", block.CurrentBlockHash)
	log.Printf("[%s] CONSENSUS: Broadcasting block proposal %s to %d peers",
		ce.nodeID, blockHash[:8], len(ce.peers))

	// Convert internal block to protobuf format for network transmission
	protoBlock := ce.blockToProto(block)

	// Send proposal to each peer concurrently
	successCount := 0
	for _, peerAddr := range ce.peers {
		go func(peer string) {
			success := ce.sendBlockProposal(peer, protoBlock)
			if success {
				successCount++
			}
		}(peerAddr)
	}

	log.Printf("[%s] CONSENSUS: Block proposal broadcast initiated to %d peers",
		ce.nodeID, len(ce.peers))
}

// sendBlockProposal sends a block proposal to a specific peer
func (ce *ConsensusEngine) sendBlockProposal(peerAddr string, protoBlock *proto.Block) bool {
	// Step 1: Establish gRPC connection to peer
	conn, err := grpc.NewClient(peerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("[%s] CONSENSUS: Failed to connect to peer %s: %v", ce.nodeID, peerAddr, err)
		return false
	}
	defer conn.Close()

	// Step 2: Create blockchain service client
	client := proto.NewBlockchainServiceClient(conn)

	// Step 3: Set timeout for the proposal request
	ctx, cancel := context.WithTimeout(context.Background(), ce.voteTimeout)
	defer cancel()

	// Step 4: Send the block proposal
	resp, err := client.ProposeBlock(ctx, &proto.ProposeBlockRequest{
		Block:      protoBlock,
		ProposerId: ce.nodeID,
	})

	if err != nil {
		log.Printf("[%s] CONSENSUS: Failed to send proposal to %s: %v", ce.nodeID, peerAddr, err)
		return false
	}

	log.Printf("[%s] CONSENSUS: Proposal sent to %s: %s",
		ce.nodeID, peerAddr, resp.Message)
	return resp.Accepted
}

// ProcessBlockProposal processes an incoming block proposal from another node
// This is called when a follower receives a proposal from the leader
func (ce *ConsensusEngine) ProcessBlockProposal(proposerID string, protoBlock *proto.Block) (bool, string) {
	blockHash := protoBlock.Hash
	log.Printf("[%s] CONSENSUS: Processing block proposal from %s, hash %s",
		ce.nodeID, proposerID, blockHash[:8])

	// Step 1: Convert protobuf block to internal format
	block := ce.protoToBlock(protoBlock)

	// Step 2: Validate the proposed block
	if !ce.validateProposedBlock(block) {
		log.Printf("[%s] CONSENSUS: Block validation failed for %s", ce.nodeID, blockHash[:8])
		return false, "Block validation failed"
	}

	// Step 3: If not leader, send vote to leader
	if !ce.isLeader {
		go ce.sendVoteToLeader(blockHash, VoteApprove)
	}

	log.Printf("[%s] CONSENSUS: Block proposal accepted: %s", ce.nodeID, blockHash[:8])
	return true, "Block proposal accepted"
}

// sendVoteToLeader sends a vote to the leader node
func (ce *ConsensusEngine) sendVoteToLeader(blockHash string, voteType VoteType) {
	// Step 1: Find the leader node address
	leaderAddr := ce.findLeaderAddress()
	if leaderAddr == "" {
		log.Printf("[%s] CONSENSUS: Leader not found in peers list", ce.nodeID)
		return
	}

	// Step 2: Establish connection to leader
	conn, err := grpc.NewClient(leaderAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("[%s] CONSENSUS: Failed to connect to leader %s: %v", ce.nodeID, leaderAddr, err)
		return
	}
	defer conn.Close()

	// Step 3: Create client and send vote
	client := proto.NewBlockchainServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), ce.voteTimeout)
	defer cancel()

	_, err = client.Vote(ctx, &proto.VoteRequest{
		BlockHash: blockHash,
		VoterId:   ce.nodeID,
		Approve:   voteType == VoteApprove,
	})

	if err != nil {
		log.Printf("[%s] CONSENSUS: Failed to send vote to leader: %v", ce.nodeID, err)
		return
	}

	log.Printf("[%s] CONSENSUS: Vote sent to leader for block %s: %v",
		ce.nodeID, blockHash[:8], voteType == VoteApprove)
}

// ProcessVote processes an incoming vote from another node
// This is typically called on the leader node
func (ce *ConsensusEngine) ProcessVote(voterID, blockHash string, approve bool) (bool, string) {
	log.Printf("[%s] CONSENSUS: Processing vote from %s for block %s: %v",
		ce.nodeID, voterID, blockHash[:8], approve)

	if !ce.isLeader {
		return false, "Only leader can process votes"
	}

	if !approve {
		log.Printf("[%s] CONSENSUS: Received rejection vote from %s", ce.nodeID, voterID)
		return true, "Vote recorded (rejected)"
	}

	// Step 1: Record the approval vote
	ce.mutex.Lock()
	ce.votes[blockHash]++
	voteCount := ce.votes[blockHash]
	ce.mutex.Unlock()

	log.Printf("[%s] CONSENSUS: Block %s now has %d votes (need %d for consensus)",
		ce.nodeID, blockHash[:8], voteCount, ce.majorityThreshold)

	// Step 2: Check if we have achieved consensus
	if voteCount >= ce.majorityThreshold {
		log.Printf("[%s] CONSENSUS: Block %s achieved consensus with %d votes!",
			ce.nodeID, blockHash[:8], voteCount)

		// Commit the block in a separate goroutine
		go ce.commitBlock(blockHash)
	}

	return true, "Vote recorded"
}

// waitForConsensus waits for consensus on a proposed block
func (ce *ConsensusEngine) waitForConsensus(blockHash string, block *blockchain.Block) {
	// Wait for consensus timeout
	time.Sleep(ce.voteTimeout * 2)

	ce.mutex.RLock()
	voteCount := ce.votes[blockHash]
	ce.mutex.RUnlock()

	if voteCount < ce.majorityThreshold {
		log.Printf("[%s] CONSENSUS: Block %s failed to achieve consensus (%d/%d votes)",
			ce.nodeID, blockHash[:8], voteCount, ce.majorityThreshold)

		// Clean up failed proposal
		ce.mutex.Lock()
		delete(ce.votes, blockHash)
		ce.mutex.Unlock()
	}
}

// commitBlock commits a block to the blockchain after achieving consensus
func (ce *ConsensusEngine) commitBlock(blockHash string) {
	log.Printf("[%s] CONSENSUS: Committing block %s to blockchain", ce.nodeID, blockHash[:8])

	// Step 1: Create a consensus block for commitment
	// In a real implementation, we would find the actual proposed block
	// For now, we create a new consensus block
	latestBlock := ce.blockchain.GetLatestBlock()

	transactions := []*blockchain.Transaction{
		{
			Sender:    []byte("consensus"),
			Receiver:  []byte("reward"),
			Amount:    1.0,
			Timestamp: time.Now().Unix(),
		},
	}

	newBlock := blockchain.NewBlock(latestBlock.Index+1, transactions, latestBlock.CurrentBlockHash)

	// Step 2: Add block to blockchain
	if err := ce.blockchain.AddBlock(newBlock); err != nil {
		log.Printf("[%s] CONSENSUS: Failed to commit block to blockchain: %v", ce.nodeID, err)
		return
	}

	log.Printf("[%s] CONSENSUS: Block %d successfully committed to blockchain",
		ce.nodeID, newBlock.Index)

	// Step 3: Clean up vote tracking
	ce.mutex.Lock()
	delete(ce.votes, blockHash)
	ce.mutex.Unlock()

	// Step 4: Notify peers about committed block (in a real implementation)
	ce.notifyPeersBlockCommitted(newBlock)
}

// notifyPeersBlockCommitted notifies all peers that a block has been committed
func (ce *ConsensusEngine) notifyPeersBlockCommitted(block *blockchain.Block) {
	log.Printf("[%s] CONSENSUS: Notifying peers about committed block %d", ce.nodeID, block.Index)
	// In a full implementation, this would send commit notifications to all peers
	// For now, peers will sync via the existing recovery mechanism
}

// Helper functions

// findLeaderAddress finds the leader node address from the peers list
func (ce *ConsensusEngine) findLeaderAddress() string {
	// Convention: node1 is always the leader
	for _, peer := range ce.peers {
		if peer == "node1:50051" || peer == "localhost:50051" || peer == "127.0.0.1:50051" {
			return peer
		}
	}
	return ""
}

// validateProposedBlock validates a proposed block before voting
func (ce *ConsensusEngine) validateProposedBlock(block *blockchain.Block) bool {
	// Basic validation checks
	if block.Index <= 0 {
		return false
	}

	// Validate each transaction
	for _, tx := range block.Transactions {
		if tx.Amount <= 0 {
			return false
		}
	}

	// Use existing blockchain validation
	return block.IsValid()
}

// Conversion helper functions

func (ce *ConsensusEngine) blockToProto(block *blockchain.Block) *proto.Block {
	var transactions []*proto.Transaction
	for _, tx := range block.Transactions {
		transactions = append(transactions, &proto.Transaction{
			Sender:    fmt.Sprintf("%x", tx.Sender),
			Receiver:  fmt.Sprintf("%x", tx.Receiver),
			Amount:    tx.Amount,
			Timestamp: tx.Timestamp,
			Signature: tx.Signature,
		})
	}

	return &proto.Block{
		Height:       int32(block.Index),
		PreviousHash: fmt.Sprintf("%x", block.PreviousBlockHash),
		MerkleRoot:   fmt.Sprintf("%x", block.MerkleRoot),
		Timestamp:    block.Timestamp,
		Transactions: transactions,
		Hash:         fmt.Sprintf("%x", block.CurrentBlockHash),
	}
}

func (ce *ConsensusEngine) protoToBlock(pb *proto.Block) *blockchain.Block {
	var transactions []*blockchain.Transaction
	for _, tx := range pb.Transactions {
		// Convert hex strings back to bytes
		sender := make([]byte, len(tx.Sender)/2)
		receiver := make([]byte, len(tx.Receiver)/2)
		fmt.Sscanf(tx.Sender, "%x", &sender)
		fmt.Sscanf(tx.Receiver, "%x", &receiver)

		transactions = append(transactions, &blockchain.Transaction{
			Sender:    sender,
			Receiver:  receiver,
			Amount:    tx.Amount,
			Timestamp: tx.Timestamp,
			Signature: tx.Signature,
		})
	}

	// Convert hex strings back to bytes
	previousHash := make([]byte, len(pb.PreviousHash)/2)
	merkleRoot := make([]byte, len(pb.MerkleRoot)/2)
	currentHash := make([]byte, len(pb.Hash)/2)

	fmt.Sscanf(pb.PreviousHash, "%x", &previousHash)
	fmt.Sscanf(pb.MerkleRoot, "%x", &merkleRoot)
	fmt.Sscanf(pb.Hash, "%x", &currentHash)

	return &blockchain.Block{
		Index:             int(pb.Height),
		PreviousBlockHash: previousHash,
		MerkleRoot:        merkleRoot,
		Timestamp:         pb.Timestamp,
		Transactions:      transactions,
		CurrentBlockHash:  currentHash,
	}
}
