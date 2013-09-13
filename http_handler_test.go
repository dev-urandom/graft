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
	server := New()
	handler := HttpHandler{server}

	listener := httptest.NewServer(handler.Handler())
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
	server := New()
	handler := HttpHandler{server}

	listener := httptest.NewServer(handler.Handler())
	defer listener.Close()

	response := telephone.Post(listener.URL+"/request_vote", "bad_message")

	test.Expect(response.StatusCode).ToEqual(422)
}
