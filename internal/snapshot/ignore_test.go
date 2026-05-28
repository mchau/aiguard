package snapshot_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mchau/aiguard/internal/snapshot"
)

func TestMatcherConfigPatterns(t *testing.T) {
	dir := t.TempDir()
	m := snapshot.NewMatcher(dir, []string{"node_modules/**", "*.log"})
	if !m.ShouldIgnore("node_modules/foo/bar.js") {
		t.Error("expected node_modules to be ignored")
	}
	if !m.ShouldIgnore("app.log") {
		t.Error("expected .log to be ignored")
	}
	if m.ShouldIgnore("main.go") {
		t.Error("expected main.go to be included")
	}
}

func TestMatcherGitignore(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("vendor/**\n# comment\n\ndist/**\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	m := snapshot.NewMatcher(dir, nil)
	if !m.ShouldIgnore("vendor/pkg/main.go") {
		t.Error("expected vendor to be ignored via .gitignore")
	}
	if !m.ShouldIgnore("dist/app.js") {
		t.Error("expected dist to be ignored via .gitignore")
	}
	if m.ShouldIgnore("src/main.go") {
		t.Error("expected src/main.go to not be ignored")
	}
}

func TestMatcherAiguardignore(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".aiguardignore"), []byte("internal/generated/**\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	m := snapshot.NewMatcher(dir, nil)
	if !m.ShouldIgnore("internal/generated/proto.go") {
		t.Error("expected .aiguardignore pattern to match")
	}
}

func TestMatcherComposition(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("*.log\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".aiguardignore"), []byte("internal/gen/**\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	m := snapshot.NewMatcher(dir, []string{"node_modules/**"})

	if !m.ShouldIgnore("node_modules/foo") {
		t.Error("config pattern should match")
	}
	if !m.ShouldIgnore("app.log") {
		t.Error(".gitignore pattern should match")
	}
	if !m.ShouldIgnore("internal/gen/pb.go") {
		t.Error(".aiguardignore pattern should match")
	}
	if m.ShouldIgnore("main.go") {
		t.Error("main.go should not be ignored")
	}
}
