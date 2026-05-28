package checks

import (
	"context"
	"fmt"

	"github.com/bmatcuk/doublestar/v4"
)

func init() { Register(&forbiddenCheck{}) }

type forbiddenCheck struct{}

func (f *forbiddenCheck) ID() string { return "forbidden_path" }

func (f *forbiddenCheck) Run(_ context.Context, input CheckInput) ([]DeterministicCheckResult, error) {
	c := cfg(input)
	if c == nil {
		return nil, nil
	}
	var results []DeterministicCheckResult
	for _, file := range input.ChangedFiles {
		for _, pattern := range c.ForbiddenPaths {
			matched, err := doublestar.Match(pattern, file)
			if err != nil {
				return nil, fmt.Errorf("match pattern %q: %w", pattern, err)
			}
			if matched {
				results = append(results, DeterministicCheckResult{
					CheckID:     f.ID(),
					Severity:    SeverityBlocked,
					File:        file,
					Pattern:     pattern,
					Message:     fmt.Sprintf("forbidden file in diff: %s matches %s", file, pattern),
					Remediation: fmt.Sprintf("remove %s from the diff; this file must never be committed", file),
				})
				break
			}
		}
	}
	return results, nil
}
