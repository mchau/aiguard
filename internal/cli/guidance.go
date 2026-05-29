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
	"github.com/mchau/aiguard/internal/requirements"
)

var guidanceCmd = &cobra.Command{
	Use:   "guidance",
	Short: "Manage technical guidance gates",
}

var guidanceForce bool
var guidanceForceReason string
var guidanceReviewer string
var guidanceApproveReason string

var guidanceCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Generate technical guidance from the approved digest",
	RunE:  runGuidanceCreate,
}

var guidanceReviewCmd = &cobra.Command{
	Use:   "review",
	Short: "Request AI review of technical guidance (single reviewer in M8)",
	RunE:  runGuidanceReview,
}

var guidanceApproveCmd = &cobra.Command{
	Use:   "approve",
	Short: "Approve the technical guidance",
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := requireRoot()
		if err != nil {
			return err
		}
		if guidanceApproveReason == "" {
			return fmt.Errorf("--reason required")
		}
		m, err := gates.Load(root)
		if err != nil {
			return err
		}
		if err := m.Set("guidance", gates.StatusApprovedByUser, guidanceApproveReason, "local-user"); err != nil {
			return err
		}
		_ = audit.Append(root, audit.Event{Command: "guidance approve", Status: "success", Inputs: []string{}, Outputs: []string{}})
		fmt.Println("Guidance gate APPROVED_BY_USER.")
		fmt.Println("Next step: run 'aiguard plan create'.")
		return nil
	},
}

func init() {
	guidanceCreateCmd.Flags().BoolVar(&guidanceForce, "force", false, "skip digest gate check")
	guidanceCreateCmd.Flags().StringVar(&guidanceForceReason, "reason", "", "reason for forcing (required with --force)")
	guidanceReviewCmd.Flags().StringVar(&guidanceReviewer, "reviewer", "claude_code", "agent for review")
	guidanceApproveCmd.Flags().StringVar(&guidanceApproveReason, "reason", "", "reason for approval (required)")
	guidanceCmd.AddCommand(guidanceCreateCmd)
	guidanceCmd.AddCommand(guidanceReviewCmd)
	guidanceCmd.AddCommand(guidanceApproveCmd)
	rootCmd.AddCommand(guidanceCmd)
}

func runGuidanceCreate(cmd *cobra.Command, args []string) error {
	root, err := requireRoot()
	if err != nil {
		return err
	}

	// Require digest approval (or --force --reason)
	if !guidanceForce {
		m, err := gates.Load(root)
		if err != nil {
			return err
		}
		if err := m.RequireApproved("digest"); err != nil {
			return err
		}
		if questions := loadQuestionsJSON(project.OpenQuestionsMD(root) + ".json"); requirements.HasBlockingOpenQuestions(questions) {
			return fmt.Errorf("there are open blocking clarification questions; run 'aiguard clarify' to resolve them (or use --force --reason)")
		}
	} else if guidanceForceReason == "" {
		return fmt.Errorf("--reason is required with --force")
	}

	acData, err := os.ReadFile(project.AcceptanceCriteriaJSON(root))
	if os.IsNotExist(err) {
		return fmt.Errorf("no acceptance-criteria.json; run 'aiguard digest create' first")
	}
	if err != nil {
		return err
	}

	cfg, err := config.Load(project.ConfigFile(root))
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	promptText, err := prompt.Render("guidance", map[string]string{
		"AcceptanceCriteria": string(acData),
		"SnapshotSummary":    loadFileOr(project.SnapshotMD(root), "*(no snapshot)*"),
	})
	if err != nil {
		return err
	}

	resp, runErr := runAgent(root, cfg, "tech_guidance", promptText, project.TechGuidanceMD(root)+".ai-raw.txt")
	if runErr != nil {
		fmt.Fprintf(os.Stderr, "WARN: AI guidance failed (%v); writing empty template.\n", runErr)
		resp = nil
	}

	guidance := "# Technical Guidance\n\n*(AI guidance pending)*\n"
	guidanceJSON := "{}"
	if resp != nil {
		guidance = "# Technical Guidance\n\n" + resp.Stdout
		guidanceJSON = extractJSON(resp.Stdout)
	}

	if err := os.WriteFile(project.TechGuidanceMD(root), []byte(guidance), 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(project.TechGuidanceJSON(root), []byte(guidanceJSON), 0o644); err != nil {
		return err
	}

	fmt.Printf("Guidance written to %s\n", project.GuidanceDir(root))
	_ = audit.Append(root, audit.Event{Command: "guidance create", Status: "success", Inputs: []string{}, Outputs: []string{project.TechGuidanceMD(root)}})
	return nil
}

func runGuidanceReview(cmd *cobra.Command, args []string) error {
	root, err := requireRoot()
	if err != nil {
		return err
	}

	m, err := gates.Load(root)
	if err != nil {
		return err
	}
	if err := m.Set("guidance", gates.StatusNeedsReview, "", "local-user"); err != nil {
		return err
	}
	fmt.Printf("Guidance gate set to NEEDS_REVIEW (reviewer: %s).\n", guidanceReviewer)
	return nil
}

func loadFileOr(path, fallback string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return fallback
	}
	return string(data)
}
