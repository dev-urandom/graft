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

func New(id string) *Server {
	serverBase := ServerBase{
		Id:            id,
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

func NewFromConfiguration(config ServerConfiguration) *Server {
	server := New(config.Id)
	server.PersistenceLocation = config.PersistenceLocation
	for _, peerLocation := range config.Peers {
		server.AddPeers(HttpPeer{URL: peerLocation})
	}
	return server
}

func (server *Server) Start() {
	server.LoadPersistedState()
	server.LoadPersistedLog()
	server.ElectionTimer.StartTimer()
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
	return server.LastLogIndex() <= message.LastLogIndex && server.lastLogTerm() <= message.LastLogTerm
}
