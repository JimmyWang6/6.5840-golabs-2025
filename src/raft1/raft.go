package raft

// The file raftapi/raft.go defines the interface that raft must
// expose to servers (or the tester), but see comments below for each
// of these functions for more details.
//
// Make() creates a new raft peer that implements the raft interface.

import (
	"log"
	//	"bytes"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	//	"6.5840/labgob"
	"6.5840/labrpc"
	"6.5840/raftapi"
	"6.5840/tester1"
)

const (
	Leader           = "leader"
	Candidate        = "candidate"
	Follower         = "follower"
	HeartbeatTimeout = 500 * time.Millisecond
)

// A Go object implementing a single Raft peer.
type Raft struct {
	mu        sync.Mutex          // Lock to protect shared access to this peer's state
	peers     []*labrpc.ClientEnd // RPC end points of all peers
	persister *tester.Persister   // Object to hold this peer's persisted state
	me        int                 // this peer's index into peers[]
	dead      int32               // set by Kill()

	Role          string
	term          int
	voteFor       int
	Logs          []Log
	lastHeartbeat time.Time
	// Your data here (3A, 3B, 3C).
	// Look at the paper's Figure 2 for a description of what
	// state a Raft server must maintain.

}

type Log struct {
	Term  int
	Index int
}

// return currentTerm and whether this server
// believes it is the leader.
func (rf *Raft) GetState() (int, bool) {

	var term int
	var isleader bool
	// Your code here (3A).
	term = rf.term
	isleader = rf.Role == Leader
	return term, isleader
}

// save Raft's persistent state to stable storage,
// where it can later be retrieved after a crash and restart.
// see paper's Figure 2 for a description of what should be persistent.
// before you've implemented snapshots, you should pass nil as the
// second argument to persister.Save().
// after you've implemented snapshots, pass the current snapshot
// (or nil if there's not yet a snapshot).
func (rf *Raft) persist() {
	// Your code here (3C).
	// Example:
	// w := new(bytes.Buffer)
	// e := labgob.NewEncoder(w)
	// e.Encode(rf.xxx)
	// e.Encode(rf.yyy)
	// raftstate := w.Bytes()
	// rf.persister.Save(raftstate, nil)
}

// restore previously persisted state.
func (rf *Raft) readPersist(data []byte) {
	if data == nil || len(data) < 1 { // bootstrap without any state?
		return
	}
	// Your code here (3C).
	// Example:
	// r := bytes.NewBuffer(data)
	// d := labgob.NewDecoder(r)
	// var xxx
	// var yyy
	// if d.Decode(&xxx) != nil ||
	//    d.Decode(&yyy) != nil {
	//   error...
	// } else {
	//   rf.xxx = xxx
	//   rf.yyy = yyy
	// }
}

// how many bytes in Raft's persisted log?
func (rf *Raft) PersistBytes() int {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	return rf.persister.RaftStateSize()
}

// the service says it has created a snapshot that has
// all info up to and including index. this means the
// service no longer needs the log through (and including)
// that index. Raft should now trim its log as much as possible.
func (rf *Raft) Snapshot(index int, snapshot []byte) {
	// Your code here (3D).

}

// example RequestVote RPC arguments structure.
// field names must start with capital letters!
type RequestVoteArgs struct {
	Term         int
	Candidate    int
	LastLogIndex int
	LastLogTerm  int
}

// example RequestVote RPC reply structure.
// field names must start with capital letters!
type RequestVoteReply struct {
	Term        int
	VoteGranted bool
}

type AppendEntriesArgs struct {
	Term         int
	LeaderID     int
	PrevLogIndex int
	PrevLogTerm  int
	Entries      []Log
	LeaderCommit int
}

type AppendEntriesReply struct {
	Term    int
	Success bool
}

