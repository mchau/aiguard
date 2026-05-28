package agents

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/mchau/aiguard/internal/config"
)

// CommandRunner implements Agent using a configured AgentConfig.
type CommandRunner struct {
	name    string
	agentCfg config.AgentConfig
	profile  config.ModelProfile
}

// NewCommandRunner creates a CommandRunner for an agent+profile pair.
func NewCommandRunner(name string, agentCfg config.AgentConfig, profile config.ModelProfile) *CommandRunner {
	return &CommandRunner{name: name, agentCfg: agentCfg, profile: profile}
}

func (r *CommandRunner) Name() string { return r.name }

// Run executes the agent command with the given request.
// Each call uses its own buffers and the request's WorkDir to avoid concurrency issues.
func (r *CommandRunner) Run(ctx context.Context, req AgentRequest) (*AgentResponse, error) {
	timeout := time.Duration(r.agentCfg.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 3600 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Build arg list
	cmdArgs := []string{}

	// model and effort args from templates
	modelArgs, err := config.RenderArgs(r.agentCfg.ModelArgTemplate, r.profile)
	if err != nil {
		return nil, fmt.Errorf("render model args: %w", err)
	}
	effortArgs, err := config.RenderArgs(r.agentCfg.EffortArgTemplate, r.profile)
	if err != nil {
		return nil, fmt.Errorf("render effort args: %w", err)
	}
	cmdArgs = append(cmdArgs, modelArgs...)
	cmdArgs = append(cmdArgs, effortArgs...)

	// Prompt: via stdin or prompt_arg
	if !r.agentCfg.SupportsStdin && r.agentCfg.PromptArg != "" {
		cmdArgs = append(cmdArgs, r.agentCfg.PromptArg, req.Prompt)
	}

	cmd := exec.CommandContext(ctx, r.agentCfg.Command, cmdArgs...) //nolint:gosec
	if req.WorkDir != "" {
		cmd.Dir = req.WorkDir
	}

	// Per-call buffers (no shared state between concurrent calls)
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	if r.agentCfg.SupportsStdin {
		cmd.Stdin = io.NopCloser(bytes.NewBufferString(req.Prompt))
		cmd.Stdout = &stdoutBuf
	} else {
		cmd.Stdout = &stdoutBuf
	}

	start := time.Now()
	runErr := cmd.Run()
	duration := time.Since(start).Seconds()

	exitCode := 0
	if runErr != nil {
		if ee, ok := runErr.(*exec.ExitError); ok {
			exitCode = ee.ExitCode()
		}
	}

	resp := &AgentResponse{
		Stdout:   stdoutBuf.String(),
		Stderr:   stderrBuf.String(),
		ExitCode: exitCode,
		Duration: duration,
	}

	// Write raw stdout to artifact file if caller provided a path
	if req.ArtifactPath != "" {
		if err := os.MkdirAll(filepath.Dir(req.ArtifactPath), 0o755); err == nil {
			_ = os.WriteFile(req.ArtifactPath, []byte(resp.Stdout), 0o644)
		}
	}

	if runErr != nil && exitCode != 0 {
		return resp, fmt.Errorf("agent %s exited %d: %s", r.name, exitCode, stderrBuf.String())
	}
	return resp, nil
}
