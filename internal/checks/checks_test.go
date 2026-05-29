package checks_test

import (
	"context"
	"os"
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

func TestSecretsAWSSecretAccessKey(t *testing.T) {
	// Regression: AWS_SECRET_ACCESS_KEY= was not caught by the original "secret=" pattern
	diff := "+AWS_SECRET_ACCESS_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY\n"
	input := defaultInput(nil, diff)
	results, err := runCheck("secret_detection", input)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) == 0 || results[0].Severity != checks.SeverityBlocked {
		t.Errorf("expected BLOCKED for AWS_SECRET_ACCESS_KEY, got %v", results)
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

// --- plan_drift scope ---

func TestPlanDriftIgnoresNonRiskFiles(t *testing.T) {
	dir := t.TempDir()
	planPath := dir + "/plan.json"
	if err := os.WriteFile(planPath, []byte(`{"steps":[{"id":"S1","expected_files":["service.go"]}]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	input := checks.TypedInput{
		Config:       config.Default(),
		ChangedFiles: []string{"service.go", "handlers/router.go"}, // router.go is unplanned but not high-risk
		PlanPath:     planPath,
	}.ToCheckInput()
	results, err := runCheck("plan_drift", input)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 0 {
		t.Errorf("expected no plan_drift on non-risk file, got %v", results)
	}
}

func TestPlanDriftFiresOnRiskPath(t *testing.T) {
	dir := t.TempDir()
	planPath := dir + "/plan.json"
	if err := os.WriteFile(planPath, []byte(`{"steps":[{"id":"S1","expected_files":["service.go"]}]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	input := checks.TypedInput{
		Config:       config.Default(),
		ChangedFiles: []string{"service.go", "auth/login.go"}, // auth/** is high-risk and unplanned
		PlanPath:     planPath,
	}.ToCheckInput()
	results, err := runCheck("plan_drift", input)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0].File != "auth/login.go" {
		t.Errorf("expected plan_drift WARN on auth/login.go, got %v", results)
	}
}

// --- AC ID normalization ---

func TestACTraceabilityNormalizesIDs(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(dir+"/.aiguard/requirements", 0o755); err != nil {
		t.Fatal(err)
	}
	acJSON := `[{"id":"AC1","text":"do X","source":"ticket_acceptance_criteria"}]`
	if err := os.WriteFile(dir+"/.aiguard/requirements/acceptance-criteria.json", []byte(acJSON), 0o644); err != nil {
		t.Fatal(err)
	}
	// Plan uses dash-form "ac-1" — should still match "AC1" via normalization
	planJSON := `{"steps":[{"id":"S1","ac_ids":["ac-1"]}]}`
	planPath := dir + "/.aiguard/plans/plan.json"
	if err := os.MkdirAll(dir+"/.aiguard/plans", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(planPath, []byte(planJSON), 0o644); err != nil {
		t.Fatal(err)
	}
	input := checks.TypedInput{
		Config:   config.Default(),
		RootDir:  dir,
		PlanPath: planPath,
	}.ToCheckInput()
	results, err := runCheck("ac_traceability", input)
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range results {
		if r.Severity == checks.SeverityHigh {
			t.Errorf("AC1 covered by 'ac-1' after normalization; should not be High: %v", r)
		}
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
