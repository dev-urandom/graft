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
	statePath := filepath.Join(persister.PersistenceLocation, "graft-state.json")
	logPath := filepath.Join(persister.PersistenceLocation, "graft-log.json")
	err := persister.PersistState(statePath)
	if err != nil {
		return err
	}
	err = persister.PersistLog(logPath)
	if err != nil {
		return err
	}
	return nil
}
