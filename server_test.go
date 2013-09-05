package graft

import (
	"errors"
	"github.com/benmills/quiz"
	"testing"
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

type SpyStateMachine struct {
	messageChan chan string
}

func (machine SpyStateMachine) Commit(data string) {
	machine.messageChan <- data
}

type SpyTimer struct {
	resetChannel chan int
}

func (timer SpyTimer) Reset() {
	timer.resetChannel <- 1
}

func (timer SpyTimer) StartTimer() {}

func TestNewServerHasEmptyEntries(t *testing.T) {
	test := quiz.Test(t)

	server := New()
	test.Expect(len(server.Log)).ToEqual(0)
}

func TestNewServerStartsAtZeroTerm(t *testing.T) {
	test := quiz.Test(t)

	server := New()
	test.Expect(server.Term).ToEqual(0)
}

func TestNewServerStartsWithEmptyVotedFor(t *testing.T) {
	test := quiz.Test(t)

	server := New()
	test.Expect(server.VotedFor).ToEqual("")
}

func TestNewServerStartsAsFollower(t *testing.T) {
	test := quiz.Test(t)

	server := New()
	test.Expect(server.State).ToEqual(Follower)
}

func TestLastLogTermDerivedFromLogEntries(t *testing.T) {
	test := quiz.Test(t)

	server := New()
	server.Log = []LogEntry{LogEntry{Term: 1, Data: "test"}, LogEntry{Term: 2, Data: "foo"}}

	test.Expect(server.lastLogTerm()).ToEqual(2)
}

func TestReceiveVoteResponseReturnsAnError(t *testing.T) {
	test := quiz.Test(t)

	server := New()
	_, err := server.ReceiveRequestVote(RequestVoteMessage{})

	test.Expect(err).ToEqual(nil)
}

func TestServersHavePeers(t *testing.T) {
	test := quiz.Test(t)

	serverA := New()
	serverB := New()

	serverA.AddPeers(serverB)

	test.Expect(serverA.Peers[0]).ToEqual(serverB)
}
