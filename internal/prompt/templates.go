package prompt

import (
	_ "embed"
	"fmt"
	"strings"
	"text/template"
)

// reviewRules is the canonical reviewer rules block from CLAUDE.md §"Prompt Requirements".
// All review templates include this block.
//
//go:embed review_rules.md
var reviewRules string

//go:embed digest.md.tmpl
var digestTmpl string

//go:embed guidance.md.tmpl
var guidanceTmpl string

//go:embed plan.md.tmpl
var planTmpl string

//go:embed review.md.tmpl
var reviewTmpl string

//go:embed test_strategy.md.tmpl
var testStrategyTmpl string

//go:embed checkpoint.md.tmpl
var checkpointTmpl string

//go:embed snapshot.md.tmpl
var snapshotTmpl string

// Render executes a named template with data, injecting the standard reviewer rules block.
func Render(name string, data interface{}) (string, error) {
	tmplSrc, ok := templateSources[name]
	if !ok {
		return "", fmt.Errorf("unknown template %q", name)
	}

	// Define a "review_rules" sub-template so templates can call {{template "review_rules" .}}
	fullSrc := `{{define "review_rules"}}` + reviewRules + `{{end}}` + "\n" + tmplSrc

	t, err := template.New(name).Parse(fullSrc)
	if err != nil {
		return "", fmt.Errorf("parse template %q: %w", name, err)
	}

	var sb strings.Builder
	if err := t.Execute(&sb, data); err != nil {
		return "", fmt.Errorf("render template %q: %w", name, err)
	}
	return sb.String(), nil
}

var templateSources = map[string]string{
	"digest":        digestTmpl,
	"guidance":      guidanceTmpl,
	"plan":          planTmpl,
	"review":        reviewTmpl,
	"test_strategy": testStrategyTmpl,
	"checkpoint":    checkpointTmpl,
	"snapshot":      snapshotTmpl,
}

// TemplateNames returns the list of registered template names.
func TemplateNames() []string {
	names := make([]string, 0, len(templateSources))
	for k := range templateSources {
		names = append(names, k)
	}
	return names
}
