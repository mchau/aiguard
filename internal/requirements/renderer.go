package requirements

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// RenderDigest writes all requirement artifacts for a digest.
func RenderDigest(rootDir string, d *RequirementDigest, digestMD, acJSON, questionsMD, assumptionsMD string) error {
	// digest.md
	if err := os.WriteFile(digestMD, []byte(renderDigestMD(d)), 0o644); err != nil {
		return fmt.Errorf("write digest.md: %w", err)
	}

	// acceptance-criteria.json
	data, err := json.MarshalIndent(d.NormalizedCriteria, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal AC: %w", err)
	}
	if err := os.WriteFile(acJSON, data, 0o644); err != nil {
		return fmt.Errorf("write acceptance-criteria.json: %w", err)
	}

	// open-questions.md
	if err := os.WriteFile(questionsMD, []byte(renderQuestionsMD(d.OpenQuestions)), 0o644); err != nil {
		return fmt.Errorf("write open-questions.md: %w", err)
	}

	// assumptions.md
	if err := os.WriteFile(assumptionsMD, []byte(renderAssumptionsMD(d.Assumptions)), 0o644); err != nil {
		return fmt.Errorf("write assumptions.md: %w", err)
	}

	return nil
}

func renderDigestMD(d *RequirementDigest) string {
	var sb strings.Builder
	sb.WriteString("# Requirement Digest\n\n")

	if d.RawSummary != "" {
		sb.WriteString("## Summary\n\n")
		sb.WriteString(d.RawSummary)
		sb.WriteString("\n\n")
	}

	sb.WriteString("## Acceptance Criteria\n\n")
	for _, ac := range d.NormalizedCriteria {
		tag := ""
		if ac.Source == "inferred" {
			tag = " *(inferred)*"
		}
		sb.WriteString(fmt.Sprintf("- **%s**%s: %s\n", ac.ID, tag, ac.Text))
	}
	sb.WriteString("\n")

	if len(d.NonGoals) > 0 {
		sb.WriteString("## Non-Goals\n\n")
		for _, ng := range d.NonGoals {
			sb.WriteString("- " + ng + "\n")
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func renderQuestionsMD(questions []ClarificationQuestion) string {
	if len(questions) == 0 {
		return "# Open Questions\n\nNo open questions.\n"
	}
	var sb strings.Builder
	sb.WriteString("# Open Questions\n\n")
	for _, q := range questions {
		blocking := ""
		if q.Blocking {
			blocking = " **[BLOCKING]**"
		}
		sb.WriteString(fmt.Sprintf("- **%s**%s: %s (status: %s)\n", q.ID, blocking, q.Text, q.Status))
		if q.Answer != "" {
			sb.WriteString(fmt.Sprintf("  - Answer: %s\n", q.Answer))
		}
	}
	return sb.String()
}

func renderAssumptionsMD(assumptions []Assumption) string {
	if len(assumptions) == 0 {
		return "# Assumptions\n\nNo assumptions recorded.\n"
	}
	var sb strings.Builder
	sb.WriteString("# Assumptions\n\n")
	for _, a := range assumptions {
		sb.WriteString(fmt.Sprintf("- **%s**: %s\n", a.ID, a.Text))
		if a.Risk != "" {
			sb.WriteString(fmt.Sprintf("  - Risk: %s\n", a.Risk))
		}
	}
	return sb.String()
}
