package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/mchau/aiguard/internal/audit"
	"github.com/mchau/aiguard/internal/config"
	"github.com/mchau/aiguard/internal/project"
)

var initForce bool
var initSourceBranch string

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize an AIGuard workspace in the current project",
	RunE:  runInit,
}

func init() {
	initCmd.Flags().BoolVar(&initForce, "force", false, "overwrite existing .aiguard/ workspace")
	initCmd.Flags().StringVar(&initSourceBranch, "source-branch", "develop", "source-of-truth branch for this project")
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	aiguardDir := project.DotAiguard
	cfgPath := project.ConfigFile(cwd)

	// 1.24a: guard against re-init without --force
	if _, err := os.Stat(cfgPath); err == nil && !initForce {
		return fmt.Errorf(".aiguard/ already exists; use --force to reinitialize, or run 'aiguard status' to inspect the current workspace")
	}

	// 1.24b: create the full .aiguard/ subtree from SPEC.md §4
	dirs := []string{
		project.SnapshotDir(cwd),
		project.RequirementsDir(cwd),
		project.TestsDir(cwd),
		project.GuidanceDir(cwd),
		project.PlansDir(cwd),
		project.ImplementationDir(cwd),
		project.CheckpointsDir(cwd),
		project.ReviewsDir(cwd),
		project.LogsDir(cwd),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return fmt.Errorf("create directory %s: %w", d, err)
		}
	}

	// 1.24c: write config.yaml from defaults, applying --source-branch flag
	cfg := config.Default()
	if initSourceBranch != "" {
		cfg.Project.SourceBranch = initSourceBranch
		cfg.Project.DefaultCompareRef = initSourceBranch
	}
	if err := config.Write(cfgPath, cfg); err != nil {
		return err
	}

	// 1.26: write audit event
	_ = audit.Append(cwd, audit.Event{
		Command: "init",
		Inputs:  []string{"--source-branch", initSourceBranch},
		Outputs: []string{aiguardDir},
		Status:  "success",
	})

	fmt.Printf("Initialized AIGuard workspace at %s\n", cfgPath)
	fmt.Printf("Source branch: %s\n", cfg.Project.SourceBranch)
	fmt.Println("Next step: run 'aiguard snapshot update' to map the codebase.")
	return nil
}
