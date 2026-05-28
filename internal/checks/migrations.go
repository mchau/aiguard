package checks

import (
	"context"
	"fmt"

	"github.com/bmatcuk/doublestar/v4"
)

func init() { Register(&migrationCheck{}) }

type migrationCheck struct{}

func (m *migrationCheck) ID() string { return "migration_changes" }

func (m *migrationCheck) Run(_ context.Context, input CheckInput) ([]DeterministicCheckResult, error) {
	c := cfg(input)
	if c == nil {
		return nil, nil
	}
	var results []DeterministicCheckResult
	for _, file := range input.ChangedFiles {
		for _, pattern := range c.MigrationPaths {
			matched, err := doublestar.Match(pattern, file)
			if err != nil {
				return nil, fmt.Errorf("match pattern %q: %w", pattern, err)
			}
			if matched {
				results = append(results, DeterministicCheckResult{
					CheckID:     m.ID(),
					Severity:    SeverityHigh,
					File:        file,
					Pattern:     pattern,
					Message:     fmt.Sprintf("migration file changed: %s", file),
					Remediation: "review migration for rollback plan and data safety; ensure it is included in the approved plan",
				})
				break
			}
		}
	}
	return results, nil
}
