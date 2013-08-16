package graft

import (
	"github.com/benmills/quiz"
	"testing"
)

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

func TestGenerateRequestVote(t *testing.T) {
	test := quiz.Test(t)

	server := New()
	newRequestVote := server.RequestVote()

	test.Expect(newRequestVote.Term).ToEqual(1)
	test.Expect(newRequestVote.CandidateId).ToEqual(server.Id)
	test.Expect(newRequestVote.LastLogIndex).ToEqual(0)
	test.Expect(newRequestVote.LastLogTerm).ToEqual(0)
}

func TestReceiveRequestVoteNotSuccessfulForSmallerTerm(t *testing.T) {
	test := quiz.Test(t)

	server := New()
	server.Term = 2
	message := RequestVoteMessage {
		Term: 1,
		CandidateId: "other_server_id",
		LastLogIndex: 0,
		LastLogTerm: 0,
	}

	voteResponse := server.ReceiveRequestVote(message)

	test.Expect(voteResponse.Term).ToEqual(2)
	test.Expect(voteResponse.VoteGranted).ToBeFalse()
}

func TestReceiveRequestVoteUpdatesServerTerm(t *testing.T) {
	test := quiz.Test(t)

	server := New()
	server.Term = 1
	message := RequestVoteMessage {
		Term: 2,
		CandidateId: "other_server_id",
		LastLogIndex: 0,
		LastLogTerm: 0,
	}

	server.ReceiveRequestVote(message)

	test.Expect(server.Term).ToEqual(2)
}

func TestRecieveRequestVoteWithHigherTermCausesVoterToStepDown(t *testing.T) {
	test := quiz.Test(t)

	server := New()
	server.Term = 1
	server.State = Candidate
	message := RequestVoteMessage {
		Term: 2,
		CandidateId: "other_server_id",
		LastLogIndex: 0,
		LastLogTerm: 0,
	}

	server.ReceiveRequestVote(message)

	test.Expect(server.State).ToEqual(Follower)
}
