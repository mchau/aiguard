package review

import (
	"fmt"
	"strings"

	"github.com/mchau/aiguard/internal/checks"
)

// GenerateFixPrompt produces the "Revise the current patch..." prompt per CLAUDE.md §"Step 15".
func GenerateFixPrompt(v *FinalVerdict) string {
	if v.Verdict == VerdictApprove {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("Revise the current patch.\n\n")

	// Blocked / critical issues
	var blocked, failed, warned []string
	for _, r := range v.DeterministicChecks {
		switch r.Severity {
		case checks.SeverityBlocked:
			blocked = append(blocked, fmt.Sprintf("- %s: %s", r.CheckID, r.Message))
		case checks.SeverityHigh, checks.SeverityCritical:
			failed = append(failed, fmt.Sprintf("- %s: %s", r.CheckID, r.Message))
		case checks.SeverityWarn, checks.SeverityMedium:
			warned = append(warned, fmt.Sprintf("- %s: %s", r.CheckID, r.Message))
		}
	}

	if len(blocked) > 0 {
		sb.WriteString("Remove these blocked changes:\n")
		sb.WriteString(strings.Join(blocked, "\n"))
		sb.WriteString("\n\n")
	}

	if len(failed) > 0 {
		sb.WriteString("Fix these high-severity issues:\n")
		sb.WriteString(strings.Join(failed, "\n"))
		sb.WriteString("\n\n")
	}

	// Missing AC coverage
	var missingAC []string
	for _, rc := range v.RequirementCoverage {
		if rc.Status == "FAIL" {
			missingAC = append(missingAC, fmt.Sprintf("- %s (%s): %s", rc.ACID, rc.Source, rc.Notes))
		}
	}
	if len(missingAC) > 0 {
		sb.WriteString("Implement these missing requirements:\n")
		sb.WriteString(strings.Join(missingAC, "\n"))
		sb.WriteString("\n\n")
	}

	if len(warned) > 0 {
		sb.WriteString("Address these warnings:\n")
		sb.WriteString(strings.Join(warned, "\n"))
		sb.WriteString("\n\n")
	}

	// Disagreements
	for _, d := range v.Disagreements {
		sb.WriteString(fmt.Sprintf("Resolve reviewer disagreement on %s:\n", d.Topic))
		for _, f := range d.Findings {
			sb.WriteString(fmt.Sprintf("  - %s [%s]: %s\n", f.Reviewer, f.Severity, f.Message))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("Do not modify files unrelated to the above issues.\n")
	return sb.String()
}
