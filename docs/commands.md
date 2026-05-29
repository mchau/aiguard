# Command Reference

Complete reference for all AIGuard commands with examples and output.

## Global Flags

```bash
aiguard --config /path/to/config.yaml <command>
  # Override the default .aiguard/config.yaml location
```

---

## init

**Initialize an AIGuard workspace in the current project.**

```bash
$ aiguard init
Initialized AIGuard workspace at /home/alice/myrepo/.aiguard/config.yaml
Source branch: main
Next step: run 'aiguard snapshot update' to map the codebase.

$ ls -la .aiguard/
config.yaml
.gitignore
snapshot/
requirements/
guidance/
...
```

### Flags

```bash
--source-branch <ref>
  # Explicitly set the source branch (default: autodetected from origin/HEAD, falls back to main/master)
  aiguard init --source-branch develop

--force
  # Overwrite an existing .aiguard/ workspace
  aiguard init --force
```

### Pain Points

- If no remote is configured, init falls back to detecting `main` or `master` locally.
- `.aiguard/.gitignore` is written automatically to exclude transient files.

---

## snapshot

**Manage the codebase map (used to guide which files are relevant for context).**

### snapshot update

**Create or refresh the codebase snapshot (tree structure, metadata, file map).**

```bash
$ aiguard snapshot update --full
Snapshot written to .aiguard/snapshot/
- snapshot.md (tree + metadata)
- snapshot-meta.json (version, commit, timestamp)
- file-map.json (file → component mapping)
- risk-map.json (high-risk file locations)
- test-map.json (test file locations)

$ head .aiguard/snapshot/snapshot.md
# Codebase Snapshot
...
## Folder Tree
.
├── cmd/
│   └── myapp/
│       └── main.go
├── internal/
...
```

### Flags

```bash
--full
  # Full snapshot (scans directory tree, slow but comprehensive)
  aiguard snapshot update --full

--no-ai
  # Skip optional AI summarization of changed areas
  aiguard snapshot update --full --no-ai
```

### snapshot status

**Check snapshot freshness.**

```bash
$ aiguard snapshot status
Snapshot at commit abc123de (version 1)
Created: 2025-05-28 14:22:00
Status: fresh (0 commits since snapshot)
```

### Pain Points

- Snapshot staleness warning fires after `config.snapshot.stale_after_commits` commits (default 20). Update periodically if your codebase is fast-moving.
- The snapshot is a **map**, not a substitute for reading code. Reviewers should reference actual files for verification.

---

## req (Requirements)

**Manage requirement ingestion and tracking.**

### req ingest

**Ingest a requirement/ticket and extract acceptance criteria.**

```bash
$ aiguard req ingest --file ticket.md
Requirement ingested to .aiguard/requirements/raw.md

$ cat ticket.md
# Implement User Profiles

## Acceptance Criteria
- AC1: Users can create a profile with name, email, avatar
- AC2: Profile data is persisted to the database
- AC3: Profiles are queryable by user ID

$ cat .aiguard/requirements/raw.md
# Implement User Profiles
...
(identical to ticket.md)
```

### Flags

```bash
--file <path>
  # Ingest from a file
  aiguard req ingest --file ticket.md

--paste
  # Ingest from stdin (paste mode)
  aiguard req ingest --paste
  # Type or paste the requirement, then Ctrl+D

--jira <issue-id>
  # Ingest from Jira (not yet implemented; returns stub)
  aiguard req ingest --jira PROJ-123
```

### req status

**Check ingestion status.**

```bash
$ aiguard req status
raw.md found (242 bytes)
```

### Pain Points

- Only the raw markdown is stored, not metadata like ticket ID or assignee. Add those to the markdown itself if you need to reference them later.

---

## digest

**Extract and normalize acceptance criteria.**

### digest create

**Extract source ACs from raw.md and optionally infer additional ACs via AI.**

