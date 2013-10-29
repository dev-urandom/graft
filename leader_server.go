package graft


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

func (s *Server) sendHeartBeat() {
	message := s.GenerateAppendEntries()
	appendResponseChan := make(chan AppendEntriesResponseMessage)
	finishedChannel := make(chan bool)
	peerFailureChannel := make(chan int)

	go s.listenForPeerResponses(appendResponseChan, peerFailureChannel, finishedChannel)

	s.broadcastToPeers(message, appendResponseChan, peerFailureChannel)

	select {
	case <-finishedChannel:
	}
}

func (s *Server) GenerateAppendEntries(data ...string) AppendEntriesMessage {
	entries := []LogEntry{}
	for _, d := range data {
		entries = append(entries, LogEntry{Term: s.Term, Data: d})
	}

	return AppendEntriesMessage{
		Term:         s.Term,
		LeaderId:     s.Id,
		PrevLogIndex: s.LastLogIndex(),
		PrevLogTerm:  s.prevLogTerm(),
		Entries:      entries,
		CommitIndex:  s.CommitIndex,
	}
}

func (s *Server) rolledBackMessage(m AppendEntriesMessage) AppendEntriesMessage {
	prevLogIndex := m.PrevLogIndex - 1
	var prevLogTerm int
	var entries []LogEntry
	if prevLogIndex <= 0 {
		prevLogTerm = 0
		entries = s.Log
	} else {
		prevLogTerm = s.Log[m.PrevLogIndex-2].Term
		entries = s.Log[m.PrevLogIndex-2:]
	}
	return AppendEntriesMessage{
		Term:         s.Term,
		LeaderId:     s.Id,
		PrevLogIndex: prevLogIndex,
		PrevLogTerm:  prevLogTerm,
		Entries:      entries,
		CommitIndex:  s.CommitIndex,
	}

}

func (s *Server) prevLogTerm() int {
	if s.LastLogIndex() > len(s.Log) || s.LastLogIndex() == 0 {
		return 0
	} else {
		return s.Log[s.LastLogIndex()-1].Term
	}
}

func (s *Server) broadcastToPeers(message AppendEntriesMessage,
	responseChannel chan AppendEntriesResponseMessage, peerFailureChannel chan int) {
	for _, peer := range s.Peers {
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
						message = s.rolledBackMessage(message)
						if message.PrevLogIndex < 0 {
							// We need to do some kind of logging. This means that for some reason
							// we can not replay a log onto a peer
							peerFailureChannel <- 0
							return
						}
					}
				}
			}
			peerFailureChannel <- 0
			return
		}(5, peer)
	}
}

func (s *Server) listenForPeerResponses(responseChannel chan AppendEntriesResponseMessage,
	peerFailureChannel chan int, finishedChannel chan bool) {
	peerCount := len(s.Peers)
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
