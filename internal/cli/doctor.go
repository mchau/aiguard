package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/mchau/aiguard/internal/audit"
	"github.com/mchau/aiguard/internal/config"
	"github.com/mchau/aiguard/internal/git"
	"github.com/mchau/aiguard/internal/project"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check AIGuard prerequisites and environment health",
	RunE:  runDoctor,
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}

type checkResult struct {
	name     string
	ok       bool
	critical bool
	detail   string
}

func runDoctor(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	cwd, _ := os.Getwd()
	var results []checkResult
	failed := false

	// 1.28a: git --version succeeds
	results = append(results, checkGitVersion(ctx))

	// 1.28b: cwd is inside a git repo
	results = append(results, checkInsideRepo(cwd))

	// Load config for remaining checks (best-effort)
	root, _ := project.FindRoot(cwd, cfgPath)
	var cfg *config.Config
	if root != "" {
		var err error
		cfg, err = config.Load(project.ConfigFile(root))
		if err != nil {
			cfg = nil
		}
	}

	// 1.28c: source_branch ref resolves
	results = append(results, checkSourceBranch(ctx, root, cfg))

	// 1.28f: config loads and validates
	results = append(results, checkConfig(root, cfg))

	// 1.28g: .aiguard/ is writable
	results = append(results, checkAiguardWritable(root))

	// 1.28d/e: agent commands on PATH
	if cfg != nil {
		for name, agent := range cfg.Agents {
			results = append(results, checkAgentCommand(name, agent.Command))
		}
	} else {
		results = append(results, checkAgentCommand("claude_code", "claude"))
		results = append(results, checkAgentCommand("codex", "codex"))
	}

	// 1.28h: test command first tokens resolve
	if cfg != nil {
		for _, tc := range cfg.TestCommands {
			results = append(results, checkTestCommand(tc.Name, tc.Command))
		}
	}

	// Print results
	for _, r := range results {
		icon := "✓"
		if !r.ok {
			if r.critical {
				icon = "✗"
				failed = true
			} else {
				icon = "!"
			}
		}
		fmt.Printf("  %s  %s", icon, r.name)
		if r.detail != "" {
			fmt.Printf(": %s", r.detail)
		}
		fmt.Println()
	}

	// 1.26: audit
	status := "success"
	if failed {
		status = "failure"
	}
	if root != "" {
		_ = audit.Append(root, audit.Event{Command: "doctor", Status: status, Inputs: []string{}, Outputs: []string{}})
	}

	if failed {
		return fmt.Errorf("doctor: one or more critical checks failed")
	}
	return nil
}

func checkGitVersion(ctx context.Context) checkResult {
	out, err := exec.CommandContext(ctx, "git", "--version").Output()
	if err != nil {
		return checkResult{name: "git installed", ok: false, critical: true, detail: err.Error()}
	}
	return checkResult{name: "git installed", ok: true, detail: strings.TrimSpace(string(out))}
}

func checkInsideRepo(cwd string) checkResult {
	gc := git.New(cwd)
	if !gc.IsRepo(context.Background()) {
		return checkResult{name: "inside git repo", ok: false, critical: true, detail: "run 'git init' or change to a git repository"}
	}
	return checkResult{name: "inside git repo", ok: true}
}

func checkSourceBranch(ctx context.Context, root string, cfg *config.Config) checkResult {
	if cfg == nil || root == "" {
		return checkResult{name: "source_branch resolves", ok: false, critical: true, detail: "config not loaded"}
	}
	gc := git.New(root)
	sha, err := gc.ResolveRef(ctx, cfg.Project.SourceBranch)
	if err != nil {
		return checkResult{name: "source_branch resolves", ok: false, critical: true,
			detail: fmt.Sprintf("%q not found; run 'git fetch' or set project.source_branch in config", cfg.Project.SourceBranch)}
	}
	return checkResult{name: "source_branch resolves", ok: true, detail: fmt.Sprintf("%s → %s", cfg.Project.SourceBranch, sha[:min(8, len(sha))])}
}

func checkConfig(root string, cfg *config.Config) checkResult {
	if root == "" {
		return checkResult{name: "config loads", ok: false, critical: true, detail: "no .aiguard/ found; run 'aiguard init'"}
	}
	if cfg == nil {
		return checkResult{name: "config loads", ok: false, critical: true, detail: "failed to load config"}
	}
	if err := config.Validate(cfg); err != nil {
		return checkResult{name: "config validates", ok: false, critical: true, detail: err.Error()}
	}
	return checkResult{name: "config loads and validates", ok: true}
}

func checkAiguardWritable(root string) checkResult {
	if root == "" {
		return checkResult{name: ".aiguard/ writable", ok: false, critical: true, detail: "no .aiguard/ found"}
	}
	aiguardDir := project.SnapshotDir(root) // any subdir works
	if err := os.MkdirAll(aiguardDir, 0o755); err != nil {
		return checkResult{name: ".aiguard/ writable", ok: false, critical: true, detail: err.Error()}
	}
	return checkResult{name: ".aiguard/ writable", ok: true}
}

func checkAgentCommand(name, command string) checkResult {
	label := fmt.Sprintf("agent %s (%s)", name, command)
	if command == "" {
		return checkResult{name: label, ok: true, detail: "no command configured (optional)"}
	}
	if _, err := exec.LookPath(command); err != nil {
		return checkResult{name: label, ok: false, critical: false, detail: fmt.Sprintf("%q not on PATH; install or disable this agent in config", command)}
	}
	return checkResult{name: label, ok: true}
}

func checkTestCommand(name, command string) checkResult {
	label := fmt.Sprintf("test command %q", name)
	fields := strings.Fields(command)
	if len(fields) == 0 {
		return checkResult{name: label, ok: false, detail: "empty command"}
	}
	if _, err := exec.LookPath(fields[0]); err != nil {
		return checkResult{name: label, ok: false, critical: false, detail: fmt.Sprintf("%q not on PATH", fields[0])}
	}
	return checkResult{name: label, ok: true}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
