package graft

type LeaderServer struct {
	Voter
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
