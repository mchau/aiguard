package config

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

// RenderArgs renders a CLI arg template string using a ModelProfile.
// Templates use Go text/template syntax: {{.ModelAlias}}, {{.Effort}}, {{.Provider}}.
// Returns the rendered string split on whitespace as individual args.
// An empty template returns nil (no args to add).
func RenderArgs(tmpl string, profile ModelProfile) ([]string, error) {
	if strings.TrimSpace(tmpl) == "" {
		return nil, nil
	}

	t, err := template.New("arg").Parse(tmpl)
	if err != nil {
		return nil, fmt.Errorf("parse arg template %q: %w", tmpl, err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, profile); err != nil {
		return nil, fmt.Errorf("render arg template %q: %w", tmpl, err)
	}

	rendered := strings.TrimSpace(buf.String())
	if rendered == "" {
		return nil, nil
	}

	return strings.Fields(rendered), nil
}
