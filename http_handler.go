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

func (handler HttpHandler) Handler() http.Handler {
	m := pat.New()

	m.Post("/request_vote", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body)
		var message RequestVoteMessage
		unmarshalError := json.Unmarshal(body, &message)
		if unmarshalError != nil {
			w.WriteHeader(422)
			return
		}

		response, _ := handler.server.ReceiveRequestVote(message)

		w.WriteHeader(200)
		json.NewEncoder(w).Encode(response)
	}))

	return m
}
