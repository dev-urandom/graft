package graft

type Server struct {
	Id string
	Log []string
	Term int
	VotedFor string
}

func New() *Server {
	return &Server{
		Id: "",
		Log: []string{},
		Term: 0,
		VotedFor: "",
	}
}

func (server *Server) RequestVote() RequestVoteMessage {
	server.Term++

	return RequestVoteMessage {
		term: server.Term,
		candidateId: server.Id,
		lastLogIndex: server.lastLogIndex(),
		lastLogTerm: server.lastLogTerm(),
	}
}

func (server *Server) lastLogIndex() int {
	return len(server.Log)
}

func (server *Server) lastLogTerm() int {
	return 0
}
