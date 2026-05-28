package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/mchau/aiguard/internal/audit"
	"github.com/mchau/aiguard/internal/checks"
	"github.com/mchau/aiguard/internal/config"
	"github.com/mchau/aiguard/internal/git"
	"github.com/mchau/aiguard/internal/project"
)

var rationaleCmd = &cobra.Command{
	Use:   "rationale",
	Short: "Manage changed-file rationale",
}

var rationaleUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Add or update rationale entries for each changed file",
	RunE:  runRationaleUpdate,
}

var rationaleCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Check that all changed files have rationale entries",
	RunE:  runRationaleCheck,
}

func init() {
	rationaleCmd.AddCommand(rationaleUpdateCmd)
	rationaleCmd.AddCommand(rationaleCheckCmd)
	rootCmd.AddCommand(rationaleCmd)
}

func runRationaleUpdate(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
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
	changedFiles, err := gc.ChangedFiles(ctx, cfg.Project.SourceBranch)
	if err != nil {
		return fmt.Errorf("get changed files: %w", err)
	}

	rationalePath := project.ChangedFilesRationaleMD(root)

	// Load existing content
	existing := loadFileOr(rationalePath, "")

	var sb strings.Builder
	sb.WriteString("# Changed Files Rationale\n\n")

	scanner := bufio.NewScanner(os.Stdin)
	for _, f := range changedFiles {
		if strings.Contains(existing, f) {
			// Already has an entry — skip
			fmt.Printf("  ✓ %s (already documented)\n", f)
			continue
		}

		fmt.Printf("\nFile: %s\n", f)
		fmt.Print("Reason (or Enter to skip): ")
		if !scanner.Scan() {
			break
		}
		reason := strings.TrimSpace(scanner.Text())
		if reason == "" {
			continue
		}

		sb.WriteString(fmt.Sprintf("## %s\n\n**Reason:** %s\n\n", f, reason))
	}

	// Append new entries to existing
	if sb.Len() > 0 {
		combined := existing + "\n" + sb.String()
		if err := os.WriteFile(rationalePath, []byte(combined), 0o644); err != nil {
			return fmt.Errorf("write rationale: %w", err)
		}
		fmt.Printf("\nRationale updated: %s\n", rationalePath)
	}

	_ = audit.Append(root, audit.Event{Command: "rationale update", Status: "success", Inputs: []string{}, Outputs: []string{rationalePath}})
	return nil
}

func runRationaleCheck(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
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
	changedFiles, err := gc.ChangedFiles(ctx, cfg.Project.SourceBranch)
	if err != nil {
		return fmt.Errorf("get changed files: %w", err)
	}

	checkInput := checks.TypedInput{
		Config:        cfg,
		ChangedFiles:  changedFiles,
		RationalePath: project.ChangedFilesRationaleMD(root),
	}.ToCheckInput()

	rc := &checks.RationaleCheck{}
	results, err := rc.Run(ctx, checkInput)
	if err != nil {
		return err
	}

	if len(results) == 0 {
		fmt.Println("All changed files have rationale entries.")
		return nil
	}

	for _, r := range results {
		fmt.Printf("  [%s] %s: %s\n", r.Severity, r.File, r.Message)
	}
	return fmt.Errorf("rationale check failed: %d file(s) missing rationale", len(results))
}
