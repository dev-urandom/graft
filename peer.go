package graft

type Peer interface {
	ReceiveRequestVote(RequestVoteMessage) (VoteResponseMessage, error)
	ReceiveAppendEntries(AppendEntriesMessage) AppendEntriesResponseMessage
}
