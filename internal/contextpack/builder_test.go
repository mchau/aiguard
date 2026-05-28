package contextpack_test

import (
	"testing"

	"github.com/mchau/aiguard/internal/config"
	"github.com/mchau/aiguard/internal/contextpack"
)

func TestRedactorCatchesSecrets(t *testing.T) {
	text := "aws key: AKIAIOSFODNN7EXAMPLE more text"
	redacted, redactions := contextpack.Redact(text, "test")
	if len(redactions) == 0 {
		t.Error("expected redaction for AWS key")
	}
	if redacted == text {
		t.Error("expected text to be redacted")
	}
}

func TestRedactorGitHubToken(t *testing.T) {
	text := "token: ghp_abcdefghijklmnopqrstuvwxyz1234567890"
	redacted, redactions := contextpack.Redact(text, "env")
	if len(redactions) == 0 {
		t.Error("expected redaction for GitHub token")
	}
	_ = redacted
}

func TestRedactorCleanText(t *testing.T) {
	text := "func hello() string { return \"world\" }"
	redacted, redactions := contextpack.Redact(text, "code")
	if len(redactions) != 0 {
		t.Errorf("expected no redactions for clean text, got %d", len(redactions))
	}
	if redacted != text {
		t.Error("clean text should not be modified")
	}
}

func TestBuilderExcludesForbiddenPaths(t *testing.T) {
	cfg := config.Default()
	opts := contextpack.BuildOptions{
		Step:         "final_review",
		ChangedFiles: []string{".env", "main.go"},
		IncludeFiles: true,
	}
	pack := contextpack.Build(cfg, opts)
	for _, f := range pack.Files {
		if f.Path == ".env" {
			t.Error("forbidden path .env must not appear in context pack")
		}
	}
}

func TestBuilderDefaultOffPrivacy(t *testing.T) {
	cfg := config.Default()
	// Default: include_actual_files_by_default = false
	opts := contextpack.BuildOptions{
		Step:         "final_review",
		ChangedFiles: []string{"main.go"},
		IncludeFiles: false, // caller does not override
	}
	pack := contextpack.Build(cfg, opts)
	if len(pack.Files) != 0 {
		t.Errorf("expected no file contents when privacy default is off, got %d files", len(pack.Files))
	}
}
