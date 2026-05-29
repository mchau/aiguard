package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/mchau/aiguard/internal/audit"
	"github.com/mchau/aiguard/internal/config"
	"github.com/mchau/aiguard/internal/git"
	"github.com/mchau/aiguard/internal/project"
	"github.com/mchau/aiguard/internal/report"
	"github.com/mchau/aiguard/internal/review"
)

var verifyBase string
var verifyLocalOnly bool
var verifyReviewers string
var verifyAutoFix bool
var verifyCopyFix bool

var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Run deterministic checks and (optionally) AI reviewers against the current diff",
	RunE:  runVerify,
}

func init() {
	verifyCmd.Flags().StringVar(&verifyBase, "base", "", "base ref to diff against (defaults to config.project.source_branch)")
	verifyCmd.Flags().BoolVar(&verifyLocalOnly, "local-only", false, "run deterministic checks only; skip AI reviewers")
	verifyCmd.Flags().StringVar(&verifyReviewers, "reviewers", "", "comma-separated list of AI reviewers (e.g. claude_code,codex)")
	verifyCmd.Flags().BoolVar(&verifyAutoFix, "auto-fix", false, "on REJECT/WARN, feed the recommended fix prompt to the configured implementation agent and save the transcript")
	verifyCmd.Flags().BoolVar(&verifyCopyFix, "copy-fix", false, "copy the recommended fix prompt to the clipboard instead of running an agent")
	rootCmd.AddCommand(verifyCmd)
}

func runVerify(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	root, err := requireRoot()
	if err != nil {
		return err
	}

	cfg, err := config.Load(project.ConfigFile(root))
	if err != nil {
		return fmt.Errorf("load config: %w\nRun 'aiguard doctor' to diagnose configuration issues.", err)
	}

	base := verifyBase
	if base == "" {
		base = cfg.Project.SourceBranch
	}

	gc := git.New(root)

	if err := os.MkdirAll(project.ReviewsDir(root), 0o755); err != nil {
		return fmt.Errorf("create reviews dir: %w", err)
	}

	var verdict *review.FinalVerdict

	if verifyLocalOnly || verifyReviewers == "" {
		verdict, err = review.RunDeterministic(ctx, cfg, gc, base, root)
		if err != nil {
			return fmt.Errorf("run deterministic checks: %w", err)
		}
	} else {
		reviewers := strings.Split(verifyReviewers, ",")
		verdict, err = review.RunMultiVerify(ctx, cfg, gc, review.MultiVerifyOptions{
			Base:      base,
			Reviewers: reviewers,
			LocalOnly: verifyLocalOnly,
			RootDir:   root,
		})
		if err != nil {
			return fmt.Errorf("run verify: %w", err)
		}

		if len(verdict.Disagreements) > 0 {
			disReport := review.RenderDisagreementReport(verdict.Disagreements)
			_ = os.WriteFile(project.DisagreementReportMD(root), []byte(disReport), 0o644)
		}
	}

	if verdict.RecommendedFixPrompt == "" {
		verdict.RecommendedFixPrompt = review.GenerateFixPrompt(verdict)
	}

	mdPath := project.FinalVerdictMD(root)
	mdContent := report.RenderFinalVerdict(verdict)
	if err := os.WriteFile(mdPath, []byte(mdContent), 0o644); err != nil {
		return fmt.Errorf("write verdict markdown: %w", err)
	}

	jsonPath := project.FinalVerdictJSON(root)
	if err := report.WriteFinalVerdictJSON(jsonPath, verdict); err != nil {
		return err
	}

	fmt.Printf("\nVerdict: %s\n", verdict.Verdict)
	fmt.Printf("Summary: %s\n", verdict.Summary)
	if verdict.RecommendedFixPrompt != "" {
		fmt.Printf("\nRecommended Fix:\n%s\n", verdict.RecommendedFixPrompt)
	}
	fmt.Printf("\nReports written to:\n  %s\n  %s\n", mdPath, jsonPath)

	if err := maybeAutoFix(ctx, root, cfg, verdict); err != nil {
		fmt.Fprintf(os.Stderr, "WARN: auto-fix step: %v\n", err)
	}

	status := "success"
	if verdict.Verdict == review.VerdictBlocked || verdict.Verdict == review.VerdictReject {
		status = "failure"
	}
	_ = audit.Append(root, audit.Event{
		Command: "verify",
		Inputs:  []string{"--base", base},
		Outputs: []string{mdPath, jsonPath},
		Status:  status,
	})

	if verdict.Verdict == review.VerdictBlocked || verdict.Verdict == review.VerdictReject {
		return fmt.Errorf("verify failed: %s — see %s for details", verdict.Verdict, mdPath)
	}
	return nil
}

// maybeAutoFix closes the loop on REJECT/WARN by handing the recommended fix
// prompt to either the clipboard (--copy-fix) or the configured implementation
// agent (--auto-fix). BLOCKED is never auto-fixed: forbidden paths and secrets
// must be addressed manually.
func maybeAutoFix(ctx context.Context, root string, cfg *config.Config, v *review.FinalVerdict) error {
	if !verifyAutoFix && !verifyCopyFix {
		return nil
	}
	if v.RecommendedFixPrompt == "" {
		return nil
	}
	if v.Verdict == review.VerdictBlocked {
		fmt.Println("\nauto-fix skipped: BLOCKED verdicts require manual remediation.")
		return nil
	}
	if v.Verdict == review.VerdictApprove {
		return nil
	}

	if verifyCopyFix {
		if err := copyToClipboard(v.RecommendedFixPrompt); err != nil {
			return err
		}
		fmt.Println("\nFix prompt copied to clipboard.")
		return nil
	}

	// --auto-fix: invoke implementation profile agent in headless mode.
	// Agents in -p mode return text suggestions, not patches — the transcript
	// is what AIGuard can deterministically produce. Apply with intent.
	artifactPath := filepath.Join(project.ReviewsDir(root), "auto-fix-transcript.md")
	resp, err := runAgent(root, cfg, "implementation", v.RecommendedFixPrompt, artifactPath)
	if err != nil {
		return fmt.Errorf("agent run: %w", err)
	}
	if resp != nil && resp.Stdout != "" {
		fmt.Printf("\nAuto-fix transcript saved to %s\n", artifactPath)
	}
	_ = ctx // headless runAgent owns its own timeout via config
	return nil
}
