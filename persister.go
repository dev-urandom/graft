package graft

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
)

type Persister struct {
	ServerBase
}

type PersistedServerState struct {
	VotedFor    string
	CurrentTerm int
}

func (persister *Persister) PersistLog(filename string) error {
	data, _ := json.Marshal(persister.Log)
	err := ioutil.WriteFile(filename, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

func (persister *Persister) PersistState(filename string) error {
	data, _ := json.Marshal(persister.stateToPersist())
	err := ioutil.WriteFile(filename, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

func (persister *Persister) stateToPersist() PersistedServerState {
	return PersistedServerState{persister.VotedFor, persister.Term}
}

func (persister *Persister) Persist() error {
	err := persister.PersistState(persister.statePath())
	if err != nil {
		return err
	}
	err = persister.PersistLog(persister.logPath())
	if err != nil {
		return err
	}
	return nil
}

func (persister *Persister) LoadPersistedState() error {
	var state PersistedServerState
	file, err := ioutil.ReadFile(persister.statePath())
	if err != nil {
		return err
	}
	json.Unmarshal(file, &state)
	persister.Term = state.CurrentTerm
	persister.VotedFor = state.VotedFor
	return nil
}

func (persister *Persister) LoadPersistedLog() error {
	var log []LogEntry
	file, err := ioutil.ReadFile(persister.logPath())
	if err != nil {
		return err
	}
	json.Unmarshal(file, &log)
	persister.Log = log
	return nil
}

func (persister Persister) statePath() string {
	return filepath.Join(persister.PersistenceLocation, "graft-state.json")
}

func (persister Persister) logPath() string {
	return filepath.Join(persister.PersistenceLocation, "graft-log.json")
}
