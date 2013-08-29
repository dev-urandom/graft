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
	FollowerServer
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
	return &Server{FollowerServer{LeaderServer{Voter{CandidateServer{Persister{serverBase}}}}}}
}

func (server *Server) Start() {
	server.ElectionTimer.StartTimer()
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