```bash
$ aiguard digest create
Extracted 3 source acceptance criteria.

$ cat .aiguard/requirements/acceptance-criteria.json
[
  {"id":"AC1","text":"Users can create a profile with name, email, avatar","source":"ticket_acceptance_criteria"},
  {"id":"AC2","text":"Profile data is persisted to the database","source":"ticket_acceptance_criteria"},
  {"id":"AC3","text":"Profiles are queryable by user ID","source":"ticket_acceptance_criteria"}
]

$ cat .aiguard/requirements/digest.md
# Requirement Digest
Extracted 3 acceptance criteria from the ticket.
...
```

### Flags

```bash
--local-only
  # Skip AI inference; use only deterministic AC extraction
  aiguard digest create --local-only
```

### digest review

**Mark the digest as needing review.**

```bash
$ aiguard digest review --reason "Reviewed by PM"
Digest gate set to NEEDS_REVIEW.
```

### digest approve

**Approve the digest (required before guidance creation).**

```bash
$ aiguard digest approve --reason "All ACs align with spec"
Digest gate APPROVED_BY_USER.
Next step: run 'aiguard guidance create'.

$ cat .aiguard/gates.json
{
  "digest": {
    "status": "APPROVED_BY_USER",
    "approved_by": "local-user",
    "approved_at": "2025-05-28T14:30:00Z",
    "reason": "All ACs align with spec"
  }
}
```

### Pain Points

- Source-provided ACs (from the ticket) are labeled `"source":"ticket_acceptance_criteria"` and are **authoritative** — they are never dropped or replaced by AI inference.
- Inferred ACs are labeled `"source":"inferred"` and are advisory.
- AC IDs are normalized for matching (AC1, ac-1, AC_001 all become "ac1"). Plan steps should use these stable IDs.

---

## guidance

**Technical guidance from approved digest.**

### guidance create

**Generate technical guidance (AI or manual).**

```bash
$ aiguard guidance create
Guidance written to .aiguard/guidance/
- tech-guidance.md (implementation approach)
- tech-guidance.json (structured guidance)

$ head .aiguard/guidance/tech-guidance.md
# Technical Guidance

## Architecture
- Use a REST API with /profiles/{id} endpoint
- Store profiles in the users table with a jsonb column
...
```

Requires: `digest` gate to be `APPROVED_BY_USER` (or `--force --reason "..."`).

### Flags

```bash
--force --reason "reason"
  # Skip the digest approval gate check (not recommended)
  aiguard guidance create --force --reason "PM approved verbally"
```

### guidance review

**Mark guidance as needing review.**

```bash
$ aiguard guidance review --reason "Awaiting technical review"
Guidance gate set to NEEDS_REVIEW (reviewer: claude_code).
```

### guidance approve

**Approve the guidance (required before plan creation).**

```bash
$ aiguard guidance approve --reason "Architecture reviewed and approved"
Guidance gate APPROVED_BY_USER.
Next step: run 'aiguard plan create'.
```

### Pain Points

- Guidance is optional if you're just using deterministic checks. You can skip it entirely and go straight to `verify --local-only`.
- The gate system is enforced; `--force --reason` bypasses but is recorded in the audit log.

---

## plan

**Implementation plan from approved guidance.**

### plan create

**Generate an implementation plan with steps mapped to ACs.**

```bash
$ aiguard plan create
Plan written to .aiguard/plans/

$ cat .aiguard/plans/plan.json
{
  "steps": [
    {
      "id": "S1",
      "ac_ids": ["AC1"],
      "expected_files": ["internal/handlers/user.go", "internal/handlers/user_test.go"],
      "description": "Create profile handler"
    },
    {
      "id": "S2",
      "ac_ids": ["AC2"],
      "expected_files": ["internal/db/migrations/001_profiles.sql", "internal/repository/profile.go"],
      "description": "Add profile persistence"
    },
    {
      "id": "S3",
      "ac_ids": ["AC3"],
      "expected_files": ["internal/repository/profile.go"],
      "description": "Implement profile query"
    }
  ]
}
```

