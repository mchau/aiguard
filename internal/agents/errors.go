package agents

import "fmt"

func errMissingAgent(name string) error {
	return fmt.Errorf("agent %q not found in config; add it to config.agents", name)
}

func errMissingProfile(name string) error {
	return fmt.Errorf("model profile %q not found in config; add it to config.model_profiles", name)
}
