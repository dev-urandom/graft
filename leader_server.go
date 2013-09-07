package graft

type LeaderServer struct {
	Voter
}

const (
	appendEntriesMaxRetries = 10
)

func (server *Server) AppendEntries(data ...string) {
	message := server.GenerateAppendEntries(data...)
	successfulAppends := 0

	for _, peer := range server.Peers {
		if server.sendEntriesToPeerWithRetries(peer, message, appendEntriesMaxRetries) {
			successfulAppends++
		}
	}

	server.updateLog(message.PrevLogIndex, message.Entries)

	if successfulAppends > (len(server.Peers) / 2) {
		server.CommitIndex++
	}
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

func (server *Server) sendEntriesToPeerWithRetries(peer Peer, message AppendEntriesMessage, retries int) bool {
		response := peer.ReceiveAppendEntries(message)
		if response.Success {
			return true
		}

		if retries > 0 {
			retries--
			return server.sendEntriesToPeerWithRetries(peer, message, retries)
		}

		return false
}
