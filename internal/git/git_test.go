package git_test

import (
	"context"
	"testing"

	"github.com/mchau/aiguard/internal/git"
)

func fakeClient(responses map[string]git.FakeResponse) *git.Client {
	return git.NewWithRunner("/fake/root", git.FakeRunner{Responses: responses})
}

func TestIsRepo(t *testing.T) {
	c := fakeClient(map[string]git.FakeResponse{
		"rev-parse --git-dir": {Output: []byte(".git\n")},
	})
	if !c.IsRepo(context.Background()) {
		t.Error("expected IsRepo to return true")
	}
}

func TestIsRepoFalse(t *testing.T) {
	c := fakeClient(map[string]git.FakeResponse{})
	if c.IsRepo(context.Background()) {
		t.Error("expected IsRepo to return false")
	}
}

func TestCurrentBranch(t *testing.T) {
	c := fakeClient(map[string]git.FakeResponse{
		"rev-parse --abbrev-ref HEAD": {Output: []byte("feature/my-branch\n")},
	})
	branch, err := c.CurrentBranch(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if branch != "feature/my-branch" {
		t.Errorf("got %q", branch)
	}
}

func TestResolveRef(t *testing.T) {
	sha := "abc123def456"
	c := fakeClient(map[string]git.FakeResponse{
		"rev-parse --verify develop": {Output: []byte(sha + "\n")},
	})
	got, err := c.ResolveRef(context.Background(), "develop")
	if err != nil {
		t.Fatal(err)
	}
	if got != sha {
		t.Errorf("got %q, want %q", got, sha)
	}
}

func TestFetch(t *testing.T) {
	c := fakeClient(map[string]git.FakeResponse{
		"fetch origin": {Output: []byte{}},
	})
	if err := c.Fetch(context.Background(), "origin"); err != nil {
		t.Fatal(err)
	}
}

func TestIsWorkingTreeClean(t *testing.T) {
	c := fakeClient(map[string]git.FakeResponse{
		"status --porcelain": {Output: []byte("")},
	})
	clean, err := c.IsWorkingTreeClean(context.Background())
	if err != nil || !clean {
		t.Errorf("expected clean tree, err=%v clean=%v", err, clean)
	}
}

func TestIsWorkingTreeDirty(t *testing.T) {
	c := fakeClient(map[string]git.FakeResponse{
		"status --porcelain": {Output: []byte(" M somefile.go\n")},
	})
	clean, err := c.IsWorkingTreeClean(context.Background())
	if err != nil || clean {
		t.Errorf("expected dirty tree, err=%v clean=%v", err, clean)
	}
}

func TestChangedFiles(t *testing.T) {
	c := fakeClient(map[string]git.FakeResponse{
		"diff --name-only develop...HEAD": {Output: []byte("a.go\nb.go\n")},
	})
	files, err := c.ChangedFiles(context.Background(), "develop")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 || files[0] != "a.go" || files[1] != "b.go" {
		t.Errorf("unexpected files: %v", files)
	}
}

func TestChangedFilesEmpty(t *testing.T) {
	c := fakeClient(map[string]git.FakeResponse{
		"diff --name-only develop...HEAD": {Output: []byte("")},
	})
	files, err := c.ChangedFiles(context.Background(), "develop")
	if err != nil || files != nil {
		t.Errorf("expected nil files, got %v, err=%v", files, err)
	}
}

func TestDiff(t *testing.T) {
	c := fakeClient(map[string]git.FakeResponse{
		"diff develop...HEAD": {Output: []byte("diff --git a/a.go b/a.go\n")},
	})
	diff, err := c.Diff(context.Background(), "develop")
	if err != nil {
		t.Fatal(err)
	}
	if diff == "" {
		t.Error("expected non-empty diff")
	}
}

func TestDiffStat(t *testing.T) {
	c := fakeClient(map[string]git.FakeResponse{
		"diff --stat develop...HEAD": {Output: []byte("1 file changed\n")},
	})
	stat, err := c.DiffStat(context.Background(), "develop")
	if err != nil || stat == "" {
		t.Errorf("unexpected: stat=%q err=%v", stat, err)
	}
}

func TestFileAtRef(t *testing.T) {
	c := fakeClient(map[string]git.FakeResponse{
		"show HEAD:main.go": {Output: []byte("package main\n")},
	})
	content, err := c.FileAtRef(context.Background(), "HEAD", "main.go")
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "package main\n" {
		t.Errorf("unexpected content: %q", content)
	}
}

func TestMergeBase(t *testing.T) {
	c := fakeClient(map[string]git.FakeResponse{
		"merge-base develop HEAD": {Output: []byte("deadbeef\n")},
	})
	base, err := c.MergeBase(context.Background(), "develop", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	if base != "deadbeef" {
		t.Errorf("got %q", base)
	}
}
