package graft

import (
	e "github.com/benmills/examples"
	"github.com/wjdix/tiktok"
	"testing"
	"time"
)

type SpyServer struct {
	electionStarted bool
	electionCount   int
}

func FakeTicker(d time.Duration) Tickable {
	return tiktok.NewTicker(d)
}

func (server *SpyServer) StartElection() {
	server.electionStarted = true
	server.electionCount += 1
}

func TestElectionTimer(t *testing.T) {
	e.Describe("NewElectionTimer", t,
		e.It("tells the server to start the election on <-ElectionChannel", func(expect e.Expectation) {
			spyServer := &SpyServer{electionStarted: false, electionCount: 0}
			timer := NewElectionTimer(1, spyServer)
			defer timer.ShutDown()

			timer.ElectionChannel <- 1

			expect(spyServer.electionStarted).ToBeTrue()
		}),
	)

	e.Describe("StartTimer", t,
		e.It("can start an election", func(expect e.Expectation) {
			spyServer := &SpyServer{electionStarted: false, electionCount: 0}
			timer := NewElectionTimer(1, spyServer)
			timer.tickerBuilder = FakeTicker
			defer tiktok.ClearTickers()

			timer.StartTimer()

			tiktok.Tick(1)

			timer.ShutDown()
			expect(spyServer.electionStarted).ToBeTrue()
		}),

		e.It("can start multiple elections", func(expect e.Expectation) {
			spyServer := &SpyServer{electionStarted: false, electionCount: 0}
			timer := NewElectionTimer(2, spyServer)
			timer.tickerBuilder = FakeTicker
			defer tiktok.ClearTickers()

			timer.StartTimer()

			tiktok.Tick(10)

			timer.ShutDown()

			expect(spyServer.electionCount).ToBeGreaterThan(1)
		}),

		e.It("does not start an election on reset", func(expect e.Expectation) {
			spyServer := &SpyServer{electionStarted: false, electionCount: 0}
			timer := NewElectionTimer(5, spyServer)
			timer.tickerBuilder = FakeTicker
			defer tiktok.ClearTickers()

			timer.StartTimer()

			tiktok.Tick(3)
			timer.Reset()

			tiktok.Tick(2)
			timer.ShutDown()
			expect(spyServer.electionCount).ToEqual(0)
		}),
	)
}
