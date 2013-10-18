package graft

import (
	"encoding/json"
	e "github.com/benmills/examples"
	"github.com/benmills/quiz"
	"io/ioutil"
	"testing"
)

func TestFollowerServer(t *testing.T) {
	e.Describe("ReceiveAppendEntries", t,
		e.It("fails when receives a lower term", func(expect e.Expectation) {
			server := New("id")
			server.Term = 3

			message := AppendEntriesMessage{
				Term:         2,
				LeaderId:     "leader_id",
				PrevLogIndex: 2,
				Entries:      []LogEntry{},
				CommitIndex:  0,
			}

			response, _ := server.ReceiveAppendEntries(message)

			expect(response.Success).ToBeFalse()
		}),

		e.It("fails when its log is out of date", func(expect e.Expectation) {
			server := New("id")

			message := AppendEntriesMessage{
				Term:         2,
				LeaderId:     "leader_id",
				PrevLogIndex: 1,
				Entries:      []LogEntry{},
				PrevLogTerm:  2,
				CommitIndex:  0,
			}

			response, _ := server.ReceiveAppendEntries(message)

			expect(response.Success).ToBeFalse()
		}),

		e.It("fails when the prev log entries term does not match PrevLogTerm", func(expect e.Expectation) {
			server := New("id")
			server.Log = []LogEntry{LogEntry{Term: 1, Data: "data"}}

			message := AppendEntriesMessage{
				Term:         2,
				LeaderId:     "leader_id",
				PrevLogIndex: 1,
				Entries:      []LogEntry{},
				PrevLogTerm:  2,
				CommitIndex:  0,
			}

			response, _ := server.ReceiveAppendEntries(message)

			expect(response.Success).ToBeFalse()
		}),

		e.It("resets ElectionTimer on successful AppendEntries", func(expect e.Expectation) {
			server := New("id")
			timer := SpyTimer{make(chan int)}
			server.ElectionTimer = timer
			message := AppendEntriesMessage{
				Term:         1,
				LeaderId:     "leader_id",
				PrevLogIndex: 0,
				Entries:      []LogEntry{},
				PrevLogTerm:  2,
				CommitIndex:  0,
			}

			shutDownChannel := make(chan int)

			go func(shutDownChannel chan int) {
				var resets int
				for {
					select {
					case <-timer.resetChannel:
						resets++
					case <-shutDownChannel:
						expect(resets).ToEqual(1)
						return
					}
				}
			}(shutDownChannel)

			server.ReceiveAppendEntries(message)
			shutDownChannel <- 1
		}),

		e.It("can accept a heartbeat", func(expect e.Expectation) {
			server := New("id")
			message := AppendEntriesMessage{
				Term:         1,
				LeaderId:     "leader_id",
				PrevLogIndex: 0,
				Entries:      []LogEntry{},
				PrevLogTerm:  2,
				CommitIndex:  0,
			}

			response, _ := server.ReceiveAppendEntries(message)

			expect(response.Success).ToBeTrue()
		}),

		e.It("will commit based on CommitIndex", func(expect e.Expectation) {
			server := New("id")
			messageChan := make(chan string)
			stateMachine := SpyStateMachine{messageChan}
			shutdownChan := make(chan int)
			server.StateMachine = stateMachine
			message := AppendEntriesMessage{
				Term:         1,
				LeaderId:     "leader_id",
				PrevLogIndex: 0,
				Entries:      []LogEntry{LogEntry{Term: 1, Data: "foo"}, LogEntry{Term: 1, Data: "Bar"}},
				PrevLogTerm:  1,
				CommitIndex:  2,
			}

			go func(messageChan chan string, shutDownChan chan int) {
				messages := []string{}
				for {
					select {
					case message := <-messageChan:
						messages = append(messages, message)
					case <-shutdownChan:
						expect(messages[0]).ToEqual("foo")
						expect(messages[1]).ToEqual("Bar")
						return
					}
				}
			}(messageChan, shutdownChan)

			server.ReceiveAppendEntries(message)
			shutdownChan <- 0
		}),

		e.It("updates the current term when receiving a higher one", func(expect e.Expectation) {
			server := New("id")

			message := AppendEntriesMessage{
				Term:         2,
				LeaderId:     "leader_id",
				PrevLogIndex: 2,
				Entries:      []LogEntry{},
				CommitIndex:  0,
			}

			server.ReceiveAppendEntries(message)

			expect(server.Term).ToEqual(2)
		}),

		e.It("steps down from being a candidate upon receiving a higher term", func(expect e.Expectation) {
			server := New("id")
			server.State = Candidate

			message := AppendEntriesMessage{
				Term:         2,
				LeaderId:     "leader_id",
				PrevLogIndex: 2,
				Entries:      []LogEntry{},
				CommitIndex:  0,
			}

			server.ReceiveAppendEntries(message)

			expect(server.State).ToEqual(Follower)
		}),

		e.It("steps down from being a leader upon receiving a higher term", func(expect e.Expectation) {
			server := New("id")
			server.State = Leader

			message := AppendEntriesMessage{
				Term:         2,
				LeaderId:     "leader_id",
				PrevLogIndex: 2,
				Entries:      []LogEntry{},
				CommitIndex:  0,
			}

			server.ReceiveAppendEntries(message)

			expect(server.State).ToEqual(Follower)
		}),

		e.It("deletes conflicting entries", func(expect e.Expectation) {
			server := New("id")
			server.Log = []LogEntry{LogEntry{Term: 1, Data: "bad"}}

			message := AppendEntriesMessage{
				Term:         1,
				LeaderId:     "leader_id",
				PrevLogIndex: 0,
				Entries:      []LogEntry{LogEntry{Term: 1, Data: "good"}},
				CommitIndex:  0,
			}

			server.ReceiveAppendEntries(message)

			entry := server.Log[0]
			expect(entry.Term).ToEqual(1)
			expect(entry.Data).ToEqual("good")
		}),
	)
}

func TestAServerPersistsBeforeRespondToAppendEntries(t *testing.T) {
	defer cleanTmpDir()
	test := quiz.Test(t)
	server := NewFromConfiguration(
		ServerConfiguration{
			Id:                  "id",
			Peers:               []string{},
			PersistenceLocation: "tmp",
		},
	)
	server.VotedFor = "other"
	message := AppendEntriesMessage{
		Term:         1,
		LeaderId:     "leader_id",
		PrevLogIndex: 0,
		Entries:      []LogEntry{LogEntry{Term: 1, Data: "good"}},
		CommitIndex:  0,
	}

	server.ReceiveAppendEntries(message)

	file, err := ioutil.ReadFile("tmp/graft-stateid.json")
	if err != nil {
		t.Fail()
	}

	var state PersistedServerState
	json.Unmarshal(file, &state)
	test.Expect(state.CurrentTerm).ToEqual(1)
	test.Expect(state.VotedFor).ToEqual("other")

	logFile, err := ioutil.ReadFile("tmp/graft-logid.json")
	if err != nil {
		t.Fail()
	}

	var log []LogEntry
	json.Unmarshal(logFile, &log)
	test.Expect(len(log)).ToEqual(1)
	test.Expect(log[0].Term).ToEqual(1)
	test.Expect(log[0].Data).ToEqual("good")
}
