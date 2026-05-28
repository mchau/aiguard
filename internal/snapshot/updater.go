package snapshot

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/mchau/aiguard/internal/git"
	"github.com/mchau/aiguard/internal/project"
)

// Incremental appends an entry to change-log.md and bumps snapshot_version.
// prevCommit and curCommit are the two source branch commits to diff.
func Incremental(ctx context.Context, rootDir string, gc *git.Client, prevCommit, curCommit string) error {
	changedFiles, err := gc.ChangedFiles(ctx, prevCommit)
	if err != nil {
		return fmt.Errorf("get changed files %s→%s: %w", prevCommit, curCommit, err)
	}

	entry := buildChangeLogEntry(prevCommit, curCommit, changedFiles)

	changeLogPath := project.ChangeLogMD(rootDir)
	f, err := os.OpenFile(changeLogPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open change-log.md: %w", err)
	}
	defer f.Close()
	if _, err := fmt.Fprint(f, entry); err != nil {
		return fmt.Errorf("write change-log.md: %w", err)
	}

	// Bump snapshot_version in meta
	metaPath := project.SnapshotMetaJSON(rootDir)
	meta, err := LoadMeta(metaPath)
	if err != nil {
		return fmt.Errorf("load snapshot meta: %w", err)
	}
	meta.SnapshotVersion++
	meta.SourceCommit = curCommit
	meta.UpdatedAt = time.Now().UTC()
	return writeMeta(metaPath, *meta)
}

func buildChangeLogEntry(from, to string, files []string) string {
	short := func(s string) string {
		if len(s) > 8 {
			return s[:8]
		}
		return s
	}
	entry := fmt.Sprintf("\n## %s → %s (%s)\n\n", short(from), short(to), time.Now().UTC().Format("2006-01-02"))
	if len(files) == 0 {
		entry += "No changes.\n"
	} else {
		for _, f := range files {
			entry += fmt.Sprintf("- %s\n", f)
		}
	}
	return entry
}
