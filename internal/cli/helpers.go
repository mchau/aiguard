package cli

import (
	"fmt"
	"os"

	"github.com/mchau/aiguard/internal/project"
)

// requireRoot finds the project root or returns an actionable error.
func requireRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}
	root, err := project.FindRoot(cwd, cfgPath)
	if err != nil {
		return "", fmt.Errorf("%w\nRun 'aiguard init' to initialize a workspace first.", err)
	}
	return root, nil
}
