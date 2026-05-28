package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mchau/aiguard/internal/config"
)

func TestLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	cfg := config.Default()
	if err := config.Write(path, cfg); err != nil {
		t.Fatalf("write: %v", err)
	}

	loaded, err := config.Load(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if err := config.Validate(loaded); err != nil {
		t.Fatalf("validate: %v", err)
	}

	if loaded.Project.SourceBranch != cfg.Project.SourceBranch {
		t.Errorf("source_branch: got %q, want %q", loaded.Project.SourceBranch, cfg.Project.SourceBranch)
	}
	if loaded.Privacy.MaxFileBytesForContext != cfg.Privacy.MaxFileBytesForContext {
		t.Errorf("max_file_bytes: got %d, want %d", loaded.Privacy.MaxFileBytesForContext, cfg.Privacy.MaxFileBytesForContext)
	}
}

func TestLoadMissing(t *testing.T) {
	_, err := config.Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoadBadYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(":\tbad yaml{{{{"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected error for bad YAML")
	}
}
