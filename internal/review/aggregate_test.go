package review_test

import (
	"testing"

	"github.com/mchau/aiguard/internal/checks"
	"github.com/mchau/aiguard/internal/review"
)

// Table-driven verdict tests per SPEC.md §16

func TestVerdictTableDriven(t *testing.T) {
	tests := []struct {
		name     string
		results  []checks.DeterministicCheckResult
		expected string
	}{
		{"empty → APPROVE", nil, review.VerdictApprove},
		{"warn only → WARN", []checks.DeterministicCheckResult{{Severity: checks.SeverityWarn}}, review.VerdictWarn},
		{"medium only → WARN", []checks.DeterministicCheckResult{{Severity: checks.SeverityMedium}}, review.VerdictWarn},
		{"high → REJECT", []checks.DeterministicCheckResult{{Severity: checks.SeverityHigh}}, review.VerdictReject},
		{"critical → REJECT", []checks.DeterministicCheckResult{{Severity: checks.SeverityCritical}}, review.VerdictReject},
		{"blocked → BLOCKED", []checks.DeterministicCheckResult{{Severity: checks.SeverityBlocked}}, review.VerdictBlocked},
		{"blocked overrides high", []checks.DeterministicCheckResult{
			{Severity: checks.SeverityHigh},
			{Severity: checks.SeverityBlocked},
		}, review.VerdictBlocked},
		{"blocked overrides warn", []checks.DeterministicCheckResult{
			{Severity: checks.SeverityWarn},
			{Severity: checks.SeverityBlocked},
		}, review.VerdictBlocked},
		{"mixed warn+high → REJECT", []checks.DeterministicCheckResult{
			{Severity: checks.SeverityWarn},
			{Severity: checks.SeverityHigh},
		}, review.VerdictReject},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := review.Aggregate(tt.results)
			if got != tt.expected {
				t.Errorf("got %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestAggregateWithReviewersBlockedUnchanged(t *testing.T) {
	det := []checks.DeterministicCheckResult{{Severity: checks.SeverityBlocked}}
	findings := []review.ReviewerFinding{{Reviewer: "r1", Severity: "Info"}}
	got := review.AggregateWithReviewers(det, findings, nil)
	if got != review.VerdictBlocked {
		t.Errorf("BLOCKED should not be overridden by reviewer, got %q", got)
	}
}

func TestAggregateWithReviewersHighFromReviewer(t *testing.T) {
	det := []checks.DeterministicCheckResult{{Severity: checks.SeverityWarn}}
	findings := []review.ReviewerFinding{{Reviewer: "r1", Severity: "High"}}
	got := review.AggregateWithReviewers(det, findings, nil)
	if got != review.VerdictReject {
		t.Errorf("reviewer High should escalate to REJECT, got %q", got)
	}
}

func TestDisagreementDetection(t *testing.T) {
	findings := []review.ReviewerFinding{
		{Reviewer: "claude", Severity: "Info", RelatedAC: "AC1"},
		{Reviewer: "codex", Severity: "High", RelatedAC: "AC1"},
	}
	_, disagreements := review.AggregateFindings(nil, findings)
	if len(disagreements) == 0 {
		t.Error("expected disagreement for AC1 with different severities from different reviewers")
	}
}

func TestNoDisagreementSameBucket(t *testing.T) {
	// Info vs Medium are both "soft" — not a real disagreement, just noise
	findings := []review.ReviewerFinding{
		{Reviewer: "claude", Severity: "Info", RelatedAC: "AC1"},
		{Reviewer: "codex", Severity: "Medium", RelatedAC: "AC1"},
	}
	_, disagreements := review.AggregateFindings(nil, findings)
	if len(disagreements) != 0 {
		t.Errorf("Info vs Medium same bucket should not be a disagreement, got %d", len(disagreements))
	}
}

func TestNoDisagreementSameReviewer(t *testing.T) {
	findings := []review.ReviewerFinding{
		{Reviewer: "claude", Severity: "Info", RelatedAC: "AC1"},
		{Reviewer: "claude", Severity: "High", RelatedAC: "AC1"},
	}
	_, disagreements := review.AggregateFindings(nil, findings)
	if len(disagreements) != 0 {
		t.Errorf("expected no disagreement when same reviewer, got %d", len(disagreements))
	}
}
