package checks

import "context"

// Severity levels for check results.
const (
	SeverityBlocked  = "BLOCKED"
	SeverityCritical = "Critical"
	SeverityHigh     = "High"
	SeverityMedium   = "Medium"
	SeverityWarn     = "WARN"
	SeverityInfo     = "Info"
)

// DeterministicCheckResult is a single finding from a deterministic check.
type DeterministicCheckResult struct {
	CheckID     string `json:"check_id"`
	Severity    string `json:"severity"`
	File        string `json:"file,omitempty"`
	Pattern     string `json:"pattern,omitempty"`
	Message     string `json:"message"`
	Remediation string `json:"remediation,omitempty"`
}

// CheckInput holds all data a check may need to inspect.
type CheckInput struct {
	Config       interface{}    // *config.Config — interface to avoid import cycle
	ChangedFiles []string
	DiffText     string
	FileContent  func(path string) ([]byte, error) // nil-safe accessor
	// Optional — nil until M8 gates are in place
	PlanPath      string
	RationalePath string
}

// Check is implemented by each deterministic check.
type Check interface {
	ID() string
	Run(ctx context.Context, input CheckInput) ([]DeterministicCheckResult, error)
}
