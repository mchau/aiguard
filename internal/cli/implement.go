package cli

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/mchau/aiguard/internal/agents"
	"github.com/mchau/aiguard/internal/audit"
	"github.com/mchau/aiguard/internal/config"
	"github.com/mchau/aiguard/internal/project"
)

var implementCmd = &cobra.Command{
	Use:   "implement",
	Short: "Invoke the configured AI agent for an implementation step",
	RunE:  runImplement,
}

var implementAgent string
var implementStep string

func init() {
	implementCmd.Flags().StringVar(&implementAgent, "agent", "claude_code", "agent to use for implementation")
	implementCmd.Flags().StringVar(&implementStep, "step", "", "plan step ID to implement (e.g. S1)")
	rootCmd.AddCommand(implementCmd)
}

func runImplement(cmd *cobra.Command, args []string) error {
	root, err := requireRoot()
	if err != nil {
		return err
	}

	cfg, err := config.Load(project.ConfigFile(root))
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Build prompt: load plan, AC, and focus on the specified step
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
`, acText, guidanceText, planText, stepOrAll(implementStep))

	profileName := cfg.StepRouting["implementation"]
	if profileName == "" {
		profileName = "coding_balanced"
	}
	profile, ok := cfg.ModelProfiles[profileName]
	if !ok {
		return fmt.Errorf("model profile %q not found", profileName)
	}

	runner, err := agents.Get(cfg, profile.Provider, profileName)
	if err != nil {
		return fmt.Errorf("get agent: %w", err)
	}

	artifactPath := project.ImplementationDir(root) + "/implement-transcript.md"
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.Agents[profile.Provider].TimeoutSeconds)*time.Second)
	defer cancel()

	resp, err := runner.Run(ctx, agents.AgentRequest{
		Prompt:       promptText,
		WorkDir:      root,
		ArtifactPath: artifactPath,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "WARN: agent returned error: %v\n", err)
	}

	if resp != nil && resp.Stdout != "" {
		fmt.Println(resp.Stdout)
	}
	fmt.Printf("Transcript saved to %s\n", artifactPath)

	_ = audit.Append(root, audit.Event{Command: "implement", Status: "success", Inputs: []string{"--step", implementStep}, Outputs: []string{artifactPath}})
	return nil
}

func stepOrAll(step string) string {
	if step == "" {
		return "Implement all plan steps."
	}
	return fmt.Sprintf("Implement plan step: %s", step)
}
