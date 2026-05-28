package checks

import (
	"context"
	"fmt"
	"strings"
)

func init() { Register(&testPresenceCheck{}) }

type testPresenceCheck struct{}

func (t *testPresenceCheck) ID() string { return "test_presence" }

func (t *testPresenceCheck) Run(_ context.Context, input CheckInput) ([]DeterministicCheckResult, error) {
	// Collect behavior files and test files separately.
	behaviorFiles := make(map[string]bool)
	testFiles := make(map[string]bool)

	for _, f := range input.ChangedFiles {
		if strings.HasSuffix(f, "_test.go") || strings.Contains(f, "_test.") ||
			strings.HasSuffix(f, ".test.ts") || strings.HasSuffix(f, ".test.js") ||
			strings.HasSuffix(f, ".spec.ts") || strings.HasSuffix(f, ".spec.js") {
			testFiles[f] = true
		} else if isBehaviorFile(f) {
			behaviorFiles[f] = true
		}
	}

	if len(behaviorFiles) == 0 || len(testFiles) > 0 {
		return nil, nil
	}

	var files []string
	for f := range behaviorFiles {
		files = append(files, f)
	}

	return []DeterministicCheckResult{{
		CheckID:     t.ID(),
		Severity:    SeverityWarn,
		Message:     fmt.Sprintf("behavior files changed without any test file changes: %s", strings.Join(files, ", ")),
		Remediation: "add or update tests for the changed behavior",
	}}, nil
}

func isBehaviorFile(f string) bool {
	ext := strings.ToLower(f)
	// Skip config, docs, and generated files
	for _, skip := range []string{".md", ".yaml", ".yml", ".json", ".toml", ".txt", ".sql"} {
		if strings.HasSuffix(ext, skip) {
			return false
		}
	}
	// Skip test files themselves (handled above)
	if strings.Contains(ext, "_test") || strings.Contains(ext, ".test.") || strings.Contains(ext, ".spec.") {
		return false
	}
	// Include common source extensions
	for _, src := range []string{".go", ".ts", ".js", ".py", ".rb", ".java", ".cs", ".rs", ".c", ".cpp", ".h"} {
		if strings.HasSuffix(ext, src) {
			return true
		}
	}
	return false
}
