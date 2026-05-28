package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Append writes ev as a JSON line to .aiguard/logs/audit.jsonl.
// It creates the file and parent directories if they don't exist.
func Append(rootDir string, ev Event) error {
	if ev.Timestamp.IsZero() {
		ev.Timestamp = time.Now().UTC()
	}

	logPath := filepath.Join(rootDir, ".aiguard", "logs", "audit.jsonl")
	if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
		return fmt.Errorf("create audit log dir: %w", err)
	}

	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open audit log: %w", err)
	}
	defer f.Close()

	line, err := json.Marshal(ev)
	if err != nil {
		return fmt.Errorf("marshal audit event: %w", err)
	}

	if _, err := f.Write(append(line, '\n')); err != nil {
		return fmt.Errorf("write audit event: %w", err)
	}
	return nil
}
