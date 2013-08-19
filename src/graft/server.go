package graft

const (
	Candidate = "candidate"
	Follower  = "follower"
	Leader    = "leader"
)

type Server struct {
	Id       string
	Log      []LogEntry
	Term     int
	VotedFor string
	State    string
}

func New() *Server {
	return &Server{
		Id:       "",
		Log:      []LogEntry{},
		Term:     0,
		VotedFor: "",
		State:    Follower,
	}
}

func (server *Server) RequestVote() RequestVoteMessage {
	server.Term++

	return RequestVoteMessage{
		Term:         server.Term,
		CandidateId:  server.Id,
		LastLogIndex: server.lastLogIndex(),
		LastLogTerm:  server.lastLogTerm(),
	}
}

func (server *Server) ReceiveRequestVote(message RequestVoteMessage) VoteResponseMessage {
	if server.Term < message.Term {
		server.stepDown()
		server.Term = message.Term

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

func (server *Server) ReceiveAppendEntries(message AppendEntriesMessage) {
	server.stepDown()
	if server.Term < message.Term {
		server.Term = message.Term
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

func (server *Server) lastCommitIndex() int {
	return server.lastLogIndex()
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
