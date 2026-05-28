# SPEC.md — Technical Specification for AIGuard

## 1. System Overview

AIGuard is a local-first CLI application written in Go. It orchestrates a requirements-gated workflow for AI-generated code.

Primary responsibilities:

1. Maintain a living codebase snapshot.
2. Ingest and normalize requirements.
3. Preserve source-provided acceptance criteria.
4. Generate and review digest, test strategy, technical guidance, and plan artifacts.
5. Analyze git diffs deterministically.
6. Invoke external AI CLI tools such as Claude Code and Codex when configured.
7. Aggregate reviewer results.
8. Produce a final spec-to-diff verification verdict.

## 2. Technology Stack

### Required

- Language: Go 1.22+
- CLI framework: Cobra
- Config: YAML
- Data encoding: JSON
- Markdown rendering: plain markdown templates
- Git interaction: shell out to `git`
- File matching: doublestar glob library
- Logging: structured JSONL audit logs
- Testing: Go unit tests

### Optional Later

- TUI: Bubble Tea
- SQLite: for long-term review history
- Jira integration: REST client
- GitHub integration: gh CLI or API client
- Local LLM: Ollama adapter
- API models: provider adapters

## 3. Repository Structure

```text
aiguard/
  cmd/
    aiguard/
      main.go

  internal/
    app/
      app.go
      errors.go

    cli/
      root.go
      init.go
      snapshot.go
      requirement.go
      digest.go
      clarify.go
      tests.go
      guidance.go
      plan.go
      implement.go
      checkpoint.go
      rationale.go
      verify.go
      status.go
      report.go
      doctor.go

    config/
      config.go
      defaults.go
      loader.go
      validator.go

    project/
      detect.go
      paths.go
      workspace.go

    git/
      git.go
      status.go
      branches.go
      diff.go
      files.go
      worktree.go

    snapshot/
      manager.go
      tree.go
      updater.go
      parser.go
      models.go
      renderer.go

    requirements/
      ingest.go
      parser.go
      acceptance.go
      digest.go
      questions.go
      models.go
      renderer.go

    gates/
      state.go
      manager.go
      approval.go
      transitions.go

    testspec/
      strategy.go
      review.go
      models.go
      renderer.go

    guidance/
      create.go
      review.go
      models.go
      renderer.go

    plan/
      create.go
      review.go
      models.go
      renderer.go

    checks/
      deterministic.go
      paths.go
      diffsize.go
      tests.go
      dependencies.go
      migrations.go
      secrets.go
      rationale.go
      traceability.go
      models.go

    agents/
      interface.go
      command_runner.go
      claude_code.go
      codex.go
      local.go
      mock.go
      models.go

    contextpack/
      builder.go
      redactor.go
      token_budget.go
      relevant_files.go
      models.go

    review/
      verify.go
      aggregate.go
      disagreement.go
      verdict.go
      models.go

    report/
      markdown.go
      json.go
      templates.go

    audit/
      audit.go
      event.go

    prompt/
      templates.go
      snapshot.go
      digest.go
      guidance.go
      plan.go
      review.go

  examples/
    config.yaml
    ticket.md

  docs/
    PRD.md
    SPEC.md

  CLAUDE.md
  README.md
  go.mod
```

## 4. Filesystem Layout in Target Project

When `aiguard init` is run in a codebase, create:

```text
.aiguard/
  config.yaml

  snapshot/
    snapshot.md
    snapshot-meta.json
    file-map.json
    risk-map.json
    test-map.json
    change-log.md

  requirements/
    raw.md
    digest.md
    acceptance-criteria.json
    open-questions.md
    assumptions.md

  tests/
    test-strategy.md
    test-requirements.json
    test-review.md

  guidance/
    tech-guidance.md
    tech-guidance.json
    guidance-review.md
    guidance-review.json

  plans/
    plan.md
    plan.json
    plan-review.md
    plan-review.json

  implementation/
    changed-files-rationale.md
    checkpoints/
      checkpoint-001.md
      checkpoint-001.json

  reviews/
    deterministic-checks.json
    claude-review.md
    codex-review.md
    disagreement-report.md
    final-verdict.md
    final-verdict.json

  logs/
    audit.jsonl
```

## 5. Configuration Schema

File:

```text
.aiguard/config.yaml
```

Schema:

