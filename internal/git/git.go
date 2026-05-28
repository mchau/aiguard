package git

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// Runner executes a git command and returns its stdout.
type Runner interface {
	Run(ctx context.Context, args ...string) ([]byte, error)
}

// ExecRunner runs git via the OS.
type ExecRunner struct {
	WorkDir string
}

func (r ExecRunner) Run(ctx context.Context, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = r.WorkDir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, stderr.String())
	}
	return out, nil
}

// Client wraps a Runner with a project root.
type Client struct {
	Root   string
	runner Runner
}

// New creates a Client using ExecRunner at root.
func New(root string) *Client {
	return &Client{Root: root, runner: ExecRunner{WorkDir: root}}
}

// NewWithRunner creates a Client with a custom runner (for testing).
func NewWithRunner(root string, r Runner) *Client {
	return &Client{Root: root, runner: r}
}

// IsRepo returns true if root is inside a git repository.
func (c *Client) IsRepo(ctx context.Context) bool {
	_, err := c.runner.Run(ctx, "rev-parse", "--git-dir")
	return err == nil
}

// CurrentBranch returns the current branch name.
func (c *Client) CurrentBranch(ctx context.Context) (string, error) {
	out, err := c.runner.Run(ctx, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// ResolveRef resolves a ref (branch, tag, or commit) to a full SHA.
func (c *Client) ResolveRef(ctx context.Context, ref string) (string, error) {
	out, err := c.runner.Run(ctx, "rev-parse", "--verify", ref)
	if err != nil {
		return "", fmt.Errorf("resolve ref %q: %w", ref, err)
	}
	return strings.TrimSpace(string(out)), nil
}

// Fetch fetches from remote (default: origin).
func (c *Client) Fetch(ctx context.Context, remote string) error {
	if remote == "" {
		remote = "origin"
	}
	_, err := c.runner.Run(ctx, "fetch", remote)
	return err
}

// IsWorkingTreeClean returns true if there are no uncommitted changes.
func (c *Client) IsWorkingTreeClean(ctx context.Context) (bool, error) {
	out, err := c.runner.Run(ctx, "status", "--porcelain")
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(string(out)) == "", nil
}

// ChangedFiles returns files changed between base and HEAD using three-dot diff.
func (c *Client) ChangedFiles(ctx context.Context, base string) ([]string, error) {
	out, err := c.runner.Run(ctx, "diff", "--name-only", base+"...HEAD")
	if err != nil {
		return nil, err
	}
	raw := strings.TrimSpace(string(out))
	if raw == "" {
		return nil, nil
	}
	return strings.Split(raw, "\n"), nil
}

// Diff returns the full diff between base and HEAD.
func (c *Client) Diff(ctx context.Context, base string) (string, error) {
	out, err := c.runner.Run(ctx, "diff", base+"...HEAD")
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// DiffStat returns the diffstat between base and HEAD.
func (c *Client) DiffStat(ctx context.Context, base string) (string, error) {
	out, err := c.runner.Run(ctx, "diff", "--stat", base+"...HEAD")
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// FileAtRef returns the content of path at the given ref.
func (c *Client) FileAtRef(ctx context.Context, ref, path string) ([]byte, error) {
	return c.runner.Run(ctx, "show", ref+":"+path)
}

// MergeBase finds the common ancestor of two commits.
func (c *Client) MergeBase(ctx context.Context, a, b string) (string, error) {
	out, err := c.runner.Run(ctx, "merge-base", a, b)
	if err != nil {
		return "", fmt.Errorf("merge-base %q %q: %w", a, b, err)
	}
	return strings.TrimSpace(string(out)), nil
}
