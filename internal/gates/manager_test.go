package gates_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mchau/aiguard/internal/gates"
)

func TestGatePersistAndLoad(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".aiguard"), 0o755); err != nil {
		t.Fatal(err)
	}

	m, err := gates.Load(dir)
	if err != nil {
		t.Fatal(err)
	}

	if err := m.Set("digest", gates.StatusApprovedByUser, "looks correct", "local-user"); err != nil {
		t.Fatal(err)
	}

	// Reload and verify persistence
	m2, err := gates.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	g := m2.Get("digest")
	if g.Status != gates.StatusApprovedByUser {
		t.Errorf("got status %q, want %q", g.Status, gates.StatusApprovedByUser)
	}
	if g.ApprovedBy != "local-user" {
		t.Errorf("got approved_by %q", g.ApprovedBy)
	}
}

func TestGateRequireApproved(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".aiguard"), 0o755); err != nil {
		t.Fatal(err)
	}

	m, _ := gates.Load(dir)

	// Not set → error
	if err := m.RequireApproved("digest"); err == nil {
		t.Error("expected error for unset gate")
	}

	// Set to IN_PROGRESS → error
	_ = m.Set("digest", gates.StatusInProgress, "", "")
	if err := m.RequireApproved("digest"); err == nil {
		t.Error("expected error for IN_PROGRESS gate")
	}

	// Set to APPROVED_BY_USER → ok
	_ = m.Set("digest", gates.StatusApprovedByUser, "ok", "user")
	if err := m.RequireApproved("digest"); err != nil {
		t.Errorf("unexpected error for approved gate: %v", err)
	}
}

func TestGateMissingFileOK(t *testing.T) {
	dir := t.TempDir()
	// No .aiguard/gates.json — should load empty without error
	m, err := gates.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if m.Get("digest").Status != "" {
		t.Error("expected empty gate for fresh load")
	}
}