```yaml
project:
  name: string
  source_branch: string
  default_compare_ref: string
  snapshot_dir: string

privacy:
  default_mode: local_first
  allow_external_ai: bool
  redact_secrets: bool
  include_actual_files_by_default: bool
  max_file_bytes_for_context: int

snapshot:
  include_tree: bool
  include_component_notes: bool
  update_from_diff: bool
  max_tree_depth: int
  ignore:
    - pattern

risk_paths:
  critical:
    - glob
  high:
    - glob
  medium:
    - glob

forbidden_paths:
  - glob

test_commands:
  - name: string
    command: string
    timeout_seconds: int

dependency_files:
  - "go.mod"
  - "go.sum"
  - "package.json"
  - "package-lock.json"
  - "pnpm-lock.yaml"
  - "yarn.lock"
  - "Gemfile"
  - "Gemfile.lock"
  - "requirements.txt"
  - "pyproject.toml"

migration_paths:
  - "migrations/**"
  - "db/migrate/**"
  - "database/migrations/**"

gates:
  require_user_approval:
    - digest
    - test_strategy
    - tech_guidance
    - plan
    - final_verdict
  block_on:
    - forbidden_path_changed
    - tests_failed
    - missing_source_acceptance_criteria
    - high_risk_unjustified_change

model_profiles:
  cheap_fast:
    provider: claude_code
    model_alias: haiku
    effort: low
  reasoning_high:
    provider: claude_code
    model_alias: opus
    effort: high
  coding_balanced:
    provider: claude_code
    model_alias: sonnet
    effort: medium

step_routing:
  snapshot_update: cheap_fast
  requirement_digest: reasoning_high
  brainstorm: reasoning_high
  test_strategy: reasoning_high
  tech_guidance: reasoning_high
  guidance_review: reasoning_high
  plan_create: reasoning_high
  plan_review: reasoning_high
  implementation: coding_balanced
  checkpoint_review: reasoning_high
  final_review: reasoning_high

agents:
  claude_code:
    enabled: true
    command: claude
    prompt_arg: "-p"
    model_arg_template: "--model {{model_alias}}"
    effort_arg_template: "--effort {{effort}}"
    supports_stdin: true
    timeout_seconds: 3600

  codex:
    enabled: true
    command: codex
    prompt_arg: ""
    model_arg_template: ""
    effort_arg_template: ""
    supports_stdin: true
    timeout_seconds: 3600
```

## 6. Core Data Models

### GateStatus

```go
type GateStatus string

const (
    GateNotStarted GateStatus = "NOT_STARTED"
    GateInProgress GateStatus = "IN_PROGRESS"
    GateNeedsReview GateStatus = "NEEDS_REVIEW"
    GatePass GateStatus = "PASS"
    GateWarn GateStatus = "WARN"
    GateFail GateStatus = "FAIL"
    GateBlocked GateStatus = "BLOCKED"
    GateApprovedByUser GateStatus = "APPROVED_BY_USER"
    GateSkippedWithReason GateStatus = "SKIPPED_WITH_REASON"
)
```

### AcceptanceCriterion

```go
type AcceptanceCriterion struct {
    ID          string   `json:"id"`
    Text        string   `json:"text"`
    Source      string   `json:"source"` // ticket_acceptance_criteria, inferred, user_added
    Priority    string   `json:"priority"` // must, should, could
    Category    string   `json:"category"` // ui, api, data, security, test, behavior
    NonGoal     bool     `json:"non_goal"`
    RelatedRefs []string `json:"related_refs"`
}
```

### RequirementDigest

```go
type RequirementDigest struct {
    RawSourcePath              string                `json:"raw_source_path"`
    Summary                    string                `json:"summary"`
    UserStory                  string                `json:"user_story"`
    SourceAcceptanceCriteria   []AcceptanceCriterion `json:"source_acceptance_criteria"`
    InferredAcceptanceCriteria []AcceptanceCriterion `json:"inferred_acceptance_criteria"`
    NormalizedCriteria         []AcceptanceCriterion `json:"normalized_criteria"`
    NonGoals                   []AcceptanceCriterion `json:"non_goals"`
    BusinessRules              []string              `json:"business_rules"`
    OpenQuestions              []ClarificationQuestion `json:"open_questions"`
    Assumptions                []Assumption          `json:"assumptions"`
}
```

### ClarificationQuestion

