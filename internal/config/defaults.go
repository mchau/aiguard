package config

// Default returns the canonical default config from CLAUDE.md §"Config Requirements".
// This is the single source of default values — do not duplicate them elsewhere.
func Default() *Config {
	return &Config{
		Project: Project{
			Name:              "",
			SourceBranch:      "main",
			DefaultCompareRef: "main",
		},
		Privacy: Privacy{
			DefaultMode:                 "local_first",
			AllowExternalAI:             false,
			RedactSecrets:               true,
			IncludeActualFilesByDefault: false,
			MaxFileBytesForContext:      50000,
		},
		Snapshot: Snapshot{
			IncludeTree:           true,
			IncludeComponentNotes: true,
			UpdateFromDiff:        true,
			MaxTreeDepth:          8,
			StaleAfterCommits:     20,
			Ignore: []string{
				".git/**",
				"node_modules/**",
				"vendor/**",
				"dist/**",
				"build/**",
				".aiguard/**",
			},
		},
		RiskPaths: RiskPaths{
			Critical: []string{
				".env",
				".env.*",
				"secrets/**",
				"infra/prod/**",
			},
			High: []string{
				"auth/**",
				"billing/**",
				"payments/**",
				"permissions/**",
				"migrations/**",
				"database/**",
				"db/migrate/**",
			},
		},
		ForbiddenPaths: []string{
			".env",
			".env.*",
			"secrets/**",
			"*.pem",
			"*.key",
		},
		TestCommands: []TestCommand{
			{
				Name:           "unit",
				Command:        "go test ./...",
				TimeoutSeconds: 600,
			},
		},
		Gates: Gates{},
		ModelProfiles: map[string]ModelProfile{
			"cheap_fast": {
				Provider:   "claude_code",
				ModelAlias: "haiku",
				Effort:     "low",
			},
			"reasoning_high": {
				Provider:   "claude_code",
				ModelAlias: "opus",
				Effort:     "high",
			},
			"coding_balanced": {
				Provider:   "claude_code",
				ModelAlias: "sonnet",
				Effort:     "medium",
			},
		},
		StepRouting: map[string]string{
			"snapshot_update":     "cheap_fast",
			"requirement_digest":  "reasoning_high",
			"brainstorm":          "reasoning_high",
			"test_strategy":       "reasoning_high",
			"tech_guidance":       "reasoning_high",
			"guidance_review":     "reasoning_high",
			"plan_create":         "reasoning_high",
			"plan_review":         "reasoning_high",
			"implementation":      "coding_balanced",
			"checkpoint_review":   "reasoning_high",
			"final_review":        "reasoning_high",
		},
		Agents: map[string]AgentConfig{
			"claude_code": {
				Enabled:           true,
				Command:           "claude",
				PromptArg:         "-p",
				ModelArgTemplate:  "--model {{.ModelAlias}}",
				EffortArgTemplate: "--effort {{.Effort}}",
				SupportsStdin:     true,
				TimeoutSeconds:    3600,
			},
			"codex": {
				Enabled:           true,
				Command:           "codex",
				PromptArg:         "",
				ModelArgTemplate:  "",
				EffortArgTemplate: "",
				SupportsStdin:     true,
				TimeoutSeconds:    3600,
			},
		},
	}
}
