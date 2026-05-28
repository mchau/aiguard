package checks

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

// SecretPatterns is the canonical set of secret detection patterns.
// Imported by internal/contextpack/redactor.go — do not duplicate.
var SecretPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)AKIA[0-9A-Z]{16}`),                       // AWS access key
	regexp.MustCompile(`(?i)-----BEGIN (RSA |EC |OPENSSH |)PRIVATE KEY`), // private keys
	regexp.MustCompile(`(?i)Bearer\s+[A-Za-z0-9\-._~+/]+=*`),        // bearer tokens
	regexp.MustCompile(`(?i)ghp_[A-Za-z0-9]{36}`),                   // GitHub personal access token
	regexp.MustCompile(`(?i)xox[baprs]-[0-9A-Za-z\-]+`),             // Slack tokens
	regexp.MustCompile(`(?i)password\s*=\s*\S+`),                     // password = ...
	regexp.MustCompile(`(?i)api_key\s*=\s*\S+`),                      // api_key = ...
	regexp.MustCompile(`(?i)secret\s*=\s*\S+`),                       // secret = ...
}

func init() { Register(&secretCheck{}) }

type secretCheck struct{}

func (s *secretCheck) ID() string { return "secret_detection" }

func (s *secretCheck) Run(_ context.Context, input CheckInput) ([]DeterministicCheckResult, error) {
	if input.DiffText == "" {
		return nil, nil
	}
	// Only scan added lines (lines starting with + but not ++)
	var added strings.Builder
	for _, line := range strings.Split(input.DiffText, "\n") {
		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			added.WriteString(line)
			added.WriteByte('\n')
		}
	}
	addedText := added.String()

	var results []DeterministicCheckResult
	for _, re := range SecretPatterns {
		if loc := re.FindStringIndex(addedText); loc != nil {
			match := addedText[loc[0]:loc[1]]
			if len(match) > 20 {
				match = match[:20] + "..."
			}
			results = append(results, DeterministicCheckResult{
				CheckID:     s.ID(),
				Severity:    SeverityBlocked,
				Pattern:     re.String(),
				Message:     fmt.Sprintf("possible secret detected in diff: %s", match),
				Remediation: "remove the secret from the diff; rotate the credential immediately; use environment variables or a secrets manager",
			})
		}
	}
	return results, nil
}
