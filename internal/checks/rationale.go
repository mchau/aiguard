package checks

import (
	"context"
	"fmt"
	"os"
	"strings"
)

// Note: rationale check is NOT registered here via init().
// It is registered in internal/cli/rationale.go when rationale update creates the file,
// following the plan rule: never register an always-pass placeholder check.

// RationaleCheck runs the rationale file check.
// Called directly by aiguard rationale check and by the CLI when rationale file exists.
type RationaleCheck struct{}

func (r *RationaleCheck) ID() string { return "rationale_missing" }

func (r *RationaleCheck) Run(_ context.Context, input CheckInput) ([]DeterministicCheckResult, error) {
	if input.RationalePath == "" {
		return nil, nil
	}

	data, err := os.ReadFile(input.RationalePath)
	if os.IsNotExist(err) {
		var results []DeterministicCheckResult
		for _, f := range input.ChangedFiles {
			if !isSafeToSkipRationale(f) {
				results = append(results, DeterministicCheckResult{
					CheckID:     r.ID(),
					Severity:    SeverityHigh,
					File:        f,
					Message:     fmt.Sprintf("no rationale found for changed file %q", f),
					Remediation: "run 'aiguard rationale update' to document the reason for each file change",
				})
			}
		}
		return results, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read rationale: %w", err)
	}

	// Check each changed file has an entry in the rationale
	content := string(data)
	var results []DeterministicCheckResult
	for _, f := range input.ChangedFiles {
		if isSafeToSkipRationale(f) {
			continue
		}
		if !strings.Contains(content, f) {
			results = append(results, DeterministicCheckResult{
				CheckID:     r.ID(),
				Severity:    SeverityHigh,
				File:        f,
				Message:     fmt.Sprintf("file %q not found in rationale", f),
				Remediation: "run 'aiguard rationale update' to add a rationale entry for this file",
			})
		}
	}
	return results, nil
}

func isSafeToSkipRationale(f string) bool {
	lower := strings.ToLower(f)
	for _, ext := range []string{"_test.go", ".test.ts", ".spec.ts", ".md", ".txt"} {
		if strings.HasSuffix(lower, ext) {
			return true
		}
	}
	return false
}
