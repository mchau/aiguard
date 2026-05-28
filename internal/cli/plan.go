package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/mchau/aiguard/internal/audit"
	"github.com/mchau/aiguard/internal/config"
	"github.com/mchau/aiguard/internal/gates"
	"github.com/mchau/aiguard/internal/project"
	"github.com/mchau/aiguard/internal/prompt"
)

var planCmd = &cobra.Command{
	Use:   "plan",
	Short: "Manage implementation plan gates",
}

var planForce bool
var planForceReason string
var planReviewer string
var planApproveReason string

var planCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Generate implementation plan from approved guidance",
	RunE:  runPlanCreate,
}

var planReviewCmd = &cobra.Command{
	Use:   "review",
	Short: "Request AI review of the plan (single reviewer in M8)",
	RunE:  runPlanReview,
}

var planApproveCmd = &cobra.Command{
	Use:   "approve",
	Short: "Approve the implementation plan",
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := requireRoot()
		if err != nil {
			return err
		}
		if planApproveReason == "" {
			return fmt.Errorf("--reason required")
		}
		m, err := gates.Load(root)
		if err != nil {
			return err
		}
		if err := m.Set("plan", gates.StatusApprovedByUser, planApproveReason, "local-user"); err != nil {
			return err
		}
		_ = audit.Append(root, audit.Event{Command: "plan approve", Status: "success", Inputs: []string{}, Outputs: []string{}})
		fmt.Println("Plan gate APPROVED_BY_USER.")
		fmt.Println("Next step: run 'aiguard implement' or 'aiguard checkpoint'.")
		return nil
	},
}

func init() {
	planCreateCmd.Flags().BoolVar(&planForce, "force", false, "skip guidance gate check")
	planCreateCmd.Flags().StringVar(&planForceReason, "reason", "", "reason for forcing (required with --force)")
	planReviewCmd.Flags().StringVar(&planReviewer, "reviewer", "claude_code", "agent for review")
	planApproveCmd.Flags().StringVar(&planApproveReason, "reason", "", "reason for approval (required)")
	planCmd.AddCommand(planCreateCmd)
	planCmd.AddCommand(planReviewCmd)
	planCmd.AddCommand(planApproveCmd)
	rootCmd.AddCommand(planCmd)
}

func runPlanCreate(cmd *cobra.Command, args []string) error {
	root, err := requireRoot()
	if err != nil {
		return err
	}

	// Require guidance approval (or --force --reason)
	if !planForce {
		m, err := gates.Load(root)
		if err != nil {
			return err
		}
		if err := m.RequireApproved("guidance"); err != nil {
			return err
		}
	} else if planForceReason == "" {
		return fmt.Errorf("--reason is required with --force")
	}

	acData, _ := os.ReadFile(project.AcceptanceCriteriaJSON(root))
	guidanceData, _ := os.ReadFile(project.TechGuidanceMD(root))

	cfg, err := config.Load(project.ConfigFile(root))
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	promptText, err := prompt.Render("plan", map[string]string{
		"AcceptanceCriteria": string(acData),
		"GuidanceSummary":    string(guidanceData),
	})
	if err != nil {
		return err
	}

	resp, runErr := runAgent(root, cfg, "plan_create", promptText, project.PlanMD(root)+".ai-raw.txt")
	planMD := "# Implementation Plan\n\n*(AI plan pending)*\n"
	planJSON := "{}"
	if runErr != nil {
		fmt.Fprintf(os.Stderr, "WARN: AI plan failed (%v); writing empty template.\n", runErr)
	} else {
		planMD = "# Implementation Plan\n\n" + resp.Stdout
		planJSON = extractJSON(resp.Stdout)
	}

	if err := os.WriteFile(project.PlanMD(root), []byte(planMD), 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(project.PlanJSON(root), []byte(planJSON), 0o644); err != nil {
		return err
	}

	fmt.Printf("Plan written to %s\n", project.PlansDir(root))
	_ = audit.Append(root, audit.Event{Command: "plan create", Status: "success", Inputs: []string{}, Outputs: []string{project.PlanMD(root)}})
	return nil
}

func runPlanReview(cmd *cobra.Command, args []string) error {
	root, err := requireRoot()
	if err != nil {
		return err
	}

	m, err := gates.Load(root)
	if err != nil {
		return err
	}
	if err := m.Set("plan", gates.StatusNeedsReview, "", "local-user"); err != nil {
		return err
	}
	fmt.Printf("Plan gate set to NEEDS_REVIEW (reviewer: %s).\n", planReviewer)
	return nil
}
