package graft

type AppendEntriesMessage struct {
	Term int
	LeaderId string
	PrevLogIndex int
	Entries []string
	CommitIndex int
}
