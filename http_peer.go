package graft

import (
	"encoding/json"
	"errors"
	"github.com/benmills/telephone"
)

type HttpPeer struct {
	URL string
}

func NewHttpPeer(url string) HttpPeer {
	return HttpPeer{url}
}

func (peer HttpPeer) ReceiveAppendEntries(message AppendEntriesMessage) AppendEntriesResponseMessage {
	return AppendEntriesResponseMessage{}
}

func (peer HttpPeer) ReceiveRequestVote(message RequestVoteMessage) (VoteResponseMessage, error) {
	var voteResponseMessage VoteResponseMessage
	body, _ := json.Marshal(message)
	request := telephone.Request{
		Url:  peer.URL + "/request_vote",
		Body: string(body),
	}
	response := request.Post()
	if !response.Success {
		return voteResponseMessage, errors.New("Could Not Communicate With Remote Peer")
	}
	json.Unmarshal([]byte(response.ParsedBody), &voteResponseMessage)

	return voteResponseMessage, nil
}