Requires: `guidance` gate to be `APPROVED_BY_USER` (or `--force --reason "..."`).

### plan review

**Mark plan as needing review.**

```bash
$ aiguard plan review --reason "Ready for technical review"
Plan gate set to NEEDS_REVIEW (reviewer: claude_code).
```

### plan approve

**Approve the plan (required before implementation).**

```bash
$ aiguard plan approve --reason "Plan aligns with architecture"
Plan gate APPROVED_BY_USER.
Next step: run 'aiguard prompt' to assemble the implementation prompt, then 'aiguard checkpoint' as you go.
```

### Pain Points

- `plan_drift` check warns if risk-path files are changed without being mentioned in the plan. Auxiliary files (test files, routers) are not checked.
- Plan AC IDs must match acceptance-criteria.json IDs (after normalization). The system handles formatting drift (AC1 vs ac-1) automatically.

---

## prompt

**Assemble the implementation prompt for your AI tool.**

**Generate a paste-ready prompt with approved digest, guidance, and plan.**

```bash
$ aiguard prompt
You are implementing a software feature. Follow the plan exactly.

## Acceptance Criteria
[
  {"id":"AC1",...},
  ...
]

## Technical Guidance
# Technical Guidance
...

## Plan
# Implementation Plan
...

## Current Task
Implement all plan steps.
```

### Flags

```bash
--step <id>
  # Focus on a specific plan step
  aiguard prompt --step S1

--copy
  # Copy the prompt to the clipboard (macOS/Linux)
  aiguard prompt --copy
  # Prompt is now in your clipboard; paste it into Claude, Codex, etc.

--out <file>
  # Write the prompt to a file instead of stdout
  aiguard prompt --out prompt.txt
```

### Usage

AIGuard does not edit your code. The prompt is meant to be pasted into Claude Code or Codex in your terminal. The AI edits files; you commit the changes.

```bash
# Typical flow:
$ aiguard prompt --copy
$ claude -p  # Paste the prompt (Cmd+V) and implement
# ...Claude makes changes...
$ git diff
# ...review...
$ git add . && git commit -m "feat: implement user profiles"
```

### Pain Points

- `prompt` is a **helper**, not an agent invocation. AIGuard doesn't run your agent or manage its output.
- For unattended workflows, use `verify --auto-fix` instead, which invokes your configured agent automatically.

---

## rationale

**Document the rationale for each changed file.**

### rationale update

**Add or update reasons for changed files.**

```bash
$ aiguard rationale update --file rationale.yaml
Rationale updated: .aiguard/implementation/changed-files-rationale.md

$ cat rationale.yaml
internal/handlers/user.go: Implements AC1 — profile creation endpoint
internal/db/migrations/001_profiles.sql: Implements AC2 — database schema
internal/repository/profile.go: Implements AC2 and AC3 — persistence and query

$ cat .aiguard/implementation/changed-files-rationale.md
# Changed Files Rationale

## internal/handlers/user.go

**Reason:** Implements AC1 — profile creation endpoint

## internal/db/migrations/001_profiles.sql

**Reason:** Implements AC2 — database schema

...
```

### Modes

Rationale update requires **one of three modes** (no stdin prompting by default to keep the command scriptable):

```bash
--file <path>
  # Load rationale from a YAML or JSON file
  aiguard rationale update --file rationale.yaml

--ai
  # Delegate to your configured AI agent
  aiguard rationale update --ai
  # Runs the guidance agent to generate one-line rationales for each file

--interactive
  # Prompt interactively via stdin
  aiguard rationale update --interactive
  # Use only for local development; breaks in CI/scripts
```

### rationale check

**Verify that all changed files have rationale entries.**

```bash
$ aiguard rationale check
[High] internal/api/v2.go: file "internal/api/v2.go" not found in rationale
Error: rationale check failed: 1 file(s) missing rationale
```

### Pain Points

- Rationale is triggered as a check by `verify`. If you skip it (by force-gating upstream commands), `rationale check` will still fail.
- Use `--file` or `--ai` for CI; `--interactive` is only safe locally.

