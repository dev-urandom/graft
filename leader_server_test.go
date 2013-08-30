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
