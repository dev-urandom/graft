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
	if server.Term < message.Term && server.logUpToDate(message) {
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

	server.updateLog(message.PrevLogIndex, message.Entries)

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

func (server *Server) invalidLog(message AppendEntriesMessage) bool {
	if message.PrevLogIndex == 0 {
		return false
	}

	return len(server.Log) < message.PrevLogIndex || server.Log[message.PrevLogIndex-1].Term != message.PrevLogTerm
}

func (server *Server) updateLog(prevLogIndex int, entries []LogEntry) {
	for i, entry := range entries {
		server.Log[i+prevLogIndex] = entry
	}
}

func (server *Server) logUpToDate(message RequestVoteMessage) bool {
	return server.lastLogIndex() <= message.LastLogIndex && server.lastLogTerm() <= message.LastLogTerm
}