```go
type ClarificationQuestion struct {
    ID       string `json:"id"`
    Text     string `json:"text"`
    Type     string `json:"type"` // blocking, non_blocking, risk, preference
    Status   string `json:"status"` // open, answered, assumed
    Answer   string `json:"answer,omitempty"`
}
```

### SnapshotMeta

```go
type SnapshotMeta struct {
    SourceBranch           string    `json:"source_branch"`
    SourceCommit           string    `json:"source_commit"`
    PreviousSnapshotCommit string    `json:"previous_snapshot_commit"`
    UpdatedAt              time.Time `json:"updated_at"`
    SnapshotVersion        int       `json:"snapshot_version"`
    StaleSections          []string  `json:"stale_sections"`
}
```

### ChangedFileRationale

```go
type ChangedFileRationale struct {
    FilePath       string   `json:"file_path"`
    Reason         string   `json:"reason"`
    RelatedACs     []string `json:"related_acs"`
    PlanSteps      []string `json:"plan_steps"`
    RiskLevel      string   `json:"risk_level"`
    ExpectedByPlan bool     `json:"expected_by_plan"`
    Status         string   `json:"status"` // expected, flagged, approved, rejected
}
```

### DeterministicCheckResult

```go
type DeterministicCheckResult struct {
    ID          string   `json:"id"`
    Name        string   `json:"name"`
    Status      string   `json:"status"` // pass, warn, fail, blocked
    Severity    string   `json:"severity"`
    Message     string   `json:"message"`
    Evidence    []string `json:"evidence"`
    Remediation string   `json:"remediation,omitempty"`
}
```

### ReviewerFinding

```go
type ReviewerFinding struct {
    ID          string   `json:"id"`
    Reviewer    string   `json:"reviewer"`
    Severity    string   `json:"severity"`
    Category    string   `json:"category"` // requirement, scope, tests, architecture, security, data
    Title       string   `json:"title"`
    Description string   `json:"description"`
    Evidence    []string `json:"evidence"`
    RelatedACs  []string `json:"related_acs"`
    Confidence  string   `json:"confidence"` // low, medium, high
}
```

### FinalVerdict

```go
type FinalVerdict struct {
    Verdict                string                       `json:"verdict"` // APPROVE, WARN, REJECT, BLOCKED
    RiskLevel              string                       `json:"risk_level"`
    RequirementCoverage    []RequirementCoverageResult  `json:"requirement_coverage"`
    DeterministicChecks    []DeterministicCheckResult    `json:"deterministic_checks"`
    ReviewerFindings       []ReviewerFinding             `json:"reviewer_findings"`
    Disagreements          []ReviewerDisagreement        `json:"disagreements"`
    ChangedFileRationales  []ChangedFileRationale        `json:"changed_file_rationales"`
    TestGaps               []string                      `json:"test_gaps"`
    RecommendedFixPrompt   string                        `json:"recommended_fix_prompt"`
}
```

## 7. Git Operations

Implement `internal/git`.

Required functions:

```go
type Client struct {
    Root string
}

func (c *Client) IsRepo(ctx context.Context) error
func (c *Client) CurrentBranch(ctx context.Context) (string, error)
func (c *Client) ResolveRef(ctx context.Context, ref string) (string, error)
func (c *Client) Fetch(ctx context.Context, remote string, branch string) error
func (c *Client) IsWorkingTreeClean(ctx context.Context) (bool, error)
func (c *Client) ChangedFiles(ctx context.Context, base string) ([]string, error)
func (c *Client) Diff(ctx context.Context, base string) (string, error)
func (c *Client) DiffStat(ctx context.Context, base string) (string, error)
func (c *Client) FileAtRef(ctx context.Context, ref, path string) ([]byte, error)
func (c *Client) MergeBase(ctx context.Context, refA, refB string) (string, error)
```

Use `git diff <base>...HEAD`.

## 8. Snapshot Manager

Command:

```bash
aiguard snapshot update --base develop
```

Algorithm:

1. Load config.
2. Validate git repo.
3. Resolve source branch.
4. Determine current source commit.
5. Load previous snapshot metadata if present.
6. If no previous snapshot, generate full snapshot:
   - Run tree-like traversal.
   - Respect `.gitignore`, `.aiguardignore`, config ignore.
   - Generate folder tree.
   - Generate initial component note placeholders.
