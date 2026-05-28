package agents

import (
	"github.com/mchau/aiguard/internal/config"
)

// NewCodex creates a CommandRunner for the codex agent.
func NewCodex(cfg *config.Config, profileName string) (*CommandRunner, error) {
	agentCfg, ok := cfg.Agents["codex"]
	if !ok {
		return nil, errMissingAgent("codex")
	}
	profile, ok := cfg.ModelProfiles[profileName]
	if !ok {
		return nil, errMissingProfile(profileName)
	}
	return NewCommandRunner("codex", agentCfg, profile), nil
}
