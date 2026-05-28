package cli_test

import (
	"context"
	"os/exec"
	"testing"
)

// TestDoctorGitAvailable checks the underlying git check works on this machine.
func TestDoctorGitAvailable(t *testing.T) {
	ctx := context.Background()
	out, err := exec.CommandContext(ctx, "git", "--version").Output()
	if err != nil {
		t.Skipf("git not available: %v", err)
	}
	if len(out) == 0 {
		t.Error("expected git --version output")
	}
}

// TestDoctorAgentLookup verifies exec.LookPath contract used in doctor checks.
func TestDoctorAgentLookup(t *testing.T) {
	// git should be resolvable; a fake binary should not.
	if _, err := exec.LookPath("git"); err != nil {
		t.Skipf("git not on PATH: %v", err)
	}
	if _, err := exec.LookPath("_aiguard_nonexistent_binary_xyz_"); err == nil {
		t.Error("expected error for nonexistent binary")
	}
}
