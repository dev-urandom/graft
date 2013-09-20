package graft

import (
	e "github.com/benmills/examples"
	"github.com/benmills/quiz"
	"github.com/wjdix/tiktok"
	"net/http/httptest"
	"testing"
	"time"
)

func NewPeerWithControlledTimeout(id string, timeoutLength time.Duration) (ChannelPeer, *ElectionTimer) {
	server := New(id)
	throwAway := make(chan string, 10)
	server.StateMachine = SpyStateMachine{throwAway}
	timer := NewElectionTimer(timeoutLength, server)
	timer.tickerBuilder = FakeTicker
	return NewChannelPeer(server), timer
}

func TestIntegration(t *testing.T) {
	e.Describe("log", t,
		e.It("will continue to pull from leaders log back in time until followers are up to date", func(expect e.Expectation) {
			leader, _ := NewPeerWithControlledTimeout("server1", 0)
			follower2, _ := NewPeerWithControlledTimeout("server2", 0)
			follower3, _ := NewPeerWithControlledTimeout("server3", 0)
			leader.server.Peers = []Peer{follower2, follower3}
			follower2.server.Peers = []Peer{leader, follower3}
			follower3.server.Peers = []Peer{leader, follower2}
			leader.Start()
			follower2.Start()
			follower3.Start()
			defer leader.ShutDown()
			defer follower2.ShutDown()
			defer follower3.ShutDown()
			leader.server.StartElection()

			follower3.Partition()

			leader.server.AppendEntries("A")
			leader.server.AppendEntries("B")

			expect(follower3.server.CommitIndex).To.Equal(0)

			follower3.HealPartition()
			follower2.Partition()

			t.Log(follower3.server.Log)

			expect(follower3.server.CommitIndex).To.Equal(0)

			leader.server.AppendEntries("C")

			time.Sleep(10 * time.Second)

			expect(leader.server.CommitIndex).To.Equal(3)
			expect(follower3.server.CommitIndex).To.Equal(2)
			expect(len(follower3.server.Log)).To.Equal(3)
			t.Log(follower3.server.Log)
		}),
	)
}

