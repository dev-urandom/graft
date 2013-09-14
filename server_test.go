package graft

import (
	"github.com/benmills/quiz"
	"testing"
)

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

	server := New("id")
	test.Expect(len(server.Log)).ToEqual(0)
}

func TestNewServerStartsAtZeroTerm(t *testing.T) {
	test := quiz.Test(t)

	server := New("id")
	test.Expect(server.Term).ToEqual(0)
}

func TestNewServerStartsWithEmptyVotedFor(t *testing.T) {
	test := quiz.Test(t)

	server := New("id")
	test.Expect(server.VotedFor).ToEqual("")
}

func TestNewServerStartsAsFollower(t *testing.T) {
	test := quiz.Test(t)

	server := New("id")
	test.Expect(server.State).ToEqual(Follower)
}

func TestNewServerFromConfigurationSetsId(t *testing.T) {
	test := quiz.Test(t)

	configuration := ServerConfiguration{
		Id:    "foo",
		Peers: []string{"localhost:4000", "localhost:3000"},
	}

	server := NewFromConfiguration(configuration)

	test.Expect(server.Id).ToEqual("foo")
}

func TestNewServerFromConfigurationsBuildsListOfPeers(t *testing.T) {
	test := quiz.Test(t)

	configuration := ServerConfiguration{
		Id:    "foo",
		Peers: []string{"localhost:4000", "localhost:3000"},
	}

	server := NewFromConfiguration(configuration)

	test.Expect(len(server.Peers)).ToEqual(2)
	test.Expect(server.Peers[0].(HttpPeer).URL).ToEqual("localhost:4000")
}

func TestLastLogTermDerivedFromLogEntries(t *testing.T) {
	test := quiz.Test(t)

	server := New("id")
	server.Log = []LogEntry{LogEntry{Term: 1, Data: "test"}, LogEntry{Term: 2, Data: "foo"}}

	test.Expect(server.lastLogTerm()).ToEqual(2)
}

func TestReceiveVoteResponseReturnsAnError(t *testing.T) {
	test := quiz.Test(t)

	server := New("id")
	_, err := server.ReceiveRequestVote(RequestVoteMessage{})

	test.Expect(err).ToEqual(nil)
}

func TestServersHavePeers(t *testing.T) {
	test := quiz.Test(t)

	serverA := New("id")
	serverB := New("id")

	serverA.AddPeers(serverB)

	test.Expect(serverA.Peers[0]).ToEqual(serverB)
}