7. If previous snapshot exists:
   - Calculate diff between previous snapshot commit and current source commit.
   - Identify changed files.
   - Update affected sections.
   - Append to change-log.md.
8. Ask AI only if configured and needed:
   - Use cheap_fast profile.
   - Prompt AI to summarize changed areas and update map notes.
9. Write snapshot.md and JSON indexes.
10. Write snapshot-meta.json.
11. Audit event.

Snapshot should never include secret file contents.

## 9. Requirement Ingestion

Sources:

### File

```bash
aiguard req ingest --file ticket.md
```

Copy to:

```text
.aiguard/requirements/raw.md
```

### Paste

```bash
aiguard req ingest --paste
```

Open editor or read stdin.

### Jira

Future implementation:

```bash
aiguard req ingest --jira ABC-123
```

Initially, create interface:

```go
type RequirementSource interface {
    Fetch(ctx context.Context, ref string) (*RawRequirement, error)
}
```

## 10. Acceptance Criteria Extraction

Critical requirement:

If the raw ticket contains acceptance criteria, preserve them.

Detection headings:

- Acceptance Criteria
- AC
- Criteria
- Requirements
- Definition of Done
- Expected Behavior
- Business Rules

Parser should do deterministic heading extraction first.

Then AI may normalize.

Output distinction:

```json
{
  "source_acceptance_criteria": [],
  "inferred_acceptance_criteria": [],
  "normalized_criteria": []
}
```

Rules:

1. Source-provided AC must not be dropped.
2. AI-inferred AC must be labeled inferred.
3. If source AC is ambiguous, keep original and add clarification question.
4. Non-goals must be extracted separately.
5. Final normalized AC IDs must be stable.

## 11. Context Pack Builder

The context pack builder decides what to send to AI.

Inputs:

- Step type.
- Token budget.
- Snapshot.
- Raw requirement.
- Digest.
- Acceptance criteria.
- Relevant changed files.
- Git diff.
- Config rules.
- Prior artifacts.

Rules:

1. Always prefer snapshot over full code when planning.
2. Use actual code files only when evidence is needed.
3. Use git diff as source of truth for final verification.
4. Redact secrets.
5. Enforce max file size.
6. Include only targeted files where possible.
7. Include metadata about source branch and commit.

Context pack model:

```go
type ContextPack struct {
    Step          string
    TokenBudget   int
    SystemContext string
    UserPrompt    string
    Files         []ContextFile
    Artifacts     []ContextArtifact
    Redactions    []Redaction
}
```

## 12. Agent Interface

```go
type Agent interface {
    Name() string
    CheckAvailable(ctx context.Context) error
    Run(ctx context.Context, req AgentRequest) (*AgentResponse, error)
}
```

```go
type AgentRequest struct {
    Step         string
    Prompt       string
    WorkDir      string
    ModelProfile ModelProfile
    Timeout      time.Duration
    Env          map[string]string
}

type AgentResponse struct {
    AgentName    string
    Step         string
    Stdout       string
    Stderr       string
    ExitCode     int
    Duration     time.Duration
    OutputPath   string
}
```

### Claude Code Adapter

Must call installed CLI.

Do not call Anthropic API directly.

Pseudo implementation:

```go
args := []string{}
if cfg.PromptArg != "" {
    args = append(args, cfg.PromptArg)
}
if model template configured {
    args = append(args, render(model_arg_template, profile))
}
if effort template configured {
    args = append(args, render(effort_arg_template, profile))
}
cmd := exec.CommandContext(ctx, cfg.Command, args...)
cmd.Dir = req.WorkDir
cmd.Stdin = strings.NewReader(req.Prompt)
```

If CLI does not support configured flags, allow user to change config.

### Codex Adapter

Same command-runner abstraction.

Do not assume fixed flags. Use configurable templates.

## 13. Prompt Templates

Prompts must force structured output.

### Digest Prompt Requirements

The digest prompt must instruct AI:

- Preserve source-provided acceptance criteria exactly.
- Do not drop or rewrite source AC without preserving original.
- Clearly mark inferred criteria.
- Extract non-goals separately.
- Create blocking questions where ambiguity exists.
- Output JSON and Markdown.

### Guidance Prompt Requirements

- Use digest and snapshot.
- Do not invent unnecessary work.
- Map every recommendation to AC IDs.
- Identify do-not-touch areas.
- Identify required tests.
- Identify risks.

