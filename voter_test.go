package graft

import (
	"github.com/benmills/quiz"
	"testing"
)

func TestReceiveRequestVoteNotSuccessfulForSmallerTerm(t *testing.T) {
	test := quiz.Test(t)

	server := New()
	server.Term = 2
	message := RequestVoteMessage{
		Term:         1,
		CandidateId:  "other_server_id",
		LastLogIndex: 0,
		LastLogTerm:  0,
	}

	voteResponse, _ := server.ReceiveRequestVote(message)

	test.Expect(voteResponse.Term).ToEqual(2)
	test.Expect(voteResponse.VoteGranted).ToBeFalse()
}

func TestReceiveRequestVoteNotSuccessfulForOutOfDateLogIndex(t *testing.T) {
	test := quiz.Test(t)

	server := New()
	server.Log = []LogEntry{LogEntry{Term: 0, Data: "some data"}}

	message := RequestVoteMessage{
		Term:         1,
		CandidateId:  "other_server_id",
		LastLogIndex: 0,
		LastLogTerm:  0,
	}

	voteResponse, _ := server.ReceiveRequestVote(message)

	test.Expect(voteResponse.VoteGranted).ToBeFalse()
}

func TestReceiveRequestVoteNotSuccessfulForOutOfDateLogTerm(t *testing.T) {
	test := quiz.Test(t)

	server := New()
	server.Term = 1
	server.Log = []LogEntry{LogEntry{Term: 1, Data: "some data"}}

	message := RequestVoteMessage{
		Term:         2,
		CandidateId:  "other_server_id",
		LastLogIndex: 1,
		LastLogTerm:  0,
	}

	voteResponse, _ := server.ReceiveRequestVote(message)

	test.Expect(voteResponse.VoteGranted).ToBeFalse()
}

func TestReceiveRequestVoteSuccess(t *testing.T) {
	test := quiz.Test(t)

	server := New()
	server.Term = 1
	message := RequestVoteMessage{
		Term:         2,
		CandidateId:  "other_server_id",
		LastLogIndex: 0,
		LastLogTerm:  0,
	}

	server.ReceiveRequestVote(message)

	test.Expect(server.Term).ToEqual(2)
	test.Expect(server.VotedFor).ToEqual("other_server_id")
}

func TestReceiveRequestVoteResetsElectionTimeout(t *testing.T) {
	test := quiz.Test(t)

	server := New()
	server.Term = 1
	timer := SpyTimer{make(chan int)}
	server.ElectionTimer = timer
	message := RequestVoteMessage{
		Term:         2,
		CandidateId:  "other_server_id",
		LastLogIndex: 0,
		LastLogTerm:  0,
	}

	shutDownChannel := make(chan int)

	go func(shutDownChannel chan int) {
		var resets int
		for {
			select {
			case <-timer.resetChannel:
				resets++
			case <-shutDownChannel:
				test.Expect(resets).ToEqual(1)
				return
			}
		}
	}(shutDownChannel)

	server.ReceiveRequestVote(message)
	shutDownChannel <- 1
}

func TestReceiveRequestVoteWithHigherTermCausesVoterToStepDown(t *testing.T) {
	test := quiz.Test(t)

	server := New()
	server.Term = 1
	server.State = Candidate
	message := RequestVoteMessage{
		Term:         2,
		CandidateId:  "other_server_id",
		LastLogIndex: 0,
		LastLogTerm:  0,
	}

	server.ReceiveRequestVote(message)

	test.Expect(server.State).ToEqual(Follower)
}
