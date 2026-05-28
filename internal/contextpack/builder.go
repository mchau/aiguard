package contextpack

import (
	"github.com/bmatcuk/doublestar/v4"
	"github.com/mchau/aiguard/internal/config"
)

// BuildOptions control what the context pack includes.
type BuildOptions struct {
	Step         string
	RootDir      string
	DiffText     string
	DiffStat     string
	ChangedFiles []string
	Artifacts    []ContextArtifact
	// IncludeFiles overrides privacy.include_actual_files_by_default when true.
	IncludeFiles bool
}

// Build assembles a ContextPack according to privacy settings.
//  - 7.4b: honor include_actual_files_by_default (default: false)
//  - 7.4c: always exclude forbidden_paths regardless of IncludeFiles
//  - 7.4d: always exclude files larger than max_file_bytes_for_context
func Build(cfg *config.Config, opts BuildOptions) *ContextPack {
	pack := &ContextPack{
		Step:      opts.Step,
		Artifacts: opts.Artifacts,
	}

	// Redact diff
	if opts.DiffText != "" {
		redactedDiff, redactions := Redact(opts.DiffText, "diff")
		pack.DiffText = redactedDiff
		pack.DiffStat = opts.DiffStat
		pack.Redactions = append(pack.Redactions, redactions...)
	}

	// File contents: only when privacy allows
	includeFiles := opts.IncludeFiles || cfg.Privacy.IncludeActualFilesByDefault
	if includeFiles && opts.RootDir != "" {
		maxBytes := cfg.Privacy.MaxFileBytesForContext
		candidates := filterForbidden(opts.ChangedFiles, cfg.ForbiddenPaths)
		files := RelevantFiles(opts.RootDir, candidates, 20, maxBytes)

		for _, f := range files {
			redactedContent, redactions := Redact(f.Content, f.Path)
			f.Content = redactedContent
			pack.Files = append(pack.Files, f)
			pack.Redactions = append(pack.Redactions, redactions...)
		}
	}

	return pack
}

// filterForbidden removes any path that matches a forbidden pattern.
func filterForbidden(files, forbiddenPatterns []string) []string {
	var result []string
	for _, f := range files {
		blocked := false
		for _, p := range forbiddenPatterns {
			if m, _ := doublestar.Match(p, f); m {
				blocked = true
				break
			}
		}
		if !blocked {
			result = append(result, f)
		}
	}
	return result
}
