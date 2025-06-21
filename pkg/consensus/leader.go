// Package consensus implements leader election mechanisms
// This file handles the selection and management of leader nodes in the blockchain network
package consensus

import (
	"log"
	"sort"
	"sync"
	"time"
)

// LeaderElectionState represents the current state of leader election
type LeaderElectionState int

const (
	StateFollower  LeaderElectionState = iota // Node is a follower
	StateCandidate                            // Node is a candidate for leadership
	StateLeader                               // Node is the current leader
)

// String returns string representation of election state
func (s LeaderElectionState) String() string {
	switch s {
	case StateFollower:
		return "FOLLOWER"
	case StateCandidate:
		return "CANDIDATE"
	case StateLeader:
		return "LEADER"
	default:
		return "UNKNOWN"
	}
}

// LeaderElection manages the leader election process
// It implements a simplified Raft-like consensus algorithm for leader selection
type LeaderElection struct {
	nodeID      string              // Unique identifier for this node
	peers       []string            // List of all peer nodes
	state       LeaderElectionState // Current state of this node
	currentTerm int                 // Current election term
	votedFor    string              // Which node this node voted for in current term
	votes       map[string]int      // Vote counts for each candidate
	mutex       sync.RWMutex        // Protects election state

	// Leader information
	currentLeader   string        // ID of current leader
	leaderTimeout   time.Duration // How long to wait for leader heartbeat
	electionTimeout time.Duration // Timeout for election process

	// Callbacks
	onLeaderChange func(string)              // Called when leader changes
	onStateChange  func(LeaderElectionState) // Called when node state changes
}

// NewLeaderElection creates a new leader election manager
func NewLeaderElection(nodeID string, peers []string) *LeaderElection {
	return &LeaderElection{
		nodeID:          nodeID,
		peers:           peers,
		state:           StateFollower,
		currentTerm:     0,
		votedFor:        "",
		votes:           make(map[string]int),
		leaderTimeout:   15 * time.Second,
		electionTimeout: 10 * time.Second,
	}
}

// StartElection initiates the leader election process
func (le *LeaderElection) StartElection() {
	le.mutex.Lock()
	defer le.mutex.Unlock()

	log.Printf("[%s] ELECTION: Starting leader election for term %d", le.nodeID, le.currentTerm+1)

	// Step 1: Increment term and become candidate
	le.currentTerm++
	le.state = StateCandidate
	le.votedFor = le.nodeID // Vote for self
	le.votes = make(map[string]int)
	le.votes[le.nodeID] = 1 // Self vote

	// Notify about state change
	if le.onStateChange != nil {
		go le.onStateChange(le.state)
	}

	log.Printf("[%s] ELECTION: Became candidate for term %d", le.nodeID, le.currentTerm)

	// Step 2: Request votes from all peers
	le.requestVotesFromPeers()

	// Step 3: Wait for election timeout
	go le.waitForElectionResult()
}

// requestVotesFromPeers sends vote requests to all peer nodes
func (le *LeaderElection) requestVotesFromPeers() {
	log.Printf("[%s] ELECTION: Requesting votes from %d peers", le.nodeID, len(le.peers))

	for _, peerID := range le.peers {
		go func(peer string) {
			success := le.sendVoteRequest(peer)
			if success {
				log.Printf("[%s] ELECTION: Vote request sent to %s", le.nodeID, peer)
			} else {
				log.Printf("[%s] ELECTION: Failed to send vote request to %s", le.nodeID, peer)
			}
		}(peerID)
	}
}

// sendVoteRequest sends a vote request to a specific peer
func (le *LeaderElection) sendVoteRequest(peerID string) bool {
	// In a real implementation, this would send an RPC call to the peer
	// For demonstration, we'll simulate the vote request
	log.Printf("[%s] ELECTION: Sending vote request to %s for term %d",
		le.nodeID, peerID, le.currentTerm)

	// Simulate network delay
	time.Sleep(100 * time.Millisecond)

	// For demonstration, we'll simulate that peers vote based on node ID ordering
	// In a real system, this would involve actual network communication
	shouldVote := le.shouldPeerVote(peerID)

	if shouldVote {
		le.receiveVote(peerID, true)
		return true
	}

	return false
}

