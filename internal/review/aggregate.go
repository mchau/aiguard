package review

import (
	"fmt"
	"strings"

	"github.com/mchau/aiguard/internal/checks"
)

// AggregateFindings groups deterministic + reviewer results and detects disagreements.
func AggregateFindings(deterministic []checks.DeterministicCheckResult, reviewerResults []ReviewerFinding) ([]checks.DeterministicCheckResult, []ReviewerDisagreement) {
	disagreements := detectDisagreements(reviewerResults)
	return deterministic, disagreements
}

// detectDisagreements finds cases where reviewers differ in verdict for the same AC.
func detectDisagreements(findings []ReviewerFinding) []ReviewerDisagreement {
	// Group by RelatedAC
	byAC := map[string][]ReviewerFinding{}
	for _, f := range findings {
		key := f.RelatedAC
		if key == "" {
			key = "general"
		}
		byAC[key] = append(byAC[key], f)
	}

	var disagreements []ReviewerDisagreement
	for ac, acFindings := range byAC {
		if len(acFindings) < 2 {
			continue
		}
		// Disagreement: different reviewers with different severities
		severities := map[string]bool{}
		reviewers := map[string]bool{}
		for _, f := range acFindings {
			severities[f.Severity] = true
			reviewers[f.Reviewer] = true
		}
		if len(severities) > 1 && len(reviewers) > 1 {
			disagreements = append(disagreements, ReviewerDisagreement{
				Topic:    fmt.Sprintf("AC: %s — severities: %s", ac, strings.Join(mapKeys(severities), ", ")),
				Findings: acFindings,
			})
		}
	}
	return disagreements
}

func mapKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
