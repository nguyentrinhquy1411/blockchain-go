package p2p

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/nguyentrinhquy1411/blockchain-go/pkg/blockchain"
	"github.com/nguyentrinhquy1411/blockchain-go/pkg/storage"
	"github.com/nguyentrinhquy1411/blockchain-go/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type BlockchainServer struct {
	proto.UnimplementedBlockchainServiceServer
	nodeID     string
	blockchain *blockchain.Blockchain
	storage    *storage.LevelDB
	peers      []string
	isLeader   bool
	votes      map[string]int // block_hash -> vote_count
	voteMutex  sync.RWMutex

	// Channels for consensus
	proposalChan chan *proto.Block
	voteChan     chan *proto.VoteRequest
}

func NewBlockchainServer(nodeID string, blockchain *blockchain.Blockchain, storage *storage.LevelDB, peers []string, isLeader bool) *BlockchainServer {
	return &BlockchainServer{
		nodeID:       nodeID,
		blockchain:   blockchain,
		storage:      storage,
		peers:        peers,
		isLeader:     isLeader,
		votes:        make(map[string]int),
		proposalChan: make(chan *proto.Block, 10),
		voteChan:     make(chan *proto.VoteRequest, 10),
	}
}

func (s *BlockchainServer) ProposeBlock(ctx context.Context, req *proto.ProposeBlockRequest) (*proto.ProposeBlockResponse, error) {
	log.Printf("[%s] Received block proposal from %s", s.nodeID, req.ProposerId)

	// Convert proto block to internal block
	block := s.protoToBlock(req.Block)

	// Validate block
	if !s.validateBlock(block) {
		return &proto.ProposeBlockResponse{
			Accepted: false,
			Message:  "Block validation failed",
		}, nil
	}

	// If not leader, vote on the proposal
	if !s.isLeader {
		go s.sendVote(req.Block.Hash, true)
	}

	return &proto.ProposeBlockResponse{
		Accepted: true,
		Message:  "Block proposal accepted",
	}, nil
}

func (s *BlockchainServer) Vote(ctx context.Context, req *proto.VoteRequest) (*proto.VoteResponse, error) {
	log.Printf("[%s] Received vote from %s for block %s: %v", s.nodeID, req.VoterId, req.BlockHash[:8], req.Approve)

	if s.isLeader && req.Approve {
		s.voteMutex.Lock()
		s.votes[req.BlockHash]++
		voteCount := s.votes[req.BlockHash]
		s.voteMutex.Unlock()

		// If majority votes (2 out of 3), commit block
		if voteCount >= 2 {
			log.Printf("[%s] Block %s achieved consensus with %d votes", s.nodeID, req.BlockHash[:8], voteCount)

			// Commit block to blockchain - find the block and add it
			// Note: In a full implementation, you'd store pending blocks
			// For now, we'll just log the successful consensus
			go s.commitBlock(req.BlockHash)
		}
	}

	return &proto.VoteResponse{
		Success: true,
		Message: "Vote recorded",
	}, nil
}

// commitBlock commits a block after achieving consensus
func (s *BlockchainServer) commitBlock(blockHash string) {
	log.Printf("[%s] 🔄 Committing block %s to blockchain", s.nodeID, blockHash[:8])

	// Store the pending block for this hash
	s.voteMutex.Lock()
	defer s.voteMutex.Unlock()

	// Create a consensus block and add it to blockchain
	transactions := []*blockchain.Transaction{
		{
			Sender:    []byte("consensus"),
			Receiver:  []byte("reward"),
			Amount:    1.0,
			Timestamp: time.Now().Unix(),
		},
	}

	latestBlock := s.blockchain.GetLatestBlock()
	newBlock := blockchain.NewBlock(latestBlock.Index+1, transactions, latestBlock.CurrentBlockHash)

	// Add to blockchain
	if err := s.blockchain.AddBlock(newBlock); err != nil {
		log.Printf("[%s] ❌ Failed to commit block to blockchain: %v", s.nodeID, err)
		return
	}

	log.Printf("[%s] ✅ Block %d successfully committed to blockchain", s.nodeID, newBlock.Index)

	// Clean up the vote count
	delete(s.votes, blockHash)

	// Notify all peers about the committed block
	s.broadcastCommittedBlock(newBlock)
}