// example RequestVote RPC handler.
func (rf *Raft) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
	log.Printf("server %d in term %d received vote request from candidate %d in term %d", rf.me, rf.term, args.Candidate, args.Term)
	// check term
	reply.Term = rf.term
	if args.Term < rf.term {
		// let candidate know current term
		log.Printf("server %d in term %d rejected vote request from candidate %d in term %d", rf.me, rf.term, args.Candidate, args.Term)
		reply.VoteGranted = false
		return
	}

	if rf.voteFor == -1 || rf.voteFor == args.Candidate {
		// Candidate’s log is at least as up-to-date as receiver’s log
		if rf.logUpToDate(args) {
			log.Printf("server %d in term %d granted vote to candidate %d in term %d", rf.me, rf.term, args.Candidate, args.Term)
			// grant vote
			rf.mu.Lock()
			rf.voteFor = args.Candidate
			rf.mu.Lock()
			reply.VoteGranted = true
		} else {
			log.Printf("server %d in term %d rejected vote request from candidate %d in term %d due to log not up-to-date", rf.me, rf.term, args.Candidate, args.Term)
			reply.VoteGranted = false
		}
	} else {
		log.Printf("server %d in term %d rejected vote request from candidate %d in term %d due to already voted for %d", rf.me, rf.term, args.Candidate, args.Term, rf.voteFor)
		reply.VoteGranted = false
	}
}

func (rf *Raft) logUpToDate(args *RequestVoteArgs) bool {
	return len(rf.Logs) == 0 || args.LastLogTerm > rf.Logs[len(rf.Logs)-1].Term ||
		(args.LastLogTerm == rf.Logs[len(rf.Logs)-1].Term && args.LastLogIndex >= len(rf.Logs))
}

// example code to send a RequestVote RPC to a server.
// server is the index of the target server in rf.peers[].
// expects RPC arguments in args.
// fills in *reply with RPC reply, so caller should
// pass &reply.
// the types of the args and reply passed to Call() must be
// the same as the types of the arguments declared in the
// handler function (including whether they are pointers).
//
// The labrpc package simulates a lossy network, in which servers
// may be unreachable, and in which requests and replies may be lost.
// Call() sends a request and waits for a reply. If a reply arrives
// within a timeout interval, Call() returns true; otherwise
// Call() returns false. Thus Call() may not return for a while.
// A false return can be caused by a dead server, a live server that
// can't be reached, a lost request, or a lost reply.
//
// Call() is guaranteed to return (perhaps after a delay) *except* if the
// handler function on the server side does not return.  Thus there
// is no need to implement your own timeouts around Call().
//
// look at the comments in ../labrpc/labrpc.go for more details.
//
// if you're having trouble getting RPC to work, check that you've
// capitalized all field names in structs passed over RPC, and
// that the caller passes the address of the reply struct with &, not
// the struct itself.
func (rf *Raft) sendRequestVote(server int, args *RequestVoteArgs, reply *RequestVoteReply) bool {
	log.Printf("server %d in term %d sending vote request to server %d: %+v", rf.me, rf.term, server, args)
	ok := rf.peers[server].Call("Raft.RequestVote", args, reply)
	return ok
}

// the service using Raft (e.g. a k/v server) wants to start
// agreement on the next command to be appended to Raft's log. if this
// server isn't the leader, returns false. otherwise start the
// agreement and return immediately. there is no guarantee that this
// command will ever be committed to the Raft log, since the leader
// may fail or lose an election. even if the Raft instance has been killed,
// this function should return gracefully.
//
// the first return value is the index that the command will appear at
// if it's ever committed. the second return value is the current
// term. the third return value is true if this server believes it is
// the leader.
func (rf *Raft) Start(command interface{}) (int, int, bool) {
	index := -1
	term := -1
	isLeader := true

	// Your code here (3B).

	return index, term, isLeader
}

