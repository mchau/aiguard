package review_test

import (
	"testing"

	"github.com/mchau/aiguard/internal/checks"
	"github.com/mchau/aiguard/internal/review"
)

func TestVerdictApprove(t *testing.T) {
	if got := review.Aggregate(nil); got != review.VerdictApprove {
		t.Errorf("empty results: got %q, want APPROVE", got)
	}
}

func TestVerdictWarn(t *testing.T) {
	results := []checks.DeterministicCheckResult{
		{CheckID: "x", Severity: checks.SeverityWarn},
	}
	if got := review.Aggregate(results); got != review.VerdictWarn {
		t.Errorf("warn result: got %q, want WARN", got)
	}
}

func TestVerdictReject(t *testing.T) {
	results := []checks.DeterministicCheckResult{
		{CheckID: "x", Severity: checks.SeverityHigh},
	}
	if got := review.Aggregate(results); got != review.VerdictReject {
		t.Errorf("high result: got %q, want REJECT", got)
	}
}

func TestVerdictBlocked(t *testing.T) {
	results := []checks.DeterministicCheckResult{
		{CheckID: "x", Severity: checks.SeverityBlocked},
	}
	if got := review.Aggregate(results); got != review.VerdictBlocked {
		t.Errorf("blocked result: got %q, want BLOCKED", got)
	}
}

func TestVerdictBlockedOverridesHigh(t *testing.T) {
	results := []checks.DeterministicCheckResult{
		{CheckID: "a", Severity: checks.SeverityHigh},
		{CheckID: "b", Severity: checks.SeverityBlocked},
		{CheckID: "c", Severity: checks.SeverityWarn},
	}
	if got := review.Aggregate(results); got != review.VerdictBlocked {
		t.Errorf("mixed: got %q, want BLOCKED", got)
	}
}
