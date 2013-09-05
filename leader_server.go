package graft

type LeaderServer struct {
	Voter
}

func (server *Server) AppendEntries(data ...string) {
	message := server.GenerateAppendEntries(data...)

	for _, peer := range server.Peers {
		peer.ReceiveAppendEntries(message)
	}

	server.updateLog(message.PrevLogIndex, message.Entries)
	server.CommitIndex++
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
