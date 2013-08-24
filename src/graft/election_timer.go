package graft

import (
	"time"
)

type Electable interface {
	StartElection()
}

type ElectionTimer struct {
	electable       Electable
	ElectionChannel chan int
	shutDownChannel chan int
	resets          int
	duration        time.Duration
}

func NewElectionTimer(duration time.Duration, electable Electable) *ElectionTimer {
	timer := &ElectionTimer{
		electable:       electable,
		ElectionChannel: make(chan int),
		shutDownChannel: make(chan int),
		duration:        duration,
	}

	go timer.waitForElection()
	return timer
}

func (timer *ElectionTimer) StartTimer() {
	go func() {
		ticker := time.NewTicker(timer.duration)
		for {
			select {
			case <-ticker.C:
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
