package review

import (
	"context"
	"fmt"

	"github.com/mchau/aiguard/internal/checks"
	"github.com/mchau/aiguard/internal/config"
	"github.com/mchau/aiguard/internal/git"
)

// RunDeterministic gathers the diff from git and runs all registered deterministic checks.
func RunDeterministic(ctx context.Context, cfg *config.Config, gc *git.Client, base string) (*FinalVerdict, error) {
	changedFiles, err := gc.ChangedFiles(ctx, base)
	if err != nil {
		return nil, fmt.Errorf("get changed files: %w", err)
	}

	diffText, err := gc.Diff(ctx, base)
	if err != nil {
		return nil, fmt.Errorf("get diff: %w", err)
	}

	input := checks.TypedInput{
		Config:       cfg,
		ChangedFiles: changedFiles,
		DiffText:     diffText,
	}.ToCheckInput()

	results, err := checks.RunAll(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("run checks: %w", err)
	}

	verdict := Aggregate(results)

	summary := buildSummary(verdict, results)

	return &FinalVerdict{
		Verdict:             verdict,
		Summary:             summary,
		DeterministicChecks: results,
	}, nil
}

func buildSummary(verdict string, results []checks.DeterministicCheckResult) string {
	if len(results) == 0 {
		return "No issues found by deterministic checks."
	}
	counts := map[string]int{}
	for _, r := range results {
		counts[r.Severity]++
	}
	s := fmt.Sprintf("Verdict: %s. Deterministic checks: ", verdict)
	for sev, n := range counts {
		s += fmt.Sprintf("%d %s ", n, sev)
	}
	return s
}
