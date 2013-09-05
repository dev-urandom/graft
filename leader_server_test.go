package graft

import (
	"github.com/benmills/quiz"
	"testing"
)

func TestGenerateAppendEntriesMessage(t *testing.T) {
	test := quiz.Test(t)

	server := New()

	message := server.GenerateAppendEntries()

	test.Expect(message.Term).ToEqual(0)
	test.Expect(message.LeaderId).ToEqual(server.Id)
	test.Expect(message.PrevLogIndex).ToEqual(0)
	test.Expect(len(message.Entries)).ToEqual(0)
	test.Expect(message.CommitIndex).ToEqual(0)
}

func TestGenerateAppendEntriesMessageWithData(t *testing.T) {
	test := quiz.Test(t)

	server := New()

	message := server.GenerateAppendEntries("foo")

	test.Expect(len(message.Entries)).ToEqual(1)
	test.Expect(message.Entries[0].Term).ToEqual(0)
	test.Expect(message.Entries[0].Data).ToEqual("foo")
}

func TestGenerateAppendEntriesMessageWithMultipleEntries(t *testing.T) {
	test := quiz.Test(t)

	server := New()

	message := server.GenerateAppendEntries("foo", "bar")

	test.Expect(len(message.Entries)).ToEqual(2)
	test.Expect(message.Entries[0].Data).ToEqual("foo")
	test.Expect(message.Entries[1].Data).ToEqual("bar")
}

func TestAppendEntriesSuccessfully(t *testing.T) {
	test := quiz.Test(t)

	leader := New()
	followerA := New()
	followerB := New()
	leader.AddPeers(followerA, followerB)

	leader.AppendEntries("foo")

	// The leader successfully commits resulting in the commit index to be
	// incremented by one.
	test.Expect(len(leader.Log)).ToEqual(1)
	test.Expect(leader.lastCommitIndex()).ToEqual(1)

	// Because the lastCommitIndex was 0 in the AppendEntriesMessage the
	// followers don't have 1 as a last commit index yet.
	test.Expect(len(followerA.Log)).ToEqual(1)
	test.Expect(followerA.lastCommitIndex()).ToEqual(0)

	test.Expect(len(followerB.Log)).ToEqual(1)
	test.Expect(followerB.lastCommitIndex()).ToEqual(0)
}

func TestAppendEntriesWontCommitWithoutMajority(t *testing.T) {
	test := quiz.Test(t)

	leader := New()
	followerA := New()
	followerB := &FailingPeer{
		numberOfFails: 1,
		failureAppendEntriesResponse: AppendEntriesResponseMessage{
			Success: false,
		},
		successfulAppendEntriesResponse: AppendEntriesResponseMessage{
			Success: true,
		},
	}
	leader.AddPeers(followerA, followerB)

	leader.AppendEntries("foo")

	test.Expect(leader.lastCommitIndex()).ToEqual(0)
	test.Expect(len(followerA.Log)).ToEqual(1)
}

func TestAppendEntriesCommitsEvenWithSomeFailures(t *testing.T) {
	test := quiz.Test(t)

	leader := New()
	followerA := New()
	followerB := New()
	followerC := &FailingPeer{
		numberOfFails: 1,
		failureAppendEntriesResponse: AppendEntriesResponseMessage{
			Success: false,
		},
		successfulAppendEntriesResponse: AppendEntriesResponseMessage{
			Success: true,
		},
	}
	leader.AddPeers(followerA, followerB, followerC)

	leader.AppendEntries("foo")

	test.Expect(leader.lastCommitIndex()).ToEqual(1)
	test.Expect(len(followerA.Log)).ToEqual(1)
	test.Expect(len(followerB.Log)).ToEqual(1)
}
