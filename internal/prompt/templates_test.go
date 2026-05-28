package prompt_test

import (
	"strings"
	"testing"

	"github.com/mchau/aiguard/internal/prompt"
)

func TestRenderAllTemplatesNonEmpty(t *testing.T) {
	for _, name := range prompt.TemplateNames() {
		t.Run(name, func(t *testing.T) {
			out, err := prompt.Render(name, map[string]string{
				"RawContent":         "test content",
				"AcceptanceCriteria": "- AC1: do X",
				"SnapshotSummary":    "test snapshot",
				"GuidanceSummary":    "test guidance",
				"PlanSummary":        "test plan",
				"SourceCriteria":     "- AC1: do X",
				"DiffText":           "+line added",
				"DiffStat":           "1 file changed",
				"Rationale":          "test rationale",
				"PlanSteps":          "S1, S2",
				"ChangedPaths":       "internal/foo/",
			})
			if err != nil {
				t.Fatalf("render %q: %v", name, err)
			}
			if strings.TrimSpace(out) == "" {
				t.Errorf("template %q produced empty output", name)
			}
		})
	}
}

func TestRenderReviewIncludesRules(t *testing.T) {
	out, err := prompt.Render("review", map[string]string{
		"SourceCriteria": "- AC1: do X",
		"PlanSummary":    "",
		"DiffText":       "+line",
		"DiffStat":       "",
		"Rationale":      "",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "You are a reviewer") {
		t.Error("review template should include reviewer rules block")
	}
}

func TestRenderUnknownTemplate(t *testing.T) {
	_, err := prompt.Render("nonexistent_template", nil)
	if err == nil {
		t.Error("expected error for unknown template")
	}
}