---

## tests

**Manage test strategy and requirements.**

### tests strategy

**Generate a test strategy mapping ACs to test types.**

```bash
$ aiguard tests strategy
Test strategy written to .aiguard/tests/

$ cat .aiguard/tests/test-strategy.md
# Test Strategy

## AC1: Users can create a profile with name, email, avatar
- **Unit**: Handler tests for `/profiles` POST endpoint
- **Integration**: Database persistence tests
- **E2E**: Full profile creation flow

## AC2: Profile data is persisted
- **Unit**: Repository layer tests
- **Integration**: DB transaction tests
...
```

### tests review

**Mark test strategy as needing review.**

```bash
$ aiguard tests review
Test gate set to NEEDS_REVIEW (reviewer: claude_code).
```

### tests approve

**Approve the test strategy.**

```bash
$ aiguard tests approve --reason "Test coverage plan reviewed"
Test gate APPROVED_BY_USER.
```

### Pain Points

- Test strategy is **planning**, not enforcement. The harness doesn't run your tests; it only warns when behavior files change without test files.
- The `test_presence` deterministic check warns on all behavior files without a `_test` file, even if tests exist elsewhere.

---

## checkpoint

**Capture mid-implementation checkpoints with a quick verdict.**

**Create a checkpoint report of current diff against the plan.**

```bash
$ aiguard checkpoint
Checkpoint 1: WARN
Report: .aiguard/implementation/checkpoints/checkpoint-001.md

$ cat .aiguard/implementation/checkpoints/checkpoint-001.md
# Checkpoint 1

**Verdict:** WARN

## Diff Stat

 internal/handlers/user.go      | 45 +++++++++++
 internal/handlers/user_test.go | 38 ++++++++++
 ...

## Deterministic Findings

- [WARN] test_presence: behavior files changed without any test file changes: internal/db/migrations/...
```

Checkpoints increment automatically (checkpoint-001, checkpoint-002, etc.).

### Usage

```bash
# During implementation, run checkpoints to verify you're on track:
$ git add some_files && git commit -m "wip: start handlers"
$ aiguard checkpoint  # checkpoint-001 created

$ git add more_files && git commit -m "wip: add tests"
$ aiguard checkpoint  # checkpoint-002 created

# When ready for final review:
$ aiguard verify  # final verdict with all checks + (optional) AI review
```

### Pain Points

- Checkpoints are **informational only**. They don't affect gates or block further work.
- Checkpoints are stored in `.aiguard/checkpoints/` which is in `.aiguard/.gitignore`, so they won't show up in your final diff.

---

## verify

**Run deterministic checks and optionally AI reviewers.**

### verify --local-only (Deterministic-Only)

**Run 11 deterministic checks against the current diff. Fast, no AI, no API calls.**

```bash
$ aiguard verify --local-only
Verdict: WARN
Summary: Verdict: WARN. Deterministic checks: 1 WARN, 2 Info

Recommended Fix:
Address these warnings:
- test_presence: behavior files changed without any test file changes: internal/handlers/user.go

Reports written to:
  .aiguard/reviews/final-verdict.md
  .aiguard/reviews/final-verdict.json
```

### verify --reviewers (Multi-Reviewer)

**Run both deterministic checks and AI reviewers concurrently.**

```bash
$ aiguard verify --reviewers claude_code,codex
Verdict: WARN
Summary: Verdict: WARN. Deterministic checks: 1 WARN...

[AI Reviews]
- claude_code: [INFO] PASS
- codex: [HIGH] suggests adding more comprehensive error handling

Reports written to:
  .aiguard/reviews/final-verdict.md
  .aiguard/reviews/disagreement-report.md
  .aiguard/reviews/{claude_code,codex}-review.md
```

### Flags

