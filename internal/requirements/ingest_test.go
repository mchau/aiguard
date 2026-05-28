package requirements_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mchau/aiguard/internal/requirements"
)

func TestIngestFile(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "ticket.md")
	if err := os.WriteFile(src, []byte("# Ticket\n\n## AC\n- Do X\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	rawMD := filepath.Join(dir, ".aiguard", "requirements", "raw.md")
	if err := requirements.IngestFile(src, rawMD); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(rawMD)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "Do X") {
		t.Errorf("expected ticket content in raw.md, got: %s", data)
	}
}

func TestIngestStdin(t *testing.T) {
	dir := t.TempDir()
	rawMD := filepath.Join(dir, "raw.md")
	r := strings.NewReader("# Stdin Ticket\n\n## Acceptance Criteria\n- AC1\n")
	if err := requirements.IngestStdin(rawMD, r); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(rawMD)
	if !strings.Contains(string(data), "AC1") {
		t.Errorf("expected AC1 in raw.md")
	}
}
