# CLAUDE.md — Instructions for Building AIGuard

You are building **AIGuard**, a local-first requirements-gated AI development harness.

AIGuard prevents AI-generated code from violating business requirements by creating a gated workflow:

```text
snapshot → requirement digest → clarification → test strategy → technical guidance → guidance review → plan → plan review → implementation checkpoints → final spec-to-diff verification
```

## Project Mission

Build a CLI tool that a professional developer can run inside a work codebase before submitting AI-generated code.

The tool must help answer:

> Does this actual git diff satisfy the business requirement and acceptance criteria?

Do not build a generic chatbot. Do not build an IDE. Do not build a SaaS. Build a local CLI-first verification harness.

## Non-Negotiable Principles

1. **The git diff is the source of truth for implementation review.**
2. **Ticket-provided acceptance criteria must be preserved.**
3. **AI-inferred acceptance criteria must be labeled as inferred.**
4. **Every changed file should have a rationale.**
5. **Every plan step should map to acceptance criteria.**
6. **Every acceptance criterion should map to implementation or a documented reason.**
7. **Use deterministic checks before AI review.**
8. **Use AI only where reasoning is needed.**
9. **Do not trust one AI reviewer blindly.**
10. **The harness owns state, gates, evidence, and verdicts.**
11. **Local-first and privacy-safe by default.**
12. **Do not upload source code unless explicitly configured.**
13. **Do not hardcode vendor model names or flags. Use config templates.**

## Recommended Stack

- Language: Go 1.22+
- CLI: Cobra
- Config: YAML
- Reports: Markdown + JSON
- Git: shell out to `git`
- Tests: Go testing package
- Optional future TUI: Bubble Tea

## Build the Full Tool, Not Just a Toy MVP

Implement the architecture in milestones, but keep the full system design intact.

Required command families:

```bash
aiguard init

aiguard snapshot update
aiguard snapshot status

aiguard req ingest
aiguard req status

aiguard digest create
aiguard digest review
aiguard digest approve

aiguard clarify

aiguard tests strategy
aiguard tests review
aiguard tests approve

aiguard guidance create
aiguard guidance review
aiguard guidance approve

aiguard plan create
aiguard plan review
aiguard plan approve

aiguard implement
aiguard checkpoint

aiguard rationale update
aiguard rationale check

aiguard verify

aiguard status
aiguard report
aiguard doctor
```

## Required Repository Structure

Create this structure:

```text
aiguard/
  cmd/
    aiguard/
      main.go

  internal/
    app/
    cli/
    config/
    project/
    git/
    snapshot/
    requirements/
    gates/
    testspec/
    guidance/
    plan/
    checks/
    agents/
    contextpack/
    review/
    report/
    audit/
    prompt/

  examples/
  docs/
  CLAUDE.md
  README.md
  go.mod
```

Keep packages small and testable.

## Target Project `.aiguard` Layout

When `aiguard init` runs, create:

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

## Config Requirements

Implement `.aiguard/config.yaml` with these sections:

