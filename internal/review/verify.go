package review

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/mchau/aiguard/internal/checks"
	"github.com/mchau/aiguard/internal/config"
	"github.com/mchau/aiguard/internal/git"
	"github.com/mchau/aiguard/internal/project"
)

// RunDeterministic gathers the diff from git and runs all registered deterministic checks.
// rootDir is the project root (containing .aiguard/); pass "" to skip path-dependent checks.
func RunDeterministic(ctx context.Context, cfg *config.Config, gc *git.Client, base, rootDir string) (*FinalVerdict, error) {
	changedFiles, err := gc.ChangedFiles(ctx, base)
	if err != nil {
		return nil, fmt.Errorf("get changed files: %w", err)
	}

	diffText, err := gc.Diff(ctx, base)
	if err != nil {
		return nil, fmt.Errorf("get diff: %w", err)
	}

	input := checks.TypedInput{
		Config:        cfg,
		ChangedFiles:  changedFiles,
		DiffText:      diffText,
		RootDir:       rootDir,
		PlanPath:      project.PlanJSON(rootDir),
		RationalePath: project.ChangedFilesRationaleMD(rootDir),
	}.ToCheckInput()

	results, err := checks.RunAll(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("run checks: %w", err)
	}

	verdict := Aggregate(results)

	return &FinalVerdict{
		Verdict:             verdict,
		Summary:             buildSummary(verdict, results),
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
	// Deterministic order: BLOCKED, High, Critical, WARN, Medium, Info, then any others
	order := []string{
		checks.SeverityBlocked, checks.SeverityHigh, checks.SeverityCritical,
		checks.SeverityWarn, checks.SeverityMedium, checks.SeverityInfo,
	}
	seen := map[string]bool{}
	var parts []string
	for _, sev := range order {
		if n, ok := counts[sev]; ok {
			parts = append(parts, fmt.Sprintf("%d %s", n, sev))
			seen[sev] = true
		}
	}
	// Append any unexpected severity values in sorted order
	var extra []string
	for sev := range counts {
		if !seen[sev] {
			extra = append(extra, sev)
		}
	}
	sort.Strings(extra)
	for _, sev := range extra {
		parts = append(parts, fmt.Sprintf("%d %s", counts[sev], sev))
	}
	return fmt.Sprintf("Verdict: %s. Deterministic checks: %s", verdict, strings.Join(parts, ", "))
}
