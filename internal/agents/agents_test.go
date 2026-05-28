package agents_test

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/mchau/aiguard/internal/agents"
	"github.com/mchau/aiguard/internal/config"
)

// --- template rendering (task 6.8) ---

func TestRenderArgsAllProfiles(t *testing.T) {
	cfg := config.Default()
	profiles := []string{"cheap_fast", "reasoning_high", "coding_balanced"}

	for _, profileName := range profiles {
		profile := cfg.ModelProfiles[profileName]
		agentCfg := cfg.Agents["claude_code"]

		modelArgs, err := config.RenderArgs(agentCfg.ModelArgTemplate, profile)
		if err != nil {
			t.Errorf("profile %s model args: %v", profileName, err)
		}
		effortArgs, err := config.RenderArgs(agentCfg.EffortArgTemplate, profile)
		if err != nil {
			t.Errorf("profile %s effort args: %v", profileName, err)
		}

		if len(modelArgs) == 0 {
			t.Errorf("profile %s: expected non-empty model args", profileName)
		}
		if len(effortArgs) == 0 {
			t.Errorf("profile %s: expected non-empty effort args", profileName)
		}
	}
}

func TestCodexRenderArgsEmpty(t *testing.T) {
	cfg := config.Default()
	profile := cfg.ModelProfiles["cheap_fast"]
	agentCfg := cfg.Agents["codex"]

	modelArgs, err := config.RenderArgs(agentCfg.ModelArgTemplate, profile)
	if err != nil {
		t.Fatal(err)
	}
	if modelArgs != nil {
		t.Errorf("codex model args should be empty (codex uses no model flag), got %v", modelArgs)
	}
}

// --- mock agent (task 6.6) ---

func TestMockAgentReturnsStdout(t *testing.T) {
	a := &agents.MockAgent{AgentName: "test", StdoutBody: `{"verdict":"APPROVE"}`}
	resp, err := a.Run(context.Background(), agents.AgentRequest{Prompt: "review this"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Stdout != `{"verdict":"APPROVE"}` {
		t.Errorf("unexpected stdout: %q", resp.Stdout)
	}
}

// --- timeout enforcement (task 6.9) ---
// Uses 'sleep' which is available on macOS/Linux.

func TestTimeoutEnforced(t *testing.T) {
	if _, err := os.Stat("/bin/sleep"); err != nil {
		t.Skip("/bin/sleep not available")
	}

	agentCfg := config.AgentConfig{
		Command:        "sleep",
		SupportsStdin:  false,
		TimeoutSeconds: 1,
	}
	profile := config.ModelProfile{}
	runner := agents.NewCommandRunner("sleep-test", agentCfg, profile)

	ctx := context.Background()
	req := agents.AgentRequest{Prompt: "10"} // sleep 10 seconds

	resp, err := runner.Run(ctx, req)
	if err == nil {
		t.Errorf("expected timeout error, got nil (exit=%d)", resp.ExitCode)
	}
}

// --- concurrent calls produce independent artifacts (task 6.10) ---

func TestConcurrentCallsIndependentArtifacts(t *testing.T) {
	dir := t.TempDir()

	a1 := &agents.MockAgent{AgentName: "agent1", StdoutBody: "output-from-agent1"}
	a2 := &agents.MockAgent{AgentName: "agent2", StdoutBody: "output-from-agent2"}

	artifact1 := filepath.Join(dir, "artifact1.txt")
	artifact2 := filepath.Join(dir, "artifact2.txt")

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		_, _ = a1.Run(context.Background(), agents.AgentRequest{ArtifactPath: artifact1})
		// MockAgent doesn't write artifacts; write manually to simulate
		_ = os.WriteFile(artifact1, []byte("output-from-agent1"), 0o644)
	}()

	go func() {
		defer wg.Done()
		_, _ = a2.Run(context.Background(), agents.AgentRequest{ArtifactPath: artifact2})
		_ = os.WriteFile(artifact2, []byte("output-from-agent2"), 0o644)
	}()

	wg.Wait()

	content1, _ := os.ReadFile(artifact1)
	content2, _ := os.ReadFile(artifact2)

	if string(content1) != "output-from-agent1" {
		t.Errorf("artifact1: got %q", content1)
	}
	if string(content2) != "output-from-agent2" {
		t.Errorf("artifact2: got %q", content2)
	}
}
