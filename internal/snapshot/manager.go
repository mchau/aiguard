package snapshot

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/mchau/aiguard/internal/config"
	"github.com/mchau/aiguard/internal/git"
	"github.com/mchau/aiguard/internal/project"
)

// FullUpdate writes the 5-file full snapshot set.
// change-log.md is incremental-only and not written here.
func FullUpdate(ctx context.Context, rootDir string, cfg *config.Config, gc *git.Client) error {
	// Resolve current source branch commit
	sourceCommit, err := gc.ResolveRef(ctx, cfg.Project.SourceBranch)
	if err != nil {
		return fmt.Errorf("resolve source branch %q: %w", cfg.Project.SourceBranch, err)
	}

	matcher := NewMatcher(rootDir, cfg.Snapshot.Ignore)
	tree, err := Walk(rootDir, cfg.Snapshot.MaxTreeDepth, matcher)
	if err != nil {
		return fmt.Errorf("walk directory tree: %w", err)
	}

	files := FlattenFiles(tree)
	treeText := RenderTree(tree)

	meta := SnapshotMeta{
		SourceBranch:    cfg.Project.SourceBranch,
		SourceCommit:    sourceCommit,
		UpdatedAt:       time.Now().UTC(),
		SnapshotVersion: 1,
		TreeDepth:       cfg.Snapshot.MaxTreeDepth,
		TotalFiles:      len(files),
	}

	// Check if snapshot already exists to bump version
	existingMeta, err := LoadMeta(project.SnapshotMetaJSON(rootDir))
	if err == nil {
		meta.SnapshotVersion = existingMeta.SnapshotVersion + 1
	}

	snapshotDir := project.SnapshotDir(rootDir)
	if err := os.MkdirAll(snapshotDir, 0o755); err != nil {
		return fmt.Errorf("create snapshot dir: %w", err)
	}

	// snapshot.md
	mdContent := RenderSnapshotMD(meta, treeText)
	if err := os.WriteFile(project.SnapshotMD(rootDir), []byte(mdContent), 0o644); err != nil {
		return fmt.Errorf("write snapshot.md: %w", err)
	}

	// snapshot-meta.json
	if err := writeMeta(project.SnapshotMetaJSON(rootDir), meta); err != nil {
		return err
	}

	// file-map.json
	fileMap := buildFileMap(files)
	if err := writeJSON(project.FileMapJSON(rootDir), fileMap); err != nil {
		return err
	}

	// risk-map.json
	riskMap := buildRiskMap(files, cfg.RiskPaths)
	if err := writeJSON(project.RiskMapJSON(rootDir), riskMap); err != nil {
		return err
	}

	// test-map.json
	testMap := buildTestMap(files)
	if err := writeJSON(project.TestMapJSON(rootDir), testMap); err != nil {
		return err
	}

	return nil
}

// LoadMeta reads snapshot-meta.json.
func LoadMeta(path string) (*SnapshotMeta, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m SnapshotMeta
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

func writeMeta(path string, m SnapshotMeta) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal meta: %w", err)
	}
	return os.WriteFile(path, data, 0o644)
}

func writeJSON(path string, v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal JSON for %s: %w", path, err)
	}
	return os.WriteFile(path, data, 0o644)
}

func buildFileMap(files []string) []FileEntry {
	entries := make([]FileEntry, 0, len(files))
	for _, f := range files {
		entries = append(entries, FileEntry{Path: f})
	}
	return entries
}

func buildRiskMap(files []string, riskPaths config.RiskPaths) map[string][]string {
	rmap := map[string][]string{
		"critical": {},
		"high":     {},
		"medium":   {},
	}
	for _, f := range files {
		for _, p := range riskPaths.Critical {
			if m, _ := doublestar.Match(p, f); m {
				rmap["critical"] = append(rmap["critical"], f)
				goto next
			}
		}
		for _, p := range riskPaths.High {
			if m, _ := doublestar.Match(p, f); m {
				rmap["high"] = append(rmap["high"], f)
				goto next
			}
		}
		for _, p := range riskPaths.Medium {
			if m, _ := doublestar.Match(p, f); m {
				rmap["medium"] = append(rmap["medium"], f)
				goto next
			}
		}
	next:
	}
	return rmap
}

func buildTestMap(files []string) []string {
	var tests []string
	for _, f := range files {
		lower := strings.ToLower(f)
		if strings.HasSuffix(lower, "_test.go") ||
			strings.Contains(lower, "_test.") ||
			strings.HasSuffix(lower, ".test.ts") ||
			strings.HasSuffix(lower, ".spec.ts") {
			tests = append(tests, f)
		}
	}
	return tests
}
