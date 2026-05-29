package review

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/mchau/aiguard/internal/agents"
	"github.com/mchau/aiguard/internal/checks"
	"github.com/mchau/aiguard/internal/config"
	"github.com/mchau/aiguard/internal/contextpack"
	"github.com/mchau/aiguard/internal/git"
	"github.com/mchau/aiguard/internal/project"
	"github.com/mchau/aiguard/internal/prompt"
)

// MultiVerifyOptions controls the full verify pipeline.
type MultiVerifyOptions struct {
	Base      string
	Reviewers []string // agent names from config.agents
	LocalOnly bool
	RootDir   string
}

// RunMultiVerify executes the full 7-stage verification pipeline.
func RunMultiVerify(ctx context.Context, cfg *config.Config, gc *git.Client, opts MultiVerifyOptions) (*FinalVerdict, error) {
	rootDir := opts.RootDir

	// Stage 1: gather inputs
	changedFiles, err := gc.ChangedFiles(ctx, opts.Base)
	if err != nil {
		return nil, fmt.Errorf("stage 1 get changed files: %w", err)
	}
	diffText, err := gc.Diff(ctx, opts.Base)
	if err != nil {
		return nil, fmt.Errorf("stage 1 get diff: %w", err)
	}
	diffStat, _ := gc.DiffStat(ctx, opts.Base)

	// Stage 2: deterministic checks
	input := checks.TypedInput{
		Config:        cfg,
		ChangedFiles:  changedFiles,
		DiffText:      diffText,
		RootDir:       rootDir,
		PlanPath:      project.PlanJSON(rootDir),
		RationalePath: project.ChangedFilesRationaleMD(rootDir),
	}.ToCheckInput()

	detResults, err := checks.RunAll(ctx, input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "WARN: check error: %v\n", err)
	}

	// Short-circuit: if BLOCKED by deterministic checks, skip AI
	if Aggregate(detResults) == VerdictBlocked || opts.LocalOnly {
		verdict := Aggregate(detResults)
		v := &FinalVerdict{
			Verdict:             verdict,
			Summary:             buildSummary(verdict, detResults),
			DeterministicChecks: detResults,
		}
		v.RecommendedFixPrompt = GenerateFixPrompt(v)
		return v, nil
	}

	// Stage 3: build context pack
	acData, _ := os.ReadFile(project.AcceptanceCriteriaJSON(rootDir))
	planData, _ := os.ReadFile(project.PlanMD(rootDir))
	rationaleData, _ := os.ReadFile(project.ChangedFilesRationaleMD(rootDir))

	pack := contextpack.Build(cfg, contextpack.BuildOptions{
		Step:         "final_review",
		RootDir:      rootDir,
		DiffText:     diffText,
		DiffStat:     diffStat,
		ChangedFiles: changedFiles,
		Artifacts: []contextpack.ContextArtifact{
			{Name: "acceptance_criteria", Content: string(acData)},
			{Name: "plan", Content: string(planData)},
			{Name: "rationale", Content: string(rationaleData)},
		},
	})

	// Render review prompt
	reviewPrompt, err := prompt.Render("review", map[string]string{
		"SourceCriteria": string(acData),
		"PlanSummary":    string(planData),
		"DiffText":       pack.DiffText,
		"DiffStat":       diffStat,
		"Rationale":      string(rationaleData),
	})
	if err != nil {
		return nil, fmt.Errorf("render review prompt: %w", err)
	}

	// Stage 4: run each reviewer concurrently
	type reviewerResult struct {
		name     string
		response *agents.AgentResponse
		err      error
	}

	results := make([]reviewerResult, len(opts.Reviewers))
	var wg sync.WaitGroup
	for i, reviewerName := range opts.Reviewers {
		wg.Add(1)
		go func(idx int, name string) {
			defer wg.Done()
			profileName := cfg.StepRouting["final_review"]
			if profileName == "" {
				profileName = "reasoning_high"
			}
			profile, ok := cfg.ModelProfiles[profileName]
			if !ok {
				results[idx] = reviewerResult{name: name, err: fmt.Errorf("profile %q not found", profileName)}
				return
			}
			runner, err := agents.Get(cfg, profile.Provider, profileName)
			if err != nil {
				results[idx] = reviewerResult{name: name, err: err}
				return
			}
			artifactPath := project.ReviewsDir(rootDir) + "/" + name + "-review.md"
			resp, err := runner.Run(ctx, agents.AgentRequest{
				Prompt:       reviewPrompt,
				WorkDir:      rootDir,
				ArtifactPath: artifactPath,
			})
			results[idx] = reviewerResult{name: name, response: resp, err: err}
		}(i, reviewerName)
	}
	wg.Wait()

	// Stage 5: parse each reviewer's JSON
	var allFindings []ReviewerFinding
	for _, r := range results {
		if r.err != nil {
			fmt.Fprintf(os.Stderr, "WARN: reviewer %s error: %v\n", r.name, r.err)
			continue
		}
		if r.response == nil {
			continue
		}
		findings, err := parseReviewerFindings(r.name, r.response.Stdout)
		if err != nil {
			fmt.Fprintf(os.Stderr, "WARN: parse %s response: %v (raw saved to reviews/)\n", r.name, err)
			// Persist raw and emit WARN finding
			allFindings = append(allFindings, ReviewerFinding{
				Reviewer: r.name, Severity: "WARN",
				Message: fmt.Sprintf("malformed response: %v", err),
			})
			continue
		}
		allFindings = append(allFindings, findings...)
	}

	// Stage 6: aggregate
	_, disagreements := AggregateFindings(detResults, allFindings)
	finalVerdict := AggregateWithReviewers(detResults, allFindings, disagreements)

	// Stage 7: build final verdict
	v := &FinalVerdict{
		Verdict:             finalVerdict,
		Summary:             buildSummary(finalVerdict, detResults),
		DeterministicChecks: detResults,
		ReviewerFindings:    allFindings,
		Disagreements:       disagreements,
	}
	v.RecommendedFixPrompt = GenerateFixPrompt(v)
	return v, nil
}

// parseReviewerFindings extracts ReviewerFinding list from an AI response.
func parseReviewerFindings(reviewerName, stdout string) ([]ReviewerFinding, error) {
	var parsed struct {
		Findings []struct {
			ACID       string `json:"ac_id"`
			Severity   string `json:"severity"`
			File       string `json:"file"`
			Message    string `json:"message"`
		} `json:"findings"`
	}
	if err := json.Unmarshal([]byte(stdout), &parsed); err != nil {
		return nil, fmt.Errorf("unmarshal reviewer JSON: %w", err)
	}
	var findings []ReviewerFinding
	for _, f := range parsed.Findings {
		findings = append(findings, ReviewerFinding{
			Reviewer:  reviewerName,
			Severity:  f.Severity,
			Message:   fmt.Sprintf("[%s] %s", f.File, f.Message),
			RelatedAC: f.ACID,
		})
	}
	return findings, nil
}

// buildSummary is defined in verify.go — already accessible within package.
// Reuse it here without re-declaration.
