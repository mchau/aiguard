package project_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mchau/aiguard/internal/project"
)

func TestFindRootViaAiguard(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "a", "b", "c")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	// Create .aiguard/config.yaml at root
	aiguardDir := filepath.Join(root, ".aiguard")
	if err := os.MkdirAll(aiguardDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(aiguardDir, "config.yaml"), []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := project.FindRoot(nested, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != root {
		t.Errorf("got %q, want %q", got, root)
	}
}

func TestFindRootViaGit(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "sub")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}

	got, err := project.FindRoot(nested, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != root {
		t.Errorf("got %q, want %q", got, root)
	}
}

func TestFindRootNotFound(t *testing.T) {
	dir := t.TempDir()
	_, err := project.FindRoot(dir, "")
	if err == nil {
		t.Fatal("expected error when no root found")
	}
}

func TestFindRootCfgOverride(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "custom", "config.yaml")
	if err := os.MkdirAll(filepath.Dir(cfgPath), 0o755); err != nil {
		t.Fatal(err)
	}

	got, err := project.FindRoot("/some/other/cwd", cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if got != filepath.Dir(cfgPath) {
		t.Errorf("got %q, want %q", got, filepath.Dir(cfgPath))
	}
}