```bash
--base <ref>
  # Base ref to diff against (default: config.project.source_branch)
  aiguard verify --base develop --local-only

--local-only
  # Run deterministic checks only; skip AI reviewers
  aiguard verify --local-only

--reviewers <names>
  # Comma-separated list of agents to run
  aiguard verify --reviewers claude_code,codex

--auto-fix
  # On REJECT/WARN (but not BLOCKED), send the recommended fix prompt
  # to the configured implementation agent and save the transcript
  aiguard verify --auto-fix

--copy-fix
  # Copy the recommended fix prompt to clipboard (alternative to --auto-fix)
  aiguard verify --copy-fix
```

### Exit Codes

```
0 = APPROVE or WARN (success, review the reports if needed)
1 = REJECT or BLOCKED (failure, fix before pushing)
```

### Verdict Precedence (SPEC.md §16)

Highest to lowest severity:

1. `BLOCKED` — forbidden file or secret detected (deterministic, never overridden)
2. `REJECT` — critical/high issue or missing source AC (deterministic or AI)
3. `WARN` — medium-severity issue or disagreement between reviewers
4. `APPROVE` — all checks passed

### Pain Points

- Deterministic `BLOCKED` is final; AI reviewer output cannot override it.
- Multi-reviewer mode runs agents concurrently. Timeouts are per-agent (3600s default).
- If an agent crashes or times out, its output is WARN-level "malformed response", not fatal.
- AC normalization happens automatically (AC1 vs ac-1), but both formats must be consistently used in the digest and plan.

---

## clarify

**Resolve open questions about the requirements.**

### clarify

**List open clarification questions from the digest.**

```bash
$ aiguard clarify
Q1 (blocking): Should profiles be public or private by default?
Q2 (blocking): Do we need profile versioning?
Q3 (non-blocking): Should avatars support WEBP?

Use 'aiguard clarify answer --id Q1' to answer a question.
```

### clarify answer

**Answer a clarification question.**

```bash
$ aiguard clarify answer --id Q1 --text "Private by default, user can opt-in to public"
Q1 answered (not assumed).

$ aiguard clarify answer --id Q2 --assume
Q2 assumed (not answered; no rationale recorded).
```

Blocking questions must be answered before downstream commands (plan, verify) proceed.

### Pain Points

- Blocking questions gate downstream commands unless `--force --reason "..."` is used.
- Questions are extracted heuristically from the digest (marked with `?` or uncertain language).
- Answer status doesn't affect the verdict; it's purely for documentation.

---

## status

**Show gate states, verdict summary, and snapshot freshness.**

```bash
$ aiguard status
## Gate States

digest: APPROVED_BY_USER (approved by local-user at 2025-05-28T14:30:00Z)
guidance: NOT_STARTED
plan: NOT_STARTED
tests: NOT_STARTED

## Last Verdict

Verdict: WARN (from 2025-05-28T15:10:30Z)
Summary: Verdict: WARN. Deterministic checks: 1 WARN, 2 Info

## Snapshot

Status: fresh (0 commits since snapshot)
Created: 2025-05-28T14:22:00Z
```

---

## report

**Re-render the latest final verdict markdown.**

```bash
$ aiguard report
# AIGuard Final Verdict

## Verdict

WARN

## Summary

Verdict: WARN. Deterministic checks: 1 WARN, 2 Info

...
(same as .aiguard/reviews/final-verdict.md)
```

---

## doctor

**Check AIGuard prerequisites and environment.**

```bash
$ aiguard doctor
Checking AIGuard health...

✓ git version 2.43.0
✓ .aiguard/ workspace found
✓ source branch 'main' resolves
✓ config.yaml loads and validates
✓ claude_code command available
⚠ codex command not found (optional)
✓ .aiguard/ is writable

All critical checks passed.
```

Exit code 0 if all critical checks pass; non-zero if any fail.

### Pain Points

- Missing optional agents (codex, claude) are warnings, not failures.
- `doctor` doesn't test network connectivity or API keys; it only checks local binaries and config.

---

## version

**Print the AIGuard version.**

```bash
$ aiguard version
v0.0.1
```
