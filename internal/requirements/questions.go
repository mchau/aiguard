package requirements

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// LoadDigest reads the digest from acceptance-criteria.json and open-questions.md.
func LoadDigest(acJSONPath, questionsPath string) (*RequirementDigest, error) {
	digest := &RequirementDigest{}

	if data, err := os.ReadFile(acJSONPath); err == nil {
		if err := json.Unmarshal(data, &digest.NormalizedCriteria); err != nil {
			return nil, fmt.Errorf("parse AC JSON: %w", err)
		}
	}

	// Questions are stored inline in the digest for simplicity; parsed from MD if needed.
	return digest, nil
}

// RenderQuestions writes an updated open-questions.md.
func RenderQuestions(questions []ClarificationQuestion) string {
	if len(questions) == 0 {
		return "# Open Questions\n\nNo open questions.\n"
	}
	var sb strings.Builder
	sb.WriteString("# Open Questions\n\n")
	for _, q := range questions {
		status := q.Status
		if status == "" {
			status = "open"
		}
		blocking := ""
		if q.Blocking {
			blocking = " **[BLOCKING]**"
		}
		sb.WriteString(fmt.Sprintf("## %s%s\n\n%s\n\nStatus: %s\n", q.ID, blocking, q.Text, status))
		if q.Answer != "" {
			sb.WriteString(fmt.Sprintf("\nAnswer: %s\n", q.Answer))
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

// HasBlockingOpenQuestions returns true if any question is blocking and open.
func HasBlockingOpenQuestions(questions []ClarificationQuestion) bool {
	for _, q := range questions {
		if q.Blocking && q.Status == "open" {
			return true
		}
	}
	return false
}