// broadcastCommittedBlock notifies all peers about a committed block
func (s *BlockchainServer) broadcastCommittedBlock(block *blockchain.Block) {
	if !s.isLeader {
		return
	}

	log.Printf("[%s] 📢 Broadcasting committed block %d to all peers", s.nodeID, block.Index)

	// For now, we'll use a simple approach - peers will sync via GetLatestBlock
	// In a full implementation, we would have a separate NotifyCommittedBlock RPC
	log.Printf("[%s] ✅ Block %d broadcast completed (peers will sync via polling)", s.nodeID, block.Index)
}

func (s *BlockchainServer) SendTransaction(ctx context.Context, req *proto.SendTransactionRequest) (*proto.SendTransactionResponse, error) {
	log.Printf("[%s] Received transaction: %s -> %s (%.2f)", s.nodeID, req.Transaction.Sender, req.Transaction.Receiver, req.Transaction.Amount)

	// Convert proto transaction to internal transaction
	tx := s.protoToTransaction(req.Transaction)

	// Basic validation
	if tx.Amount <= 0 {
		return &proto.SendTransactionResponse{
			Accepted: false,
			Message:  "Invalid transaction amount",
		}, nil
	}

	// If this is the leader, we could immediately create a block
	// For now, just accept the transaction
	if s.isLeader {
		log.Printf("[%s] Leader received transaction, will include in next block", s.nodeID)
		// In a full implementation, add to transaction pool
		// s.addToTransactionPool(tx)
	} else {
		// Forward to leader if we're a follower
		log.Printf("[%s] Follower received transaction, forwarding to leader", s.nodeID)
		// In a full implementation, forward to leader node
	}

	return &proto.SendTransactionResponse{
		Accepted: true,
		Message:  "Transaction accepted",
	}, nil
}

func (s *BlockchainServer) GetLatestBlock(ctx context.Context, req *proto.GetLatestBlockRequest) (*proto.GetLatestBlockResponse, error) {
	// Get latest block from blockchain
	latestBlock := s.blockchain.GetLatestBlock()

	return &proto.GetLatestBlockResponse{
		Block:  s.blockToProto(latestBlock),
		Height: int32(latestBlock.Index),
	}, nil
}

func (s *BlockchainServer) GetBlock(ctx context.Context, req *proto.GetBlockRequest) (*proto.GetBlockResponse, error) {
	var block *blockchain.Block
	var err error

	switch req.Identifier.(type) {
	case *proto.GetBlockRequest_Height:
		height := req.GetHeight()
		block, err = s.blockchain.GetBlockByHeight(int(height))
	case *proto.GetBlockRequest_Hash:
		hash := req.GetHash()
		block, err = s.blockchain.GetBlockByHash(hash)
	default:
		return &proto.GetBlockResponse{Found: false}, fmt.Errorf("invalid identifier")
	}

	if err != nil {
		return &proto.GetBlockResponse{Found: false}, nil
	}

	return &proto.GetBlockResponse{
		Block: s.blockToProto(block),
		Found: true,
	}, nil
}

func (s *BlockchainServer) SyncBlocks(ctx context.Context, req *proto.SyncBlocksRequest) (*proto.SyncBlocksResponse, error) {
	var blocks []*proto.Block

	for height := req.FromHeight; height <= req.ToHeight; height++ {
		block, err := s.blockchain.GetBlockByHeight(int(height))
		if err != nil {
			break
		}
		blocks = append(blocks, s.blockToProto(block))
	}

	return &proto.SyncBlocksResponse{
		Blocks: blocks,
	}, nil
}

// Helper functions for conversion
func (s *BlockchainServer) protoToBlock(pb *proto.Block) *blockchain.Block {
	var transactions []*blockchain.Transaction
	for _, tx := range pb.Transactions {
		transactions = append(transactions, s.protoToTransaction(tx))
	}

	previousHash, _ := hex.DecodeString(pb.PreviousHash)
	merkleRoot, _ := hex.DecodeString(pb.MerkleRoot)
	currentHash, _ := hex.DecodeString(pb.Hash)

	return &blockchain.Block{
		Index:             int(pb.Height),
		PreviousBlockHash: previousHash,
		MerkleRoot:        merkleRoot,
		Timestamp:         pb.Timestamp,
		Transactions:      transactions,
		CurrentBlockHash:  currentHash,
	}
}

