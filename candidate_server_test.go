package graft

import (
	"github.com/benmills/quiz"
	"github.com/wjdix/tiktok"
	"testing"
)

func TestGenerateRequestVoteDerivedFromLog(t *testing.T) {
	test := quiz.Test(t)

	server := New("id")
	server.Log = []LogEntry{LogEntry{Term: 1, Data: "test"}, LogEntry{Term: 1, Data: "foo"}}
	newRequestVote := server.RequestVote()

	test.Expect(newRequestVote.Term).ToEqual(1)
	test.Expect(newRequestVote.CandidateId).ToEqual(server.Id)
	test.Expect(newRequestVote.LastLogIndex).ToEqual(2)
	test.Expect(newRequestVote.LastLogTerm).ToEqual(1)
}

func TestReceiveVoteResponseEndsElectionForHigherTerm(t *testing.T) {
	test := quiz.Test(t)

	server := New("id")
	server.Term = 0
	server.State = Candidate

	server.ReceiveVoteResponse(VoteResponseMessage{
		VoteGranted: false,
		Term:        2,
	})

	test.Expect(server.State).ToEqual(Follower)
	test.Expect(server.Term).ToEqual(2)
	test.Expect(server.VotesGranted).ToEqual(0)
}

func TestReceiveVoteResponseTalliesVoteGranted(t *testing.T) {
	test := quiz.Test(t)

	server := New("id")
	server.Term = 0
	server.State = Candidate

	server.ReceiveVoteResponse(VoteResponseMessage{
		VoteGranted: true,
		Term:        0,
	})

	test.Expect(server.VotesGranted).ToEqual(1)
	test.Expect(server.State).ToEqual(Candidate)
	test.Expect(server.Term).ToEqual(0)
}

func TestReceiveVoteResponseWithoutGrantingVoteDoesNotTaillyButContinuesTheElection(t *testing.T) {
	test := quiz.Test(t)

	server := New("id")
	server.Term = 0
	server.State = Candidate

	server.ReceiveVoteResponse(VoteResponseMessage{
		VoteGranted: false,
		Term:        0,
	})

	test.Expect(server.State).ToEqual(Candidate)
	test.Expect(server.Term).ToEqual(0)
	test.Expect(server.VotesGranted).ToEqual(0)
}

func TestServerCanWinElection(t *testing.T) {
	test := quiz.Test(t)

	serverA := New("id")
	serverB := New("id")
	serverC := New("id")
	serverA.AddPeers(serverB, serverC)

	serverA.StartElection()

	test.Expect(serverA.State).ToEqual(Leader)
	test.Expect(serverA.Term).ToEqual(1)
	test.Expect(serverA.VotesGranted).ToEqual(2)

	test.Expect(serverA.VotedFor).ToEqual(serverA.Id)
	test.Expect(serverB.VotedFor).ToEqual(serverA.Id)
	test.Expect(serverC.VotedFor).ToEqual(serverA.Id)
}

func TestServerCanWinElectionWithRetries(t *testing.T) {
	test := quiz.Test(t)

	serverA := New("id")
	serverB := New("id")
	serverC := &FailingPeer{
		numberOfFails: 1,
		successfulResponse: VoteResponseMessage{
			Term:        0,
			VoteGranted: true,
		},
	}

	serverA.AddPeers(serverB, serverC)

	serverA.StartElection()

	test.Expect(serverA.State).ToEqual(Leader)
	test.Expect(serverA.Term).ToEqual(1)
	test.Expect(serverA.VotesGranted).ToEqual(2)
}

func TestServerCanStartAndWinElectionAfterElectionTimeout(t *testing.T) {
	test := quiz.Test(t)

	serverA := New("id")
	timer := NewElectionTimer(1, serverA)
	timer.tickerBuilder = FakeTicker
	defer tiktok.ClearTickers()
	serverA.ElectionTimer = timer
	serverB := New("id")
	serverC := New("id")
	serverA.AddPeers(serverB, serverC)

	serverA.Start()
	tiktok.Tick(1)

	timer.ShutDown()
	test.Expect(serverA.State).ToEqual(Leader)
	test.Expect(serverA.Term).ToEqual(1)
	test.Expect(serverA.VotesGranted).ToEqual(2)

	test.Expect(serverB.VotedFor).ToEqual(serverA.Id)
	test.Expect(serverC.VotedFor).ToEqual(serverA.Id)
}
