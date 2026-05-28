package requirements_test

import (
	"testing"

	"github.com/mchau/aiguard/internal/requirements"
)

func TestSourceACPreservedWithInferred(t *testing.T) {
	raw := `## Acceptance Criteria
- AC from ticket
- Another AC from ticket
`
	inferred := []requirements.AcceptanceCriterion{
		{Text: "AI inferred this", Source: "inferred"},
	}

	digest := requirements.BuildDigest(raw, inferred)

	// Source ACs must be present and labeled correctly
	sourceCount := 0
	for _, ac := range digest.NormalizedCriteria {
		if ac.Source == "ticket_acceptance_criteria" {
			sourceCount++
		}
	}
	if sourceCount != 2 {
		t.Errorf("expected 2 source ACs, got %d", sourceCount)
	}

	// Inferred AC must be present too
	inferredCount := 0
	for _, ac := range digest.NormalizedCriteria {
		if ac.Source == "inferred" {
			inferredCount++
		}
	}
	if inferredCount != 1 {
		t.Errorf("expected 1 inferred AC, got %d", inferredCount)
	}

	// Total
	if len(digest.NormalizedCriteria) != 3 {
		t.Errorf("expected 3 total ACs, got %d", len(digest.NormalizedCriteria))
	}
}

func TestSourceACPreservedWithConflict(t *testing.T) {
	raw := `## Acceptance Criteria
- Must do X
`
	// Even if inferred says something different, source must remain
	inferred := []requirements.AcceptanceCriterion{
		{Text: "Must NOT do X", Source: "inferred"},
	}

	digest := requirements.BuildDigest(raw, inferred)

	found := false
	for _, ac := range digest.NormalizedCriteria {
		if ac.Source == "ticket_acceptance_criteria" && ac.Text == "Must do X" {
			found = true
		}
	}
	if !found {
		t.Error("source AC 'Must do X' was lost")
	}
}
