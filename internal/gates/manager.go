package gates

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Manager handles gate state persistence and transitions.
type Manager struct {
	path  string
	gates GateFile
}

// Load reads .aiguard/gates.json (creates empty if missing).
func Load(rootDir string) (*Manager, error) {
	path := filepath.Join(rootDir, ".aiguard", "gates.json")
	m := &Manager{path: path, gates: GateFile{}}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return m, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read gates.json: %w", err)
	}
	if err := json.Unmarshal(data, &m.gates); err != nil {
		return nil, fmt.Errorf("parse gates.json: %w", err)
	}
	return m, nil
}

// Save writes the current gate state to disk.
func (m *Manager) Save() error {
	if err := os.MkdirAll(filepath.Dir(m.path), 0o755); err != nil {
		return fmt.Errorf("create gates dir: %w", err)
	}
	data, err := json.MarshalIndent(m.gates, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal gates: %w", err)
	}
	return os.WriteFile(m.path, data, 0o644)
}

// Get returns a gate by name (zero value if not set).
func (m *Manager) Get(name string) Gate {
	return m.gates[name]
}

// Set updates a gate and persists immediately.
func (m *Manager) Set(name, status, reason, by string) error {
	g := Gate{
		Status: status,
		Reason: reason,
	}
	if status == StatusApprovedByUser || status == StatusPass {
		g.ApprovedAt = time.Now().UTC()
		g.ApprovedBy = by
	}
	m.gates[name] = g
	return m.Save()
}

// RequireApproved returns an actionable error if the named gate is not approved.
func (m *Manager) RequireApproved(name string) error {
	g := m.gates[name]
	switch g.Status {
	case StatusApprovedByUser, StatusPass, StatusSkippedWithReason:
		return nil
	default:
		return fmt.Errorf("gate %q is %q — run 'aiguard %s approve' to proceed (or use --force --reason to override)",
			name, g.Status, name)
	}
}
