package graft

import (
	"time"
)

type Electable interface {
	StartElection()
}

type Tickable interface {
	Stop()
	Chan() <-chan time.Time
}

type ElectionTimer struct {
	electable       Electable
	ElectionChannel chan int
	shutDownChannel chan int
	resets          int
	duration        time.Duration
	tickerBuilder   func(time.Duration) Tickable
}

type WrappedTicker struct {
	ticker *time.Ticker
}

func (ticker *WrappedTicker) Chan() <-chan time.Time {
	return ticker.ticker.C
}

func (ticker *WrappedTicker) Stop() {
	ticker.ticker.Stop()
}

func DefaultTicker(d time.Duration) Tickable {
	return &WrappedTicker{
		ticker: time.NewTicker(d),
	}
}

func NewElectionTimer(duration time.Duration, electable Electable) *ElectionTimer {
	timer := &ElectionTimer{
		electable:       electable,
		ElectionChannel: make(chan int),
		shutDownChannel: make(chan int),
		duration:        duration,
		tickerBuilder:   DefaultTicker,
	}

	go timer.waitForElection()
	return timer
}

func (timer *ElectionTimer) StartTimer() {
	go func() {
		ticker := timer.tickerBuilder(timer.duration)
		for {
			select {
			case <-ticker.Chan():
				timer.ElectionChannel <- 1
			}
		}
	}()
}

func (timer *ElectionTimer) ShutDown() {
	timer.shutDownChannel <- 1
}

func (timer *ElectionTimer) startElection() {
	timer.electable.StartElection()
}

func (timer *ElectionTimer) waitForElection() {
	for {
		select {
		case <-timer.ElectionChannel:
			timer.startElection()
		case <-timer.shutDownChannel:
			return
		}
	}
}
