package checks

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/mchau/aiguard/internal/project"
	"github.com/mchau/aiguard/internal/requirements"
)

func init() { Register(&traceabilityCheck{}) }

type traceabilityCheck struct{}

func (t *traceabilityCheck) ID() string { return "ac_traceability" }

func (t *traceabilityCheck) Run(_ context.Context, input CheckInput) ([]DeterministicCheckResult, error) {
	if input.RootDir == "" {
		return []DeterministicCheckResult{{
			CheckID:  t.ID(),
			Severity: SeverityInfo,
			Message:  "traceability check skipped (no project root available)",
		}}, nil
	}

	acPath := project.AcceptanceCriteriaJSON(input.RootDir)
	data, err := os.ReadFile(acPath)
	if os.IsNotExist(err) {
		return []DeterministicCheckResult{{
			CheckID:  t.ID(),
			Severity: SeverityInfo,
			Message:  "acceptance-criteria.json not found; run 'aiguard digest create' to enable traceability checks",
		}}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read acceptance-criteria.json: %w", err)
	}

	var acs []requirements.AcceptanceCriterion
	if err := json.Unmarshal(data, &acs); err != nil {
		return nil, fmt.Errorf("parse acceptance-criteria.json: %w", err)
	}

	// Load plan steps if plan.json exists
	coveredByPlan := loadPlanACIDs(input.PlanPath)
	hasPlan := len(coveredByPlan) > 0

	var results []DeterministicCheckResult
	for _, ac := range acs {
		if hasPlan && ac.Source == "ticket_acceptance_criteria" {
			if !coveredByPlan[normalizeACID(ac.ID)] {
				results = append(results, DeterministicCheckResult{
					CheckID:     t.ID(),
					Severity:    SeverityHigh,
					Message:     fmt.Sprintf("%s (%s): %q not covered by any plan step", ac.ID, ac.Source, ac.Text),
					Remediation: fmt.Sprintf("add a plan step for %s or document why it is not applicable", ac.ID),
				})
			}
		} else {
			results = append(results, DeterministicCheckResult{
				CheckID:  t.ID(),
				Severity: SeverityInfo,
				Message:  fmt.Sprintf("%s (%s): %s — no plan yet", ac.ID, ac.Source, ac.Text),
			})
		}
	}
	return results, nil
}

// loadPlanACIDs parses plan.json and returns a set of AC IDs covered by plan steps.
func loadPlanACIDs(planPath string) map[string]bool {
	if planPath == "" {
		return nil
	}
	data, err := os.ReadFile(planPath)
	if err != nil {
		return nil
	}
	var plan struct {
		Steps []struct {
			ACIDs []string `json:"ac_ids"`
		} `json:"steps"`
	}
	if err := json.Unmarshal(data, &plan); err != nil {
		return nil
	}
	covered := map[string]bool{}
	for _, step := range plan.Steps {
		for _, id := range step.ACIDs {
			covered[normalizeACID(id)] = true
		}
	}
	return covered
}

// normalizeACID makes AC ID matching robust to formatting drift in AI output
// ("AC1", "ac-1", "AC_001" all collapse to "ac1").
func normalizeACID(id string) string {
	s := strings.ToLower(id)
	s = strings.ReplaceAll(s, "-", "")
	s = strings.ReplaceAll(s, "_", "")
	s = strings.ReplaceAll(s, " ", "")
	// Strip leading zeros after the AC prefix
	if strings.HasPrefix(s, "ac") {
		rest := strings.TrimLeft(s[2:], "0")
		if rest == "" {
			rest = "0"
		}
		s = "ac" + rest
	}
	return s
}
