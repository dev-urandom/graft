package graft

type ServerBase struct {
	Id                   string
	Log                  []LogEntry
	Term                 int
	VotedFor             string
	VotesGranted         int
	State                string
	Peers                []Peer
	ElectionTimer        Timable
	StateMachine         Commiter
	CommitIndex          int
	electionFinishedChan chan int
	PersistenceLocation  string
}

func (server *ServerBase) LastLogIndex() int {
	return len(server.Log)
}

func (server *ServerBase) lastLogTerm() int {
	if len(server.Log) == 0 {
		return 0
	} else {
		return (server.Log[len(server.Log)-1]).Term
	}
}

func (server *ServerBase) AddPeers(peers ...Peer) {
	for _, peer := range peers {
		server.Peers = append(server.Peers, peer)
	}
}
