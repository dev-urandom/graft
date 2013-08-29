package graft

import (
	"encoding/json"
	"github.com/benmills/quiz"
	"io/ioutil"
	"testing"
)

func TestPersistLogWritesLogToDisk(t *testing.T) {
	test := quiz.Test(t)

	server := ServerBase{
		Id:            "hello",
		Log:           []LogEntry{LogEntry{Term: 1, Data: "Foo"}, LogEntry{Term: 2, Data: "Bar"}},
		Term:          1,
		VotedFor:      "hello",
		VotesGranted:  0,
		State:         Candidate,
		Peers:         []Peer{},
		ElectionTimer: NullTimer{},
		CommitIndex:   0,
	}
	persister := Persister{server}

	persister.PersistLog("tmp/graft-test.log")

	file, err := ioutil.ReadFile("tmp/graft-test.log")
	if err != nil {
		t.Fail()
	}

	var log []LogEntry

	json.Unmarshal(file, &log)
	test.Expect(len(log)).ToEqual(2)
	test.Expect(log[0].Term).ToEqual(1)
	test.Expect(log[0].Data).ToEqual("Foo")
	test.Expect(log[1].Term).ToEqual(2)
	test.Expect(log[1].Data).ToEqual("Bar")

}

func TestPersistServerStateIncludesCurrentTermAndVotedFor(t *testing.T) {
	test := quiz.Test(t)

	server := ServerBase{
		Id:            "hello",
		Log:           []LogEntry{LogEntry{Term: 1, Data: "Foo"}, LogEntry{Term: 2, Data: "Bar"}},
		Term:          1,
		VotedFor:      "hello",
		VotesGranted:  0,
		State:         Candidate,
		Peers:         []Peer{},
		ElectionTimer: NullTimer{},
		CommitIndex:   0,
	}
	persister := Persister{server}

	persister.PersistState("tmp/graft-state.json")

	file, err := ioutil.ReadFile("tmp/graft-state.json")
	if err != nil {
		t.Fail()
	}

	var state PersistedServerState

	json.Unmarshal(file, &state)
	test.Expect(state.VotedFor).ToEqual("hello")
	test.Expect(state.CurrentTerm).ToEqual(1)
}
