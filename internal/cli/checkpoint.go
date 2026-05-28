package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/mchau/aiguard/internal/audit"
	"github.com/mchau/aiguard/internal/checks"
	"github.com/mchau/aiguard/internal/config"
	"github.com/mchau/aiguard/internal/git"
	"github.com/mchau/aiguard/internal/project"
	"github.com/mchau/aiguard/internal/prompt"
	"github.com/mchau/aiguard/internal/report"
	"github.com/mchau/aiguard/internal/review"
)

var checkpointCmd = &cobra.Command{
	Use:   "checkpoint",
	Short: "Capture a mid-implementation checkpoint with deterministic checks and optional AI review",
	RunE:  runCheckpoint,
}

func init() {
	rootCmd.AddCommand(checkpointCmd)
}

func runCheckpoint(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
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
	base := cfg.Project.SourceBranch

	// Gather diff
	changedFiles, _ := gc.ChangedFiles(ctx, base)
	diffText, _ := gc.Diff(ctx, base)
	diffStat, _ := gc.DiffStat(ctx, base)

	// Run deterministic checks
	input := checks.TypedInput{
		Config:       cfg,
		ChangedFiles: changedFiles,
		DiffText:     diffText,
		PlanPath:     project.PlanJSON(root),
	}.ToCheckInput()

	results, err := checks.RunAll(ctx, input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "WARN: check error: %v\n", err)
	}

	verdict := review.Aggregate(results)

	// Find next checkpoint number
	n := nextCheckpointNumber(project.CheckpointsDir(root))
	cpDir := project.CheckpointsDir(root)
	if err := os.MkdirAll(cpDir, 0o755); err != nil {
		return err
	}
	baseName := fmt.Sprintf("checkpoint-%03d", n)

	// Build checkpoint prompt for AI
	planText := loadFileOr(project.PlanMD(root), "*(no plan)*")
	cpPrompt, _ := prompt.Render("checkpoint", map[string]string{
		"PlanSteps": planText,
		"DiffText":  diffText,
	})

	// Optionally call AI
	var aiSummary string
	if agentCfg, ok := cfg.Agents["claude_code"]; ok && agentCfg.Enabled {
		resp, runErr := runAgent(root, cfg, "checkpoint_review", cpPrompt, filepath.Join(cpDir, baseName+".ai-raw.txt"))
		if runErr == nil && resp != nil {
			aiSummary = resp.Stdout
		}
	}

	// Write checkpoint MD
	cpContent := buildCheckpointMD(n, verdict, results, diffStat, aiSummary)
	if err := os.WriteFile(filepath.Join(cpDir, baseName+".md"), []byte(cpContent), 0o644); err != nil {
		return err
	}

	// Write checkpoint JSON (final verdict style)
	v := &review.FinalVerdict{
		Verdict:             verdict,
		Summary:             fmt.Sprintf("Checkpoint %d: %s", n, verdict),
		DeterministicChecks: results,
	}
	if err := report.WriteFinalVerdictJSON(filepath.Join(cpDir, baseName+".json"), v); err != nil {
		return err
	}

	fmt.Printf("Checkpoint %d: %s\n", n, verdict)
	fmt.Printf("Report: %s\n", filepath.Join(cpDir, baseName+".md"))

	_ = audit.Append(root, audit.Event{Command: "checkpoint", Status: "success", Inputs: []string{}, Outputs: []string{filepath.Join(cpDir, baseName+".md")}})
	return nil
}

func nextCheckpointNumber(cpDir string) int {
	entries, err := os.ReadDir(cpDir)
	if err != nil {
		return 1
	}
	max := 0
	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, "checkpoint-") && strings.HasSuffix(name, ".md") {
			var n int
			fmt.Sscanf(name, "checkpoint-%d.md", &n)
			if n > max {
				max = n
			}
		}
	}
	return max + 1
}

func buildCheckpointMD(n int, verdict string, results []checks.DeterministicCheckResult, diffStat, aiSummary string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Checkpoint %d\n\n", n))
	sb.WriteString(fmt.Sprintf("**Verdict:** %s\n\n", verdict))
	sb.WriteString("## Diff Stat\n\n```\n")
	sb.WriteString(diffStat)
	sb.WriteString("\n```\n\n")
	if len(results) > 0 {
		sb.WriteString("## Deterministic Findings\n\n")
		for _, r := range results {
			sb.WriteString(fmt.Sprintf("- [%s] %s: %s\n", r.Severity, r.CheckID, r.Message))
		}
		sb.WriteString("\n")
	}
	if aiSummary != "" {
		sb.WriteString("## AI Review\n\n")
		sb.WriteString(aiSummary)
		sb.WriteString("\n")
	}
	return sb.String()
}