### Plan Prompt Requirements

- Each step must map to AC IDs.
- Each expected file change must have rationale.
- Include tests.
- Do not add unrelated refactors.
- Keep implementation narrow.

### Final Review Prompt Requirements

- Review actual diff.
- Check every AC.
- Check explicit non-goals.
- Flag missing tests.
- Flag scope drift.
- Flag risky files.
- Output structured findings.
- Do not assume code is correct because summary says so.

## 14. Deterministic Check Engine

Interface:

```go
type Check interface {
    ID() string
    Run(ctx context.Context, input CheckInput) ([]DeterministicCheckResult, error)
}
```

Required checks:

### Working Tree Check

Warn if uncommitted unrelated changes exist before starting.

### Snapshot Freshness Check

Fail or warn if snapshot commit differs from source branch by threshold.

Config:

```yaml
snapshot:
  stale_after_commits: 20
```

### Forbidden Path Check

Block if changed files match forbidden paths.

### Risk Path Check

Warn/high if changed files match risk paths.

### Diff Size Check

Warn if diff exceeds configured thresholds.

### Dependency Change Check

Warn/high if dependency files changed.

### Migration Check

Warn/high if migration paths changed.

### Test Presence Check

Warn if behavior files changed but no tests changed.

### Rationale Check

Fail if changed files lack rationale.

### Plan Drift Check

Warn/fail if changed files were not expected by approved plan.

### AC Traceability Check

Fail if an AC has no planned implementation and no documented non-applicability reason.

### Secret Redaction Check

Block if diff contains likely secrets.

## 15. Reviewer Aggregation

Inputs:

- Deterministic checks.
- Reviewer findings from Claude.
- Reviewer findings from Codex.
- Optional local reviewer.
- Final diff.
- AC list.

Algorithm:

1. Normalize reviewer findings.
2. Group by category and related AC.
3. Identify agreements.
4. Identify disagreements.
5. Give deterministic evidence higher weight.
6. If any critical deterministic check fails, verdict BLOCKED.
7. If any source-provided AC is missing, verdict REJECT.
8. If high-risk file changed without rationale, verdict REJECT or WARN depending config.
9. If tests fail, verdict REJECT.
10. If reviewers disagree on core AC, verdict WARN at minimum.
11. If only minor findings, verdict WARN or APPROVE.
12. Generate recommended fix prompt.

## 16. Final Verdict Rules

Default rules:

### BLOCKED

- Forbidden path changed.
- Secret detected.
- Required prior gate missing.
- Tests failed and config blocks on test failure.

### REJECT

- Source-provided AC not implemented.
- Explicit non-goal violated.
- High-risk file changed with no rationale.
- Critical or high reviewer finding supported by evidence.
- Plan drift touches critical/high risk area.

### WARN

- Reviewer disagreement.
- Missing non-critical tests.
- Medium scope drift.
- Unclear requirement ambiguity.
- Snapshot stale but not critical.

### APPROVE

- All source AC satisfied.
- No critical/high findings.
- Tests pass or test status documented.
- All changed files have rationale.
- No unapproved high-risk changes.
- No unresolved blocking questions.

## 17. Reports

### Markdown Final Verdict

Sections:

1. Verdict.
2. Executive summary.
3. Requirement coverage.
4. Traceability matrix.
5. Deterministic checks.
6. Reviewer findings.
7. Reviewer disagreements.
8. Changed file rationale.
9. Test adequacy.
10. Scope drift.
11. Risk areas.
12. Recommended fix prompt.
13. PR readiness checklist.

### JSON Final Verdict

Must match `FinalVerdict` struct.

## 18. CLI Command Specifications

### `aiguard init`

Creates `.aiguard/config.yaml`.

Flags:

```bash
--source-branch develop
--force
```

### `aiguard snapshot update`

Flags:

```bash
--base develop
--no-ai
--full
```

### `aiguard req ingest`

Flags:

```bash
--file path
--paste
--jira key
```

### `aiguard digest create`

Flags:

```bash
--reviewer claude
--local-only
```

### `aiguard digest approve`

Writes gate approval.

Flags:

```bash
--reason string
```

### `aiguard guidance create`

Requires digest approved unless `--force`.

### `aiguard guidance review`

Flags:

```bash
--reviewers claude,codex
```

### `aiguard plan create`

Requires guidance approved unless `--force`.

