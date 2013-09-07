package graft

import (
	"github.com/benmills/quiz"
	"github.com/wjdix/tiktok"
	"testing"
	"time"
)

func NewPeerWithControlledTimeout(timeoutLength time.Duration) (ChannelPeer, *ElectionTimer) {
	server := New()
	timer := NewElectionTimer(timeoutLength, server)
	timer.tickerBuilder = FakeTicker
	return NewChannelPeer(server), timer
}

func TestA3NodeClusterElectsTheFirstNodeToCallForElection(t *testing.T) {
	test := quiz.Test(t)
	peer1, timer1 := NewPeerWithControlledTimeout(2)
	peer2, timer2 := NewPeerWithControlledTimeout(9)
	peer3, timer3 := NewPeerWithControlledTimeout(9)
	peer1.server.Peers = []Peer{peer2, peer3}
	peer1.server.Id = "server1"
	peer2.server.Peers = []Peer{peer1, peer3}
	peer2.server.Id = "server2"
	peer3.server.Peers = []Peer{peer1, peer2}
	peer3.server.Id = "server3"
	peer1.Start()
	timer1.StartTimer()
	peer2.Start()
	timer2.StartTimer()
	peer3.Start()
	timer3.StartTimer()
	tiktok.Tick(3)

	timer1.ShutDown()
	timer2.ShutDown()
	timer3.ShutDown()
	peer1.ShutDown()
	peer2.ShutDown()
	peer3.ShutDown()
	tiktok.ClearTickers()

	test.Expect(peer1.server.State).ToEqual(Leader)
}
