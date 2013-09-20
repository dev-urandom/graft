package graft

import (
	e "github.com/benmills/examples"
	"github.com/benmills/quiz"
	"github.com/wjdix/tiktok"
	"testing"
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
	c := newCluster(3).withChannelPeers().withTimeouts(2, 9, 9)
	c.startChannelPeers()
	c.startElectionTimers()

	tiktok.Tick(3)

	c.shutdown()

	test.Expect(c.server(1).State).ToEqual(Leader)
}

func TestStartElectionIsLiveWith2FailingNodes(t *testing.T) {
	test := quiz.Test(t)
	c := newCluster(3).withChannelPeers().withTimeouts(2, 9, 9)
	c.startChannelPeers()
	c.startElectionTimers()
	c.partition(2, 3)

	c.server(1).StartElection()

	c.shutdown()

	test.Expect(c.server(1).State).To.Equal(Follower)
}

func TestA5NodeClusterCanElectLeaderIf2NodesPartitioned(t *testing.T) {
	test := quiz.Test(t)
	c := newCluster(5).withChannelPeers().withTimeouts(2, 9, 9, 9, 9)
	c.startChannelPeers()
	c.partition(4, 5)
	c.startElectionTimers()

	tiktok.Tick(3)

	c.shutdown()

	test.Expect(c.server(1).State).ToEqual(Leader)
}

func TestA5NodeClusterWillEndAnElectionEarlyUnderAPartitionDueToHigherTerm(t *testing.T) {
	test := quiz.Test(t)
	c := newCluster(5).withChannelPeers().withTimeouts(2, 9, 9, 9, 9)

	// server 3 lead before 4 and 5 were partitioned
	c.server(3).Term = 2
	c.server(4).Term = 2
	c.server(5).Term = 2

	c.startChannelPeers()
	c.partition(4, 5)
	c.startElectionTimers()

	tiktok.Tick(3)

	c.shutdown()

	test.Expect(c.server(1).State).ToEqual(Follower)
	test.Expect(c.server(1).Term).ToEqual(2)
}

func TestHttpElection(t *testing.T) {
	test := quiz.Test(t)
	c := newCluster(3).withHttpPeers()
	defer c.closeHttpServers()

	c.server(1).StartElection()

	test.Expect(c.server(1).State).ToEqual(Leader)
	test.Expect(c.server(1).VotedFor).ToEqual("server1")
	test.Expect(c.server(1).Term).ToEqual(1)

	test.Expect(c.server(2).VotedFor).ToEqual("server1")
	test.Expect(c.server(2).Term).ToEqual(1)

	test.Expect(c.server(3).VotedFor).ToEqual("server1")
	test.Expect(c.server(3).Term).ToEqual(1)
}

func TestCanCommitAcrossA3NodeHttpCluster(t *testing.T) {
	test := quiz.Test(t)
	c := newCluster(3).withHttpPeers()
	defer c.closeHttpServers()
	c.electLeader(1)

	c.server(1).AppendEntries("foo")

	test.Expect(c.server(1).CommitIndex).ToEqual(1)
}

func TestCannotCommitAcrossA3NodeClusterIfTwoNodesArePartitioned(t *testing.T) {
	test := quiz.Test(t)
	c := newCluster(3).withChannelPeers().withTimeouts(2, 9, 9)
	c.startChannelPeers()
	c.electLeader(1)

	c.partition(2, 3)

	c.server(1).AppendEntries("foo")

	c.shutdown()

	test.Expect(c.server(1).CommitIndex).ToEqual(0)
}
