package snapshot

import (
	"fmt"
	"strings"
	"time"
)

// RenderSnapshotMD builds the snapshot.md content per CLAUDE.md §"Step 5".
func RenderSnapshotMD(meta SnapshotMeta, treeText string) string {
	var sb strings.Builder

	sb.WriteString("# Codebase Snapshot\n\n")
	sb.WriteString("## Metadata\n\n")
	sb.WriteString(fmt.Sprintf("- Source branch: %s\n", meta.SourceBranch))
	sb.WriteString(fmt.Sprintf("- Source commit: %s\n", meta.SourceCommit))
	sb.WriteString(fmt.Sprintf("- Updated at: %s\n", meta.UpdatedAt.Format(time.RFC3339)))
	sb.WriteString(fmt.Sprintf("- Snapshot version: %d\n", meta.SnapshotVersion))
	sb.WriteString(fmt.Sprintf("- Total files: %d\n", meta.TotalFiles))
	sb.WriteString("\n")

	sb.WriteString("## Folder Tree\n\n")
	sb.WriteString("```\n")
	sb.WriteString(treeText)
	sb.WriteString("```\n\n")

	sb.WriteString("## Component Notes\n\n")
	sb.WriteString("*(Run `aiguard snapshot update` with an AI profile to populate component notes.)*\n\n")

	sb.WriteString("## Architecture Patterns\n\n")
	sb.WriteString("*(Not yet populated.)*\n\n")

	sb.WriteString("## High-Risk Areas\n\n")
	sb.WriteString("*(Derived from risk_paths config — see risk-map.json.)*\n\n")

	sb.WriteString("## Test Locations\n\n")
	sb.WriteString("*(See test-map.json.)*\n\n")

	sb.WriteString("## Recent Changes\n\n")
	sb.WriteString("*(Run `aiguard snapshot update --incremental` to populate.)*\n")

	return sb.String()
}
