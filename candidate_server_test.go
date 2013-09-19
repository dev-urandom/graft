package graft

import (
	e "github.com/lionelbarrow/examples"
	"github.com/wjdix/tiktok"
	"testing"
)

func TestCandidateServer(t *testing.T) {
	e.Describe("RequestVote", t,
		e.It("builds a RequestVote message derived from log", func(expect e.Expectation) {
			server := New("id")
			server.Log = []LogEntry{LogEntry{Term: 1, Data: "test"}, LogEntry{Term: 1, Data: "foo"}}
			newRequestVote := server.RequestVote()

			expect(newRequestVote.Term).To.Equal(1)
			expect(newRequestVote.CandidateId).To.Equal(server.Id)
			expect(newRequestVote.LastLogIndex).To.Equal(2)
			expect(newRequestVote.LastLogTerm).To.Equal(1)
		}),
	)

	e.Describe("ReceiveVoteResponse", t,
		e.It("ends election if response has higher term", func(expect e.Expectation) {
			server := New("id")
			server.Term = 0
			server.State = Candidate

			server.ReceiveVoteResponse(VoteResponseMessage{
				VoteGranted: false,
				Term:        2,
			})

			expect(server.State).ToEqual(Follower)
			expect(server.Term).ToEqual(2)
			expect(server.VotesGranted).ToEqual(0)
		}),

		e.It("will tally votes granted", func(expect e.Expectation) {
			server := New("id")
			server.Term = 0
			server.State = Candidate

			server.ReceiveVoteResponse(VoteResponseMessage{
				VoteGranted: true,
				Term:        0,
			})

			expect(server.VotesGranted).ToEqual(1)
			expect(server.State).ToEqual(Candidate)
			expect(server.Term).ToEqual(0)
		}),

		e.It("will continue the election even if VoteGranted is false", func(expect e.Expectation) {
			server := New("id")
			server.Term = 0
			server.State = Candidate

			server.ReceiveVoteResponse(VoteResponseMessage{
				VoteGranted: false,
				Term:        0,
			})

			expect(server.State).ToEqual(Candidate)
			expect(server.Term).ToEqual(0)
			expect(server.VotesGranted).ToEqual(0)
		}),
	)

	e.Describe("StartElection", t,
		e.It("can win the election", func(expect e.Expectation) {
			serverA := New("id")
			serverB := New("id")
			serverC := New("id")
			serverA.AddPeers(serverB, serverC)

			serverA.StartElection()

			expect(serverA.State).ToEqual(Leader)
			expect(serverA.Term).ToEqual(1)
			expect(serverA.VotesGranted).ToEqual(2)

			expect(serverA.VotedFor).ToEqual(serverA.Id)
			expect(serverB.VotedFor).ToEqual(serverA.Id)
			expect(serverC.VotedFor).ToEqual(serverA.Id)
		}),

		e.It("will retry sending RequestVote to peers if errors occur", func(expect e.Expectation) {
			serverA := New("id")
			serverB := New("id")
			serverC := &FailingPeer{
				numberOfFails: 1,
				successfulResponse: VoteResponseMessage{
					Term:        0,
					VoteGranted: true,
				},
			}

			serverA.AddPeers(serverB, serverC)

			serverA.StartElection()

			expect(serverA.State).ToEqual(Leader)
			expect(serverA.Term).ToEqual(1)
			expect(serverA.VotesGranted).ToEqual(2)
		}),

		e.It("can start and win an election after election timeout", func(expect e.Expectation) {
			serverA := New("id")
			timer := NewElectionTimer(1, serverA)
			timer.tickerBuilder = FakeTicker
			defer tiktok.ClearTickers()
			serverA.ElectionTimer = timer
			serverB := New("id")
			serverC := New("id")
			serverA.AddPeers(serverB, serverC)

			serverA.Start()
			tiktok.Tick(1)

			timer.ShutDown()
			expect(serverA.State).ToEqual(Leader)
			expect(serverA.Term).ToEqual(1)
			expect(serverA.VotesGranted).ToEqual(2)

			expect(serverB.VotedFor).ToEqual(serverA.Id)
			expect(serverC.VotedFor).ToEqual(serverA.Id)

		}),
	)
}
