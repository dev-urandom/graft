package graft

type RequestVoteMessage struct {
	Term         int
	CandidateId  string
	LastLogIndex int
	LastLogTerm  int
}
