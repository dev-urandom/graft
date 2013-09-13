package graft

type CandidateServer struct {
	Persister
}

func (server *CandidateServer) RequestVote() RequestVoteMessage {
	server.State = Candidate
	server.Term++

	return RequestVoteMessage{
		Term:         server.Term,
		CandidateId:  server.Id,
		LastLogIndex: server.lastLogIndex(),
		LastLogTerm:  server.lastLogTerm(),
	}
}

func (server *CandidateServer) StartElection() {
	requestVoteMessage := server.RequestVote()
	server.VotedFor = server.Id
	receiveVoteChan := make(chan VoteResponseMessage)
	electionFinishedChan := make(chan int)
	failedPeerChan := make(chan int)
	go func(peercount int) {
		received := 0
		for received < peercount {
			select {
			case <-failedPeerChan:
				received++
			case response := <-receiveVoteChan:
				received++
				ended := server.ReceiveVoteResponse(response)
				if ended {
					electionFinishedChan <- 1
					return
				}

			}
		}
		electionFinishedChan <- 1
		return
	}(len(server.Peers))

	for i, peer := range server.Peers {
		go func(maxFailures int, peerN int, target Peer) {
			failureCount := 0
			for failureCount < maxFailures {
				response, err := target.ReceiveRequestVote(requestVoteMessage)
				if err != nil {
					failureCount++
				} else {
					receiveVoteChan <- response
					return
				}
			}
			failedPeerChan <- 1
			return
		}(5, i, peer)
	}

	select {
	case <-electionFinishedChan:
		if server.VotesGranted >= (len(server.Peers) / 2) {
			server.State = Leader
		} else {
			server.State = Follower
		}
	}
}

func (server *CandidateServer) ReceiveVoteResponse(message VoteResponseMessage) bool {
	if message.VoteGranted {
		server.VotesGranted++
		if server.VotesGranted > (len(server.Peers) / 2) {
			return true
		}
	} else if server.Term < message.Term {
		server.Term = message.Term
		server.State = Follower
		return true
	}
	return false
}
