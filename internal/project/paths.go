package project

import "path/filepath"

// All paths are relative to the project root (the dir containing .aiguard/).

const DotAiguard = ".aiguard"

func dir(parts ...string) string { return filepath.Join(parts...) }

// Config
func ConfigFile(root string) string { return dir(root, DotAiguard, "config.yaml") }

// Gates
func GatesFile(root string) string { return dir(root, DotAiguard, "gates.json") }

// Snapshot
func SnapshotDir(root string) string         { return dir(root, DotAiguard, "snapshot") }
func SnapshotMD(root string) string          { return dir(root, DotAiguard, "snapshot", "snapshot.md") }
func SnapshotMetaJSON(root string) string    { return dir(root, DotAiguard, "snapshot", "snapshot-meta.json") }
func FileMapJSON(root string) string         { return dir(root, DotAiguard, "snapshot", "file-map.json") }
func RiskMapJSON(root string) string         { return dir(root, DotAiguard, "snapshot", "risk-map.json") }
func TestMapJSON(root string) string         { return dir(root, DotAiguard, "snapshot", "test-map.json") }
func ChangeLogMD(root string) string         { return dir(root, DotAiguard, "snapshot", "change-log.md") }

// Requirements
func RequirementsDir(root string) string         { return dir(root, DotAiguard, "requirements") }
func RawMD(root string) string                   { return dir(root, DotAiguard, "requirements", "raw.md") }
func DigestMD(root string) string                { return dir(root, DotAiguard, "requirements", "digest.md") }
func AcceptanceCriteriaJSON(root string) string  { return dir(root, DotAiguard, "requirements", "acceptance-criteria.json") }
func OpenQuestionsMD(root string) string         { return dir(root, DotAiguard, "requirements", "open-questions.md") }
func AssumptionsMD(root string) string           { return dir(root, DotAiguard, "requirements", "assumptions.md") }

// Tests
func TestsDir(root string) string          { return dir(root, DotAiguard, "tests") }
func TestStrategyMD(root string) string    { return dir(root, DotAiguard, "tests", "test-strategy.md") }
func TestRequirementsJSON(root string) string { return dir(root, DotAiguard, "tests", "test-requirements.json") }
func TestReviewMD(root string) string      { return dir(root, DotAiguard, "tests", "test-review.md") }

// Guidance
func GuidanceDir(root string) string        { return dir(root, DotAiguard, "guidance") }
func TechGuidanceMD(root string) string     { return dir(root, DotAiguard, "guidance", "tech-guidance.md") }
func TechGuidanceJSON(root string) string   { return dir(root, DotAiguard, "guidance", "tech-guidance.json") }
func GuidanceReviewMD(root string) string   { return dir(root, DotAiguard, "guidance", "guidance-review.md") }
func GuidanceReviewJSON(root string) string { return dir(root, DotAiguard, "guidance", "guidance-review.json") }

// Plans
func PlansDir(root string) string       { return dir(root, DotAiguard, "plans") }
func PlanMD(root string) string         { return dir(root, DotAiguard, "plans", "plan.md") }
func PlanJSON(root string) string       { return dir(root, DotAiguard, "plans", "plan.json") }
func PlanReviewMD(root string) string   { return dir(root, DotAiguard, "plans", "plan-review.md") }
func PlanReviewJSON(root string) string { return dir(root, DotAiguard, "plans", "plan-review.json") }

// Implementation
func ImplementationDir(root string) string      { return dir(root, DotAiguard, "implementation") }
func ChangedFilesRationaleMD(root string) string { return dir(root, DotAiguard, "implementation", "changed-files-rationale.md") }
func CheckpointsDir(root string) string          { return dir(root, DotAiguard, "implementation", "checkpoints") }

// Reviews
func ReviewsDir(root string) string             { return dir(root, DotAiguard, "reviews") }
func DeterministicChecksJSON(root string) string { return dir(root, DotAiguard, "reviews", "deterministic-checks.json") }
func ClaudeReviewMD(root string) string          { return dir(root, DotAiguard, "reviews", "claude-review.md") }
func CodexReviewMD(root string) string           { return dir(root, DotAiguard, "reviews", "codex-review.md") }
func DisagreementReportMD(root string) string    { return dir(root, DotAiguard, "reviews", "disagreement-report.md") }
func FinalVerdictMD(root string) string          { return dir(root, DotAiguard, "reviews", "final-verdict.md") }
func FinalVerdictJSON(root string) string        { return dir(root, DotAiguard, "reviews", "final-verdict.json") }

// Logs
func LogsDir(root string) string    { return dir(root, DotAiguard, "logs") }
func AuditJSONL(root string) string { return dir(root, DotAiguard, "logs", "audit.jsonl") }
