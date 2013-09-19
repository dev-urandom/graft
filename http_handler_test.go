package graft

import (
	"encoding/json"
	"github.com/benmills/telephone"
	e "github.com/lionelbarrow/examples"
	"net/http/httptest"
	"testing"
)

func TestHttpHandler(t *testing.T) {
	e.Describe("POST /request_vote", t,
		e.It("can be asked to vote", func(expect e.Expectation) {
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

			expect(response.StatusCode).ToEqual(200)
			expect(server.Term).ToEqual(1)
			expect(server.VotedFor).ToEqual("Foo")

			var responseMessage VoteResponseMessage
			json.Unmarshal([]byte(response.ParsedBody), &responseMessage)

			expect(responseMessage.Term).ToEqual(1)
			expect(responseMessage.VoteGranted).ToBeTrue()
		}),

		e.It("will respond with a 422 if the message is malformed", func(expect e.Expectation) {
			server := New("id")

			listener := httptest.NewServer(HttpHandler(server))
			defer listener.Close()

			response := telephone.Post(listener.URL+"/request_vote", "bad_message")

			expect(response.StatusCode).ToEqual(422)
		}),
	)

	e.Describe("POST /append_entries", t,
		e.It("can be asked to append entries", func(expect e.Expectation) {
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

			expect(response.StatusCode).ToEqual(200)
			expect(server.Term).ToEqual(1)
			expect(server.Log).ToContain("foo")

			var responseMessage AppendEntriesResponseMessage
			json.Unmarshal([]byte(response.ParsedBody), &responseMessage)

			expect(responseMessage.Success).ToBeTrue()
		}),

		e.It("will respond with a 422 if the message is malformed", func(expect e.Expectation) {
			server := New("id")

			listener := httptest.NewServer(HttpHandler(server))
			defer listener.Close()

			response := telephone.Post(listener.URL+"/append_entries", "bad_message")

			expect(response.StatusCode).ToEqual(422)
		}),

		e.It("can have a prefix", func(expect e.Expectation) {
			server := New("id")

			listener := httptest.NewServer(PrefixedHttpHandler(server, "/prefix"))
			defer listener.Close()

			response := telephone.Post(listener.URL+"/prefix/append_entries", "")

			expect(response.StatusCode).ToEqual(422)
		}),
	)
}
