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
		Term: server.Term,
		CandidateId: server.Id,
		LastLogIndex: server.lastLogIndex(),
		LastLogTerm: server.lastLogTerm(),
	}
}

func (server *Server) lastLogIndex() int {
	return len(server.Log)
}

func (server *Server) lastLogTerm() int {
	return 0
}

func (server *Server) ReceiveRequestVote(message RequestVoteMessage) VoteResponseMessage {
	if server.Term < message.Term {
		server.Term = message.Term

		return VoteResponseMessage {
			Term: server.Term,
			VoteGranted: true,
		}
	} else {
		return VoteResponseMessage {
			Term: server.Term,
			VoteGranted: false,
		}
	}
}
