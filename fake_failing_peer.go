package graft

import (
	"errors"
)

type FailingPeer struct {
	numberOfFails      int
	successfulResponse VoteResponseMessage
	failureAppendEntriesResponse AppendEntriesResponseMessage
	successfulAppendEntriesResponse AppendEntriesResponseMessage
	Log []LogEntry
}

func (peer *FailingPeer) ReceiveAppendEntries(message AppendEntriesMessage) AppendEntriesResponseMessage {
	if peer.shouldFail() {
		return peer.failureAppendEntriesResponse
	}

	for _, entry := range message.Entries {
		peer.Log = append(peer.Log, entry)
	}

	return peer.successfulAppendEntriesResponse
}

func (peer *FailingPeer) ReceiveRequestVote(message RequestVoteMessage) (VoteResponseMessage, error) {
	if peer.shouldFail() {
		return VoteResponseMessage{}, errors.New("boom")
	}

	return peer.successfulResponse, nil
}

func (peer *FailingPeer) shouldFail() bool {
	if peer.numberOfFails > 0 {
		peer.numberOfFails--
		return true
	} else if peer.numberOfFails == -1 {
		return true
	} else {
		return false
	}
}
