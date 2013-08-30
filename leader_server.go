package graft

type LeaderServer struct {
	Voter
}

func (server *Server) GenerateAppendEntries(data ...string) AppendEntriesMessage {
	entries := []LogEntry{}
	for _, d := range(data) {
		entries = append(entries, LogEntry{Term: server.Term, Data: d})
	}

	return AppendEntriesMessage{
		Term:         server.Term,
		LeaderId:     server.Id,
		PrevLogIndex: server.lastLogIndex(),
		Entries:      entries,
		CommitIndex:  server.lastCommitIndex(),
	}
}