// the tester doesn't halt goroutines created by Raft after each test,
// but it does call the Kill() method. your code can use killed() to
// check whether Kill() has been called. the use of atomic avoids the
// need for a lock.
//
// the issue is that long-running goroutines use memory and may chew
// up CPU time, perhaps causing later tests to fail and generating
// confusing debug output. any goroutine with a long-running loop
// should call killed() to check whether it should stop.
func (rf *Raft) Kill() {
	atomic.StoreInt32(&rf.dead, 1)
	// Your code here, if desired.
}

func (rf *Raft) killed() bool {
	z := atomic.LoadInt32(&rf.dead)
	return z == 1
}

func (rf *Raft) ticker() {
	for rf.killed() == false {
		// pause for a random amount of time between 50 and 350
		// milliseconds.
		ms := 50 + (rand.Int63() % 300)
		time.Sleep(time.Duration(ms) * time.Millisecond)

		// Your code here (3A)
		// Check if a leader election should be started.
		if rf.Role == Leader {
			// send heartbeat
		} else if rf.Role == Candidate {
			// initial state, send RequestVote RPCs
			rf.Elect()
		} else {
			// follower status
			if rf.HeartbeatExpire() {
				// election timeout, start election
				rf.Elect()
			}
		}
	}
}

func (rf *Raft) Elect() {
	total := len(rf.peers)
	// vote himself first
	votes := 0
	// increase term
	rf.mu.Lock()
	rf.Role = Candidate
	rf.term += 1
	rf.voteFor = rf.me
	votes += 1
	rf.mu.Unlock()
	for server := range rf.peers {
		if (rf.peers == nil) || server == rf.me {
			continue
		}
		request := &RequestVoteArgs{
			Term:         rf.term,
			Candidate:    rf.me,
			LastLogIndex: len(rf.Logs),
			LastLogTerm: func() int {
				if len(rf.Logs) == 0 {
					return 0
				}
				return rf.Logs[len(rf.Logs)-1].Term
			}(),
		}
		reply := &RequestVoteReply{}
		rf.sendRequestVote(server, request, reply)
		log.Printf("Received vote reply at server %d in term %d from server %d: %+v", rf.me, rf.term, server, reply)
		if reply.VoteGranted {
			votes += 1
			log.Printf("server %d in term %d has now %d/%d votes", rf.me, rf.term, votes, total)
			if votes >= total/2 {
				rf.mu.Lock()
				rf.Role = Leader
				rf.mu.Unlock()
				// should send initial empty AppendEntries RPCs (heartbeats) to each server
				log.Printf("server %d becomes leader in term %d with majority votes %d/%d", rf.me, rf.term, votes, total)
			}
		}
		if reply.Term > rf.term {
			rf.mu.Lock()
			rf.term = reply.Term
			rf.Role = Follower
			rf.voteFor = -1
			rf.mu.Unlock()
			// Return to follower and Failed
			return
		}
		log.Printf("server %d in term %d received vote reply from server %d: %+v", rf.me, rf.term, server, reply)
	}
}

func (rf *Raft) HeartbeatExpire() bool {
	return time.Since(rf.lastHeartbeat) > HeartbeatTimeout
}

// the service or tester wants to create a Raft server. the ports
// of all the Raft servers (including this one) are in peers[]. this
// server's port is peers[me]. all the servers' peers[] arrays
// have the same order. persister is a place for this server to
// save its persistent state, and also initially holds the most
// recent saved state, if any. applyCh is a channel on which the
// tester or service expects Raft to send ApplyMsg messages.
// Make() must return quickly, so it should start goroutines
// for any long-running work.
func Make(peers []*labrpc.ClientEnd, me int,
	persister *tester.Persister, applyCh chan raftapi.ApplyMsg) raftapi.Raft {
	rf := &Raft{voteFor: -1, Role: Candidate}
	rf.peers = peers
	rf.persister = persister
	rf.me = me

	// Your initialization code here (3A, 3B, 3C).

	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())

	// start ticker goroutine to start elections
	go rf.ticker()

	return rf
}
