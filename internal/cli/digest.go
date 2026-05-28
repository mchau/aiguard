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
	"github.com/mchau/aiguard/internal/requirements"
)

var digestCmd = &cobra.Command{
	Use:   "digest",
	Short: "Manage the requirement digest",
}

var digestLocalOnly bool

var digestCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Extract acceptance criteria from raw.md and build the digest",
	RunE:  runDigestCreate,
}

var digestReviewReason string
var digestApproveReason string

var digestReviewCmd = &cobra.Command{
	Use:   "review",
	Short: "Mark the digest as needing review (NEEDS_REVIEW)",
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := requireRoot()
		if err != nil {
			return err
		}
		m, err := gates.Load(root)
		if err != nil {
			return err
		}
		if err := m.Set("digest", gates.StatusNeedsReview, digestReviewReason, "local-user"); err != nil {
			return err
		}
		fmt.Println("Digest gate set to NEEDS_REVIEW.")
		return nil
	},
}

var digestApproveCmd = &cobra.Command{
	Use:   "approve",
	Short: "Approve the digest (APPROVED_BY_USER)",
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := requireRoot()
		if err != nil {
			return err
		}
		if digestApproveReason == "" {
			return fmt.Errorf("--reason is required; describe why you approve this digest")
		}
		m, err := gates.Load(root)
		if err != nil {
			return err
		}
		if err := m.Set("digest", gates.StatusApprovedByUser, digestApproveReason, "local-user"); err != nil {
			return err
		}
		_ = audit.Append(root, audit.Event{Command: "digest approve", Inputs: []string{"--reason", digestApproveReason}, Outputs: []string{}, Status: "success"})
		fmt.Println("Digest gate APPROVED_BY_USER.")
		fmt.Println("Next step: run 'aiguard guidance create'.")
		return nil
	},
}

func init() {
	digestCreateCmd.Flags().BoolVar(&digestLocalOnly, "local-only", false, "skip AI digest; use only deterministic AC extraction")
	digestReviewCmd.Flags().StringVar(&digestReviewReason, "reason", "", "reason for review request")
	digestApproveCmd.Flags().StringVar(&digestApproveReason, "reason", "", "reason for approval (required)")
	digestCmd.AddCommand(digestCreateCmd)
	digestCmd.AddCommand(digestReviewCmd)
	digestCmd.AddCommand(digestApproveCmd)
	rootCmd.AddCommand(digestCmd)
}

func runDigestCreate(cmd *cobra.Command, args []string) error {
	root, err := requireRoot()
	if err != nil {
		return err
	}

	rawMD := project.RawMD(root)
	data, err := os.ReadFile(rawMD)
	if os.IsNotExist(err) {
		return fmt.Errorf("no raw.md found; run 'aiguard req ingest' first")
	}
	if err != nil {
		return fmt.Errorf("read raw.md: %w", err)
	}

	// Deterministic AC extraction
	digest := requirements.BuildDigest(string(data), nil)

	// Render artifacts
	if err := os.MkdirAll(project.RequirementsDir(root), 0o755); err != nil {
		return err
	}
	if err := requirements.RenderDigest(
		root,
		digest,
		project.DigestMD(root),
		project.AcceptanceCriteriaJSON(root),
		project.OpenQuestionsMD(root),
		project.AssumptionsMD(root),
	); err != nil {
		return err
	}

	fmt.Printf("Extracted %d source acceptance criteria.\n", len(digest.NormalizedCriteria))

	if !digestLocalOnly {
		cfg, err := config.Load(project.ConfigFile(root))
		if err == nil {
			inferred, err := runAIDigest(root, cfg, string(data))
			if err != nil {
				// Malformed AI output → WARN, never crash
				fmt.Fprintf(os.Stderr, "WARN: AI digest failed (%v); using deterministic ACs only.\n", err)
			} else if len(inferred) > 0 {
				digest = requirements.BuildDigest(string(data), inferred)
				if err := requirements.RenderDigest(root, digest,
					project.DigestMD(root), project.AcceptanceCriteriaJSON(root),
					project.OpenQuestionsMD(root), project.AssumptionsMD(root)); err != nil {
					fmt.Fprintf(os.Stderr, "WARN: re-render with inferred ACs failed: %v\n", err)
				} else {
					fmt.Printf("AI added %d inferred criteria.\n", len(inferred))
				}
			}
		} else {
			fmt.Println("AI digest skipped (config unavailable); use --local-only to suppress this message.")
		}
	}

	fmt.Printf("Artifacts written to %s\n", project.RequirementsDir(root))

	_ = audit.Append(root, audit.Event{Command: "digest create", Inputs: []string{rawMD}, Outputs: []string{project.DigestMD(root)}, Status: "success"})
	return nil
}

// runAIDigest calls the configured AI agent to infer additional ACs.
// On malformed output, returns nil, error — caller emits WARN, never crashes.
func runAIDigest(root string, cfg *config.Config, rawContent string) ([]requirements.AcceptanceCriterion, error) {
	profileName, ok := cfg.StepRouting["requirement_digest"]
	if !ok {
		return nil, nil
	}
	agentName, ok2 := func() (string, bool) {
		p, exists := cfg.ModelProfiles[profileName]
		return p.Provider, exists
	}()
	if !ok2 {
		return nil, nil
	}

	runner, err := agents.Get(cfg, agentName, profileName)
	if err != nil {
		return nil, fmt.Errorf("get agent: %w", err)
	}

	promptText, err := prompt.Render("digest", map[string]string{"RawContent": rawContent})
	if err != nil {
		return nil, fmt.Errorf("render digest prompt: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.Agents[agentName].TimeoutSeconds)*time.Second)
	defer cancel()

	artifactPath := project.DigestMD(root) + ".ai-raw.txt"
	resp, err := runner.Run(ctx, agents.AgentRequest{
		Prompt:       promptText,
		WorkDir:      root,
		ArtifactPath: artifactPath,
	})
	if err != nil {
		return nil, fmt.Errorf("agent run: %w", err)
	}

	// Parse JSON from AI response
	var parsed struct {
		InferredCriteria []requirements.AcceptanceCriterion `json:"inferred_criteria"`
	}
	if err := json.Unmarshal([]byte(resp.Stdout), &parsed); err != nil {
		return nil, fmt.Errorf("parse AI output (raw saved to %s): %w", artifactPath, err)
	}
	return parsed.InferredCriteria, nil
}
