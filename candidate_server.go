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
	for _, peer := range server.Peers {
		response := requestVoteFromPeer(peer, requestVoteMessage)
		server.ReceiveVoteResponse(response)
	}

	if server.VotesGranted > (len(server.Peers) / 2) {
		server.State = Leader
	} else {
		server.State = Follower
	}
}

func (server *CandidateServer) ReceiveVoteResponse(message VoteResponseMessage) {
	if message.VoteGranted {
		server.VotesGranted++
	} else if server.Term < message.Term {
		server.Term = message.Term
		server.State = Follower
	}
}
