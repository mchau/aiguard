package config_test

import (
	"testing"

	"github.com/mchau/aiguard/internal/config"
)

func TestRenderArgs(t *testing.T) {
	profile := config.ModelProfile{Provider: "claude_code", ModelAlias: "opus", Effort: "high"}

	args, err := config.RenderArgs("--model {{.ModelAlias}} --effort {{.Effort}}", profile)
	if err != nil {
		t.Fatal(err)
	}
	if len(args) != 4 || args[1] != "opus" || args[3] != "high" {
		t.Errorf("unexpected args: %v", args)
	}
}

func TestRenderArgsEmpty(t *testing.T) {
	profile := config.ModelProfile{}
	args, err := config.RenderArgs("", profile)
	if err != nil {
		t.Fatal(err)
	}
	if args != nil {
		t.Errorf("expected nil for empty template, got %v", args)
	}
}

func TestRenderArgsBadTemplate(t *testing.T) {
	_, err := config.RenderArgs("{{.Bad", config.ModelProfile{})
	if err == nil {
		t.Fatal("expected error for invalid template")
	}
}
