package graft

type LeaderServer struct {
	Voter
}

const (
	appendEntriesMaxRetries = 0
)

func (server *Server) AppendEntries(data ...string) {
	message := server.GenerateAppendEntries(data...)
	successfulAppends := 1
	appendResponseChan := make(chan AppendEntriesResponseMessage)
	finishedChan := make(chan bool)
	failedPeerChan := make(chan int)
	go func(peercount int) {
		received := 0
		for received < peercount {
			select {
			case <-failedPeerChan:
				received++
			case response := <-appendResponseChan:
				received++
				if response.Success {
					successfulAppends++
				}
				if successfulAppends > (peercount / 2) {
					finishedChan <- true
					return
				}
			}
		}
		finishedChan <- false
		return
	}(len(server.Peers))

	for _, peer := range server.Peers {
		go func(maxFailures int, target Peer) {
			failureCount := 0
			for failureCount < maxFailures {
				response, err := target.ReceiveAppendEntries(message)
				if err != nil {
					failureCount++
				} else {
					appendResponseChan <- response
					return
				}
			}
			failedPeerChan <- 0
			return
		}(5, peer)
	}

	select {
	case success := <-finishedChan:
		if success {
			server.updateLog(message.PrevLogIndex, message.Entries)
			server.CommitIndex++
		}
	}
}

func (server *Server) GenerateAppendEntries(data ...string) AppendEntriesMessage {
	entries := []LogEntry{}
	for _, d := range data {
		entries = append(entries, LogEntry{Term: server.Term, Data: d})
	}

	return AppendEntriesMessage{
		Term:         server.Term,
		LeaderId:     server.Id,
		PrevLogIndex: server.LastLogIndex(),
		Entries:      entries,
		CommitIndex:  server.LastCommitIndex(),
	}
}
