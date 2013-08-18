package graft

type AppendEntriesMessage struct {
	Term int
	LeaderId string
	PrevLogIndex int
	Entries []LogEntry
	CommitIndex int
}
