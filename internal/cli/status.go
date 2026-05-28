package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/mchau/aiguard/internal/gates"
	"github.com/mchau/aiguard/internal/project"
	"github.com/mchau/aiguard/internal/review"
	"github.com/mchau/aiguard/internal/snapshot"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show gate states, last verdict, and snapshot freshness",
	RunE:  runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	root, err := requireRoot()
	if err != nil {
		return err
	}

	fmt.Print("## Gate States\n\n")
	m, err := gates.Load(root)
	if err != nil {
		fmt.Printf("  (could not load gates: %v)\n", err)
	} else {
		for _, name := range []string{"digest", "tests", "guidance", "plan"} {
			g := m.Get(name)
			status := g.Status
			if status == "" {
				status = gates.StatusNotStarted
			}
			fmt.Printf("  %-10s %s\n", name, status)
		}
	}

	fmt.Print("\n## Last Verdict\n\n")
	verdictPath := project.FinalVerdictJSON(root)
	if data, err := os.ReadFile(verdictPath); err == nil {
		var v review.FinalVerdict
		if err := json.Unmarshal(data, &v); err == nil {
			fmt.Printf("  Verdict: %s\n", v.Verdict)
			fmt.Printf("  Summary: %s\n", v.Summary)
		}
	} else {
		fmt.Println("  No verdict yet. Run 'aiguard verify'.")
	}

	fmt.Print("\n## Snapshot\n\n")
	meta, err := snapshot.LoadMeta(project.SnapshotMetaJSON(root))
	if err != nil {
		fmt.Println("  No snapshot. Run 'aiguard snapshot update --full'.")
	} else {
		age := time.Since(meta.UpdatedAt).Round(time.Minute)
		fmt.Printf("  Version:  %d\n", meta.SnapshotVersion)
		fmt.Printf("  Branch:   %s @ %s\n", meta.SourceBranch, shortSHA(meta.SourceCommit))
		fmt.Printf("  Updated:  %s ago\n", age)
		fmt.Printf("  Files:    %d\n", meta.TotalFiles)
	}
	return nil
}