```yaml
project:
  name: ""
  source_branch: develop
  default_compare_ref: develop

privacy:
  default_mode: local_first
  allow_external_ai: false
  redact_secrets: true
  include_actual_files_by_default: false
  max_file_bytes_for_context: 50000

snapshot:
  include_tree: true
  include_component_notes: true
  update_from_diff: true
  max_tree_depth: 8
  stale_after_commits: 20
  ignore:
    - ".git/**"
    - "node_modules/**"
    - "vendor/**"
    - "dist/**"
    - "build/**"
    - ".aiguard/**"

risk_paths:
  critical:
    - ".env"
    - ".env.*"
    - "secrets/**"
    - "infra/prod/**"
  high:
    - "auth/**"
    - "billing/**"
    - "payments/**"
    - "permissions/**"
    - "migrations/**"
    - "database/**"
    - "db/migrate/**"

forbidden_paths:
  - ".env"
  - ".env.*"
  - "secrets/**"
  - "*.pem"
  - "*.key"

test_commands:
  - name: unit
    command: "go test ./..."
    timeout_seconds: 600

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

Important: CLI flags for external tools may change. Do not hardcode Claude/Codex flags. Use configurable templates.

## Important User Workflow Details to Preserve

The user’s current manual workflow has these ideas. Build them into AIGuard.

### 1. Source-of-truth branch

The work codebase uses `develop` as the source-of-truth branch. AIGuard must support configurable source branches.

### 2. Living codebase map

The snapshot is not a substitute for the code. It is a token-saving map of where folders/files/components live and what they do.

Use snapshot to route context. Use actual files and git diff when evidence is required.

### 3. Incremental snapshot update

When many tickets have merged, update the snapshot by diffing the old source commit against the new source commit.

### 4. Ticket acceptance criteria

If a ticket already defines acceptance criteria, preserve them and treat them as authoritative. Do not replace them with AI-generated criteria.

### 5. Model routing

The user prefers different model strengths for different steps:

- Snapshot update: cheap/fast model.
- Brainstorm, tech guidance, plan, code review: strongest reasoning model, high effort.
- Coding: balanced coding model, medium effort.

Implement this as configurable routing profiles, not hardcoded names.

### 6. Multi-reviewer bias reduction

Claude may create and review its own work. AIGuard must support independent reviewers such as Codex to reduce same-model bias.

### 7. Gates

Brainstorm, technical guidance, plan, implementation, and code review must become gates with PASS/WARN/FAIL/BLOCKED statuses.

## Implementation Instructions

### Step 1 — CLI Skeleton

Build the Cobra CLI with all command stubs.

Each command must:

- Load config.
- Validate project root.
- Write audit log.
- Return actionable errors.

### Step 2 — Config

Implement:

- Default config creation.
- YAML load.
- Validation.
- Path expansion.
- Glob pattern validation.
- Config-driven model routing.
- Config-driven external agent command templates.

### Step 3 — Git Package

Implement git wrapper.

Required commands:

- current branch
- resolve ref
- merge-base
- changed files
- diff
- diff stat
- is working tree clean
- file content at ref
- fetch

Use context timeouts.

### Step 4 — Audit Log

Every command writes a JSONL event:

```json
{
  "timestamp": "...",
  "command": "verify",
  "git_branch": "...",
  "git_head": "...",
  "source_branch": "develop",
  "source_commit": "...",
  "inputs": [],
  "outputs": [],
  "status": "success"
}
```

### Step 5 — Snapshot Manager

Implement `aiguard snapshot update`.

Full snapshot:

- Walk directory tree.
- Respect ignore rules.
- Generate markdown tree.
- Generate file-map.json.
- Generate snapshot-meta.json.

Incremental snapshot:

- Load previous snapshot commit.
- Resolve current source branch commit.
- Diff old vs new.
- Update change-log.md.
- Mark changed modules.
- Optionally call AI profile `cheap_fast` to summarize changed areas.

Snapshot markdown should include:

```markdown
# Codebase Snapshot

## Metadata

- Source branch:
- Source commit:
- Updated at:
- Snapshot version:

## Folder Tree

## Component Notes

## Architecture Patterns

## High-Risk Areas

## Test Locations

## Recent Changes
```

### Step 6 — Requirement Ingestion

Implement `aiguard req ingest`.

For file input, copy content to `.aiguard/requirements/raw.md`.

For paste input, read stdin.

For Jira input, create interface but allow stub implementation.

### Step 7 — Acceptance Criteria Parser

Implement deterministic extraction first.

Detect headings:

- Acceptance Criteria
- AC
- Criteria
- Requirements
- Definition of Done
- Expected Behavior
- Business Rules

Extract bullet lists, numbered lists, checkboxes.

Store source-provided criteria as:

```json
{
  "source": "ticket_acceptance_criteria"
}
```

Then allow AI digest to infer additional criteria, labeled:

```json
{
  "source": "inferred"
}
```

Never drop source-provided AC.

### Step 8 — Agent Layer

Implement generic command-runner agents.

Do not call APIs directly.

The agent adapter must:

- Check command exists.
- Build args from templates.
- Pass prompt through stdin where supported.
- Capture stdout, stderr, exit code, duration.
- Write raw output to artifact file.
- Respect timeout.

### Step 9 — Prompt Templates

Create prompt templates for:

- Requirement digest.
- Clarification questions.
- Test strategy.
- Technical guidance.
- Guidance review.
- Plan.
- Plan review.
- Checkpoint review.
- Final verification.

All prompts must request both markdown-friendly content and structured JSON where possible.

### Step 10 — Gates

Implement gate state in:

```text
.aiguard/gates.json
```

Gate model:

```json
{
  "digest": {
    "status": "APPROVED_BY_USER",
    "approved_at": "...",
    "approved_by": "local-user",
    "reason": "Looks correct"
  }
}
```

Commands that require previous gates must enforce them.

### Step 11 — Deterministic Checks

Implement checks:

- forbidden path
- risk path
- diff size
- dependency changes
- migration changes
- test presence
- rationale missing
- plan drift
- AC traceability
- secret detection
- snapshot freshness

All checks output structured JSON.

### Step 12 — Guidance and Plan

Implement:

```bash
aiguard guidance create
aiguard guidance review
aiguard guidance approve