// shouldPeerVote determines if a peer should vote for this candidate
// This is a simplified logic for demonstration
func (le *LeaderElection) shouldPeerVote(peerID string) bool {
	// Simple logic: nodes with lower ID have higher priority
	// In a real system, this would consider factors like:
	// - Network partition detection
	// - Node health status
	// - Last known good state

	allNodes := append(le.peers, le.nodeID)
	sort.Strings(allNodes)

	// Find the candidate with lowest ID (highest priority)
	if len(allNodes) > 0 && allNodes[0] == le.nodeID {
		return true
	}

	return false
}

// receiveVote processes a vote received from a peer
func (le *LeaderElection) receiveVote(voterID string, approved bool) {
	le.mutex.Lock()
	defer le.mutex.Unlock()

	if le.state != StateCandidate {
		log.Printf("[%s] ELECTION: Received vote from %s but not a candidate", le.nodeID, voterID)
		return
	}

	log.Printf("[%s] ELECTION: Received vote from %s: %v", le.nodeID, voterID, approved)

	if approved {
		le.votes[le.nodeID]++
		totalVotes := le.votes[le.nodeID]

		log.Printf("[%s] ELECTION: Current vote count: %d/%d",
			le.nodeID, totalVotes, le.getMajorityThreshold())

		// Check if we have majority
		if totalVotes >= le.getMajorityThreshold() {
			le.becomeLeader()
		}
	}
}

// getMajorityThreshold calculates the minimum votes needed to become leader
func (le *LeaderElection) getMajorityThreshold() int {
	totalNodes := len(le.peers) + 1 // +1 for this node
	return (totalNodes / 2) + 1
}

// becomeLeader transitions this node to leader state
func (le *LeaderElection) becomeLeader() {
	log.Printf("[%s] ELECTION: Becoming leader for term %d", le.nodeID, le.currentTerm)

	le.state = StateLeader
	le.currentLeader = le.nodeID

	// Notify about state change
	if le.onStateChange != nil {
		go le.onStateChange(le.state)
	}

	// Notify about leader change
	if le.onLeaderChange != nil {
		go le.onLeaderChange(le.nodeID)
	}

	// Start sending heartbeats to maintain leadership
	go le.startHeartbeat()

	log.Printf("[%s] ELECTION: Successfully became leader", le.nodeID)
}

// startHeartbeat starts sending periodic heartbeats to maintain leadership
func (le *LeaderElection) startHeartbeat() {
	log.Printf("[%s] ELECTION: Starting heartbeat process", le.nodeID)

	// Send heartbeat every 5 seconds
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		le.mutex.RLock()
		if le.state != StateLeader {
			le.mutex.RUnlock()
			log.Printf("[%s] ELECTION: No longer leader, stopping heartbeat", le.nodeID)
			return
		}
		le.mutex.RUnlock()

		le.sendHeartbeat()
	}
}

// sendHeartbeat sends heartbeat to all peers
func (le *LeaderElection) sendHeartbeat() {
	le.mutex.RLock()
	currentTerm := le.currentTerm
	le.mutex.RUnlock()

	log.Printf("[%s] ELECTION: Sending heartbeat to %d peers (term %d)",
		le.nodeID, len(le.peers), currentTerm)

	for _, peerID := range le.peers {
		go func(peer string) {
			// In a real implementation, this would send RPC heartbeat
			log.Printf("[%s] ELECTION: Heartbeat sent to %s", le.nodeID, peer)
		}(peerID)
	}
}

// receiveHeartbeat processes a heartbeat from the current leader
func (le *LeaderElection) receiveHeartbeat(leaderID string, term int) {
	le.mutex.Lock()
	defer le.mutex.Unlock()

	log.Printf("[%s] ELECTION: Received heartbeat from %s (term %d)",
		le.nodeID, leaderID, term)

	// If we receive heartbeat from a leader with higher or equal term
	if term >= le.currentTerm {
		if le.state != StateFollower {
			log.Printf("[%s] ELECTION: Stepping down to follower", le.nodeID)
			le.state = StateFollower

			if le.onStateChange != nil {
				go le.onStateChange(le.state)
			}
		}

		le.currentTerm = term
		le.currentLeader = leaderID
		le.votedFor = ""

		// Reset election timeout
		go le.resetElectionTimeout()

		// Notify about leader change if it's a new leader
		if le.currentLeader != leaderID && le.onLeaderChange != nil {
			go le.onLeaderChange(leaderID)
		}
	}
}

