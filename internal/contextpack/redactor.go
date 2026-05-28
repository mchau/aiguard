package contextpack

import (
	"fmt"

	"github.com/mchau/aiguard/internal/checks"
)

// Redact scans text for secret patterns and replaces matches with [REDACTED].
// Patterns are imported from internal/checks/secrets.go (single source of truth).
// Returns the redacted text and a list of what was masked.
func Redact(text, location string) (string, []Redaction) {
	var redactions []Redaction
	result := text

	for i, re := range checks.SecretPatterns {
		matches := re.FindAllStringIndex(result, -1)
		if len(matches) == 0 {
			continue
		}
		replacement := fmt.Sprintf("[REDACTED-pattern-%d]", i)
		redacted := re.ReplaceAllString(result, replacement)
		if redacted != result {
			result = redacted
			redactions = append(redactions, Redaction{
				Pattern:     re.String(),
				Location:    location,
				Replacement: replacement,
			})
		}
	}

	return result, redactions
}
