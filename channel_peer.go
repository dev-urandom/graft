package graft

import (
	"errors"
)

type ChannelPeer struct {
	requestVoteChan           chan RequestVoteMessage
	server                    *Server
	voteResponseChan          chan VoteResponseMessage
	shutDownChan              chan int
	appendEntriesChan         chan AppendEntriesMessage
	appendEntriesResponseChan chan AppendEntriesResponseMessage
	partitionChan             chan bool
	errorChan                 chan error
}

func NewChannelPeer(server *Server) ChannelPeer {
	return ChannelPeer{
		requestVoteChan:           make(chan RequestVoteMessage),
		server:                    server,
		voteResponseChan:          make(chan VoteResponseMessage),
		shutDownChan:              make(chan int),
		appendEntriesChan:         make(chan AppendEntriesMessage),
		appendEntriesResponseChan: make(chan AppendEntriesResponseMessage),
		partitionChan:             make(chan bool),
		errorChan:                 make(chan error),
	}
}

func (peer ChannelPeer) ReceiveRequestVote(message RequestVoteMessage) (response VoteResponseMessage, err error) {
	peer.requestVoteChan <- message
	for {
		select {
		case response = <-peer.voteResponseChan:
			err = nil
			return
		case err = <-peer.errorChan:
			return
		}
	}
}

func (peer ChannelPeer) ReceiveAppendEntries(message AppendEntriesMessage) (response AppendEntriesResponseMessage, err error) {
	peer.appendEntriesChan <- message
	for {
		select {
		case response = <-peer.appendEntriesResponseChan:
			return
		default:
			return
		}
	}
}

func (peer ChannelPeer) Start() {
	peer.server.Start()
	go func(partitioned bool) {
		for {
			select {
			case message := <-peer.requestVoteChan:
				if partitioned {
					peer.errorChan <- errors.New("Partitioned")
				} else {
					response, _ := peer.server.ReceiveRequestVote(message)
					peer.voteResponseChan <- response
				}
			case message := <-peer.appendEntriesChan:
				if partitioned != true {
					response, _ := peer.server.ReceiveAppendEntries(message)
					peer.appendEntriesResponseChan <- response
				} else {
					peer.errorChan <- errors.New("Partitioned")
				}
			case partitioned = <-peer.partitionChan:
			case <-peer.shutDownChan:
				return
			}
		}
	}(false)
}

func (peer ChannelPeer) Partition() {
	peer.partitionChan <- true
}

func (peer ChannelPeer) HealPartition() {
	peer.partitionChan <- false
}

func (peer ChannelPeer) ShutDown() {
	peer.shutDownChan <- 0
}
