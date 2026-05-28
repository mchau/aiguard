package config

// Config is the top-level structure for .aiguard/config.yaml.
type Config struct {
	Project      Project                 `yaml:"project"`
	Privacy      Privacy                 `yaml:"privacy"`
	Snapshot     Snapshot                `yaml:"snapshot"`
	RiskPaths    RiskPaths               `yaml:"risk_paths"`
	ForbiddenPaths  []string             `yaml:"forbidden_paths"`
	DependencyFiles []string             `yaml:"dependency_files"`
	MigrationPaths  []string             `yaml:"migration_paths"`
	TestCommands []TestCommand           `yaml:"test_commands"`
	Gates        Gates                   `yaml:"gates"`
	ModelProfiles map[string]ModelProfile `yaml:"model_profiles"`
	StepRouting  map[string]string       `yaml:"step_routing"`
	Agents       map[string]AgentConfig  `yaml:"agents"`
}

// Project holds repository and branching configuration.
type Project struct {
	Name               string `yaml:"name"`
	SourceBranch       string `yaml:"source_branch"`
	DefaultCompareRef  string `yaml:"default_compare_ref"`
}

// Privacy controls what data leaves the local machine.
type Privacy struct {
	DefaultMode                    string `yaml:"default_mode"`
	AllowExternalAI                bool   `yaml:"allow_external_ai"`
	RedactSecrets                  bool   `yaml:"redact_secrets"`
	IncludeActualFilesByDefault    bool   `yaml:"include_actual_files_by_default"`
	MaxFileBytesForContext         int    `yaml:"max_file_bytes_for_context"`
}

// Snapshot controls how the codebase map is built.
type Snapshot struct {
	IncludeTree          bool     `yaml:"include_tree"`
	IncludeComponentNotes bool    `yaml:"include_component_notes"`
	UpdateFromDiff       bool     `yaml:"update_from_diff"`
	MaxTreeDepth         int      `yaml:"max_tree_depth"`
	StaleAfterCommits    int      `yaml:"stale_after_commits"`
	Ignore               []string `yaml:"ignore"`
}

// RiskPaths classifies file paths by risk level.
type RiskPaths struct {
	Critical []string `yaml:"critical"`
	High     []string `yaml:"high"`
	Medium   []string `yaml:"medium"`
}

// TestCommand defines a runnable test suite.
type TestCommand struct {
	Name           string `yaml:"name"`
	Command        string `yaml:"command"`
	TimeoutSeconds int    `yaml:"timeout_seconds"`
}

// Gates holds gate-related configuration (thresholds, etc.).
type Gates struct{}

// ModelProfile describes how to invoke an AI model via an agent.
type ModelProfile struct {
	Provider   string `yaml:"provider"`
	ModelAlias string `yaml:"model_alias"`
	Effort     string `yaml:"effort"`
}

// AgentConfig describes how to invoke an external AI agent CLI.
type AgentConfig struct {
	Enabled              bool   `yaml:"enabled"`
	Command              string `yaml:"command"`
	PromptArg            string `yaml:"prompt_arg"`
	ModelArgTemplate     string `yaml:"model_arg_template"`
	EffortArgTemplate    string `yaml:"effort_arg_template"`
	SupportsStdin        bool   `yaml:"supports_stdin"`
	TimeoutSeconds       int    `yaml:"timeout_seconds"`
}
