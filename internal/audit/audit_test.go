package audit_test

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/mchau/aiguard/internal/audit"
)

func TestAppendTwoEvents(t *testing.T) {
	dir := t.TempDir()

	e1 := audit.Event{Command: "init", Status: "success", Inputs: []string{}, Outputs: []string{}}
	e2 := audit.Event{Command: "verify", Status: "success", Inputs: []string{"--base", "develop"}, Outputs: []string{}}

	if err := audit.Append(dir, e1); err != nil {
		t.Fatalf("append e1: %v", err)
	}
	if err := audit.Append(dir, e2); err != nil {
		t.Fatalf("append e2: %v", err)
	}

	logPath := filepath.Join(dir, ".aiguard", "logs", "audit.jsonl")
	f, err := os.Open(logPath)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	var events []audit.Event
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var ev audit.Event
		if err := json.Unmarshal(scanner.Bytes(), &ev); err != nil {
			t.Fatalf("parse line: %v", err)
		}
		events = append(events, ev)
	}

	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
	if events[0].Command != "init" {
		t.Errorf("first event: got %q, want %q", events[0].Command, "init")
	}
	if events[1].Command != "verify" {
		t.Errorf("second event: got %q, want %q", events[1].Command, "verify")
	}
}
