package graft

type ServerBase struct {
	Id            string
	Log           []LogEntry
	Term          int
	VotedFor      string
	VotesGranted  int
	State         string
	Peers         []Peer
	ElectionTimer Timable
	StateMachine  Commiter
	CommitIndex   int
}

func (server *ServerBase) lastLogIndex() int {
	return len(server.Log)
}

func (server *ServerBase) lastLogTerm() int {
	if len(server.Log) == 0 {
		return 0
	} else {
		return (server.Log[len(server.Log)-1]).Term
	}
}

func (server *ServerBase) AddPeer(peer Peer) {
	server.Peers = append(server.Peers, peer)
}
