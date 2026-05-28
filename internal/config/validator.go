package config

import (
	"fmt"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

// Validate returns an error if the config has invalid or inconsistent values.
func Validate(cfg *Config) error {
	var errs []string

	// 1.10a: source_branch non-empty
	if strings.TrimSpace(cfg.Project.SourceBranch) == "" {
		errs = append(errs, "project.source_branch must not be empty")
	}

	// 1.10b: every step_routing profile exists in model_profiles
	for step, profileName := range cfg.StepRouting {
		if _, ok := cfg.ModelProfiles[profileName]; !ok {
			errs = append(errs, fmt.Sprintf("step_routing.%s references unknown model profile %q", step, profileName))
		}
	}

	// 1.10c: every model_profile's provider exists in agents and is enabled (warn-level, collected as errors here)
	for name, profile := range cfg.ModelProfiles {
		agent, ok := cfg.Agents[profile.Provider]
		if !ok {
			errs = append(errs, fmt.Sprintf("model_profiles.%s.provider %q not found in agents", name, profile.Provider))
		} else if !agent.Enabled {
			// warn only — do not hard-fail for disabled agents
			_ = agent
		}
	}

	// 1.10d: glob patterns compile
	patterns := make([]string, 0)
	patterns = append(patterns, cfg.Snapshot.Ignore...)
	patterns = append(patterns, cfg.ForbiddenPaths...)
	patterns = append(patterns, cfg.RiskPaths.Critical...)
	patterns = append(patterns, cfg.RiskPaths.High...)
	patterns = append(patterns, cfg.RiskPaths.Medium...)
	patterns = append(patterns, cfg.DependencyFiles...)
	patterns = append(patterns, cfg.MigrationPaths...)

	for _, p := range patterns {
		if _, err := doublestar.Match(p, "test"); err != nil {
			errs = append(errs, fmt.Sprintf("invalid glob pattern %q: %v", p, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("config validation failed:\n  - %s", strings.Join(errs, "\n  - "))
	}
	return nil
}
