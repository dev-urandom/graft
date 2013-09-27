package graft

import (
	"github.com/benmills/quiz"
	"testing"
)

func TestGenerateAppendEntriesMessage(t *testing.T) {
	test := quiz.Test(t)

	server := New("id")

	message := server.GenerateAppendEntries()

	test.Expect(message.Term).ToEqual(0)
	test.Expect(message.LeaderId).ToEqual(server.Id)
	test.Expect(message.PrevLogIndex).ToEqual(0)
	test.Expect(len(message.Entries)).ToEqual(0)
	test.Expect(message.CommitIndex).ToEqual(0)
}

func TestGenerateAppendEntriesMessageWithData(t *testing.T) {
	test := quiz.Test(t)

	server := New("id")

	message := server.GenerateAppendEntries("foo")

	test.Expect(len(message.Entries)).ToEqual(1)
	test.Expect(message.Entries[0].Term).ToEqual(0)
	test.Expect(message.Entries[0].Data).ToEqual("foo")
}

func TestGenerateAppendEntriesMessageWithMultipleEntries(t *testing.T) {
	test := quiz.Test(t)

	server := New("id")

	message := server.GenerateAppendEntries("foo", "bar")

	test.Expect(len(message.Entries)).ToEqual(2)
	test.Expect(message.Entries[0].Data).ToEqual("foo")
	test.Expect(message.Entries[1].Data).ToEqual("bar")
}

func TestGenerateAppendEntriesIncludesLastLogTerm(t *testing.T) {
	test := quiz.Test(t)

	server := New("id")
	server.Log = []LogEntry{LogEntry{Term: 1, Data: "baz"}}

	message := server.GenerateAppendEntries("foo", "bar")

	test.Expect(len(message.Entries)).ToEqual(2)
	test.Expect(message.Entries[0].Data).ToEqual("foo")
	test.Expect(message.Entries[1].Data).ToEqual("bar")
	test.Expect(message.PrevLogTerm).ToEqual(1)
}

func TestRollBackAppendEntriesMessage(t *testing.T) {
	test := quiz.Test(t)

	server := New("id")
	server.Log = []LogEntry{LogEntry{Term: 1, Data: "baz"}, LogEntry{Term: 2, Data: "bar"}}

	m := AppendEntriesMessage{
		Term:         2,
		LeaderId:     "id",
		PrevLogIndex: 1,
		PrevLogTerm:  1,
		Entries:      []LogEntry{LogEntry{Term: 2, Data: "bar"}},
		CommitIndex:  0,
	}

	m2 := server.rolledBackMessage(m)

	test.Expect(m2.PrevLogIndex).ToEqual(0)
	test.Expect(m2.PrevLogTerm).ToEqual(0)
}

func TestAppendEntriesSuccessfully(t *testing.T) {
	test := quiz.Test(t)

	leader := New("id")
	followerA := New("id")
	followerB := New("id")
	leader.AddPeers(followerA, followerB)

	leader.AppendEntries("foo")

	// The leader successfully commits resulting in the commit index to be
	// incremented by one.
	test.Expect(len(leader.Log)).ToEqual(1)
	test.Expect(leader.CommitIndex).ToEqual(1)

	// Because the LastCommitIndex was 0 in the AppendEntriesMessage the
	// followers don't have 1 as a last commit index yet.
	test.Expect(len(followerA.Log)).ToEqual(1)
	test.Expect(followerA.CommitIndex).ToEqual(0)

	test.Expect(len(followerB.Log)).ToEqual(1)
	test.Expect(followerB.CommitIndex).ToEqual(0)
}

func TestAppendEntriesWontCommitWithoutMajority(t *testing.T) {
	test := quiz.Test(t)

	leader := New("id")
	followerA := &FailingPeer{
		numberOfFails: -1,
		failureAppendEntriesResponse: AppendEntriesResponseMessage{
			Success: false,
		},
		successfulAppendEntriesResponse: AppendEntriesResponseMessage{
			Success: true,
		},
	}
	followerB := &FailingPeer{
		numberOfFails: -1,
		failureAppendEntriesResponse: AppendEntriesResponseMessage{
			Success: false,
		},
		successfulAppendEntriesResponse: AppendEntriesResponseMessage{
			Success: true,
		},
	}
	leader.AddPeers(followerA, followerB)

	leader.AppendEntries("foo")

	test.Expect(leader.CommitIndex).ToEqual(0)
	test.Expect(len(followerA.Log)).ToEqual(0)
	test.Expect(len(followerB.Log)).ToEqual(0)
}

func TestAppendEntriesCommitsEvenWithSomeFailures(t *testing.T) {
	test := quiz.Test(t)

	leader := New("id")
	followerA := New("id")
	followerB := New("id")
	followerC := &FailingPeer{
		numberOfFails: -1,
		failureAppendEntriesResponse: AppendEntriesResponseMessage{
			Success: false,
		},
		successfulAppendEntriesResponse: AppendEntriesResponseMessage{
			Success: true,
		},
	}
	leader.AddPeers(followerA, followerB, followerC)

	leader.AppendEntries("foo")

	test.Expect(leader.CommitIndex).ToEqual(1)
	test.Expect(len(followerA.Log)).ToEqual(1)
	test.Expect(len(followerB.Log)).ToEqual(1)
	test.Expect(len(followerC.Log)).ToEqual(0)
}
