package review

import (
	"fmt"
	"strings"
)

// RenderDisagreementReport renders a markdown disagreement report.
func RenderDisagreementReport(disagreements []ReviewerDisagreement) string {
	if len(disagreements) == 0 {
		return "# Reviewer Disagreements\n\nNo disagreements detected.\n"
	}

	var sb strings.Builder
	sb.WriteString("# Reviewer Disagreements\n\n")
	sb.WriteString(fmt.Sprintf("%d disagreement(s) detected.\n\n", len(disagreements)))

	for _, d := range disagreements {
		sb.WriteString(fmt.Sprintf("## %s\n\n", d.Topic))
		for _, f := range d.Findings {
			sb.WriteString(fmt.Sprintf("- **%s** [%s]: %s\n", f.Reviewer, f.Severity, f.Message))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("## Harness Conclusion\n\nWARN: reviewer disagreement detected. Human review recommended before approval.\n")
	return sb.String()
}
