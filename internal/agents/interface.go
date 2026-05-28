package agents

import "context"

// Agent runs a prompt through an AI agent CLI.
type Agent interface {
	Name() string
	Run(ctx context.Context, req AgentRequest) (*AgentResponse, error)
}

// AgentRequest is the input to an agent invocation.
type AgentRequest struct {
	Prompt      string
	WorkDir     string
	ArtifactPath string // caller-supplied path for raw stdout output
}

// AgentResponse captures the output of an agent invocation.
type AgentResponse struct {
	Stdout    string
	Stderr    string
	ExitCode  int
	Duration  float64 // seconds
	Truncated bool
}
