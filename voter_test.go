package graft

import (
	"encoding/json"
	"github.com/benmills/quiz"
	"io/ioutil"
	"testing"
)

func TestReceiveRequestVoteNotSuccessfulForSmallerTerm(t *testing.T) {
	test := quiz.Test(t)

	server := New("id")
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

	server := New("id")
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

	server := New("id")
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

	server := New("id")
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

	server := New("id")
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

	server := New("id")
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

func TestPersistsStateBeforeVoting(t *testing.T) {
	defer cleanTmpDir()
	test := quiz.Test(t)
	server := NewFromConfiguration(
		ServerConfiguration{
			Id:                  "id",
			Peers:               []string{},
			PersistenceLocation: "tmp",
		},
	)

	message := RequestVoteMessage{
		Term:         1,
		CandidateId:  "other_server_id",
		LastLogIndex: 0,
		LastLogTerm:  0,
	}

	server.ReceiveRequestVote(message)

	file, err := ioutil.ReadFile("tmp/graft-stateid.json")
	if err != nil {
		t.Fail()
	}

	var state PersistedServerState
	json.Unmarshal(file, &state)
	test.Expect(state.CurrentTerm).ToEqual(1)
	test.Expect(state.VotedFor).ToEqual("other_server_id")
	test.Expect(server.State).ToEqual(Follower)
}
