package review

import (
	"github.com/mchau/aiguard/internal/checks"
)

// Aggregate applies the verdict precedence rules from SPEC.md §16:
//   BLOCKED > REJECT > WARN > APPROVE
//
// Deterministic BLOCKED checks can never be overridden by AI reviewer output.
// This function covers the deterministic-only path used in M2; M11 extends it.
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
