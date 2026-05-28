package snapshot

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

// Matcher decides whether a file path should be excluded from the snapshot.
type Matcher struct {
	patterns []string
}

// NewMatcher builds a composable matcher from:
//  1. Patterns from config.snapshot.ignore
//  2. Lines from .gitignore at rootDir (if present)
//  3. Lines from .aiguardignore at rootDir (if present)
func NewMatcher(rootDir string, configPatterns []string) *Matcher {
	patterns := make([]string, 0, len(configPatterns))
	patterns = append(patterns, configPatterns...)
	patterns = append(patterns, readIgnoreFile(filepath.Join(rootDir, ".gitignore"))...)
	patterns = append(patterns, readIgnoreFile(filepath.Join(rootDir, ".aiguardignore"))...)
	return &Matcher{patterns: patterns}
}

// ShouldIgnore returns true if path matches any ignore pattern.
func (m *Matcher) ShouldIgnore(path string) bool {
	for _, p := range m.patterns {
		matched, _ := doublestar.Match(p, path)
		if matched {
			return true
		}
	}
	return false
}

func readIgnoreFile(path string) []string {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()

	var patterns []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		patterns = append(patterns, line)
	}
	return patterns
}
