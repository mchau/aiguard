package review

import "github.com/mchau/aiguard/internal/checks"

// FinalVerdict is the output of a verify run.
type FinalVerdict struct {
	Verdict             string                              `json:"verdict"`
	Summary             string                              `json:"summary"`
	DeterministicChecks []checks.DeterministicCheckResult   `json:"deterministic_checks"`
	// Fields populated in M11:
	ReviewerFindings    []ReviewerFinding                   `json:"reviewer_findings,omitempty"`
	Disagreements       []ReviewerDisagreement              `json:"disagreements,omitempty"`
	RequirementCoverage []RequirementCoverageResult         `json:"requirement_coverage,omitempty"`
	RecommendedFixPrompt string                             `json:"recommended_fix_prompt,omitempty"`
}

// Verdict constants (precedence per SPEC.md §16: BLOCKED > REJECT > WARN > APPROVE)
const (
	VerdictBlocked = "BLOCKED"
	VerdictReject  = "REJECT"
	VerdictWarn    = "WARN"
	VerdictApprove = "APPROVE"
)

// ReviewerFinding is populated in M11.
type ReviewerFinding struct {
	Reviewer  string `json:"reviewer"`
	Severity  string `json:"severity"`
	Message   string `json:"message"`
	RelatedAC string `json:"related_ac,omitempty"`
}

// ReviewerDisagreement is populated in M11.
type ReviewerDisagreement struct {
	Topic    string `json:"topic"`
	Findings []ReviewerFinding `json:"findings"`
}

// RequirementCoverageResult is populated in M11.
type RequirementCoverageResult struct {
	ACID     string `json:"ac_id"`
	Source   string `json:"source"`
	Status   string `json:"status"`
	Evidence string `json:"evidence,omitempty"`
	Notes    string `json:"notes,omitempty"`
}
