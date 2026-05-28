package cli_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/mchau/aiguard/internal/review"
)

// TestVerifyForbiddenFileBlocked runs a full integration: temp repo, forbidden file diff, verify BLOCKED.
func TestVerifyForbiddenFileBlocked(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}

	dir := t.TempDir()
	run := func(args ...string) string {
		t.Helper()
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Logf("cmd %v: %s", args, out)
		}
		return string(out)
	}

	// Bootstrap a git repo with a base commit on 'main'
	run("git", "init", "-b", "main")
	run("git", "config", "user.email", "test@test.com")
	run("git", "config", "user.name", "Test")
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	run("git", "add", "main.go")
	run("git", "commit", "-m", "init")

	// Create a feature branch with a forbidden file
	run("git", "checkout", "-b", "feature")
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte("SECRET=hunter2\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	run("git", "add", ".env")
	run("git", "commit", "-m", "add forbidden file")

	// Run aiguard init
	aiguardBin := filepath.Join(os.TempDir(), "aiguard_test_bin")
	buildCmd := exec.Command("go", "build", "-o", aiguardBin, "./cmd/aiguard")
	buildCmd.Dir = findModRoot(t)
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("build aiguard: %s", out)
	}

	initCmd := exec.Command(aiguardBin, "init", "--source-branch", "main")
	initCmd.Dir = dir
	if out, err := initCmd.CombinedOutput(); err != nil {
		t.Fatalf("aiguard init: %s", out)
	}

	// Run verify --local-only --base main
	verifyCmd := exec.Command(aiguardBin, "verify", "--local-only", "--base", "main")
	verifyCmd.Dir = dir
	verifyOut, _ := verifyCmd.CombinedOutput()
	t.Logf("verify output: %s", verifyOut)

	// Assert verdict file exists and is BLOCKED
	verdictPath := filepath.Join(dir, ".aiguard", "reviews", "final-verdict.json")
	if _, err := os.Stat(verdictPath); err != nil {
		t.Fatalf("final-verdict.json not found: %v", err)
	}
	data, err := os.ReadFile(verdictPath)
	if err != nil {
		t.Fatal(err)
	}
	var v review.FinalVerdict
	if err := json.Unmarshal(data, &v); err != nil {
		t.Fatalf("parse verdict: %v", err)
	}
	if v.Verdict != review.VerdictBlocked {
		t.Errorf("expected BLOCKED, got %q", v.Verdict)
	}

	// Assert markdown report exists
	mdPath := filepath.Join(dir, ".aiguard", "reviews", "final-verdict.md")
	if _, err := os.Stat(mdPath); err != nil {
		t.Fatalf("final-verdict.md not found: %v", err)
	}
}

// findModRoot walks up from this file to find the go.mod directory.
func findModRoot(t *testing.T) string {
	t.Helper()
	dir, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find go.mod")
		}
		dir = parent
	}
}
