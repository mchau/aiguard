package report

import (
	"fmt"
	"strings"

	"github.com/mchau/aiguard/internal/checks"
	"github.com/mchau/aiguard/internal/review"
)

// RenderFinalVerdict renders the markdown report for a FinalVerdict.
// Covers sections 1, 2, 5, 12, 13 from SPEC.md §17.
// Remaining sections are populated in later milestones.
func RenderFinalVerdict(v *review.FinalVerdict) string {
	var sb strings.Builder

	// Section 1: header and verdict
	sb.WriteString("# AIGuard Final Verdict\n\n")
	sb.WriteString("## Verdict\n\n")
	sb.WriteString(v.Verdict)
	sb.WriteString("\n\n")

	// Section 2: summary
	sb.WriteString("## Summary\n\n")
	sb.WriteString(v.Summary)
	sb.WriteString("\n\n")

	// Section 5: deterministic check findings
	if len(v.DeterministicChecks) > 0 {
		sb.WriteString("## Deterministic Check Findings\n\n")
		sb.WriteString("| Check | Severity | File | Message |\n")
		sb.WriteString("|---|---|---|---|\n")
		for _, r := range v.DeterministicChecks {
			sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n",
				r.CheckID, r.Severity, r.File, r.Message))
		}
		sb.WriteString("\n")

		// Remediations
		sb.WriteString("### Remediations\n\n")
		for _, r := range v.DeterministicChecks {
			if r.Remediation != "" {
				sb.WriteString(fmt.Sprintf("- **%s** (%s): %s\n", r.CheckID, r.File, r.Remediation))
			}
		}
		sb.WriteString("\n")
	}

	// Section 12: requirement coverage (placeholder until M11)
	if len(v.RequirementCoverage) > 0 {
		sb.WriteString("## Requirement Coverage\n\n")
		sb.WriteString("| AC | Source | Status | Evidence | Notes |\n")
		sb.WriteString("|---|---|---|---|---|\n")
		for _, rc := range v.RequirementCoverage {
			sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s |\n",
				rc.ACID, rc.Source, rc.Status, rc.Evidence, rc.Notes))
		}
		sb.WriteString("\n")
	}

	// Section 13: recommended fix prompt
	if v.RecommendedFixPrompt != "" {
		sb.WriteString("## Recommended Fix Prompt\n\n")
		sb.WriteString("```\n")
		sb.WriteString(v.RecommendedFixPrompt)
		sb.WriteString("\n```\n\n")
	}

	// Blocked/rejected hint
	if v.Verdict == review.VerdictBlocked || v.Verdict == review.VerdictReject {
		sb.WriteString("---\n\n")
		sb.WriteString(blockedHint(v))
	}

	return sb.String()
}

func blockedHint(v *review.FinalVerdict) string {
	var blocked []string
	for _, r := range v.DeterministicChecks {
		if r.Severity == checks.SeverityBlocked {
			blocked = append(blocked, fmt.Sprintf("- %s: %s", r.CheckID, r.Message))
		}
	}
	if len(blocked) == 0 {
		return ""
	}
	return "**Blocking issues must be resolved before this change can be submitted:**\n\n" +
		strings.Join(blocked, "\n") + "\n"
}
