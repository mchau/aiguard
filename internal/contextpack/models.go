package contextpack

// ContextPack is the structured input passed to AI reviewers.
type ContextPack struct {
	Step       string          `json:"step"`
	DiffText   string          `json:"diff_text,omitempty"`
	DiffStat   string          `json:"diff_stat,omitempty"`
	Artifacts  []ContextArtifact `json:"artifacts,omitempty"`
	Files      []ContextFile   `json:"files,omitempty"`
	Redactions []Redaction     `json:"redactions,omitempty"`
}

// ContextArtifact is a named text artifact (digest, plan, rationale, etc.).
type ContextArtifact struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

// ContextFile is an included source file (only when privacy allows).
type ContextFile struct {
	Path    string `json:"path"`
	Content string `json:"content"`
	Bytes   int    `json:"bytes"`
}

// Redaction records what was masked from context.
type Redaction struct {
	Pattern     string `json:"pattern"`
	Location    string `json:"location"`
	Replacement string `json:"replacement"`
}
