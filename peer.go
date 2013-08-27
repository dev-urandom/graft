package graft

type Peer interface {
	ReceiveRequestVote(RequestVoteMessage) VoteResponseMessage
}