func TestA3NodeClusterElectsTheFirstNodeToCallForElection(t *testing.T) {
	test := quiz.Test(t)
	peer1, timer1 := NewPeerWithControlledTimeout("server1", 2)
	peer2, timer2 := NewPeerWithControlledTimeout("server2", 9)
	peer3, timer3 := NewPeerWithControlledTimeout("server3", 9)
	peer1.server.Peers = []Peer{peer2, peer3}
	peer2.server.Peers = []Peer{peer1, peer3}
	peer3.server.Peers = []Peer{peer1, peer2}
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

func TestStartElectionIsLiveWith2FailingNodes(t *testing.T) {
	test := quiz.Test(t)
	peer1, _ := NewPeerWithControlledTimeout("server1", 2)
	peer2, _ := NewPeerWithControlledTimeout("server2", 9)
	peer3, _ := NewPeerWithControlledTimeout("server3", 9)
	peer1.server.Peers = []Peer{peer2, peer3}
	peer2.server.Peers = []Peer{peer1, peer3}
	peer3.server.Peers = []Peer{peer1, peer2}
	peer1.Start()
	peer2.Start()
	peer2.Partition()
	peer3.Start()
	peer3.Partition()

	peer1.server.StartElection()

	peer1.ShutDown()
	peer2.ShutDown()
	peer3.ShutDown()

	test.Expect(peer1.server.State).ToEqual(Follower)
}

func TestA5NodeClusterCanElectLeaderIf2NodesPartitioned(t *testing.T) {
	test := quiz.Test(t)
	peer1, timer1 := NewPeerWithControlledTimeout("server1", 2)
	peer2, timer2 := NewPeerWithControlledTimeout("server2", 9)
	peer3, timer3 := NewPeerWithControlledTimeout("server3", 9)
	peer4, timer4 := NewPeerWithControlledTimeout("server4", 9)
	peer5, timer5 := NewPeerWithControlledTimeout("server5", 9)
	peer1.server.Peers = []Peer{peer2, peer3, peer4, peer5}
	peer1.Start()
	timer1.StartTimer()
	peer2.Start()
	timer2.StartTimer()
	peer3.Start()
	timer3.StartTimer()
	peer4.Start()
	peer4.Partition()
	timer4.StartTimer()
	peer5.Start()
	peer5.Partition()
	timer5.StartTimer()

	tiktok.Tick(3)

	timer1.ShutDown()
	timer2.ShutDown()
	timer3.ShutDown()
	timer4.ShutDown()
	timer5.ShutDown()
	peer1.ShutDown()
	peer2.ShutDown()
	peer3.ShutDown()
	peer4.ShutDown()
	peer5.ShutDown()
	tiktok.ClearTickers()

	test.Expect(peer1.server.State).ToEqual(Leader)
}

func TestA5NodeClusterWillEndAnElectionEarlyUnderAPartitionDueToHigherTerm(t *testing.T) {
	test := quiz.Test(t)
	peer1, timer1 := NewPeerWithControlledTimeout("server1", 2)
	peer2, timer2 := NewPeerWithControlledTimeout("server2", 9)
	peer3, timer3 := NewPeerWithControlledTimeout("server3", 9)
	peer4, timer4 := NewPeerWithControlledTimeout("server4", 9)
	peer5, timer5 := NewPeerWithControlledTimeout("server5", 9)
	peer1.server.Peers = []Peer{peer2, peer3, peer4, peer5}

	// peer3 lead before 4 and 5 were partitioned
	peer3.server.Term = 2
	peer4.server.Term = 2
	peer5.server.Term = 2

	peer1.Start()
	timer1.StartTimer()
	peer2.Start()
	timer2.StartTimer()
	peer3.Start()
	timer3.StartTimer()
	peer4.Start()
	peer4.Partition()
	timer4.StartTimer()
	peer5.Start()
	peer5.Partition()
	timer5.StartTimer()
	tiktok.Tick(3)

	timer1.ShutDown()
	timer2.ShutDown()
	timer3.ShutDown()
	timer4.ShutDown()
	timer5.ShutDown()
	peer1.ShutDown()
	peer2.ShutDown()
	peer3.ShutDown()
	peer4.ShutDown()
	peer5.ShutDown()
	tiktok.ClearTickers()

	test.Expect(peer1.server.State).ToEqual(Follower)
	test.Expect(peer1.server.Term).ToEqual(2)
}

func TestHttpElection(t *testing.T) {
	test := quiz.Test(t)

	serverA := New("A")
	serverB := New("B")
	serverC := New("C")

	listenerA := httptest.NewServer(HttpHandler(serverA))
	defer listenerA.Close()
	listenerB := httptest.NewServer(HttpHandler(serverB))
	defer listenerB.Close()
	listenerC := httptest.NewServer(HttpHandler(serverC))
	defer listenerC.Close()

	serverA.AddPeers(HttpPeer{listenerB.URL}, HttpPeer{listenerC.URL})
	serverB.AddPeers(HttpPeer{listenerA.URL}, HttpPeer{listenerC.URL})
	serverC.AddPeers(HttpPeer{listenerA.URL}, HttpPeer{listenerB.URL})

	serverA.StartElection()

	test.Expect(serverA.State).ToEqual(Leader)
	test.Expect(serverA.VotedFor).ToEqual("A")
	test.Expect(serverA.Term).ToEqual(1)

	test.Expect(serverB.VotedFor).ToEqual("A")
	test.Expect(serverB.Term).ToEqual(1)

	test.Expect(serverC.VotedFor).ToEqual("A")
	test.Expect(serverC.Term).ToEqual(1)
}

func TestCanCommitAcrossA3NodeCluster(t *testing.T) {
	test := quiz.Test(t)

	serverA := New("A")
	serverB := New("B")
	serverC := New("C")

	listenerA := httptest.NewServer(HttpHandler(serverA))
	defer listenerA.Close()
	listenerB := httptest.NewServer(HttpHandler(serverB))
	defer listenerB.Close()
	listenerC := httptest.NewServer(HttpHandler(serverC))
	defer listenerC.Close()

	serverA.AddPeers(HttpPeer{listenerB.URL}, HttpPeer{listenerC.URL})
	serverB.AddPeers(HttpPeer{listenerA.URL}, HttpPeer{listenerC.URL})
	serverC.AddPeers(HttpPeer{listenerA.URL}, HttpPeer{listenerB.URL})

	serverA.StartElection()

	serverA.AppendEntries("foo")

	test.Expect(serverA.CommitIndex).ToEqual(1)
}

func TestCannotCommitAcrossA3NodeClusterIfTwoNodesArePartitioned(t *testing.T) {
	test := quiz.Test(t)
	peer1, _ := NewPeerWithControlledTimeout("server1", 2)
	peer2, _ := NewPeerWithControlledTimeout("server2", 9)
	peer3, _ := NewPeerWithControlledTimeout("server3", 9)
	peer1.server.Peers = []Peer{peer2, peer3}
	peer2.server.Peers = []Peer{peer1, peer3}
	peer3.server.Peers = []Peer{peer1, peer2}
	peer1.Start()
	peer2.Start()
	peer3.Start()

	peer1.server.StartElection()

	peer2.Partition()
	peer3.Partition()

	peer1.server.AppendEntries("foo")
	peer1.ShutDown()
	peer2.ShutDown()
	peer3.ShutDown()

	test.Expect(peer1.server.CommitIndex).ToEqual(0)
}
