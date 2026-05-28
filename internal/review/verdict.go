package review

import (
	"github.com/mchau/aiguard/internal/checks"
)

// Aggregate applies the verdict precedence rules from SPEC.md §16:
//
//	BLOCKED > REJECT > WARN > APPROVE
//
// Rules:
//  1. Any BLOCKED deterministic check → BLOCKED (never overridden by AI reviewers)
//  2. Any FAIL/Critical/High from deterministic checks → at least REJECT
//  3. Source-provided AC with no coverage → REJECT (enforced by caller)
//  4. Any WARN / Medium / disagreement → at least WARN
//  5. All must-have ACs covered, no high/critical issues → APPROVE
//
// Deterministic BLOCKED checks can never be overridden by AI reviewer output.
func Aggregate(results []checks.DeterministicCheckResult) string {
	verdict := VerdictApprove
	for _, r := range results {
		switch r.Severity {
		case checks.SeverityBlocked:
			return VerdictBlocked // short-circuit: nothing overrides BLOCKED
		case checks.SeverityHigh, checks.SeverityCritical:
			if verdict != VerdictBlocked {
				verdict = VerdictReject
			}
		case checks.SeverityWarn, checks.SeverityMedium:
			if verdict == VerdictApprove {
				verdict = VerdictWarn
			}
		}
	}
	return verdict
}

// AggregateWithReviewers extends Aggregate with AI reviewer findings per SPEC.md §16.
// Deterministic BLOCKED always wins regardless of reviewer output.
func AggregateWithReviewers(detResults []checks.DeterministicCheckResult, reviewerFindings []ReviewerFinding, disagreements []ReviewerDisagreement) string {
	base := Aggregate(detResults)
	if base == VerdictBlocked {
		return VerdictBlocked // deterministic BLOCKED is final
	}

	// Reviewer Critical/High → REJECT (unless already BLOCKED)
	for _, f := range reviewerFindings {
		sev := f.Severity
		if sev == checks.SeverityCritical || sev == "Critical" || sev == checks.SeverityHigh || sev == "High" {
			if base == VerdictApprove || base == VerdictWarn {
				base = VerdictReject
			}
		} else if sev == checks.SeverityWarn || sev == checks.SeverityMedium || sev == "Medium" || sev == "WARN" {
			if base == VerdictApprove {
				base = VerdictWarn
			}
		}
	}

	// Disagreements → at least WARN
	if len(disagreements) > 0 && base == VerdictApprove {
		base = VerdictWarn
	}

	return base
}
