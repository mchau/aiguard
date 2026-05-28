package snapshot_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/mchau/aiguard/internal/config"
	"github.com/mchau/aiguard/internal/git"
	"github.com/mchau/aiguard/internal/project"
	"github.com/mchau/aiguard/internal/snapshot"
)

func TestFullUpdateProducesAllFiles(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	dir := t.TempDir()

	// Init git repo with one commit
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("cmd %v: %s: %v", args, out, err)
		}
	}
	run("git", "init", "-b", "main")
	run("git", "config", "user.email", "t@t.com")
	run("git", "config", "user.name", "T")
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	run("git", "add", ".")
	run("git", "commit", "-m", "init")

	cfg := config.Default()
	cfg.Project.SourceBranch = "main"
	gc := git.New(dir)

	// Create snapshot dir
	if err := os.MkdirAll(project.SnapshotDir(dir), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := snapshot.FullUpdate(context.Background(), dir, cfg, gc); err != nil {
		t.Fatalf("FullUpdate: %v", err)
	}

	// Assert all 5 files exist
	required := []string{
		project.SnapshotMD(dir),
		project.SnapshotMetaJSON(dir),
		project.FileMapJSON(dir),
		project.RiskMapJSON(dir),
		project.TestMapJSON(dir),
	}
	for _, p := range required {
		if _, err := os.Stat(p); err != nil {
			t.Errorf("missing file: %s", p)
		}
	}

	// Assert meta SHA matches HEAD
	meta, err := snapshot.LoadMeta(project.SnapshotMetaJSON(dir))
	if err != nil {
		t.Fatal(err)
	}
	headSHA, _ := gc.ResolveRef(context.Background(), "HEAD")
	if meta.SourceCommit != headSHA {
		t.Errorf("meta.SourceCommit=%q, HEAD=%q", meta.SourceCommit, headSHA)
	}
}
