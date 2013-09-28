package graft

import "fmt"

type LeaderServer struct {
	Voter
}

func (server *Server) AppendEntries(data ...string) {
	message := server.GenerateAppendEntries(data...)
	appendResponseChan := make(chan AppendEntriesResponseMessage)
	finishedChannel := make(chan bool)
	peerFailureChannel := make(chan int)

	go server.listenForPeerResponses(appendResponseChan, peerFailureChannel, finishedChannel)

	server.broadcastToPeers(message, appendResponseChan, peerFailureChannel)

	select {
	case success := <-finishedChannel:
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
		PrevLogTerm:  server.prevLogTerm(),
		Entries:      entries,
		CommitIndex:  server.CommitIndex,
	}
}

func (server *Server) rolledBackMessage(m AppendEntriesMessage) AppendEntriesMessage {
	prevLogIndex := m.PrevLogIndex - 1
	var prevLogTerm int
	var entries []LogEntry
	if prevLogIndex <= 0 {
		prevLogTerm = 0
		entries = server.Log
	} else {
		prevLogTerm = server.Log[m.PrevLogIndex-2].Term
		entries = server.Log[m.PrevLogIndex-2:]
	}
	return AppendEntriesMessage{
		Term:         server.Term,
		LeaderId:     server.Id,
		PrevLogIndex: prevLogIndex,
		PrevLogTerm:  prevLogTerm,
		Entries:      entries,
		CommitIndex:  server.CommitIndex,
	}

}

func (server *Server) prevLogTerm() int {
	if server.LastLogIndex() > len(server.Log) || server.LastLogIndex() == 0 {
		return 0
	} else {
		return server.Log[server.LastLogIndex()-1].Term
	}
}

func (server *Server) broadcastToPeers(message AppendEntriesMessage,
	responseChannel chan AppendEntriesResponseMessage, peerFailureChannel chan int) {
	for _, peer := range server.Peers {
		go func(maxFailures int, target Peer) {
			failureCount := 0
			for failureCount < maxFailures {
				response, err := target.ReceiveAppendEntries(message)
				if err != nil {
					failureCount++
				} else {
					if response.Success {
						responseChannel <- response
						return
					} else {
						message = server.rolledBackMessage(message)
						if message.PrevLogIndex < 0 {
							// We need to do some kind of logging. This means that for some reason
							// we can not replay a log onto a peer
							peerFailureChannel <- 0
							return
						}
						fmt.Println(message)
					}
				}
			}
			peerFailureChannel <- 0
			return
		}(5, peer)
	}
}

func (server *Server) listenForPeerResponses(responseChannel chan AppendEntriesResponseMessage,
	peerFailureChannel chan int, finishedChannel chan bool) {
	peerCount := len(server.Peers)
	received := 0
	successfulAppends := 1
	for received < peerCount {
		select {
		case <-peerFailureChannel:
			received++
		case response := <-responseChannel:
			received++
			if response.Success {
				successfulAppends++
			}
			if successfulAppends > (peerCount / 2) {
				finishedChannel <- true
				return
			}
		}
	}
	finishedChannel <- false
	return
}
