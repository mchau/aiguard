package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/mchau/aiguard/internal/project"
	"github.com/mchau/aiguard/internal/report"
	"github.com/mchau/aiguard/internal/review"
)

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Re-render the latest final verdict markdown report",
	RunE:  runReport,
}

func init() {
	rootCmd.AddCommand(reportCmd)
}

func runReport(cmd *cobra.Command, args []string) error {
	root, err := requireRoot()
	if err != nil {
		return err
	}

	verdictPath := project.FinalVerdictJSON(root)
	data, err := os.ReadFile(verdictPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("no final-verdict.json found; run 'aiguard verify' first")
	}
	if err != nil {
		return fmt.Errorf("read verdict JSON: %w", err)
	}

	var v review.FinalVerdict
	if err := json.Unmarshal(data, &v); err != nil {
		return fmt.Errorf("parse verdict JSON: %w", err)
	}

	mdContent := report.RenderFinalVerdict(&v)
	mdPath := project.FinalVerdictMD(root)
	if err := os.WriteFile(mdPath, []byte(mdContent), 0o644); err != nil {
		return fmt.Errorf("write report: %w", err)
	}

	fmt.Println(mdContent)
	fmt.Printf("\nReport written to %s\n", mdPath)
	return nil
}
