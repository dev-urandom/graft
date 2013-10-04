package graft

import (
	"github.com/benmills/quiz"
	"github.com/wjdix/tiktok"
	"testing"
)

func buildThrowAwayStateMachine() Commiter {
	throwAway := make(chan string, 10)
	return SpyStateMachine{throwAway}
}

func TestALeaderCanOverwriteItsLogToPartitionedServerAfterHeal(t *testing.T) {
	test := quiz.Test(t)
	c := newCluster(3).withChannelPeers().withStateMachine(buildThrowAwayStateMachine).withTimeouts(2, 9, 9)
	c.startChannelPeers()
	c.startElectionTimers()

	c.server(1).StartElection()
	c.partition(3)

	c.server(1).AppendEntries("A")
	c.server(1).AppendEntries("B")

	test.Expect(c.server(3).CommitIndex).To.Equal(0)

	c.healPartition(3)
	c.partition(2)

	test.Expect(c.server(3).CommitIndex).To.Equal(0)

	c.server(1).AppendEntries("C")

	test.Expect(c.server(1).CommitIndex).To.Equal(3)
	test.Expect(c.server(3).CommitIndex).To.Equal(2)
	test.Expect(len(c.server(3).Log)).To.Equal(3)
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
