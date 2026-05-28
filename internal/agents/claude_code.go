package agents

import (
	"github.com/mchau/aiguard/internal/config"
)

// NewClaudeCode creates a CommandRunner for the claude_code agent.
func NewClaudeCode(cfg *config.Config, profileName string) (*CommandRunner, error) {
	agentCfg, ok := cfg.Agents["claude_code"]
	if !ok {
		return nil, errMissingAgent("claude_code")
	}
	profile, ok := cfg.ModelProfiles[profileName]
	if !ok {
		return nil, errMissingProfile(profileName)
	}
	return NewCommandRunner("claude_code", agentCfg, profile), nil
}
