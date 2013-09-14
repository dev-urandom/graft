package graft

import (
	"encoding/json"
	"github.com/benmills/quiz"
	"github.com/benmills/telephone"
	"net/http/httptest"
	"testing"
)

func TestHttpHandlerWillRespondToRequestVote(t *testing.T) {
	test := quiz.Test(t)
	server := New("id")

	listener := httptest.NewServer(HttpHandler(server))
	defer listener.Close()

	voteRequestMessage := RequestVoteMessage{
		Term:         1,
		CandidateId:  "Foo",
		LastLogIndex: 0,
		LastLogTerm:  0,
	}
	body, _ := json.Marshal(voteRequestMessage)

	response := telephone.Post(listener.URL+"/request_vote", string(body))

	test.Expect(response.StatusCode).ToEqual(200)
	test.Expect(server.Term).ToEqual(1)
	test.Expect(server.VotedFor).ToEqual("Foo")

	var responseMessage VoteResponseMessage
	json.Unmarshal([]byte(response.ParsedBody), &responseMessage)

	test.Expect(responseMessage.Term).ToEqual(1)
	test.Expect(responseMessage.VoteGranted).ToBeTrue()
}

func TestHttpHandlerWillRespondWith422OnBadMessage(t *testing.T) {
	test := quiz.Test(t)
	server := New("id")

	listener := httptest.NewServer(HttpHandler(server))
	defer listener.Close()

	response := telephone.Post(listener.URL+"/request_vote", "bad_message")

	test.Expect(response.StatusCode).ToEqual(422)
}

func TestHttpHandlerWillRespondToAppendEntries(t *testing.T) {
	test := quiz.Test(t)
	server := New("id")

	listener := httptest.NewServer(HttpHandler(server))
	defer listener.Close()

	message := AppendEntriesMessage{
		Term:         1,
		LeaderId:     "leader_id",
		PrevLogIndex: 0,
		PrevLogTerm:  0,
		Entries:      []LogEntry{LogEntry{1, "foo"}},
		CommitIndex:  0,
	}
	body, _ := json.Marshal(message)

	response := telephone.Post(listener.URL+"/append_entries", string(body))

	test.Expect(response.StatusCode).ToEqual(200)
	test.Expect(server.Term).ToEqual(1)
	test.Expect(server.Log).ToContain("foo")

	var responseMessage AppendEntriesResponseMessage
	json.Unmarshal([]byte(response.ParsedBody), &responseMessage)

	test.Expect(responseMessage.Success).ToBeTrue()
}

func TestHttpHandlerWillRespondToAppendEntriesWithA422OnBadMessage(t *testing.T) {
	test := quiz.Test(t)
	server := New("id")

	listener := httptest.NewServer(HttpHandler(server))
	defer listener.Close()

	response := telephone.Post(listener.URL+"/append_entries", "bad_message")

	test.Expect(response.StatusCode).ToEqual(422)
}

func TestHttpHandlerPrefix(t *testing.T) {
	test := quiz.Test(t)
	server := New("id")

	listener := httptest.NewServer(PrefixedHttpHandler(server, "/prefix"))
	defer listener.Close()

	response := telephone.Post(listener.URL+"/prefix/append_entries", "")

	test.Expect(response.StatusCode).ToEqual(422)
}
