package graft

import (
	"encoding/json"
	"github.com/benmills/quiz"
	"io/ioutil"
	"testing"
)

func TestServerStartReadsFromPersistedState(t *testing.T) {
	test := quiz.Test(t)

	log := []LogEntry{LogEntry{1, "foo"}}
	logData, _ := json.Marshal(log)
	ioutil.WriteFile("tmp/graft-log.json", logData, 0644)

	state := PersistedServerState{"id", 1}
	data, _ := json.Marshal(state)
	ioutil.WriteFile("tmp/graft-state.json", data, 0644)

	configuration := ServerConfiguration{
		Id:                  "foo",
		Peers:               []string{"localhost:4000", "localhost:3000"},
		PersistenceLocation: "tmp",
	}

	server := NewFromConfiguration(configuration)

	server.Start()

	test.Expect(server.Term).ToEqual(1)
	test.Expect(server.VotedFor).ToEqual("id")
	test.Expect(len(server.Log)).ToEqual(1)
}

func TestLoadPersistedState(t *testing.T) {
	test := quiz.Test(t)

	state := PersistedServerState{"id", 1}
	data, _ := json.Marshal(state)
	ioutil.WriteFile("tmp/graft-state.json", data, 0644)

	configuration := ServerConfiguration{
		Id:                  "foo",
		Peers:               []string{"localhost:4000", "localhost:3000"},
		PersistenceLocation: "tmp",
	}

	server := NewFromConfiguration(configuration)

	server.LoadPersistedState()

	test.Expect(server.Term).ToEqual(1)
	test.Expect(server.VotedFor).ToEqual("id")
}
