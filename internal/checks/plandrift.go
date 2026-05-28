package checks

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func init() { Register(&planDriftCheck{}) }

type planDriftCheck struct{}

func (p *planDriftCheck) ID() string { return "plan_drift" }

type planStep struct {
	ID            string   `json:"id"`
	ExpectedFiles []string `json:"expected_files"`
}

type planJSON struct {
	Steps []planStep `json:"steps"`
}

func (p *planDriftCheck) Run(_ context.Context, input CheckInput) ([]DeterministicCheckResult, error) {
	if input.PlanPath == "" {
		return nil, nil
	}

	data, err := os.ReadFile(input.PlanPath)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read plan: %w", err)
	}

	var plan planJSON
	if err := json.Unmarshal(data, &plan); err != nil {
		return nil, nil // plan not yet in JSON format; skip
	}

	// Build set of expected files from plan
	expected := map[string]bool{}
	for _, step := range plan.Steps {
		for _, f := range step.ExpectedFiles {
			expected[f] = true
		}
	}

	if len(expected) == 0 {
		return nil, nil
	}

	var results []DeterministicCheckResult
	for _, changed := range input.ChangedFiles {
		if !expected[changed] && !isTestOrConfig(changed) {
			results = append(results, DeterministicCheckResult{
				CheckID:     p.ID(),
				Severity:    SeverityWarn,
				File:        changed,
				Message:     fmt.Sprintf("file %q changed but not in approved plan", changed),
				Remediation: "update the plan to include this file, or revert the change if unintentional",
			})
		}
	}
	return results, nil
}

func isTestOrConfig(f string) bool {
	lower := strings.ToLower(f)
	for _, ext := range []string{"_test.go", ".test.ts", ".spec.ts", ".yaml", ".yml", ".json", ".md", ".toml"} {
		if strings.HasSuffix(lower, ext) {
			return true
		}
	}
	return false
}
