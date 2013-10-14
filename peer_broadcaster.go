package graft

import (
	"sync"
)

type broadcastResponse struct {
	VoteRes VoteResponseMessage
	p       Peer
	b       *broadcast
}

func (r *broadcastResponse) Done() {
	r.b.Done()
}

func (r *broadcastResponse) Resend() {
	go r.b.send(r.p)
}

type broadcast struct {
	*sync.WaitGroup
	Response          chan *broadcastResponse
	End               chan int
	m                 interface{}
	peers             []Peer
	defaultRetryCount int
}

func PeerBroadcast(m interface{}, peers []Peer) *broadcast {
	b := &broadcast{
		WaitGroup:         &sync.WaitGroup{},
		Response:          make(chan *broadcastResponse),
		End:               make(chan int),
		m:                 m,
		peers:             peers,
		defaultRetryCount: 5,
	}
	go b.sendToPeers()
	return b
}

func (b *broadcast) sendToPeers() {
	b.Add(len(b.peers))

	for _, p := range b.peers {
		go b.send(p)
	}

	b.Wait()
	b.End <- 1
}

func (b *broadcast) send(p Peer) {
	retries := b.defaultRetryCount

	for retries != 0 {
		switch m := b.m.(type) {
		case RequestVoteMessage:
			res, err := p.ReceiveRequestVote(m)
			if err == nil {
				b.Response <- &broadcastResponse{p: p, b: b, VoteRes: res}
				return
			} else {
				retries--
			}
		}
	}

	b.Done()
}
