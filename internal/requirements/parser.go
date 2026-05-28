package requirements

import (
	"fmt"
	"strings"
)

// acHeadings is the set of heading names that introduce acceptance criteria.
var acHeadings = []string{
	"acceptance criteria",
	"ac",
	"criteria",
	"requirements",
	"definition of done",
	"expected behavior",
	"business rules",
}

// ExtractSourceAC parses markdown and returns all AC items found under
// recognized AC headings. Each item gets a stable ID (AC1, AC2...) and
// Source set to "ticket_acceptance_criteria".
func ExtractSourceAC(markdown string) []AcceptanceCriterion {
	lines := strings.Split(markdown, "\n")
	inACSection := false
	var results []AcceptanceCriterion
	counter := 1

	for _, line := range lines {
		// Detect heading (# ## ###)
		if strings.HasPrefix(line, "#") {
			headingText := strings.TrimSpace(strings.TrimLeft(line, "#"))
			inACSection = isACHeading(headingText)
			continue
		}

		if !inACSection {
			continue
		}

		// Extract list items: - item, * item, 1. item, - [ ] item
		text := extractListItem(line)
		if text == "" {
			continue
		}

		results = append(results, AcceptanceCriterion{
			ID:     fmt.Sprintf("AC%d", counter),
			Text:   text,
			Source: "ticket_acceptance_criteria",
			Status: "open",
		})
		counter++
	}
	return results
}

func isACHeading(text string) bool {
	lower := strings.ToLower(strings.TrimSpace(text))
	for _, h := range acHeadings {
		if lower == h {
			return true
		}
	}
	return false
}

func extractListItem(line string) string {
	s := strings.TrimSpace(line)
	if s == "" {
		return ""
	}

	// Checkbox: - [ ] text or - [x] text
	if strings.HasPrefix(s, "- [ ]") || strings.HasPrefix(s, "- [x]") || strings.HasPrefix(s, "- [X]") {
		return strings.TrimSpace(s[5:])
	}

	// Bullet: - text or * text
	if strings.HasPrefix(s, "- ") {
		return strings.TrimSpace(s[2:])
	}
	if strings.HasPrefix(s, "* ") {
		return strings.TrimSpace(s[2:])
	}

	// Numbered: 1. text
	if len(s) >= 3 && s[0] >= '0' && s[0] <= '9' {
		rest := s
		for len(rest) > 0 && rest[0] >= '0' && rest[0] <= '9' {
			rest = rest[1:]
		}
		if strings.HasPrefix(rest, ". ") {
			return strings.TrimSpace(rest[2:])
		}
	}

	return ""
}
