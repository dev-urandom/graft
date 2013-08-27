package graft

const (
	Candidate = "candidate"
	Follower  = "follower"
	Leader    = "leader"
)

type NullTimer struct{}

func (timer NullTimer) Reset()      {}
func (timer NullTimer) StartTimer() {}

type Timable interface {
	Reset()
	StartTimer()
}

type Commiter interface {
	Commit(string)
}

type Server struct {
	Id            string
	Log           []LogEntry
	Term          int
	VotedFor      string
	VotesGranted  int
	State         string
	Peers         []Peer
	ElectionTimer Timable
	StateMachine  Commiter
	CommitIndex   int
}

func New() *Server {
	return &Server{
		Id:            "",
		Log:           []LogEntry{},
		Term:          0,
		VotedFor:      "",
		VotesGranted:  0,
		State:         Follower,
		Peers:         []Peer{},
		ElectionTimer: NullTimer{},
	}
}

func (server *Server) AddPeer(peer Peer) {
	server.Peers = append(server.Peers, peer)
}

func (server *Server) Start() {
	server.ElectionTimer.StartTimer()
}

func (server *Server) StartElection() {
	requestVoteMessage := server.RequestVote()
	for _, peer := range server.Peers {
		response := peer.ReceiveRequestVote(requestVoteMessage)
		server.ReceiveVoteResponse(response)
	}

	if server.VotesGranted > (len(server.Peers) / 2) {
		server.State = Leader
	}
}

func (server *Server) RequestVote() RequestVoteMessage {
	server.State = Candidate
	server.Term++

	return RequestVoteMessage{
		Term:         server.Term,
		CandidateId:  server.Id,
		LastLogIndex: server.lastLogIndex(),
		LastLogTerm:  server.lastLogTerm(),
	}
}

func (server *Server) ReceiveRequestVote(message RequestVoteMessage) VoteResponseMessage {
	if server.Term < message.Term && server.logUpToDate(message) {
		server.stepDown()
		server.Term = message.Term
		server.ElectionTimer.Reset()

		return VoteResponseMessage{
			Term:        server.Term,
			VoteGranted: true,
		}
	} else {
		return VoteResponseMessage{
			Term:        server.Term,
			VoteGranted: false,
		}
	}
}

func (server *Server) ReceiveVoteResponse(message VoteResponseMessage) {
	if message.VoteGranted {
		server.VotesGranted++
	} else {
		server.Term = message.Term
		server.State = Follower
	}
}

func (server *Server) ReceiveAppendEntries(message AppendEntriesMessage) AppendEntriesResponseMessage {
	server.stepDown()
	if server.Term < message.Term {
		server.Term = message.Term
	}

	if server.Term > message.Term || server.invalidLog(message) {
		return AppendEntriesResponseMessage{
			Success: false,
		}
	}

	server.ElectionTimer.Reset()
	server.updateLog(message.PrevLogIndex, message.Entries)
	server.commitTo(message.CommitIndex)

	return AppendEntriesResponseMessage{
		Success: true,
	}
}

func (server *Server) AppendEntries() AppendEntriesMessage {
	return AppendEntriesMessage{
		Term:         server.Term,
		LeaderId:     server.Id,
		PrevLogIndex: server.lastLogIndex(),
		Entries:      []LogEntry{},
		CommitIndex:  server.lastCommitIndex(),
	}
}

func (server *Server) commitTo(i int) {
	if server.CommitIndex == 0 && i > 0 {
		server.StateMachine.Commit(server.Log[0].Data)
		server.CommitIndex = 1
	}
}

func (server *Server) lastCommitIndex() int {
	return server.CommitIndex
}

func (server *Server) lastLogIndex() int {
	return len(server.Log)
}

func (server *Server) lastLogTerm() int {
	if len(server.Log) == 0 {
		return 0
	} else {
		return (server.Log[len(server.Log)-1]).Term
	}
}

func (server *Server) stepDown() {
	server.State = Follower
}

func (server *Server) invalidLog(message AppendEntriesMessage) bool {
	if message.PrevLogIndex == 0 {
		return false
	}

	return len(server.Log) < message.PrevLogIndex || server.Log[message.PrevLogIndex-1].Term != message.PrevLogTerm
}

func (server *Server) updateLog(prevLogIndex int, entries []LogEntry) {
	if len(server.Log) == 0 {
		server.Log = entries
	}
	for i, entry := range entries {
		server.Log[i+prevLogIndex] = entry
	}
}

func (server *Server) logUpToDate(message RequestVoteMessage) bool {
	return server.lastLogIndex() <= message.LastLogIndex && server.lastLogTerm() <= message.LastLogTerm
}
