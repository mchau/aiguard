package checks

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/mchau/aiguard/internal/project"
	"github.com/mchau/aiguard/internal/snapshot"
)

func init() { Register(&snapshotFreshCheck{}) }

type snapshotFreshCheck struct{}

func (s *snapshotFreshCheck) ID() string { return "snapshot_freshness" }

func (s *snapshotFreshCheck) Run(_ context.Context, input CheckInput) ([]DeterministicCheckResult, error) {
	c := cfg(input)
	if c == nil {
		return nil, nil
	}

	// We need the root dir. Use RationalePath as a proxy for root until M7.
	// If neither is available, skip this check gracefully.
	rootDir := rootDirFromInput(input)
	if rootDir == "" {
		return nil, nil
	}

	meta, err := snapshot.LoadMeta(project.SnapshotMetaJSON(rootDir))
	if err != nil {
		return []DeterministicCheckResult{{
			CheckID:     s.ID(),
			Severity:    SeverityInfo,
			Message:     "snapshot-meta.json not found; run 'aiguard snapshot update --full' to create a snapshot",
			Remediation: "run 'aiguard snapshot update --full'",
		}}, nil
	}

	// Count commits since snapshot source commit using the diff.
	// We use a simple heuristic: count "commit " lines in PlanPath field if populated.
	// Full implementation uses git.Client in M7; for now we just check if meta is non-empty.
	_ = meta

	// Warn if stale_after_commits is configured but we can't check (no git client passed here).
	if c.Snapshot.StaleAfterCommits > 0 {
		return []DeterministicCheckResult{{
			CheckID:  s.ID(),
			Severity: SeverityInfo,
			Message:  fmt.Sprintf("snapshot at commit %s (version %d) — freshness check requires git client (M7)", shortSHA(meta.SourceCommit), meta.SnapshotVersion),
		}}, nil
	}

	return nil, nil
}

func shortSHA(sha string) string {
	if len(sha) > 8 {
		return sha[:8]
	}
	return sha
}

// rootDirFromInput extracts the project root from CheckInput.
// RationalePath holds the root dir as a workaround until M7's context pack.
func rootDirFromInput(input CheckInput) string {
	return input.RationalePath
}

// commitCount is a helper for counting commits in a log output.
func commitCount(logOutput string) int {
	count := 0
	for _, line := range strings.Split(logOutput, "\n") {
		if strings.HasPrefix(line, "commit ") {
			count++
		}
	}
	return count
}

var _ = strconv.Itoa // suppress unused import
var _ = commitCount  // suppress unused warning
