package checks_test

import (
	"context"
	"strings"
	"testing"

	"github.com/mchau/aiguard/internal/checks"
	"github.com/mchau/aiguard/internal/config"
)

func defaultInput(files []string, diff string) checks.CheckInput {
	return checks.TypedInput{
		Config:       config.Default(),
		ChangedFiles: files,
		DiffText:     diff,
	}.ToCheckInput()
}

// --- forbidden path ---

func TestForbiddenPathBlocked(t *testing.T) {
	input := defaultInput([]string{".env"}, "")
	results, err := runCheck("forbidden_path", input)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) == 0 || results[0].Severity != checks.SeverityBlocked {
		t.Errorf("expected BLOCKED, got %v", results)
	}
}

func TestForbiddenPathClean(t *testing.T) {
	input := defaultInput([]string{"main.go"}, "")
	results, err := runCheck("forbidden_path", input)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 0 {
		t.Errorf("expected no results, got %v", results)
	}
}

// --- risk path ---

func TestRiskPathCritical(t *testing.T) {
	input := defaultInput([]string{"secrets/token.txt"}, "")
	results, err := runCheck("risk_path", input)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) == 0 || results[0].Severity != checks.SeverityBlocked {
		t.Errorf("expected BLOCKED for critical risk, got %v", results)
	}
}

func TestRiskPathHigh(t *testing.T) {
	input := defaultInput([]string{"auth/login.go"}, "")
	results, err := runCheck("risk_path", input)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) == 0 || results[0].Severity != checks.SeverityHigh {
		t.Errorf("expected High for auth path, got %v", results)
	}
}

func TestRiskPathClean(t *testing.T) {
	input := defaultInput([]string{"utils/helper.go"}, "")
	results, err := runCheck("risk_path", input)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 0 {
		t.Errorf("expected no results for safe path, got %v", results)
	}
}

// --- diff size ---

func TestDiffSizeLarge(t *testing.T) {
	diff := strings.Repeat("+line\n", 600)
	input := defaultInput(nil, diff)
	results, err := runCheck("diff_size", input)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) == 0 {
		t.Error("expected warn for large diff")
	}
}

func TestDiffSizeSmall(t *testing.T) {
	input := defaultInput(nil, "+one line\n")
	results, err := runCheck("diff_size", input)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 0 {
		t.Errorf("expected no results for small diff, got %v", results)
	}
}

// --- test presence ---

func TestTestPresenceMissingTest(t *testing.T) {
	input := defaultInput([]string{"service.go"}, "")
	results, err := runCheck("test_presence", input)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) == 0 {
		t.Error("expected warn for missing test")
	}
}

func TestTestPresenceWithTest(t *testing.T) {
	input := defaultInput([]string{"service.go", "service_test.go"}, "")
	results, err := runCheck("test_presence", input)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 0 {
		t.Errorf("expected no results when test is present, got %v", results)
	}
}

// --- secrets ---

func TestSecretsAWSKey(t *testing.T) {
	diff := "+AKIAIOSFODNN7EXAMPLE\n"
	input := defaultInput(nil, diff)
	results, err := runCheck("secret_detection", input)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) == 0 || results[0].Severity != checks.SeverityBlocked {
		t.Errorf("expected BLOCKED for AWS key, got %v", results)
	}
}

func TestSecretsGitHubToken(t *testing.T) {
	diff := "+ghp_abcdefghijklmnopqrstuvwxyz1234567890\n"
	input := defaultInput(nil, diff)
	results, err := runCheck("secret_detection", input)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) == 0 {
		t.Error("expected detection of GitHub token")
	}
}

func TestSecretsClean(t *testing.T) {
	diff := "+func hello() { return \"world\" }\n"
	input := defaultInput(nil, diff)
	results, err := runCheck("secret_detection", input)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 0 {
		t.Errorf("expected no secrets, got %v", results)
	}
}

func TestSecretsNotInRemovedLines(t *testing.T) {
	// Removed lines (starting with -) should not trigger
	diff := "-AKIAIOSFODNN7EXAMPLE\n"
	input := defaultInput(nil, diff)
	results, err := runCheck("secret_detection", input)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 0 {
		t.Errorf("expected no results for removed lines, got %v", results)
	}
}

// helper: run a specific check by ID
func runCheck(id string, input checks.CheckInput) ([]checks.DeterministicCheckResult, error) {
	// Use RunAll but filter — simpler than exposing individual checks
	// We rely on registration order to get a superset then filter
	all, err := checks.RunAll(context.Background(), input)
	if err != nil {
		return nil, err
	}
	var out []checks.DeterministicCheckResult
	for _, r := range all {
		if r.CheckID == id {
			out = append(out, r)
		}
	}
	return out, nil
}
