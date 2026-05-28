package project

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// FindRoot walks up from cwd looking for a .aiguard/config.yaml or .git/ directory.
// If cfgOverride is non-empty, its parent directory is returned directly.
// Returns an actionable error if neither marker is found.
func FindRoot(cwd, cfgOverride string) (string, error) {
	if cfgOverride != "" {
		abs, err := filepath.Abs(cfgOverride)
		if err != nil {
			return "", fmt.Errorf("resolve --config path %q: %w", cfgOverride, err)
		}
		return filepath.Dir(abs), nil
	}

	dir := cwd
	for {
		if _, err := os.Stat(filepath.Join(dir, DotAiguard, "config.yaml")); err == nil {
			return dir, nil
		}
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", errors.New("no .aiguard/ or .git/ found; run 'aiguard init' to initialize this project")
}
