package graft

type RequestVoteMessage struct {
	term int
	candidateId string
	lastLogIndex int
	lastLogTerm int
}
