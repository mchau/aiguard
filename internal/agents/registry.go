package agents

import (
	"fmt"
	"os/exec"

	"github.com/mchau/aiguard/internal/config"
)

// Get returns a CommandRunner for the named agent using the given profile.
func Get(cfg *config.Config, agentName, profileName string) (*CommandRunner, error) {
	agentCfg, ok := cfg.Agents[agentName]
	if !ok {
		return nil, errMissingAgent(agentName)
	}
	profile, ok := cfg.ModelProfiles[profileName]
	if !ok {
		return nil, errMissingProfile(profileName)
	}
	return NewCommandRunner(agentName, agentCfg, profile), nil
}

// CheckAvailable returns nil if the agent command is on PATH.
func CheckAvailable(cfg *config.Config, agentName string) error {
	agentCfg, ok := cfg.Agents[agentName]
	if !ok {
		return errMissingAgent(agentName)
	}
	if agentCfg.Command == "" {
		return fmt.Errorf("agent %q has no command configured", agentName)
	}
	if _, err := exec.LookPath(agentCfg.Command); err != nil {
		return fmt.Errorf("agent %q command %q not found on PATH: %w", agentName, agentCfg.Command, err)
	}
	return nil
}
