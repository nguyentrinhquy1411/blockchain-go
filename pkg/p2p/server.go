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
			// TODO: Commit block to blockchain
		}
	}

	return &proto.VoteResponse{
		Success: true,
		Message: "Vote recorded",
	}, nil
}

func (s *BlockchainServer) SendTransaction(ctx context.Context, req *proto.SendTransactionRequest) (*proto.SendTransactionResponse, error) {
	log.Printf("[%s] Received transaction: %s -> %s (%.2f)", s.nodeID, req.Transaction.Sender, req.Transaction.Receiver, req.Transaction.Amount)

	// Convert proto transaction to internal transaction
	_ = s.protoToTransaction(req.Transaction)

	// Add to pending transactions
	// TODO: Add to transaction pool

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
	// Send vote to leader (assuming first peer is leader)
	if len(s.peers) > 0 {
		// TODO: Implement gRPC client call to send vote
		log.Printf("[%s] Sending vote for block %s: %v", s.nodeID, blockHash[:8], approve)
	}
}

func (s *BlockchainServer) StartServer(port string) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	proto.RegisterBlockchainServiceServer(grpcServer, s)

	log.Printf("[%s] Starting gRPC server on port %s (Leader: %v)", s.nodeID, port, s.isLeader)

	// Start consensus routine if leader
	if s.isLeader {
		go s.consensusLoop()
	}

	return grpcServer.Serve(lis)
}

func (s *BlockchainServer) consensusLoop() {
	ticker := time.NewTicker(10 * time.Second) // Create new block every 10 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if s.isLeader {
				s.proposeNewBlock()
			}
		}
	}
}

func (s *BlockchainServer) proposeNewBlock() {
	log.Printf("[%s] Proposing new block...", s.nodeID)

	// Create a simple block with dummy transaction
	transactions := []*blockchain.Transaction{
		{
			Sender:    []byte("system"),
			Receiver:  []byte("reward"),
			Amount:    1.0,
			Timestamp: time.Now().Unix(),
		},
	}

	// Get latest block
	latestBlock := s.blockchain.GetLatestBlock()

	// Create new block using the existing constructor
	newBlock := blockchain.NewBlock(latestBlock.Index+1, transactions, latestBlock.CurrentBlockHash)

	// TODO: Send to followers for voting
	log.Printf("[%s] Created block %d with hash %x", s.nodeID, newBlock.Index, newBlock.CurrentBlockHash[:8])
}
