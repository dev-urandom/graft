package graft

type ChannelPeer struct {
	requestVoteChan           chan RequestVoteMessage
	server                    *Server
	voteResponseChan          chan VoteResponseMessage
	shutDownChan              chan int
	appendEntriesChan         chan AppendEntriesMessage
	appendEntriesResponseChan chan AppendEntriesResponseMessage
}

func NewChannelPeer(server *Server) ChannelPeer {
	return ChannelPeer{
		requestVoteChan:           make(chan RequestVoteMessage),
		server:                    server,
		voteResponseChan:          make(chan VoteResponseMessage),
		shutDownChan:              make(chan int),
		appendEntriesChan:         make(chan AppendEntriesMessage),
		appendEntriesResponseChan: make(chan AppendEntriesResponseMessage),
	}
}

func (peer ChannelPeer) ReceiveRequestVote(message RequestVoteMessage) (response VoteResponseMessage, err error) {
	peer.requestVoteChan <- message
	for {
		select {
		case response = <-peer.voteResponseChan:
			return
		}
	}
}

func (peer ChannelPeer) ReceiveAppendEntries(message AppendEntriesMessage) (response AppendEntriesResponseMessage) {
	peer.appendEntriesChan <- message
	for {
		select {
		case response = <-peer.appendEntriesResponseChan:
			return
		}
	}
}

func (peer ChannelPeer) Start() {
	peer.server.Start()
	go func() {
		for {
			select {
			case message := <-peer.requestVoteChan:
				response, _ := peer.server.ReceiveRequestVote(message)
				peer.voteResponseChan <- response
			case message := <-peer.appendEntriesChan:
				response := peer.server.ReceiveAppendEntries(message)
				peer.appendEntriesResponseChan <- response
			case <-peer.shutDownChan:
				return
			}
		}
	}()
}

func (peer ChannelPeer) ShutDown() {
	peer.shutDownChan <- 0
}
