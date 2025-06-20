package raft

// this is an outline of the API that raft must expose to
// the service (or tester). see comments below for
// each of these functions for more details.
//
// rf = Make(...)
//   create a new Raft server.
// rf.Start(command interface{}) (index, term, isleader)
//   start agreement on a new log entry
// rf.GetState() (term, isLeader)
//   ask a Raft for its current term, and whether it thinks it is leader
// ApplyMsg
//   each time a new entry is committed to the log, each Raft peer
//   should send an ApplyMsg to the service (or tester)
//   in the same server.

import "sync"
import "labrpc"
import "time"
import "math/rand"

// import "bytes"
// import "encoding/gob"

// as each Raft peer becomes aware that successive log entries are
// committed, the peer should send an ApplyMsg to the service (or
// tester) on the same server, via the applyCh passed to Make().

type ApplyMsg struct {
	Index       int
	Command     interface{}
	UseSnapshot bool   // ignore for lab2; only used in lab3
	Snapshot    []byte // ignore for lab2; only used in lab3
}

type PeerState int

const (
	Follower  PeerState = 0
	Candidate PeerState = 1
	Leader    PeerState = 2
)

// A Go object implementing a single Raft peer.

type Raft struct {
	mu        sync.Mutex          // Lock to protect shared access to this peer's state
	peers     []*labrpc.ClientEnd // RPC end points of all peers
	persister *Persister          // Object to hold this peer's persisted state
	me        int                 // this peer's index into peers[]

	// Your data here (2A, 2B, 2C).
	// Look at the paper's Figure 2 for a description of what
	// state a Raft server must maintain.

	currentTerm   int       // latest term server has seen (initialized to 0 on first boot, increases monotonically)
	peerState     PeerState // Follower, Candidate or Leader state
	votedFor      int       // candidateId that received vote in current term (or null if none)
	numberOfVotes int
	heartbeat     chan bool
}

// return currentTerm and whether this server
// believes it is the leader.
func (rf *Raft) GetState() (int, bool) {

	var term int
	var isleader bool
	// Your code here (2A).
	term = rf.currentTerm
	if rf.peerState == Leader {
		isleader = true
	} else {
		isleader = false
	}

	return term, isleader
}

// save Raft's persistent state to stable storage,
// where it can later be retrieved after a crash and restart.
// see paper's Figure 2 for a description of what should be persistent.
func (rf *Raft) persist() {
	// Your code here (2C).
	// Example:
	// w := new(bytes.Buffer)
	// e := gob.NewEncoder(w)
	// e.Encode(rf.xxx)
	// e.Encode(rf.yyy)
	// data := w.Bytes()
	// rf.persister.SaveRaftState(data)
}

//
// restore previously persisted state.
//
func (rf *Raft) readPersist(data []byte) {
	// Your code here (2C).
	// Example:
	// r := bytes.NewBuffer(data)
	// d := gob.NewDecoder(r)
	// d.Decode(&rf.xxx)
	// d.Decode(&rf.yyy)
	if data == nil || len(data) < 1 { // bootstrap without any state?
		return
	}
}


// example RequestVote RPC arguments structure.
// field names must start with capital letters!
type RequestVoteArgs struct {
	// Your data here (2A, 2B).
	Term         int // candidate's term
	CandidateId  int // candidate requesting vote
	LastLogIndex int
	LastLogTerm  int
}

// example RequestVote RPC reply structure.
// field names must start with capital letters!
type RequestVoteReply struct {
	// Your data here (2A).
	Term        int  // currentTerm
	VoteGranted bool // true means candidate received vote
}


// example RequestVote RPC handler.
func (rf *Raft) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
	// Your code here (2A, 2B).

	if args.Term < rf.currentTerm { 
		// If Candidate term is less than Follower term, update Candidate term but don't receive vote
		// Reply false if term < currentTerm
		reply.Term = rf.currentTerm
		reply.VoteGranted = false
		rf.votedFor = -1
		return
	}

	if args.Term > rf.currentTerm { 
		// If Candidate term is greater than Follower/Candidate term, updates Follower/Candidate term and Follower/Candidate becomes Follower
		rf.currentTerm = args.Term
		rf.peerState = Follower
		reply.VoteGranted = false
		rf.votedFor = -1
	}

	reply.Term = rf.currentTerm // Follower and Candidate are in the same term

	if (rf.votedFor == -1 || rf.votedFor == args.CandidateId) && args.LastLogTerm <= rf.currentTerm {
		// If votedFor is null or candidateId, and candidate’s log is at least as up-to-date as receiver’s log
		rf.votedFor = args.CandidateId
		reply.VoteGranted = true
	}
}

//
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
//
func (rf *Raft) sendRequestVote(server int, args *RequestVoteArgs, reply *RequestVoteReply) bool {
	ok := rf.peers[server].Call("Raft.RequestVote", args, reply)

	if ok == true {
		if rf.currentTerm < reply.Term { // Candidate term updates and Candidate becomes follower
			rf.currentTerm = reply.Term
			rf.peerState = Follower
			rf.votedFor = -1
		}

		if reply.VoteGranted == true { // If Candidate received a vote from a Follower
			rf.numberOfVotes += 1
			if rf.peerState == Candidate && rf.numberOfVotes > len(rf.peers)/2 { // If is Candidate and received majority of Follower votes, become Leader
				rf.peerState = Leader
			}
		}
	}

	return ok
}

