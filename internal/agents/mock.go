package agents

import "context"

// MockAgent returns canned stdout for use in tests.
type MockAgent struct {
	AgentName  string
	StdoutBody string
	Err        error
}

func (m *MockAgent) Name() string { return m.AgentName }

func (m *MockAgent) Run(_ context.Context, req AgentRequest) (*AgentResponse, error) {
	if m.Err != nil {
		return &AgentResponse{ExitCode: 1, Stderr: m.Err.Error()}, m.Err
	}
	return &AgentResponse{Stdout: m.StdoutBody, ExitCode: 0}, nil
}
