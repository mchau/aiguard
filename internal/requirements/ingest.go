package requirements

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// IngestFile copies src to .aiguard/requirements/raw.md.
func IngestFile(src, rawMDPath string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("read requirement file %s: %w", src, err)
	}
	if err := os.MkdirAll(filepath.Dir(rawMDPath), 0o755); err != nil {
		return fmt.Errorf("create requirements dir: %w", err)
	}
	if err := os.WriteFile(rawMDPath, data, 0o644); err != nil {
		return fmt.Errorf("write raw.md: %w", err)
	}
	return nil
}

// IngestStdin reads from r and writes to rawMDPath.
func IngestStdin(rawMDPath string, r io.Reader) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("read stdin: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(rawMDPath), 0o755); err != nil {
		return fmt.Errorf("create requirements dir: %w", err)
	}
	if err := os.WriteFile(rawMDPath, data, 0o644); err != nil {
		return fmt.Errorf("write raw.md: %w", err)
	}
	return nil
}

// RequirementSource abstracts external ticket systems.
type RequirementSource interface {
	Fetch(id string) (string, error)
}

// ErrNotImplemented is returned by stub sources.
var ErrNotImplemented = fmt.Errorf("not implemented")

// JiraSource is a stub for future Jira integration.
type JiraSource struct{}

func (j JiraSource) Fetch(id string) (string, error) {
	return "", fmt.Errorf("Jira source: %w", ErrNotImplemented)
}
