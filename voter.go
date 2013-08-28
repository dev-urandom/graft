package graft

type Voter struct {
	CandidateServer
}

func (server *Server) ReceiveRequestVote(message RequestVoteMessage) (VoteResponseMessage, error) {
	if server.Term < message.Term && server.logUpToDate(message) {
		server.stepDown()
		server.Term = message.Term
		server.ElectionTimer.Reset()

		return VoteResponseMessage{
			Term:        server.Term,
			VoteGranted: true,
		}, nil
	} else {
		return VoteResponseMessage{
			Term:        server.Term,
			VoteGranted: false,
		}, nil
	}
}
