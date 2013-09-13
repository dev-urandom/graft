package graft

import (
	"encoding/json"
	"github.com/bmizerany/pat"
	"io/ioutil"
	"net/http"
)

type HttpHandler struct {
	server *Server
}

func extractMessage(r *http.Request, message interface{}) error {
	body, _ := ioutil.ReadAll(r.Body)
	return json.Unmarshal(body, message)
}

func (handler HttpHandler) Handler() http.Handler {
	m := pat.New()

	m.Post("/request_vote", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var message RequestVoteMessage
		err := extractMessage(r, &message)
		if err != nil {
			w.WriteHeader(422)
			return
		}

		response, _ := handler.server.ReceiveRequestVote(message)

		w.WriteHeader(200)
		json.NewEncoder(w).Encode(response)
	}))

	m.Post("/append_entries", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var message AppendEntriesMessage
		err := extractMessage(r, &message)
		if err != nil {
			w.WriteHeader(422)
			return
		}

		response, _ := handler.server.ReceiveAppendEntries(message)

		w.WriteHeader(200)
		json.NewEncoder(w).Encode(response)
	}))

	return m
}
