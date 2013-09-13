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
	server.commitTo(message.CommitIndex)

	return AppendEntriesResponseMessage{
		Success: true,
	}, nil
}

func (server *Server) commitTo(i int) {
	if server.CommitIndex == 0 && i > 0 {
		server.StateMachine.Commit(server.Log[0].Data)
		server.CommitIndex = 1
	}
}

func (server *Server) updateLog(prevLogIndex int, entries []LogEntry) {
	if len(server.Log) == 0 {
		server.Log = entries
	}
	for i, entry := range entries {
		server.Log[i+prevLogIndex] = entry
	}
}
