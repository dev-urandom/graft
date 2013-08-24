package graft

import (
	"github.com/benmills/quiz"
	"testing"
	"time"
)

type SpyServer struct {
	electionStarted bool
	electionCount   int
}

func (server *SpyServer) StartElection() {
	server.electionStarted = true
	server.electionCount += 1
}

func TestTimerTellsServerToStartElectionWhenReceivingOnTimeoutChannel(t *testing.T) {
	test := quiz.Test(t)

	spyServer := &SpyServer{electionStarted: false, electionCount: 0}
	timer := NewElectionTimer(1, spyServer)
	timer.ElectionChannel <- 1
	test.Expect(spyServer.electionStarted).ToBeTrue()
	timer.ShutDown()
}

func TestTimer(t *testing.T) {
	test := quiz.Test(t)

	spyServer := &SpyServer{electionStarted: false, electionCount: 0}
	timer := NewElectionTimer(1, spyServer)

	timer.StartTimer()

	time.Sleep(100000)
	test.Expect(spyServer.electionStarted).ToBeTrue()
	timer.ShutDown()
}

func TestTimerStartsMultipleElections(t *testing.T) {
	test := quiz.Test(t)

	spyServer := &SpyServer{electionStarted: false, electionCount: 0}
	timer := NewElectionTimer(2, spyServer)

	timer.StartTimer()

	time.Sleep(100000)
	test.Expect(spyServer.electionCount).ToBeGreaterThan(1)
	timer.ShutDown()
}
