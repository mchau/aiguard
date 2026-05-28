package gates

import "time"

// GateStatus values per SPEC.md §6 and PRD §13.
const (
	StatusNotStarted       = "NOT_STARTED"
	StatusInProgress       = "IN_PROGRESS"
	StatusNeedsReview      = "NEEDS_REVIEW"
	StatusPass             = "PASS"
	StatusWarn             = "WARN"
	StatusFail             = "FAIL"
	StatusBlocked          = "BLOCKED"
	StatusApprovedByUser   = "APPROVED_BY_USER"
	StatusSkippedWithReason = "SKIPPED_WITH_REASON"
)

// Gate records the state of a workflow gate.
type Gate struct {
	Status     string    `json:"status"`
	ApprovedAt time.Time `json:"approved_at,omitempty"`
	ApprovedBy string    `json:"approved_by,omitempty"`
	Reason     string    `json:"reason,omitempty"`
}

// GateFile is the persisted structure in .aiguard/gates.json.
type GateFile map[string]Gate
