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

func TestGenerateRequestVote(t *testing.T) {
	test := quiz.Test(t)

	server := New()
	newRequestVote := server.RequestVote()

	test.Expect(newRequestVote.term).ToEqual(1)
	test.Expect(newRequestVote.candidateId).ToEqual(server.Id)
	test.Expect(newRequestVote.lastLogIndex).ToEqual(0)
	test.Expect(newRequestVote.lastLogTerm).ToEqual(0)
}
