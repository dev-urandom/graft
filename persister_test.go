package graft

import (
	"encoding/json"
	"github.com/benmills/quiz"
	"io/ioutil"
	"os"
	"testing"
)

func cleanTmpDir() {
	os.Remove("tmp/graft-state.json")
	os.Remove("tmp/graft-test.log")
	os.Remove("tmp/graft-log.json")
	os.Mkdir("tmp", 0755)
}

func TestPersistReturnsErrorIfWriteFails(t *testing.T) {
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

	err := persister.PersistLog("/no-way-this-file-exists.log")
	test.Expect(err != nil).ToBeTrue()
}

func TestPersistLogWritesLogToDisk(t *testing.T) {
	cleanTmpDir()
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
	cleanTmpDir()
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

func TestPersistServerReturnsErrorIfFails(t *testing.T) {
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

	err := persister.PersistState("/this-file-cant-possibly-exist")
	test.Expect(err != nil).ToBeTrue()
}

func TestPersistStateForConfiguredServer(t *testing.T) {
	cleanTmpDir()
	test := quiz.Test(t)

	config := ServerConfiguration{
		Id:                  "foo",
		Peers:               []string{},
		PersistenceLocation: "tmp",
	}

	server := NewFromConfiguration(config)
	server.VotedFor = "hello"
	server.Term = 1
	server.Log = []LogEntry{LogEntry{Term: 1, Data: "Foo"}, LogEntry{Term: 2, Data: "Bar"}}

	server.Persist()

	file, err := ioutil.ReadFile("tmp/graft-state.json")
	if err != nil {
		t.Fail()
	}

	var state PersistedServerState

	json.Unmarshal(file, &state)
	test.Expect(state.VotedFor).ToEqual("hello")
	test.Expect(state.CurrentTerm).ToEqual(1)

	file, err = ioutil.ReadFile("tmp/graft-log.json")
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