type AppendEntries struct {
	Term         int   // leader’s term
	LeaderId     int
	PrevLogIndex int
	PrevLogTerm  int
	Entries      []int // log entries to store (empty for heartbeat; may send more than one for efficiency)
	LeaderCommit int   // leader’s commitIndex
}

type AppendEntriesReply struct {
	Term    int  // currentTerm, for leader to update itself
	Success bool
}

func (rf *Raft) AppendEntries(args *AppendEntries, reply *AppendEntriesReply) {

	rf.heartbeat <- true // Update heartbeat

	if args.Term < rf.currentTerm { // If Leader term is less than Follower term, reply term updates Leader with Follower term and reply Success false. Leader will become a updated Follower
		reply.Term = rf.currentTerm
		reply.Success = false
		return
	}

	if args.Term > rf.currentTerm { // If Leader term is greater than Follower term, Follower term is updated with Leader term
		rf.currentTerm = args.Term
		rf.peerState = Follower
		rf.votedFor = -1
	}

	reply.Term = args.Term
	reply.Success = true
}

func (rf *Raft) sendAppendEntries(server int, args *AppendEntries, reply *AppendEntriesReply) bool {
	ok := rf.peers[server].Call("Raft.AppendEntries", args, reply)

	if ok == true {
		if reply.Success == false {
			rf.currentTerm = reply.Term
			rf.peerState = Follower
			rf.votedFor = -1
		}
	}

	return ok
}

//
// the service using Raft (e.g. a k/v server) wants to start
// agreement on the next command to be appended to Raft's log. if this
// server isn't the leader, returns false. otherwise start the
// agreement and return immediately. there is no guarantee that this
// command will ever be committed to the Raft log, since the leader
// may fail or lose an election.
//
// the first return value is the index that the command will appear at
// if it's ever committed. the second return value is the current
// term. the third return value is true if this server believes it is
// the leader.
//
func (rf *Raft) Start(command interface{}) (int, int, bool) {
	index := -1
	term := -1
	isLeader := true

	// Your code here (2B).


	return index, term, isLeader
}

//
// the tester calls Kill() when a Raft instance won't
// be needed again. you are not required to do anything
// in Kill(), but it might be convenient to (for example)
// turn off debug output from this instance.
//
func (rf *Raft) Kill() {
	// Your code here, if desired.
	rf.mu.Lock()
	defer rf.mu.Unlock()
	
	// Close the heartbeat channel to signal the election loop to stop
	close(rf.heartbeat)
	
	// Reset state
	rf.peerState = Follower
	rf.currentTerm = 0
	rf.votedFor = -1
	rf.numberOfVotes = 0
}

//
// the service or tester wants to create a Raft server. the ports
// of all the Raft servers (including this one) are in peers[]. this
// server's port is peers[me]. all the servers' peers[] arrays
// have the same order. persister is a place for this server to
// save its persistent state, and also initially holds the most
// recent saved state, if any. applyCh is a channel on which the
// tester or service expects Raft to send ApplyMsg messages.
// Make() must return quickly, so it should start goroutines
// for any long-running work.
//
func Make(peers []*labrpc.ClientEnd, me int,
	persister *Persister, applyCh chan ApplyMsg) *Raft {
	rf := &Raft{}
	rf.peers = peers
	rf.persister = persister
	rf.me = me

	// Your initialization code here (2A, 2B, 2C).
	rf.peerState = Follower
	rf.heartbeat = make(chan bool)
	go main_loop_election(rf)

	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())


	return rf
}

func main_loop_election(rf *Raft) {
	for { // Infinite loop
		switch rf.peerState {
		case Follower:
			select { // "Switch case for channels". Allows any "ready" alternative to proceed among
			case <-rf.heartbeat: 
			case <-time.After(time.Duration(250+rand.Intn(250)) * time.Millisecond): // Election timeout
				rf.peerState = Candidate
			}

		case Candidate:
			rf.broadcastRequestVote()

			select {
			case <-rf.heartbeat:
				rf.peerState = Follower
			case <-time.After(time.Duration(250+rand.Intn(250)) * time.Millisecond): // Election timeout
			}

		case Leader:
			rf.broadcastAppendEntries()

			time.Sleep(150 * time.Millisecond)
		}
	}
}

func (rf *Raft) broadcastRequestVote() {
	rf.currentTerm += 1
	rf.votedFor = rf.me
	rf.numberOfVotes = 1

	args := &RequestVoteArgs{
		Term:        rf.currentTerm,
		CandidateId: rf.me,
	}

	reply := &RequestVoteReply{
		Term:        rf.currentTerm,
		VoteGranted: false,
	}

	for server := 0; server < len(rf.peers); server++ {
		if server != rf.me {
			go rf.sendRequestVote(server, args, reply)
		}
	}
}

func (rf *Raft) broadcastAppendEntries() {
	for server := 0; server < len(rf.peers); server++ {
		if server != rf.me {
			args := &AppendEntries{
				Term: rf.currentTerm,
			}
			reply := &AppendEntriesReply{
				Term:    rf.currentTerm,
				Success: false,
			}
			go rf.sendAppendEntries(server, args, reply)
		}
	}
}
