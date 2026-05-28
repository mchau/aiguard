package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/mchau/aiguard/internal/audit"
	"github.com/mchau/aiguard/internal/project"
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

var digestReviewCmd = &cobra.Command{
	Use:   "review",
	Short: "Request AI review of the digest (sets gate to NEEDS_REVIEW)",
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("digest review: gate system not yet implemented (M4)")
	},
}

var digestApproveCmd = &cobra.Command{
	Use:   "approve",
	Short: "Approve the digest (sets gate to APPROVED_BY_USER)",
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("digest approve: gate system not yet implemented (M4)")
	},
}

func init() {
	digestCreateCmd.Flags().BoolVar(&digestLocalOnly, "local-only", false, "skip AI digest; use only deterministic AC extraction")
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
		fmt.Println("AI digest pending (configure agent and re-run, or use --local-only).")
	}
	fmt.Printf("Artifacts written to %s\n", project.RequirementsDir(root))

	_ = audit.Append(root, audit.Event{Command: "digest create", Inputs: []string{rawMD}, Outputs: []string{project.DigestMD(root)}, Status: "success"})
	return nil
}
