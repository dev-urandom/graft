package graft

import (
	/*"github.com/benmills/quiz"*/
	"testing"
)

func TestHttpChannelCanBeAddedToAServersListOfPeers(t *testing.T) {
	peer := NewHttpPeer("http://localhost:4040")
	server2 := New()
	server2.Peers = []Peer{peer}
}
