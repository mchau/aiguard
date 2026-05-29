package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/mchau/aiguard/internal/audit"
	"github.com/mchau/aiguard/internal/checks"
	"github.com/mchau/aiguard/internal/config"
	"github.com/mchau/aiguard/internal/git"
	"github.com/mchau/aiguard/internal/project"
	"github.com/mchau/aiguard/internal/prompt"
)

var rationaleCmd = &cobra.Command{
	Use:   "rationale",
	Short: "Manage changed-file rationale",
}

var (
	rationaleFile        string
	rationaleAI          bool
	rationaleInteractive bool
)

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
	rationaleUpdateCmd.Flags().StringVar(&rationaleFile, "file", "", "path to YAML or JSON file mapping changed paths to rationale strings")
	rationaleUpdateCmd.Flags().BoolVar(&rationaleAI, "ai", false, "delegate rationale generation to the configured guidance agent")
	rationaleUpdateCmd.Flags().BoolVar(&rationaleInteractive, "interactive", false, "prompt the user for each file via stdin")
	rationaleCmd.AddCommand(rationaleUpdateCmd)
	rationaleCmd.AddCommand(rationaleCheckCmd)
	rootCmd.AddCommand(rationaleCmd)
}

func runRationaleUpdate(cmd *cobra.Command, args []string) error {
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
	changedFiles, err := gc.ChangedFiles(ctx, cfg.Project.SourceBranch)
	if err != nil {
		return fmt.Errorf("get changed files: %w", err)
	}

	rationalePath := project.ChangedFilesRationaleMD(root)
	existing := loadFileOr(rationalePath, "")

	var newEntries map[string]string
	switch {
	case rationaleFile != "":
		newEntries, err = loadRationaleFromFile(rationaleFile, changedFiles)
	case rationaleAI:
		newEntries, err = generateRationaleViaAgent(ctx, root, cfg, gc, changedFiles)
	case rationaleInteractive:
		newEntries, err = promptRationaleStdin(changedFiles, existing)
	default:
		return fmt.Errorf("pick a mode: --file <path>, --ai, or --interactive (no stdin prompting by default to keep the command scriptable)")
	}
	if err != nil {
		return err
	}

	var sb strings.Builder
	sb.WriteString("# Changed Files Rationale\n\n")
	for _, f := range changedFiles {
		if strings.Contains(existing, "## "+f) {
			fmt.Printf("  ✓ %s (already documented)\n", f)
			continue
		}
		reason, ok := newEntries[f]
		if !ok || strings.TrimSpace(reason) == "" {
			continue
		}
		sb.WriteString(fmt.Sprintf("## %s\n\n**Reason:** %s\n\n", f, reason))
	}

	if sb.Len() > len("# Changed Files Rationale\n\n") {
		combined := existing + "\n" + sb.String()
		if err := os.WriteFile(rationalePath, []byte(combined), 0o644); err != nil {
			return fmt.Errorf("write rationale: %w", err)
		}
		fmt.Printf("\nRationale updated: %s\n", rationalePath)
	} else {
		fmt.Println("No new rationale entries to add.")
	}

	_ = audit.Append(root, audit.Event{Command: "rationale update", Status: "success", Inputs: []string{}, Outputs: []string{rationalePath}})
	return nil
}

// loadRationaleFromFile parses YAML or JSON map[string]string keyed by changed path.
// Unknown paths in the file are ignored; missing paths from the diff just produce no entry.
func loadRationaleFromFile(path string, changedFiles []string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read rationale file: %w", err)
	}
	var raw map[string]string
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse rationale file (YAML or JSON map[string]string): %w", err)
	}
	changedSet := map[string]bool{}
	for _, f := range changedFiles {
		changedSet[f] = true
	}
	out := map[string]string{}
	for path, reason := range raw {
		if changedSet[path] {
			out[path] = reason
		}
	}
	return out, nil
}

// generateRationaleViaAgent invokes the configured guidance profile to produce
// a YAML rationale map covering the changed files. Malformed output → error,
// caller surfaces it without crashing.
func generateRationaleViaAgent(ctx context.Context, root string, cfg *config.Config, gc *git.Client, changedFiles []string) (map[string]string, error) {
	diffText, _ := gc.Diff(ctx, cfg.Project.SourceBranch)
	promptText, err := prompt.Render("guidance", map[string]string{
		"AcceptanceCriteria": loadFileOr(project.AcceptanceCriteriaJSON(root), "[]"),
		"SnapshotSummary":    loadFileOr(project.SnapshotMD(root), "*(no snapshot)*"),
	})
	if err != nil {
		return nil, err
	}
	promptText += "\n\nThe diff under review is:\n\n```\n" + diffText + "\n```\n\n" +
		"For each changed file below, provide a one-line rationale connecting it to an AC or plan step. " +
		"Output ONLY a YAML map of path → rationale string, no preamble:\n\n" +
		strings.Join(changedFiles, "\n")

	resp, err := runAgent(root, cfg, "tech_guidance", promptText, "")
	if err != nil {
		return nil, fmt.Errorf("agent: %w", err)
	}
	var out map[string]string
	if err := yaml.Unmarshal([]byte(resp.Stdout), &out); err != nil {
		return nil, fmt.Errorf("parse agent YAML: %w", err)
	}
	// Filter to changed files only — agents sometimes hallucinate extra paths
	changedSet := map[string]bool{}
	for _, f := range changedFiles {
		changedSet[f] = true
	}
	filtered := map[string]string{}
	for k, v := range out {
		if changedSet[k] {
			filtered[k] = v
		}
	}
	return filtered, nil
}

func promptRationaleStdin(changedFiles []string, existing string) (map[string]string, error) {
	scanner := bufio.NewScanner(os.Stdin)
	out := map[string]string{}
	for _, f := range changedFiles {
		if strings.Contains(existing, "## "+f) {
			continue
		}
		fmt.Printf("\nFile: %s\nReason (or Enter to skip): ", f)
		if !scanner.Scan() {
			break
		}
		if reason := strings.TrimSpace(scanner.Text()); reason != "" {
			out[f] = reason
		}
	}
	return out, nil
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