### `aiguard plan review`

Flags:

```bash
--reviewers claude,codex
```

### `aiguard implement`

Flags:

```bash
--agent claude
--step step-id
```

### `aiguard checkpoint`

Analyzes current diff and writes checkpoint.

### `aiguard verify`

Flags:

```bash
--base develop
--reviewers claude,codex
--local-only
--skip-tests
--run-tests
```

### `aiguard doctor`

Checks:

- Git installed.
- In git repo.
- Source branch exists.
- Claude CLI availability.
- Codex CLI availability.
- Config valid.
- `.aiguard` writable.
- Test commands valid.

## 19. Error Handling

Errors must be actionable.

Bad:

```text
failed
```

Good:

```text
Cannot run final verification because no requirement digest exists.
Run: aiguard digest create
```

## 20. Testing Requirements

Unit tests:

- Config loading.
- Glob path matching.
- AC parser.
- Snapshot metadata.
- Git command wrapper with fake executor.
- Deterministic checks.
- Verdict aggregation.
- Prompt rendering.
- Report rendering.

Integration tests:

- Temporary git repo.
- Init.
- Snapshot.
- Requirement ingest.
- Diff creation.
- Verify.
- Report generation.

Use mock agents for tests.

## 21. Security Requirements

- Redact secret patterns before AI context.
- Never include forbidden files.
- Respect `.aiguardignore`.
- Log external AI usage.
- Warn when `allow_external_ai` is false but external reviewers requested.
- Do not store API keys in config.
- Read tokens from environment only if adapters need them.
- Do not print secrets.

Secret patterns:

- AWS keys.
- Private keys.
- Bearer tokens.
- GitHub tokens.
- Slack tokens.
- Generic `password=`, `api_key=`, `secret=`.

## 22. Performance Requirements

- Should run deterministic checks on large diffs within seconds.
- Snapshot update should use diff-based incremental update.
- Avoid loading entire repo into memory.
- Stream large command outputs.
- Enforce max prompt size.
- Write large diffs to temp files if needed.

## 23. Implementation Milestones

### Milestone 1 — Foundation

- CLI skeleton.
- Config.
- `.aiguard` project layout.
- Audit logger.
- Git wrapper.
- Doctor command.

### Milestone 2 — Deterministic Verify

- Changed files.
- Diff analyzer.
- Risk path checks.
- Forbidden path checks.
- Test presence checks.
- Markdown final report.

### Milestone 3 — Requirement Artifacts

- Requirement ingestion.
- Acceptance criteria extraction.
- Digest generation with AI adapter.
- Digest approval gate.

### Milestone 4 — Snapshot

- Snapshot creation.
- Snapshot metadata.
- Incremental update.
- Snapshot status.

### Milestone 5 — Guidance/Plan Gates

- Guidance create/review.
- Plan create/review.
- Gate transitions.
- User approvals.

### Milestone 6 — Agent Adapters

- Claude Code adapter.
- Codex adapter.
- Model profile routing.
- Prompt templates.

### Milestone 7 — Final Verification

- Context pack builder.
- Multi-reviewer final review.
- Disagreement report.
- JSON final verdict.
- Recommended fix prompt.

### Milestone 8 — Implementation Checkpoints

- Changed file rationale.
- Checkpoint command.
- Plan drift detection.

### Milestone 9 — Polish

- Better errors.
- Tests.
- README.
- Examples.
- TUI optional.

## 24. Acceptance Criteria for AIGuard Itself

AIGuard is acceptable when:

1. It can initialize in a git repo.
2. It can create a snapshot from `develop`.
3. It can ingest a requirement file.
4. It preserves ticket-provided acceptance criteria separately from inferred criteria.
5. It can generate a digest.
6. It can require user approval gates.
7. It can generate tech guidance.
8. It can generate an implementation plan.
9. It can analyze `git diff develop...HEAD`.
10. It can flag forbidden and high-risk changed paths.
11. It can detect when tests were not changed.
12. It can require changed-file rationale.
13. It can invoke Claude Code CLI through configurable command templates.
14. It can invoke Codex CLI through configurable command templates.
15. It can run multi-reviewer final verification.
16. It can produce reviewer disagreement reports.
17. It can output final verdict markdown and JSON.
18. It can produce a recommended fix prompt.
19. It can run in local-only deterministic mode.
20. It redacts secrets before AI context.
