package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/mchau/aiguard/internal/agents"
	"github.com/mchau/aiguard/internal/audit"
	"github.com/mchau/aiguard/internal/config"
	"github.com/mchau/aiguard/internal/gates"
	"github.com/mchau/aiguard/internal/project"
	"github.com/mchau/aiguard/internal/prompt"
)

var testsCmd = &cobra.Command{
	Use:   "tests",
	Short: "Manage test strategy gates",
}

var testsStrategyLocalOnly bool
var testsReviewReason string
var testsApproveReason string
var testsReviewer string

var testsStrategyCmd = &cobra.Command{
	Use:   "strategy",
	Short: "Generate test strategy from acceptance criteria",
	RunE:  runTestsStrategy,
}

var testsReviewCmd = &cobra.Command{
	Use:   "review",
	Short: "Request AI review of test strategy (single reviewer in M8)",
	RunE:  runTestsReview,
}

var testsApproveCmd = &cobra.Command{
	Use:   "approve",
	Short: "Approve the test strategy",
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := requireRoot()
		if err != nil {
			return err
		}
		if testsApproveReason == "" {
			return fmt.Errorf("--reason required")
		}
		m, err := gates.Load(root)
		if err != nil {
			return err
		}
		if err := m.Set("tests", gates.StatusApprovedByUser, testsApproveReason, "local-user"); err != nil {
			return err
		}
		_ = audit.Append(root, audit.Event{Command: "tests approve", Status: "success", Inputs: []string{}, Outputs: []string{}})
		fmt.Println("Tests gate APPROVED_BY_USER.")
		return nil
	},
}

func init() {
	testsStrategyCmd.Flags().BoolVar(&testsStrategyLocalOnly, "local-only", false, "skip AI, output template only")
	testsReviewCmd.Flags().StringVar(&testsReviewReason, "reason", "", "reason for review")
	testsReviewCmd.Flags().StringVar(&testsReviewer, "reviewer", "claude_code", "agent to use for review")
	testsApproveCmd.Flags().StringVar(&testsApproveReason, "reason", "", "reason for approval (required)")
	testsCmd.AddCommand(testsStrategyCmd)
	testsCmd.AddCommand(testsReviewCmd)
	testsCmd.AddCommand(testsApproveCmd)
	rootCmd.AddCommand(testsCmd)
}

func runTestsStrategy(cmd *cobra.Command, args []string) error {
	root, err := requireRoot()
	if err != nil {
		return err
	}

	acData, err := os.ReadFile(project.AcceptanceCriteriaJSON(root))
	if os.IsNotExist(err) {
		return fmt.Errorf("no acceptance-criteria.json found; run 'aiguard digest create' first")
	}
	if err != nil {
		return fmt.Errorf("read acceptance-criteria.json: %w", err)
	}

	if testsStrategyLocalOnly {
		if err := os.WriteFile(project.TestStrategyMD(root), []byte("# Test Strategy\n\n*(AI strategy pending)*\n"), 0o644); err != nil {
			return err
		}
		if err := os.WriteFile(project.TestRequirementsJSON(root), []byte("[]"), 0o644); err != nil {
			return err
		}
		fmt.Println("Test strategy template written (--local-only).")
		return nil
	}

	cfg, err := config.Load(project.ConfigFile(root))
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	promptText, err := prompt.Render("test_strategy", map[string]string{"AcceptanceCriteria": string(acData)})
	if err != nil {
		return err
	}

	resp, runErr := runAgent(root, cfg, "test_strategy", promptText, project.TestStrategyMD(root)+".ai-raw.txt")
	if runErr != nil {
		fmt.Fprintf(os.Stderr, "WARN: AI test strategy failed (%v); writing empty template.\n", runErr)
		resp = &agents.AgentResponse{Stdout: ""}
	}

	if err := os.WriteFile(project.TestStrategyMD(root), []byte("# Test Strategy\n\n"+resp.Stdout), 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(project.TestRequirementsJSON(root), []byte(extractJSON(resp.Stdout)), 0o644); err != nil {
		return err
	}

	fmt.Printf("Test strategy written to %s\n", project.TestStrategyMD(root))
	_ = audit.Append(root, audit.Event{Command: "tests strategy", Status: "success", Inputs: []string{}, Outputs: []string{project.TestStrategyMD(root)}})
	return nil
}

func runTestsReview(cmd *cobra.Command, args []string) error {
	root, err := requireRoot()
	if err != nil {
		return err
	}

	m, err := gates.Load(root)
	if err != nil {
		return err
	}
	if err := m.Set("tests", gates.StatusNeedsReview, testsReviewReason, "local-user"); err != nil {
		return err
	}
	fmt.Printf("Tests gate set to NEEDS_REVIEW (reviewer: %s).\n", testsReviewer)
	return nil
}

// runAgent invokes the AI agent for a given step and returns the response.
func runAgent(root string, cfg *config.Config, step, promptText, artifactPath string) (*agents.AgentResponse, error) {
	profileName, ok := cfg.StepRouting[step]
	if !ok {
		return nil, fmt.Errorf("no step_routing entry for %q", step)
	}
	profile, ok := cfg.ModelProfiles[profileName]
	if !ok {
		return nil, fmt.Errorf("model profile %q not found", profileName)
	}
	runner, err := agents.Get(cfg, profile.Provider, profileName)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.Agents[profile.Provider].TimeoutSeconds)*time.Second)
	defer cancel()

	return runner.Run(ctx, agents.AgentRequest{
		Prompt:       promptText,
		WorkDir:      root,
		ArtifactPath: artifactPath,
	})
}

// extractJSON tries to find a JSON object in text; returns "{}" on failure.
func extractJSON(text string) string {
	var m json.RawMessage
	if err := json.Unmarshal([]byte(text), &m); err == nil {
		return text
	}
	return "{}"
}
