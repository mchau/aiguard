package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/mchau/aiguard/internal/audit"
	"github.com/mchau/aiguard/internal/config"
	"github.com/mchau/aiguard/internal/git"
	"github.com/mchau/aiguard/internal/project"
	"github.com/mchau/aiguard/internal/snapshot"
)

var snapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "Manage the codebase snapshot (living map of the repo)",
}

var snapshotFull bool
var snapshotNoAI bool

var snapshotUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update the codebase snapshot",
	RunE:  runSnapshotUpdate,
}

var snapshotStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show snapshot metadata",
	RunE:  runSnapshotStatus,
}

func init() {
	snapshotUpdateCmd.Flags().BoolVar(&snapshotFull, "full", false, "force full snapshot (default when no snapshot exists)")
	snapshotUpdateCmd.Flags().BoolVar(&snapshotNoAI, "no-ai", false, "skip optional AI summarization step")
	snapshotCmd.AddCommand(snapshotUpdateCmd)
	snapshotCmd.AddCommand(snapshotStatusCmd)
	rootCmd.AddCommand(snapshotCmd)
}

func runSnapshotUpdate(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	root, err := requireRoot()
	if err != nil {
		return err
	}

	cfg, err := config.Load(project.ConfigFile(root))
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	gc := git.New(root)

	if err := snapshot.FullUpdate(ctx, root, cfg, gc); err != nil {
		return fmt.Errorf("snapshot update: %w", err)
	}

	meta, _ := snapshot.LoadMeta(project.SnapshotMetaJSON(root))
	if meta != nil {
		fmt.Printf("Snapshot updated: version %d, %d files, commit %s\n",
			meta.SnapshotVersion, meta.TotalFiles, shortSHA(meta.SourceCommit))
	}

	_ = audit.Append(root, audit.Event{Command: "snapshot update", Inputs: []string{}, Outputs: []string{project.SnapshotMD(root)}, Status: "success"})
	return nil
}

func runSnapshotStatus(cmd *cobra.Command, args []string) error {
	root, err := requireRoot()
	if err != nil {
		return err
	}

	meta, err := snapshot.LoadMeta(project.SnapshotMetaJSON(root))
	if err != nil {
		fmt.Println("No snapshot found. Run 'aiguard snapshot update --full'.")
		return nil
	}

	fmt.Printf("Snapshot version: %d\n", meta.SnapshotVersion)
	fmt.Printf("Source branch:    %s\n", meta.SourceBranch)
	fmt.Printf("Source commit:    %s\n", meta.SourceCommit)
	fmt.Printf("Updated at:       %s\n", meta.UpdatedAt.Format(time.RFC3339))
	fmt.Printf("Total files:      %d\n", meta.TotalFiles)
	return nil
}

func shortSHA(sha string) string {
	if len(sha) > 8 {
		return sha[:8]
	}
	return sha
}
