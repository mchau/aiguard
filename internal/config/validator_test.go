package config_test

import (
	"strings"
	"testing"

	"github.com/mchau/aiguard/internal/config"
)

func TestValidateEmptySourceBranch(t *testing.T) {
	cfg := config.Default()
	cfg.Project.SourceBranch = ""
	err := config.Validate(cfg)
	if err == nil || !strings.Contains(err.Error(), "source_branch") {
		t.Errorf("expected source_branch error, got: %v", err)
	}
}

func TestValidateUnknownStepRoutingProfile(t *testing.T) {
	cfg := config.Default()
	cfg.StepRouting["test_step"] = "nonexistent_profile"
	err := config.Validate(cfg)
	if err == nil || !strings.Contains(err.Error(), "nonexistent_profile") {
		t.Errorf("expected unknown profile error, got: %v", err)
	}
}

func TestValidateUnknownAgentProvider(t *testing.T) {
	cfg := config.Default()
	cfg.ModelProfiles["bad_profile"] = config.ModelProfile{
		Provider:   "nonexistent_agent",
		ModelAlias: "model",
		Effort:     "low",
	}
	err := config.Validate(cfg)
	if err == nil || !strings.Contains(err.Error(), "nonexistent_agent") {
		t.Errorf("expected unknown agent error, got: %v", err)
	}
}

func TestValidateBadGlobPattern(t *testing.T) {
	cfg := config.Default()
	cfg.ForbiddenPaths = append(cfg.ForbiddenPaths, "[invalid-glob")
	err := config.Validate(cfg)
	if err == nil || !strings.Contains(err.Error(), "invalid glob") {
		t.Errorf("expected invalid glob error, got: %v", err)
	}
}

func TestValidateDefaultConfigPasses(t *testing.T) {
	if err := config.Validate(config.Default()); err != nil {
		t.Errorf("default config should be valid: %v", err)
	}
}
