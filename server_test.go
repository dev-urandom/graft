package graft

import (
	"github.com/benmills/quiz"
	"testing"
)

type SpyTimer struct {
	resetChannel chan int
}

func (timer SpyTimer) Reset() {
	timer.resetChannel <- 1
}

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

func TestGenerateRequestVoteDerivedFromLog(t *testing.T) {
	test := quiz.Test(t)

	server := New()
	server.Log = []LogEntry{LogEntry{Term: 1, Data: "test"}, LogEntry{Term: 1, Data: "foo"}}
	newRequestVote := server.RequestVote()

	test.Expect(newRequestVote.Term).ToEqual(1)
	test.Expect(newRequestVote.CandidateId).ToEqual(server.Id)
	test.Expect(newRequestVote.LastLogIndex).ToEqual(2)
	test.Expect(newRequestVote.LastLogTerm).ToEqual(1)
}

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

	voteResponse := server.ReceiveRequestVote(message)

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

	voteResponse := server.ReceiveRequestVote(message)

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

	voteResponse := server.ReceiveRequestVote(message)

	test.Expect(voteResponse.VoteGranted).ToBeFalse()
}

func TestReceiveRequestVoteUpdatesServerTerm(t *testing.T) {
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
}

func TestRecieveRequestVoteWithHigherTermCausesVoterToStepDown(t *testing.T) {
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

func TestGenerateAppendEntriesMessage(t *testing.T) {
	test := quiz.Test(t)

	server := New()

	message := server.AppendEntries()

	test.Expect(message.Term).ToEqual(0)
	test.Expect(message.LeaderId).ToEqual(server.Id)
	test.Expect(message.PrevLogIndex).ToEqual(0)
	test.Expect(len(message.Entries)).ToEqual(0)
	test.Expect(message.CommitIndex).ToEqual(0)
}

func TestAppendEntriesFailsWhenReceivedTermIsLessThanCurrentTerm(t *testing.T) {
	test := quiz.Test(t)

	server := New()
	server.Term = 3

	message := AppendEntriesMessage{
		Term:         2,
		LeaderId:     "leader_id",
		PrevLogIndex: 2,
		Entries:      []LogEntry{},
		CommitIndex:  0,
	}

	response := server.ReceiveAppendEntries(message)

	test.Expect(response.Success).ToBeFalse()
}

func TestAppendEntiesFailsWhenLogContainsNothingAtPrevLogIndex(t *testing.T) {
	test := quiz.Test(t)

	server := New()

	message := AppendEntriesMessage{
		Term:         2,
		LeaderId:     "leader_id",
		PrevLogIndex: 1,
		Entries:      []LogEntry{},
		PrevLogTerm:  2,
		CommitIndex:  0,
	}

	response := server.ReceiveAppendEntries(message)

	test.Expect(response.Success).ToBeFalse()
}

func TestAppendEntriesFailsWhenLogDoesNotContainEntryAtPrevLogIndexMatchingPrevLogTerm(t *testing.T) {
	test := quiz.Test(t)

	server := New()
	server.Log = []LogEntry{LogEntry{Term: 1, Data: "data"}}

	message := AppendEntriesMessage{
		Term:         2,
		LeaderId:     "leader_id",
		PrevLogIndex: 1,
		Entries:      []LogEntry{},
		PrevLogTerm:  2,
		CommitIndex:  0,
	}

	response := server.ReceiveAppendEntries(message)

	test.Expect(response.Success).ToBeFalse()
}

func TestSuccessfulAppendEntriesResetsElectionTimer(t *testing.T) {
	test := quiz.Test(t)

	server := New()
	timer := SpyTimer{make(chan int)}
	server.ElectionTimer = timer
	message := AppendEntriesMessage{
		Term:         1,
		LeaderId:     "leader_id",
		PrevLogIndex: 0,
		Entries:      []LogEntry{},
		PrevLogTerm:  2,
		CommitIndex:  0,
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

	server.ReceiveAppendEntries(message)
	shutDownChannel <- 1
}

func TestAppendEntriesSucceedsWhenHeartbeatingOnAnEmptyLog(t *testing.T) {
	test := quiz.Test(t)

	server := New()
	message := AppendEntriesMessage{
		Term:         1,
		LeaderId:     "leader_id",
		PrevLogIndex: 0,
		Entries:      []LogEntry{},
		PrevLogTerm:  2,
		CommitIndex:  0,
	}

	response := server.ReceiveAppendEntries(message)

	test.Expect(response.Success).ToBeTrue()
}

func TestTermUpdatesWhenReceivingHigherTermInAppendEntries(t *testing.T) {
	test := quiz.Test(t)

	server := New()

	message := AppendEntriesMessage{
		Term:         2,
		LeaderId:     "leader_id",
		PrevLogIndex: 2,
		Entries:      []LogEntry{},
		CommitIndex:  0,
	}

	server.ReceiveAppendEntries(message)

	test.Expect(server.Term).ToEqual(2)
}

func TestCandidateStepsDownWhenReceivingAppendEntriesMessage(t *testing.T) {
	test := quiz.Test(t)

	server := New()
	server.State = Candidate

	message := AppendEntriesMessage{
		Term:         2,
		LeaderId:     "leader_id",
		PrevLogIndex: 2,
		Entries:      []LogEntry{},
		CommitIndex:  0,
	}

	server.ReceiveAppendEntries(message)

	test.Expect(server.State).ToEqual(Follower)
}

func TestLeaderStepsDownWhenReceivingAppendEntriesMessage(t *testing.T) {
	test := quiz.Test(t)

	server := New()
	server.State = Leader

	message := AppendEntriesMessage{
		Term:         2,
		LeaderId:     "leader_id",
		PrevLogIndex: 2,
		Entries:      []LogEntry{},
		CommitIndex:  0,
	}

	server.ReceiveAppendEntries(message)

	test.Expect(server.State).ToEqual(Follower)
}

func TestServerDeletesConflictingEntriesWhenReceivingAppendEntriesMessage(t *testing.T) {
	test := quiz.Test(t)

	server := New()
	server.Log = []LogEntry{LogEntry{Term: 1, Data: "bad"}}

	message := AppendEntriesMessage{
		Term:         1,
		LeaderId:     "leader_id",
		PrevLogIndex: 0,
		Entries:      []LogEntry{LogEntry{Term: 1, Data: "good"}},
		CommitIndex:  0,
	}

	server.ReceiveAppendEntries(message)

	entry := server.Log[0]
	test.Expect(entry.Term).ToEqual(1)
	test.Expect(entry.Data).ToEqual("good")
}

func TestRecieveVoteResponseEndsElectionForHigherTerm(t *testing.T) {
	test := quiz.Test(t)

	server := New()
	server.Term = 0
	server.State = Candidate

	server.RecieveVoteResponse(VoteResponseMessage{
		VoteGranted: false,
		Term:        2,
	})

	test.Expect(server.State).ToEqual(Follower)
	test.Expect(server.Term).ToEqual(2)
	test.Expect(server.VotesGranted).ToEqual(0)
}

func TestRecieveVoteResponseTalliesVoteGranted(t *testing.T) {
	test := quiz.Test(t)

	server := New()
	server.Term = 0
	server.State = Candidate

	server.RecieveVoteResponse(VoteResponseMessage{
		VoteGranted: true,
		Term:        0,
	})

	test.Expect(server.VotesGranted).ToEqual(1)
	test.Expect(server.State).ToEqual(Candidate)
	test.Expect(server.Term).ToEqual(0)
}

func TestServersHavePeers(t *testing.T) {
	test := quiz.Test(t)

	serverA := New()
	serverB := New()

	serverA.AddPeer(serverB)

	test.Expect(serverA.Peers[0]).ToEqual(serverB)
}

func TestServerCanWinElection(t *testing.T) {
	test := quiz.Test(t)

	serverA := New()
	serverB := New()
	serverC := New()
	serverA.AddPeer(serverB)
	serverA.AddPeer(serverC)

	serverA.StartElection()

	test.Expect(serverA.State).ToEqual(Leader)
	test.Expect(serverA.Term).ToEqual(1)
	test.Expect(serverA.VotesGranted).ToEqual(2)

	test.Expect(serverB.VotedFor).ToEqual(serverA.Id)
	test.Expect(serverC.VotedFor).ToEqual(serverA.Id)
}

func TestServerCanLoseElectionForPeerWithHigherTerm(t *testing.T) {
	test := quiz.Test(t)

	serverA := New()
	serverB := New()
	serverC := New()
	serverA.AddPeer(serverB)
	serverA.AddPeer(serverC)

	serverB.Term = 2

	serverA.StartElection()

	test.Expect(serverA.State).ToEqual(Follower)
	test.Expect(serverA.Term).ToEqual(2)
	test.Expect(serverA.VotesGranted).ToEqual(1)

	test.Expect(serverB.VotedFor).ToEqual("")
	test.Expect(serverC.VotedFor).ToEqual(serverA.Id)
}

func TestServerCanLoseElectionDueToOutOfDateLog(t *testing.T) {
	test := quiz.Test(t)

	serverA := New()
	serverB := New()
	serverC := New()
	serverA.AddPeer(serverB)
	serverA.AddPeer(serverC)

	serverB.Log = []LogEntry{LogEntry{Term: 1, Data: "some data"}}

	serverA.StartElection()

	test.Expect(serverA.State).ToEqual(Follower)
	test.Expect(serverA.Term).ToEqual(0)
	test.Expect(serverA.VotesGranted).ToEqual(1)

	test.Expect(serverB.VotedFor).ToEqual("")
	test.Expect(serverC.VotedFor).ToEqual(serverA.Id)
}
