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

func TestAppendEntriesCommitsToStateMachineBasedOnCommitIndex(t *testing.T) {
	test := quiz.Test(t)

	server := New()
	messageChan := make(chan string)
	stateMachine := SpyStateMachine{messageChan}
	shutdownChan := make(chan int)
	server.StateMachine = stateMachine
	message := AppendEntriesMessage{
		Term:         1,
		LeaderId:     "leader_id",
		PrevLogIndex: 0,
		Entries:      []LogEntry{LogEntry{Term: 1, Data: "foo"}},
		PrevLogTerm:  1,
		CommitIndex:  1,
	}

	go func(messageChan chan string, shutDownChan chan int) {
		messages := []string{}
		for {
			select {
			case message := <-messageChan:
				messages = append(messages, message)
			case <-shutdownChan:
				test.Expect(messages[0]).ToEqual("foo")
				return
			}
		}
	}(messageChan, shutdownChan)

	server.ReceiveAppendEntries(message)
	shutdownChan <- 0
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

func TestServersHavePeers(t *testing.T) {
	test := quiz.Test(t)

	serverA := New()
	serverB := New()

	serverA.AddPeer(serverB)

	test.Expect(serverA.Peers[0]).ToEqual(serverB)
}
