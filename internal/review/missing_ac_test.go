package review_test

import (
	"strings"
	"testing"

	"github.com/mchau/aiguard/internal/checks"
	"github.com/mchau/aiguard/internal/review"
)

// TestMissingACProducesRejectWithFixPrompt verifies that a REJECT verdict produces
// a recommended fix prompt mentioning the missing AC IDs.
func TestMissingACProducesRejectWithFixPrompt(t *testing.T) {
	// Simulate: AC2 not covered (High severity finding from traceability)
	detResults := []checks.DeterministicCheckResult{
		{CheckID: "ac_traceability", Severity: checks.SeverityHigh, Message: "AC2 not covered by any plan step or implementation", File: ""},
	}
	verdict := review.Aggregate(detResults)
	if verdict != review.VerdictReject {
		t.Fatalf("expected REJECT for missing AC, got %q", verdict)
	}

	v := &review.FinalVerdict{
		Verdict:             verdict,
		DeterministicChecks: detResults,
		RequirementCoverage: []review.RequirementCoverageResult{
			{ACID: "AC2", Source: "ticket_acceptance_criteria", Status: "FAIL", Notes: "no implementation found"},
		},
	}
	v.RecommendedFixPrompt = review.GenerateFixPrompt(v)

	if v.RecommendedFixPrompt == "" {
		t.Error("expected non-empty fix prompt for REJECT verdict")
	}
	if !strings.Contains(v.RecommendedFixPrompt, "AC2") {
		t.Errorf("fix prompt should mention AC2, got: %s", v.RecommendedFixPrompt)
	}
}
