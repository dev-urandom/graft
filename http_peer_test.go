package graft

import (
	"encoding/json"
	"github.com/benmills/quiz"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func fakeRemoteServer() *httptest.Server {
	server := http.NewServeMux()
	server.HandleFunc("/request_vote", func(w http.ResponseWriter, r *http.Request) {
		var requestVoteMessage RequestVoteMessage
		rawBody, _ := ioutil.ReadAll(r.Body)
		json.Unmarshal(rawBody, &requestVoteMessage)
		response := VoteResponseMessage{
			Term:        requestVoteMessage.Term,
			VoteGranted: true,
		}
		enc := json.NewEncoder(w)
		enc.Encode(response)
	})
	server.HandleFunc("/append_entries", func(w http.ResponseWriter, r *http.Request) {
		var appendEntriesMessage AppendEntriesMessage
		rawBody, _ := ioutil.ReadAll(r.Body)
		json.Unmarshal(rawBody, &appendEntriesMessage)
		response := AppendEntriesResponseMessage{true}

		enc := json.NewEncoder(w)
		enc.Encode(response)
	})
	return httptest.NewServer(server)

}

func TestHttpPeerCanBeAddedToAServersListOfPeers(t *testing.T) {
	peer := NewHttpPeer("http://localhost:4040")
	server2 := New()
	server2.Peers = []Peer{peer}
}

func TestHttpPeerRequestsVoteOverHttp(t *testing.T) {
	test := quiz.Test(t)
	remoteServer := fakeRemoteServer()
	defer remoteServer.Close()

	peer := NewHttpPeer(remoteServer.URL)

	voteRequest := RequestVoteMessage{
		Term:         1,
		CandidateId:  "foo",
		LastLogIndex: 0,
		LastLogTerm:  0,
	}

	response, err := peer.ReceiveRequestVote(voteRequest)
	test.Expect(err == nil).ToBeTrue()
	test.Expect(response.Term).ToEqual(1)
	test.Expect(response.VoteGranted).ToBeTrue()
}

func TestHttpPeerAppendsEntriesOverHttp(t *testing.T) {
	test := quiz.Test(t)
	remoteServer := fakeRemoteServer()
	defer remoteServer.Close()

	peer := NewHttpPeer(remoteServer.URL)

	appendEntriesMessage := AppendEntriesMessage{
		Term:         1,
		LeaderId:     "foo",
		PrevLogIndex: 0,
		PrevLogTerm:  0,
		Entries:      []LogEntry{},
		CommitIndex:  0,
	}

	response := peer.ReceiveAppendEntries(appendEntriesMessage)

	test.Expect(response.Success).ToBeTrue()
}
