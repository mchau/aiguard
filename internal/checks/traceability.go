package checks

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/mchau/aiguard/internal/requirements"
)

func init() { Register(&traceabilityCheck{}) }

type traceabilityCheck struct{}

func (t *traceabilityCheck) ID() string { return "ac_traceability" }

func (t *traceabilityCheck) Run(_ context.Context, input CheckInput) ([]DeterministicCheckResult, error) {
	c := cfg(input)
	if c == nil {
		return nil, nil
	}

	// Attempt to load acceptance-criteria.json if the path helper is available.
	// The check is registered now but promoted to FAIL in M8.9.
	// If no AC file exists, emit INFO to note that traceability is not yet configured.
	acPath := acPathFromInput(input)
	if acPath == "" {
		return []DeterministicCheckResult{{
			CheckID:  t.ID(),
			Severity: SeverityInfo,
			Message:  "acceptance-criteria.json not found; run 'aiguard digest create' to enable traceability checks",
		}}, nil
	}

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

	// INFO: list ACs — no plan exists yet to check against.
	// Promoted to FAIL in M8.9 when plan exists.
	var results []DeterministicCheckResult
	for _, ac := range acs {
		results = append(results, DeterministicCheckResult{
			CheckID:  t.ID(),
			Severity: SeverityInfo,
			Message:  fmt.Sprintf("%s (%s): %s — no plan step yet", ac.ID, ac.Source, ac.Text),
		})
	}
	return results, nil
}

// acPathFromInput extracts the AC JSON path from CheckInput when available.
// The path is stored in PlanPath field (overloaded) or derived from RationalePath.
// A proper context pack will populate this in M7.
func acPathFromInput(input CheckInput) string {
	// PlanPath is used to pass the AC path until M7 builds a proper context pack.
	return input.PlanPath
}
