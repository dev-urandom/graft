package graft

import (
	e "github.com/benmills/examples"
	"testing"
)

func TestPeerBroadcast(t *testing.T) {
	e.When("voting", t,
		e.It("can send successful RequestVoteMessages", func(expect e.Expectation) {
			peer := &FailingPeer{numberOfFails: 0}
			votes := 0
			broadcast := PeerBroadcast(RequestVoteMessage{}, []Peer{peer})

			func() {
				for {
					select {
					case r := <-broadcast.Response:
						votes++
						r.Done()
					case <-broadcast.End:
						return
					}
				}
			}()

			expect(votes).To.Equal(1)
		}),

		e.It("can handle failures", func(expect e.Expectation) {
			peer := &FailingPeer{numberOfFails: -1}
			votes := 0
			broadcast := PeerBroadcast(RequestVoteMessage{}, []Peer{peer})

			func() {
				for {
					select {
					case r := <-broadcast.Response:
						votes++
						r.Done()
					case <-broadcast.End:
						return
					}
				}
			}()

			expect(votes).To.Equal(0)
		}),

		e.It("will retry on failure", func(expect e.Expectation) {
			goodPeer := &FailingPeer{numberOfFails: 0}
			badPeer := &FailingPeer{numberOfFails: -1}
			soSoPeer := &FailingPeer{numberOfFails: 4}

			votes := 0
			broadcast := PeerBroadcast(RequestVoteMessage{}, []Peer{goodPeer, badPeer, soSoPeer})
			broadcast.defaultRetryCount = 5

			func() {
				for {
					select {
					case r := <-broadcast.Response:
						votes++
						r.Done()
					case <-broadcast.End:
						return
					}
				}
			}()

			expect(votes).To.Equal(2)
		}),

		e.It("can resend messages", func(expect e.Expectation) {
			receivedMessages := 0
			recordingPeer := &FailingPeer{
				numberOfFails: 0,
				receiveRequestVoteCallback: func() {
					receivedMessages++
				},
			}

			broadcast := PeerBroadcast(RequestVoteMessage{}, []Peer{recordingPeer})

			func() {
				for {
					select {
					case r := <-broadcast.Response:
						receivedMessages++
						if receivedMessages < 2 {
							r.Resend()
						} else {
							r.Done()
						}
					case <-broadcast.End:
						return
					}
				}
			}()

			expect(receivedMessages).To.Equal(2)
		}),
	)
}
