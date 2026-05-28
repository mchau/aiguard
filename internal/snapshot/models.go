package snapshot

import "time"

// SnapshotMeta is persisted to snapshot-meta.json.
type SnapshotMeta struct {
	SourceBranch    string    `json:"source_branch"`
	SourceCommit    string    `json:"source_commit"`
	UpdatedAt       time.Time `json:"updated_at"`
	SnapshotVersion int       `json:"snapshot_version"`
	TreeDepth       int       `json:"tree_depth"`
	TotalFiles      int       `json:"total_files"`
}

// FileEntry is one entry in file-map.json.
type FileEntry struct {
	Path     string `json:"path"`
	IsDir    bool   `json:"is_dir"`
	RiskLevel string `json:"risk_level,omitempty"`
	IsTest   bool   `json:"is_test,omitempty"`
}