// resetElectionTimeout resets the election timeout
func (le *LeaderElection) resetElectionTimeout() {
	// Wait for leader timeout
	time.Sleep(le.leaderTimeout)

	le.mutex.RLock()
	state := le.state
	le.mutex.RUnlock()

	// If we're still a follower and haven't heard from leader, start election
	if state == StateFollower {
		log.Printf("[%s] ELECTION: Leader timeout, starting new election", le.nodeID)
		le.StartElection()
	}
}

// waitForElectionResult waits for the election to complete
func (le *LeaderElection) waitForElectionResult() {
	time.Sleep(le.electionTimeout)

	le.mutex.Lock()
	defer le.mutex.Unlock()

	if le.state == StateCandidate {
		// Election timeout - become follower and wait for next election
		log.Printf("[%s] ELECTION: Election timeout, becoming follower", le.nodeID)
		le.state = StateFollower

		if le.onStateChange != nil {
			go le.onStateChange(le.state)
		}
		// Start new election after a random delay
		go func() {
			// Random delay between 5-10 seconds to avoid simultaneous elections
			delay := time.Duration(5+len(le.nodeID)%5) * time.Second
			time.Sleep(delay)

			le.mutex.RLock()
			stillFollower := le.state == StateFollower
			le.mutex.RUnlock()

			if stillFollower {
				le.StartElection()
			}
		}()
	}
}

// GetCurrentLeader returns the current leader node ID
func (le *LeaderElection) GetCurrentLeader() string {
	le.mutex.RLock()
	defer le.mutex.RUnlock()
	return le.currentLeader
}

// GetCurrentState returns the current state of this node
func (le *LeaderElection) GetCurrentState() LeaderElectionState {
	le.mutex.RLock()
	defer le.mutex.RUnlock()
	return le.state
}

// GetCurrentTerm returns the current election term
func (le *LeaderElection) GetCurrentTerm() int {
	le.mutex.RLock()
	defer le.mutex.RUnlock()
	return le.currentTerm
}

// IsLeader returns true if this node is currently the leader
func (le *LeaderElection) IsLeader() bool {
	le.mutex.RLock()
	defer le.mutex.RUnlock()
	return le.state == StateLeader
}

// SetOnLeaderChange sets the callback for leader changes
func (le *LeaderElection) SetOnLeaderChange(callback func(string)) {
	le.onLeaderChange = callback
}

// SetOnStateChange sets the callback for state changes
func (le *LeaderElection) SetOnStateChange(callback func(LeaderElectionState)) {
	le.onStateChange = callback
}

// ForceLeaderElection forces a new leader election
func (le *LeaderElection) ForceLeaderElection() {
	log.Printf("[%s] ELECTION: Forcing new leader election", le.nodeID)

	le.mutex.Lock()
	le.state = StateFollower
	le.currentLeader = ""
	le.votedFor = ""
	le.mutex.Unlock()

	// Start election immediately
	go le.StartElection()
}

// GetElectionStatus returns detailed election status
func (le *LeaderElection) GetElectionStatus() map[string]interface{} {
	le.mutex.RLock()
	defer le.mutex.RUnlock()

	return map[string]interface{}{
		"node_id":            le.nodeID,
		"state":              le.state.String(),
		"current_term":       le.currentTerm,
		"current_leader":     le.currentLeader,
		"voted_for":          le.votedFor,
		"peers_count":        len(le.peers),
		"votes":              le.votes,
		"majority_threshold": le.getMajorityThreshold(),
	}
}

// StepDownAsLeader forces the current leader to step down
func (le *LeaderElection) StepDownAsLeader() {
	le.mutex.Lock()
	defer le.mutex.Unlock()

	if le.state == StateLeader {
		log.Printf("[%s] ELECTION: Stepping down as leader", le.nodeID)
		le.state = StateFollower
		le.currentLeader = ""

		if le.onStateChange != nil {
			go le.onStateChange(le.state)
		}

		if le.onLeaderChange != nil {
			go le.onLeaderChange("")
		}
	}
}
