package graft

import "fmt"
import "sync"

type LeaderServer struct {
	Voter
}

const (
	appendEntriesMaxRetries = 0
)

func (server *Server) AppendEntries(data ...string) {
	message := server.GenerateAppendEntries(data...)
	successfulAppends := 1
	appendResponseChan := make(chan AppendEntriesResponseMessage, 1)
	finishedChan := make(chan bool)
	failedPeerChan := make(chan int, 1)
	go func(peercount int) {
		received := 0
		for received < peercount {
			fmt.Println("still running")
			select {
			case <-failedPeerChan:
				received++
			case response := <-appendResponseChan:
				received++
				if response.Success {
					successfulAppends++
				}
				if successfulAppends > (peercount / 2) {
					finishedChan <- true
					return
				}
			}
		}
		finishedChan <- false
		return
	}(len(server.Peers))
	waitGroup := sync.WaitGroup{}
	for i, peer := range server.Peers {
		go func(maxFailures int, target Peer, peerNum int) {
			waitGroup.Add(1)
			failureCount := 0
			for failureCount < maxFailures {
				response, err := target.ReceiveAppendEntries(message)
				if err != nil {
					fmt.Println("this guy failed: ", peerNum+1)
					failureCount++
					fmt.Println("failurecount: ", failureCount)
				} else {
					fmt.Println("this guy succeeded: ", peerNum+1)
					appendResponseChan <- response
					waitGroup.Done()
					return
				}
			}
			fmt.Println("sending to failed peer")
			failedPeerChan <- 0
			fmt.Println("failed peer")
			waitGroup.Done()
			return
		}(1, peer, i)
	}
	select {
	case success := <-finishedChan:
		if success {
			server.updateLog(message.PrevLogIndex, message.Entries)
			server.CommitIndex++
		}
	}
	fmt.Println("Waiting")
	waitGroup.Wait()
}

func (server *Server) GenerateAppendEntries(data ...string) AppendEntriesMessage {
	entries := []LogEntry{}
	for _, d := range data {
		entries = append(entries, LogEntry{Term: server.Term, Data: d})
	}

	return AppendEntriesMessage{
		Term:         server.Term,
		LeaderId:     server.Id,
		PrevLogIndex: server.LastLogIndex(),
		PrevLogTerm:  server.prevLogTerm(),
		Entries:      entries,
		CommitIndex:  server.LastCommitIndex(),
	}
}

func (server *Server) prevLogTerm() int {
	if server.LastLogIndex() > len(server.Log) || server.LastLogIndex() == 0 {
		return 0
	} else {
		return server.Log[server.LastLogIndex()-1].Term
	}
}
