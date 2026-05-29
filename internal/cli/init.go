package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

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
	initCmd.Flags().StringVar(&initSourceBranch, "source-branch", "", "source-of-truth branch (autodetected from origin/HEAD, falls back to 'main')")
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

	branch := initSourceBranch
	if branch == "" {
		branch = detectDefaultBranch(cwd)
	}

	cfg := config.Default()
	cfg.Project.SourceBranch = branch
	cfg.Project.DefaultCompareRef = branch
	if err := config.Write(cfgPath, cfg); err != nil {
		return err
	}

	// Transient artifacts (audit log, reviewer transcripts, checkpoint reports)
	// would otherwise show up in every diff and trigger the rationale check.
	// Persistent gate/digest/plan artifacts are intentionally kept tracked.
	gitignorePath := cwd + "/" + project.DotAiguard + "/.gitignore"
	gitignoreBody := "logs/\nreviews/\ncheckpoints/\n*.ai-raw.txt\n"
	if err := os.WriteFile(gitignorePath, []byte(gitignoreBody), 0o644); err != nil {
		return fmt.Errorf("write .aiguard/.gitignore: %w", err)
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

// detectDefaultBranch reads origin/HEAD to find the remote's default branch,
// then falls back to the first of {main, master} that exists locally,
// and finally to "main".
func detectDefaultBranch(dir string) string {
	cmd := exec.Command("git", "symbolic-ref", "--short", "refs/remotes/origin/HEAD")
	cmd.Dir = dir
	if out, err := cmd.Output(); err == nil {
		ref := strings.TrimSpace(string(out))
		if name := strings.TrimPrefix(ref, "origin/"); name != "" {
			return name
		}
	}
	for _, candidate := range []string{"main", "master"} {
		c := exec.Command("git", "show-ref", "--verify", "--quiet", "refs/heads/"+candidate)
		c.Dir = dir
		if c.Run() == nil {
			return candidate
		}
	}
	return "main"
}