aiguard plan create
aiguard plan review
aiguard plan approve
```

Guidance and plan must map to AC IDs.

Review must fail if source-provided ACs are missing.

### Step 13 — Changed File Rationale

Implement:

```bash
aiguard rationale update
aiguard rationale check
```

This should inspect changed files and create/update:

```text
.aiguard/implementation/changed-files-rationale.md
```

If a file cannot be mapped to AC or plan step, flag it.

### Step 14 — Final Verification

Implement:

```bash
aiguard verify --base develop --reviewers claude,codex
```

Process:

1. Load config.
2. Resolve base.
3. Run deterministic checks.
4. Build context pack.
5. Run configured reviewers.
6. Parse reviewer findings.
7. Aggregate findings.
8. Detect disagreements.
9. Produce final verdict.
10. Write markdown and JSON reports.

Final verdict rules:

- BLOCKED if forbidden path or secret detected.
- REJECT if source-provided AC missing.
- REJECT if explicit non-goal violated.
- REJECT if high-risk file lacks rationale.
- WARN if reviewer disagreement on core requirement.
- WARN if tests are missing but not required to block.
- APPROVE only if all must-have ACs are covered and no high/critical issues exist.

### Step 15 — Recommended Fix Prompt

Every WARN/REJECT/BLOCKED report must include a prompt the developer can paste back into Claude Code/Codex.

Example:

```text
Revise the current patch. Keep these working changes: ...
Fix these missing requirements: ...
Revert these unrelated changes: ...
Add these tests: ...
Do not modify: ...
```

This is a core feature.

## Prompt Requirements

When creating prompts for AI reviewers, include these rules:

- You are a reviewer, not an implementer.
- Do not assume the patch is correct.
- The git diff is the source of truth.
- Source-provided acceptance criteria are authoritative.
- Label missing requirements clearly.
- Flag unrelated changes.
- Flag risky files.
- Flag insufficient tests.
- Provide evidence from diff paths or artifacts.
- Output severity: Critical, High, Medium, Minor, Info.
- Output final recommendation: APPROVE, WARN, REJECT, BLOCKED.

## Code Quality Requirements

- Keep functions small.
- Avoid global state.
- Use interfaces for filesystem, git runner, and agent runner where practical.
- Make deterministic checks testable without real AI.
- Do not let AI output parsing crash the whole command.
- If AI output is malformed, include raw output and produce WARN.
- Use clear actionable errors.
- Add tests for every check.

## Testing Strategy

Write tests for:

- Config loading.
- Init command.
- Git wrapper with fake command runner.
- AC extraction from markdown headings.
- Source AC preservation.
- Glob matching.
- Forbidden path check.
- Risk path check.
- Test presence check.
- Secret detection.
- Verdict aggregation.
- Disagreement detection.
- Markdown report rendering.

Integration test:

1. Create temp git repo.
2. Commit base files.
3. Run `aiguard init`.
4. Run `aiguard snapshot update`.
5. Ingest sample ticket with AC.
6. Modify files.
7. Run `aiguard verify --local-only`.
8. Assert final report exists and flags expected issues.

## Do Not Do These Things

- Do not build a cloud backend.
- Do not require an API key.
- Do not upload code by default.
- Do not hardcode Claude model names into business logic.
- Do not assume `main` is the source branch.
- Do not ignore ticket-provided acceptance criteria.
- Do not let AI reviewer output override deterministic blocked checks.
- Do not treat snapshot.md as proof of code behavior.
- Do not include `.env`, secrets, private keys, or tokens in prompts.
- Do not create vague reports. Every finding needs evidence and remediation.

## Example Final Report Shape

```markdown
# AIGuard Final Verdict

## Verdict

REJECT

## Summary

The patch partially implements the requested feature but misses AC3 and modifies a high-risk file without approved rationale.

## Requirement Coverage

| AC | Source | Status | Evidence | Notes |
|---|---|---|---|---|
| AC1 | ticket | PASS | ui/form.tsx | User can enter reason |
| AC2 | ticket | PASS | service.go | Reason passed to service |
| AC3 | ticket | FAIL | none | No persistence found |

## Risk Findings

- High-risk file changed: billing/refund_policy.go
- No related AC or plan step found.

## Test Gaps

- No test proves AC3.
- No regression test for refund behavior.

## Reviewer Disagreements

Claude approved AC2.
Codex warned AC2 may not persist to DB.
Harness conclusion: WARN because no migration/repository diff found.

## Recommended Fix Prompt

Revise the patch to implement AC3 persistence. Revert billing/refund_policy.go unless explicitly required. Add tests for persistence and refund behavior unchanged.
```

## Build Order

Implement in this order:

1. CLI skeleton and config.
2. Init and project layout.
3. Git wrapper.
4. Deterministic checks.
5. Requirement ingestion and AC parser.
6. Local-only final verification.
7. Snapshot manager.
8. Agent command runner.
9. Digest/guidance/plan prompts.
10. Gate system.
11. Multi-reviewer final verification.
12. Disagreement report.
13. Implementation checkpoints.
14. Polish and tests.

Even though this is the build order, keep the full architecture in place from the start.

## Final Reminder

AIGuard exists because AI-generated code can look correct while violating business requirements.

The final product must help a developer avoid submitting bad AI-generated code at work.

Be strict. Be evidence-based. Be local-first. Prefer deterministic proof over AI confidence.
