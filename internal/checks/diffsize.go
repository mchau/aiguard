package checks

import (
	"context"
	"fmt"
	"strings"
)

func init() { Register(&diffSizeCheck{}) }

const (
	diffSizeWarnLines = 500
	diffSizeFailLines = 2000
)

type diffSizeCheck struct{}

func (d *diffSizeCheck) ID() string { return "diff_size" }

func (d *diffSizeCheck) Run(_ context.Context, input CheckInput) ([]DeterministicCheckResult, error) {
	if input.DiffText == "" {
		return nil, nil
	}
	lines := strings.Count(input.DiffText, "\n")
	if lines >= diffSizeFailLines {
		return []DeterministicCheckResult{{
			CheckID:     d.ID(),
			Severity:    SeverityHigh,
			Message:     fmt.Sprintf("diff is very large: %d lines (threshold: %d)", lines, diffSizeFailLines),
			Remediation: "break the change into smaller, focused commits",
		}}, nil
	}
	if lines >= diffSizeWarnLines {
		return []DeterministicCheckResult{{
			CheckID:     d.ID(),
			Severity:    SeverityWarn,
			Message:     fmt.Sprintf("diff is large: %d lines (threshold: %d)", lines, diffSizeWarnLines),
			Remediation: "consider splitting into smaller commits for easier review",
		}}, nil
	}
	return nil, nil
}
