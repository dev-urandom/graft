package graft

import (
	"errors"
)

type FailingPeer struct {
	numberOfFails      int
	successfulResponse VoteResponseMessage
}

func (peer *FailingPeer) ReceiveAppendEntries(message AppendEntriesMessage) AppendEntriesResponseMessage {
	return AppendEntriesResponseMessage{}
}

func (peer *FailingPeer) ReceiveRequestVote(message RequestVoteMessage) (VoteResponseMessage, error) {
	if peer.numberOfFails > 0 {
		peer.numberOfFails--
		return VoteResponseMessage{}, errors.New("boom")
	}

	return peer.successfulResponse, nil
}
