package checks

import "github.com/mchau/aiguard/internal/config"

// TypedInput is a concrete version of CheckInput with a typed Config.
// Callers build TypedInput and call ToCheckInput() before passing to RunAll.
type TypedInput struct {
	Config        *config.Config
	ChangedFiles  []string
	DiffText      string
	FileContent   func(path string) ([]byte, error)
	RootDir       string
	PlanPath      string
	RationalePath string
}

// ToCheckInput converts to the interface-typed CheckInput for the registry.
func (t TypedInput) ToCheckInput() CheckInput {
	return CheckInput{
		Config:        t.Config,
		ChangedFiles:  t.ChangedFiles,
		DiffText:      t.DiffText,
		FileContent:   t.FileContent,
		RootDir:       t.RootDir,
		PlanPath:      t.PlanPath,
		RationalePath: t.RationalePath,
	}
}

// cfg extracts the typed Config from a CheckInput.
func cfg(input CheckInput) *config.Config {
	if input.Config == nil {
		return nil
	}
	c, _ := input.Config.(*config.Config)
	return c
}
