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
		LastLogIndex: server.LastLogIndex(),
		LastLogTerm:  server.lastLogTerm(),
	}
}

func (server *Server) StartElection() {
	requestVoteMessage := server.RequestVote()
	server.VotedFor = server.Id

	server.collectVotes(PeerBroadcast(requestVoteMessage, server.Peers))

	if server.VotesGranted >= (len(server.Peers) / 2) {
		server.State = Leader
		server.sendHeartBeat()
	} else {
		server.State = Follower
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

func (server *CandidateServer) collectVotes(b *broadcast) {
	for {
		select {
		case r := <-b.Response:
			endElection := server.ReceiveVoteResponse(r.VoteRes)
			r.Done()
			if endElection {
				return
			}
		case <-b.End:
			return
		}
	}
}
