package server

type Server struct {
	Log []string
	Term int
	VotedFor string
}

func New() *Server {
	return &Server{
		Log: []string{},
		Term: 0,
		VotedFor: "",
	}
}
