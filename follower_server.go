package graft

type FollowerServer struct {
	LeaderServer
}

func (server *Server) ReceiveAppendEntries(message AppendEntriesMessage) (AppendEntriesResponseMessage, error) {
	server.stepDown()
	if server.Term < message.Term {
		server.Term = message.Term
	}

	if server.Term > message.Term || server.invalidLog(message) {
		return AppendEntriesResponseMessage{
			Success: false,
		}, nil
	}

	server.ElectionTimer.Reset()
	server.updateLog(message.PrevLogIndex, message.Entries)

	if message.CommitIndex > 0 {
		server.commitTo(message.CommitIndex)
	}

	return AppendEntriesResponseMessage{
		Success: true,
	}, nil
}

func (server *Server) commitTo(i int) {
	for i > server.CommitIndex {
		server.StateMachine.Commit(server.Log[server.CommitIndex].Data)
		server.CommitIndex++
	}
}

func (server *Server) updateLog(prevLogIndex int, entries []LogEntry) {
	if len(server.Log) == 0 {
		server.Log = entries
	} else {
		for i, entry := range entries {
			if i+prevLogIndex >= len(server.Log) {
				server.Log = append(server.Log, entry)
			} else {
				server.Log[i+prevLogIndex] = entry
			}
		}
	}
}
