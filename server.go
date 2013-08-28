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
	CandidateServer
}

func New() *Server {
	serverBase := ServerBase{
		Id:            "",
		Log:           []LogEntry{},
		Term:          0,
		VotedFor:      "",
		VotesGranted:  0,
		State:         Follower,
		Peers:         []Peer{},
		ElectionTimer: NullTimer{},
	}
	return &Server{CandidateServer{serverBase}}
}

func (server *Server) Start() {
	server.ElectionTimer.StartTimer()
}

func (server *Server) ReceiveRequestVote(message RequestVoteMessage) (VoteResponseMessage, error) {
	if server.Term < message.Term && server.logUpToDate(message) {
		server.stepDown()
		server.Term = message.Term
		server.ElectionTimer.Reset()

		return VoteResponseMessage{
			Term:        server.Term,
			VoteGranted: true,
		}, nil
	} else {
		return VoteResponseMessage{
			Term:        server.Term,
			VoteGranted: false,
		}, nil
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

func requestVoteFromPeer(peer Peer, message RequestVoteMessage) VoteResponseMessage {
	response, err := peer.ReceiveRequestVote(message)
	if err != nil {
		return requestVoteFromPeer(peer, message)
	}

	return response
}
