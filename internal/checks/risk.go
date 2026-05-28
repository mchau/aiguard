package checks

import (
	"context"
	"fmt"

	"github.com/bmatcuk/doublestar/v4"
)

func init() { Register(&riskCheck{}) }

type riskCheck struct{}

func (r *riskCheck) ID() string { return "risk_path" }

func (r *riskCheck) Run(_ context.Context, input CheckInput) ([]DeterministicCheckResult, error) {
	c := cfg(input)
	if c == nil {
		return nil, nil
	}

	type level struct {
		severity string
		patterns []string
	}
	levels := []level{
		{SeverityBlocked, c.RiskPaths.Critical},
		{SeverityHigh, c.RiskPaths.High},
		{SeverityMedium, c.RiskPaths.Medium},
	}

	var results []DeterministicCheckResult
	for _, file := range input.ChangedFiles {
		for _, lv := range levels {
			for _, pattern := range lv.patterns {
				matched, err := doublestar.Match(pattern, file)
				if err != nil {
					return nil, fmt.Errorf("match pattern %q: %w", pattern, err)
				}
				if matched {
					results = append(results, DeterministicCheckResult{
						CheckID:     r.ID(),
						Severity:    lv.severity,
						File:        file,
						Pattern:     pattern,
						Message:     fmt.Sprintf("risk path changed (%s): %s", lv.severity, file),
						Remediation: "ensure this file has an approved rationale entry and corresponding plan step",
					})
					goto nextFile
				}
			}
		}
	nextFile:
	}
	return results, nil
}
