package agent

type Server struct {
	agent *Agent
}

func (s *Server) AddLog(request AddLogRequest, response *AddLogResponse) error {
	return s.agent.AddLog(request, response)
}

func (s *Server) RequestVote(request RequestVoteRequest, response *RequestVoteResponse) error {
	return s.agent.RequestVote(request, response)
}

func (s *Server) AppendEntries(request AppendEntriesRequest, response *AppendEntriesResponse) error {
	return s.agent.AppendEntries(request, response)
}
