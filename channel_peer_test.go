package graft

import (
	"github.com/benmills/quiz"
	"testing"
)

func TestChannelPeerCanBeAddedToAServersListOfPeers(t *testing.T) {
	server := New()
	peer := NewChannelPeer(server)
	server2 := New()
	server2.Peers = []Peer{peer}
}

func TestChannelPeerRespondsToVoteMessages(t *testing.T) {
	test := quiz.Test(t)

	server := New()
	peer := NewChannelPeer(server)
	peer.Start()
	defer peer.ShutDown()
	requestVote := RequestVoteMessage{
		Term:         1,
		CandidateId:  "foo",
		LastLogIndex: 0,
		LastLogTerm:  0,
	}
	response, err := peer.ReceiveRequestVote(requestVote)

	test.Expect(err).ToEqual(nil)
	test.Expect(response.VoteGranted).ToBeTrue()
	test.Expect(server.Term).ToEqual(1)
}

func TestChannelPeerRespondsToAppendEntriesMessages(t *testing.T) {
	test := quiz.Test(t)

	server := New()
	peer := NewChannelPeer(server)
	peer.Start()
	defer peer.ShutDown()
	message := AppendEntriesMessage{
		Term:         2,
		LeaderId:     "leader_id",
		PrevLogIndex: 2,
		Entries:      []LogEntry{},
		CommitIndex:  0,
	}

	peer.ReceiveAppendEntries(message)

	test.Expect(server.Term).ToEqual(2)
}
