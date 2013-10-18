// Cluster is a testing tool that reduces the repetitiveness of setting
// up an integration tests with a number of nodes.

package graft

import (
	"fmt"
	"github.com/wjdix/tiktok"
	"log"
	"net/http/httptest"
	"os"
	"time"
)

type cluster struct {
	peers          []Peer
	servers        []*Server
	electionTimers []*ElectionTimer
	httpServers    []*httptest.Server
}

type CommiterBuilder func() Commiter

func newCluster(size int) *cluster {
	c := &cluster{}

	for len(c.servers) != size {
		server := New(fmt.Sprintf("server%v", len(c.servers)+1))
		c.servers = append(c.servers, server)
	}

	return c
}

func (c *cluster) withChannelPeers() *cluster {
	for _, s := range c.servers {
		c.peers = append(c.peers, NewChannelPeer(s))
	}

	c.addPeers()

	return c
}

func (c *cluster) withHttpPeers() *cluster {
	c.startHttpServers()

	for _, s := range c.httpServers {
		c.peers = append(c.peers, NewHttpPeer(s.URL))
	}

	c.addPeers()

	return c
}

func (c *cluster) withStateMachine(machineBuilder CommiterBuilder) *cluster {
	for _, s := range c.servers {
		s.StateMachine = machineBuilder()
	}

	return c
}

func (c *cluster) addPeers() {
	for si, s := range c.servers {
		for pi, p := range c.peers {
			if pi != si {
				s.AddPeers(p)
			}
		}
	}
}

func (c *cluster) partition(peerIds ...int) {
	for _, peerId := range peerIds {
		c.peer(peerId).(ChannelPeer).Partition()
	}
}

func (c *cluster) healPartition(peerIds ...int) {
	for _, peerId := range peerIds {
		c.peer(peerId).(ChannelPeer).HealPartition()
	}
}

func (c *cluster) server(id int) *Server {
	return c.servers[id-1]
}

func (c *cluster) peer(id int) Peer {
	return c.peers[id-1]
}

func (c *cluster) startChannelPeers() {
	for _, p := range c.peers {
		p.(ChannelPeer).Start()
	}
}

func (c *cluster) startHttpServers() {
	c.closeHttpServers()

	for _, s := range c.servers {
		c.httpServers = append(c.httpServers, httptest.NewServer(HttpHandler(s)))
	}
}

func (c *cluster) startElectionTimers() {
	for i, _ := range c.electionTimers {
		c.electionTimers[i].StartTimer()
	}
}

func (c *cluster) shutdown() {
	for _, t := range c.electionTimers {
		t.ShutDown()
	}

	for _, p := range c.peers {
		p.(ChannelPeer).ShutDown()
	}

	c.closeHttpServers()

	tiktok.ClearTickers()
}

func (c *cluster) closeHttpServers() {
	for _, s := range c.httpServers {
		s.Close()
	}

	c.httpServers = []*httptest.Server{}
}

func (c *cluster) electLeader(serverId int) {
	c.server(serverId).StartElection()
	if c.server(serverId).State != Leader {
		log.Panicf("server%v did not become leader", serverId)
	}
}

func (c *cluster) withTimeouts(timeouts ...time.Duration) *cluster {
	if len(timeouts) != len(c.servers) {
		log.Panicf("Cannot setup %v timers on a cluster with %v servers", len(timeouts), len(c.servers))
	}

	for i, timeoutLength := range timeouts {
		timer := NewElectionTimer(timeoutLength, c.servers[i])
		timer.tickerBuilder = func(d time.Duration) Tickable {
			return tiktok.NewTicker(d)
		}
		c.electionTimers = append(c.electionTimers, timer)
	}

	return c
}

func (c *cluster) cleanUp() {
	for _, s := range c.servers {
		os.Remove(s.statePath())
		os.Remove(s.logPath())

	}
}