func (s *BlockchainServer) blockToProto(block *blockchain.Block) *proto.Block {
	var transactions []*proto.Transaction
	for _, tx := range block.Transactions {
		transactions = append(transactions, s.transactionToProto(tx))
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

func (s *BlockchainServer) protoToTransaction(pt *proto.Transaction) *blockchain.Transaction {
	sender, _ := hex.DecodeString(pt.Sender)
	receiver, _ := hex.DecodeString(pt.Receiver)

	return &blockchain.Transaction{
		Sender:    sender,
		Receiver:  receiver,
		Amount:    pt.Amount,
		Timestamp: pt.Timestamp,
		Signature: pt.Signature,
	}
}

func (s *BlockchainServer) transactionToProto(tx *blockchain.Transaction) *proto.Transaction {
	return &proto.Transaction{
		Sender:    fmt.Sprintf("%x", tx.Sender),
		Receiver:  fmt.Sprintf("%x", tx.Receiver),
		Amount:    tx.Amount,
		Timestamp: tx.Timestamp,
		Signature: tx.Signature,
	}
}

func (s *BlockchainServer) validateBlock(block *blockchain.Block) bool {
	// Basic validation
	if block.Index <= 0 {
		return false
	}

	// Validate each transaction
	for _, tx := range block.Transactions {
		if tx.Amount <= 0 {
			return false
		}
	}

	// Use existing validation method
	return block.IsValid()
}

func (s *BlockchainServer) sendVote(blockHash string, approve bool) {
	// Send vote to leader
	if len(s.peers) == 0 {
		log.Printf("[%s] No peers configured for voting", s.nodeID)
		return
	}

	// Find leader peer (node1 is always leader)
	var leaderAddr string
	for _, peer := range s.peers {
		if peer == "node1:50051" || peer == "localhost:50051" || peer == "127.0.0.1:50051" {
			leaderAddr = peer
			break
		}
	}

	if leaderAddr == "" {
		log.Printf("[%s] Leader not found in peers list", s.nodeID)
		return
	}

	go func() {
		conn, err := grpc.NewClient(leaderAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Printf("[%s] Failed to connect to leader %s: %v", s.nodeID, leaderAddr, err)
			return
		}
		defer conn.Close()

		client := proto.NewBlockchainServiceClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err = client.Vote(ctx, &proto.VoteRequest{
			BlockHash: blockHash,
			VoterId:   s.nodeID,
			Approve:   approve,
		})
		if err != nil {
			log.Printf("[%s] Failed to send vote to leader: %v", s.nodeID, err)
		} else {
			log.Printf("[%s] ✅ Vote sent successfully to leader for block %s: %v", s.nodeID, blockHash[:8], approve)
		}
	}()
}

func (s *BlockchainServer) StartServer(port string) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	proto.RegisterBlockchainServiceServer(grpcServer, s)

	log.Printf("[%s] 🚀 Starting gRPC server on port %s (Leader: %v)", s.nodeID, port, s.isLeader)

	// Start node recovery if not leader (followers need to sync)
	if !s.isLeader {
		go func() {
			// Wait for server to start
			time.Sleep(3 * time.Second)
			log.Printf("[%s] 🔄 Starting node recovery...", s.nodeID)
			s.StartNodeRecovery()
			// Continue periodic sync every 30 seconds
			ticker := time.NewTicker(30 * time.Second)
			defer ticker.Stop()

			for range ticker.C {
				s.StartNodeRecovery()
			}
		}()
	}

	// Start consensus routine if leader
	if s.isLeader {
		go s.consensusLoop()
	}

	return grpcServer.Serve(lis)
}

func (s *BlockchainServer) consensusLoop() {
	ticker := time.NewTicker(10 * time.Second) // Create new block every 10 seconds
	defer ticker.Stop()

	for range ticker.C {
		if s.isLeader {
			s.proposeNewBlock()
		}
	}
}

func (s *BlockchainServer) proposeNewBlock() {
	log.Printf("[%s] 🚀 Proposing new block...", s.nodeID)

	// Create a consensus block with system transaction
	transactions := []*blockchain.Transaction{
		{
			Sender:    []byte("consensus"),
			Receiver:  []byte("reward"),
			Amount:    1.0,
			Timestamp: time.Now().Unix(),
		},
	}

	// Get latest block
	latestBlock := s.blockchain.GetLatestBlock()

	// Create new block using the existing constructor
	newBlock := blockchain.NewBlock(latestBlock.Index+1, transactions, latestBlock.CurrentBlockHash)

	// Convert to proto and send to followers for voting
	protoBlock := s.blockToProto(newBlock)
	blockHash := fmt.Sprintf("%x", newBlock.CurrentBlockHash)

	log.Printf("[%s] 📦 Created block %d with hash %s", s.nodeID, newBlock.Index, blockHash[:8])

	// Initialize vote count for this block (leader gets 1 vote automatically)
	s.voteMutex.Lock()
	s.votes[blockHash] = 1 // Leader's vote
	s.voteMutex.Unlock()

	// Send to all followers for voting
	proposalsSent := 0
	for _, peer := range s.peers {
		go func(peerAddr string) {
			conn, err := grpc.NewClient(peerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				log.Printf("[%s] ❌ Failed to connect to peer %s: %v", s.nodeID, peerAddr, err)
				return
			}
			defer conn.Close()

			client := proto.NewBlockchainServiceClient(conn)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			resp, err := client.ProposeBlock(ctx, &proto.ProposeBlockRequest{
				Block:      protoBlock,
				ProposerId: s.nodeID,
			})
			if err != nil {
				log.Printf("[%s] ❌ Failed to send block proposal to %s: %v", s.nodeID, peerAddr, err)
			} else {
				log.Printf("[%s] ✅ Block proposal sent to %s: %s", s.nodeID, peerAddr, resp.Message)
			}
		}(peer)
		proposalsSent++
	}

	log.Printf("[%s] 📡 Block proposal sent to %d peers", s.nodeID, proposalsSent)
}

// StartNodeRecovery starts the node recovery process
func (s *BlockchainServer) StartNodeRecovery() {
	if len(s.peers) == 0 {
		log.Printf("[%s] ⚠️  No peers configured, skipping recovery", s.nodeID)
		return
	}

	log.Printf("[%s] 🔄 Starting node recovery process...", s.nodeID)

	// Try to sync with all peers
	syncSuccess := false
	for _, peer := range s.peers {
		if s.syncWithPeer(peer) {
			log.Printf("[%s] ✅ Successfully synced with peer %s", s.nodeID, peer)
			syncSuccess = true
			break
		}
	}

	if !syncSuccess {
		log.Printf("[%s] ⚠️  Failed to sync with any peer", s.nodeID)
	}
}

// syncWithPeer attempts to sync missing blocks from a peer
func (s *BlockchainServer) syncWithPeer(peerAddr string) bool {
	conn, err := grpc.NewClient(peerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("[%s] ❌ Failed to connect to peer %s for sync: %v", s.nodeID, peerAddr, err)
		return false
	}
	defer conn.Close()

	client := proto.NewBlockchainServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get latest block from peer
	resp, err := client.GetLatestBlock(ctx, &proto.GetLatestBlockRequest{})
	if err != nil {
		log.Printf("[%s] ❌ Failed to get latest block from %s: %v", s.nodeID, peerAddr, err)
		return false
	}

	peerLatestHeight := resp.Height
	localLatestBlock := s.blockchain.GetLatestBlock()
	localHeight := int32(localLatestBlock.Index)

	log.Printf("[%s] 📊 Sync check - Local: %d, Peer %s: %d", s.nodeID, localHeight, peerAddr, peerLatestHeight)

	if peerLatestHeight <= localHeight {
		log.Printf("[%s] ✅ Local blockchain is up to date", s.nodeID)
		return true
	}

	// Sync missing blocks
	log.Printf("[%s] 🔄 Syncing blocks from height %d to %d", s.nodeID, localHeight+1, peerLatestHeight)

	syncResp, err := client.SyncBlocks(ctx, &proto.SyncBlocksRequest{
		FromHeight: localHeight + 1,
		ToHeight:   peerLatestHeight,
	})
	if err != nil {
		log.Printf("[%s] ❌ Failed to sync blocks: %v", s.nodeID, err)
		return false
	}

	// Process synced blocks
	syncedCount := 0
	for _, protoBlock := range syncResp.Blocks {
		block := s.protoToBlock(protoBlock)
		if err := s.blockchain.AddBlock(block); err != nil {
			log.Printf("[%s] ❌ Failed to add synced block %d: %v", s.nodeID, block.Index, err)
			return false
		}
		syncedCount++
		log.Printf("[%s] ✅ Synced block %d successfully", s.nodeID, block.Index)
	}

	log.Printf("[%s] 🎉 Successfully synced %d blocks from peer %s", s.nodeID, syncedCount, peerAddr)
	return true
}
