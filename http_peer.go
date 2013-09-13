package graft

type HttpPeer struct {
}

func NewHttpPeer(url string) HttpPeer {
	return HttpPeer{}
}

func (peer HttpPeer) ReceiveAppendEntries(message AppendEntriesMessage) AppendEntriesResponseMessage {
	return AppendEntriesResponseMessage{}
}

func (peer HttpPeer) ReceiveRequestVote(message RequestVoteMessage) (VoteResponseMessage, error) {
	return VoteResponseMessage{}, nil
}
