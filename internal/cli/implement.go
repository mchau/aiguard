package cli

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"github.com/mchau/aiguard/internal/audit"
	"github.com/mchau/aiguard/internal/project"
)

// 'prompt' assembles a prompt from the approved digest/guidance/plan
// so the developer can paste it into their AI tool of choice.
//
// AIGuard does not patch files itself: agent CLIs in `-p` mode return text,
// not edits. Misleadingly running an agent and dumping stdout was renamed
// to surface this honestly.
var promptCmd = &cobra.Command{
	Use:   "prompt",
	Short: "Assemble the implementation prompt for the configured AI agent",
	RunE:  runPromptAssemble,
}

var promptStep string
var promptCopy bool
var promptOut string

func init() {
	promptCmd.Flags().StringVar(&promptStep, "step", "", "plan step ID to focus on (e.g. S1); blank emits the full plan")
	promptCmd.Flags().BoolVar(&promptCopy, "copy", false, "copy the assembled prompt to the system clipboard")
	promptCmd.Flags().StringVar(&promptOut, "out", "", "write the assembled prompt to this file instead of stdout")
	rootCmd.AddCommand(promptCmd)
}

func runPromptAssemble(cmd *cobra.Command, args []string) error {
	root, err := requireRoot()
	if err != nil {
		return err
	}

	planText := loadFileOr(project.PlanMD(root), "*(no plan)*")
	acText := loadFileOr(project.AcceptanceCriteriaJSON(root), "[]")
	guidanceText := loadFileOr(project.TechGuidanceMD(root), "*(no guidance)*")

	promptText := fmt.Sprintf(`You are implementing a software feature. Follow the plan exactly.

## Acceptance Criteria
%s

## Technical Guidance
%s

## Plan
%s

## Current Task
%s

Implement the code changes needed. Do not modify files outside the plan scope.
`, acText, guidanceText, planText, focusOrAll(promptStep))

	switch {
	case promptOut != "":
		if err := os.WriteFile(promptOut, []byte(promptText), 0o644); err != nil {
			return fmt.Errorf("write prompt: %w", err)
		}
		fmt.Printf("Prompt written to %s\n", promptOut)
	case promptCopy:
		if err := copyToClipboard(promptText); err != nil {
			return fmt.Errorf("copy to clipboard: %w (use --out to write to a file instead)", err)
		}
		fmt.Println("Prompt copied to clipboard. Paste it into your AI tool.")
	default:
		fmt.Print(promptText)
	}

	_ = audit.Append(root, audit.Event{Command: "prompt", Status: "success", Inputs: []string{"--step", promptStep}, Outputs: []string{}})
	return nil
}

func focusOrAll(step string) string {
	if step == "" {
		return "Implement all plan steps."
	}
	return fmt.Sprintf("Implement plan step: %s", step)
}

func copyToClipboard(text string) error {
	var bin string
	switch runtime.GOOS {
	case "darwin":
		bin = "pbcopy"
	case "linux":
		bin = "xclip"
	default:
		return fmt.Errorf("clipboard not supported on %s", runtime.GOOS)
	}
	cmd := exec.Command(bin)
	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}
