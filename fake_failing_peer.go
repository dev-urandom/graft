package graft

import (
	"errors"
)

type FailingPeer struct {
	numberOfFails      int
	successfulResponse VoteResponseMessage
	failureAppendEntriesResponse AppendEntriesResponseMessage
	successfulAppendEntriesResponse AppendEntriesResponseMessage
}

func (peer *FailingPeer) ReceiveAppendEntries(message AppendEntriesMessage) AppendEntriesResponseMessage {
	if peer.numberOfFails > 0 {
		peer.numberOfFails--
		return peer.failureAppendEntriesResponse
	}

	return peer.successfulAppendEntriesResponse
}

func (peer *FailingPeer) ReceiveRequestVote(message RequestVoteMessage) (VoteResponseMessage, error) {
	if peer.numberOfFails > 0 {
		peer.numberOfFails--
		return VoteResponseMessage{}, errors.New("boom")
	}

	return peer.successfulResponse, nil
}
