package report

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/mchau/aiguard/internal/review"
)

// WriteFinalVerdictJSON serializes the verdict to path as indented JSON.
func WriteFinalVerdictJSON(path string, v *review.FinalVerdict) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal verdict: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write verdict JSON %s: %w", path, err)
	}
	return nil
}
