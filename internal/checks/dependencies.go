package checks

import (
	"context"
	"fmt"

	"github.com/bmatcuk/doublestar/v4"
)

func init() { Register(&dependencyCheck{}) }

type dependencyCheck struct{}

func (d *dependencyCheck) ID() string { return "dependency_changes" }

func (d *dependencyCheck) Run(_ context.Context, input CheckInput) ([]DeterministicCheckResult, error) {
	c := cfg(input)
	if c == nil {
		return nil, nil
	}
	var results []DeterministicCheckResult
	for _, file := range input.ChangedFiles {
		for _, pattern := range c.DependencyFiles {
			matched, err := doublestar.Match(pattern, file)
			if err != nil {
				return nil, fmt.Errorf("match pattern %q: %w", pattern, err)
			}
			if matched {
				results = append(results, DeterministicCheckResult{
					CheckID:     d.ID(),
					Severity:    SeverityWarn,
					File:        file,
					Pattern:     pattern,
					Message:     fmt.Sprintf("dependency file changed: %s", file),
					Remediation: "verify dependency change is intentional and review for supply-chain risk",
				})
				break
			}
		}
	}
	return results, nil
}
