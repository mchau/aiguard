package audit

import "time"

// Event is a single audit log entry per CLAUDE.md §"Step 4 — Audit Log".
type Event struct {
	Timestamp    time.Time `json:"timestamp"`
	Command      string    `json:"command"`
	GitBranch    string    `json:"git_branch,omitempty"`
	GitHead      string    `json:"git_head,omitempty"`
	SourceBranch string    `json:"source_branch,omitempty"`
	SourceCommit string    `json:"source_commit,omitempty"`
	Inputs       []string  `json:"inputs"`
	Outputs      []string  `json:"outputs"`
	Status       string    `json:"status"`
	Error        string    `json:"error,omitempty"`
}
