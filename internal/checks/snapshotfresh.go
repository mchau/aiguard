package checks

import (
	"context"
	"fmt"

	"github.com/mchau/aiguard/internal/project"
	"github.com/mchau/aiguard/internal/snapshot"
)

func init() { Register(&snapshotFreshCheck{}) }

type snapshotFreshCheck struct{}

func (s *snapshotFreshCheck) ID() string { return "snapshot_freshness" }

func (s *snapshotFreshCheck) Run(_ context.Context, input CheckInput) ([]DeterministicCheckResult, error) {
	if input.RootDir == "" {
		return nil, nil
	}

	meta, err := snapshot.LoadMeta(project.SnapshotMetaJSON(input.RootDir))
	if err != nil {
		return []DeterministicCheckResult{{
			CheckID:     s.ID(),
			Severity:    SeverityInfo,
			Message:     "snapshot-meta.json not found; run 'aiguard snapshot update --full' to create a snapshot",
			Remediation: "run 'aiguard snapshot update --full'",
		}}, nil
	}

	c := cfg(input)
	if c != nil && c.Snapshot.StaleAfterCommits > 0 {
		return []DeterministicCheckResult{{
			CheckID:  s.ID(),
			Severity: SeverityInfo,
			Message:  fmt.Sprintf("snapshot at commit %s (version %d)", shortSHA(meta.SourceCommit), meta.SnapshotVersion),
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
