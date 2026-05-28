# PRD.md — AIGuard: Requirements-Gated AI Development Harness

## 1. Product Summary

**AIGuard** is a local-first gated development harness that prevents AI-generated code from drifting away from business requirements. It converts Jira tickets, pasted requirements, or documentation into structured acceptance criteria, technical guidance, implementation plans, test strategy, implementation checkpoints, and final spec-to-diff verification.

AIGuard is not primarily a coding assistant. It is a **requirements-aware verification and workflow gate** for developers using Claude Code, Codex, Cursor, Copilot, Aider, or manual coding.

The core promise:

> AIGuard helps developers prove that AI-generated code satisfies the business requirement before the code reaches a pull request.

## 2. Problem Statement

Developers increasingly use AI coding tools at work, but AI-generated code can be dangerous because it often:

- Implements the wrong requirement.
- Misses explicit acceptance criteria.
- Changes unrelated code.
- Passes tests while violating business logic.
- Creates shallow or misleading tests.
- Modifies risky areas such as auth, billing, permissions, data migrations, or shared infrastructure.
- Produces convincing explanations that hide incorrect assumptions.
- Causes embarrassment or professional risk during code review.

The user has already experienced real workplace consequences from AI-generated bad code. The product must prioritize correctness, traceability, local execution, evidence, and human approval.

## 3. Target Users

### Primary User

A professional software developer using AI coding tools in a real work codebase.

Characteristics:

- Works with Jira tickets, docs, business requirements, and feature branches.
- Uses Claude Code, Codex, Cursor, Copilot, or similar AI tools.
- Needs to avoid submitting bad AI-generated PRs.
- Wants to reduce token usage.
- Wants workflow discipline without manually managing dozens of markdown files.

### Secondary Users

- Senior developers reviewing AI-assisted pull requests.
- Engineering managers adopting AI coding tools.
- Small teams using Claude Code or Codex but worried about quality.
- Agencies producing client code using AI.

## 4. Product Positioning

AIGuard is:

> A local-first requirements-gated AI development harness.

Alternative short positioning:

> Spec-to-diff verification for AI-generated code.

What AIGuard is not:

- Not a generic chatbot.
- Not a replacement for Claude Code or Codex.
- Not a cloud SaaS by default.
- Not a generic linter.
- Not only an AI code reviewer.
- Not a test runner replacement.
- Not a guarantee that code is perfect.

## 5. Goals

### Product Goals

1. Ensure AI-generated code stays aligned with business requirements.
2. Convert vague tickets into structured, traceable acceptance criteria.
3. Preserve explicit acceptance criteria from the ticket when provided.
4. Generate and validate technical guidance before code is written.
5. Generate and validate implementation plans before code is written.
6. Use a living codebase map to reduce token usage.
7. Use deterministic checks wherever possible.
8. Use AI only where reasoning is needed.
9. Support multiple AI reviewers to reduce single-model bias.
10. Produce an audit trail showing what was checked and approved.
11. Produce a final submit/warn/reject verdict before PR submission.

### Developer Goals

1. Run locally on a work machine.
2. Avoid uploading proprietary code unless explicitly configured.
3. Use current CLI tools such as Claude Code or Codex.
4. Keep the source-of-truth branch, often `develop`, as the baseline.
5. Make it easy to run on every ticket.
6. Be useful even before full automation is complete.

## 6. Non-Goals

AIGuard v1 is not intended to:

- Automatically guarantee production correctness.
- Replace human engineering judgment.
- Replace official CI/CD.
- Replace code owners or team review.
- Automatically modify remote Jira unless explicitly configured.
- Require cloud hosting.
- Require direct Claude or OpenAI API usage.
- Store proprietary source code outside the local machine.
- Force a specific AI model or provider.
- Depend on one vendor’s model naming forever.

## 7. Core Concept

AIGuard turns a ticket into a chain of gated artifacts:

```text
source branch snapshot
→ raw requirement
→ requirement digest
→ acceptance criteria
→ clarification questions
→ test strategy
→ technical guidance
→ guidance review
→ implementation plan
→ plan review
→ implementation checkpoints
→ final diff verification
→ submit / warn / reject verdict
```

Every major artifact must trace back to the original requirement.

The final review must be based on the actual git diff, not on the AI’s summary of what it did.

## 8. Workflow Overview

### Phase 0 — Project Initialization

Command:

```bash
aiguard init
```

Creates:

```text
.aiguard/
  config.yaml
  snapshot/
  requirements/
  guidance/
  plans/
  implementation/
  reviews/
  logs/
```

The user configures:

- Source-of-truth branch, for example `develop`.
- Risk paths.
- Forbidden paths.
- Test commands.
- AI reviewer commands.
- Model routing preferences.
- Gate rules.

### Phase 1 — Living Codebase Map / Snapshot

Command:

```bash
aiguard snapshot update --base develop
```

AIGuard creates or updates a living codebase map.

The snapshot is a **map**, not a substitute for source code.

It includes:

- Folder tree.
- Component/module notes.
- Known code ownership boundaries.
- Where business logic lives.
- Where API routes live.
- Where UI components live.
- Where tests live.
- High-risk folders.
- Common architecture patterns.
- Recently changed areas.
- Last source branch commit SHA.

The snapshot reduces token usage and helps route AI to relevant files.

The snapshot must include metadata:

```yaml
source_branch: develop
source_commit: <git-sha>
previous_snapshot_commit: <git-sha>
updated_at: <timestamp>
snapshot_version: <number>
```

Snapshot update strategy:

1. Fetch latest source branch.
2. Determine the previous snapshot commit.
3. Diff previous snapshot commit against current source branch.
4. Update only changed map sections where possible.
5. Mark uncertain sections as stale.
6. Store the new commit SHA.

### Phase 2 — Requirement Ingestion

Commands:

```bash
aiguard req ingest --file ticket.md
aiguard req ingest --paste
aiguard req ingest --jira ABC-123
```

AIGuard ingests raw requirements from:

- Jira ticket text.
- Markdown file.
- Pasted text.
- Documentation file.
- Future: Jira API connector.

Important behavior:

If the original ticket already includes an **Acceptance Criteria** section, AIGuard must preserve it, label it as source-provided, and use it as the highest-priority input when creating the final acceptance criteria.

The tool must separate:

- Source-provided acceptance criteria.
- AI-inferred acceptance criteria.
- Non-goals / do-not-change constraints.
- Business rules.
- Open questions.
- Assumptions.

### Phase 3 — Requirement Digest

Command:

```bash
aiguard digest create
```

Output:

```text
.aiguard/requirements/raw.md
.aiguard/requirements/digest.md
.aiguard/requirements/acceptance-criteria.json
.aiguard/requirements/open-questions.md
.aiguard/requirements/assumptions.md
```

The digest must include:

- User story.
- Summary.
- Source-provided acceptance criteria.
- AI-inferred acceptance criteria.
- Combined normalized acceptance criteria.
- Explicit non-goals.
- Business rules.
- API/data/UI/test implications.
- Clarification questions.
- Assumptions.

Acceptance criteria must have stable IDs:

```text
AC1, AC2, AC3...
```

Each criterion must include its source:

```json
{
  "id": "AC2",
  "text": "Cancellation reason must be persisted.",
  "source": "ticket_acceptance_criteria",
  "priority": "must",
  "status": "pending"
}
```

### Phase 4 — Clarification Gate

Command:

```bash
aiguard clarify
```

AIGuard asks questions only when needed.

Questions must be classified:

- Blocking.
- Non-blocking.
- Risk-related.
- Implementation preference.

The workflow must not proceed when unresolved blocking questions exist unless the user explicitly records an assumption.

### Phase 5 — Test Strategy

Command:

```bash
aiguard tests strategy
```

AIGuard produces:

```text
.aiguard/tests/test-strategy.md
.aiguard/tests/test-requirements.json
```

The test strategy maps each acceptance criterion to expected test coverage.

Test categories:

- Unit tests.
- Integration tests.
- E2E tests.
- Regression tests for explicit non-goals.
- Negative/error tests.
- Data migration tests.
- Permission/security tests.

Important:

AIGuard should usually produce test strategy before implementation, but actual test creation can happen during or after implementation to avoid overfitting tests to guessed implementation details.

### Phase 6 — Technical Guidance

Command:

```bash
aiguard guidance create
```

Inputs:

- Snapshot map.
- Raw requirement.
- Digest.
- Acceptance criteria.
- Open questions and assumptions.
- Targeted relevant files, when necessary.
- Project rules.

Output:

```text
.aiguard/guidance/tech-guidance.md
.aiguard/guidance/tech-guidance.json
```

Technical guidance must include:

- Requirement mapping.
- Relevant code areas.
- Existing patterns to follow.
- Proposed technical approach.
- Non-goals / do-not-touch areas.
- Data model impact.
- API impact.
- UI impact.
- Test guidance.
- Risk areas.
- Rollback considerations.

### Phase 7 — Technical Guidance Review Gate

Command:

```bash
aiguard guidance review --reviewers claude,codex
```

AIGuard reviews guidance against:

- Raw requirement.
- Source-provided acceptance criteria.
- Normalized acceptance criteria.
- Snapshot.
- Project rules.

The review must answer:

- Does guidance satisfy every AC?
- Does guidance ignore source-provided AC?
- Does guidance invent unrelated work?
- Does guidance violate non-goals?
- Does guidance identify risky areas?
- Does guidance include tests?

Output:

```text
.aiguard/guidance/guidance-review.md
.aiguard/guidance/guidance-review.json
```

Verdict:

- PASS
- WARN
- FAIL

### Phase 8 — Implementation Plan

Command:

```bash
aiguard plan create
```

Output:

```text
.aiguard/plans/plan.md
.aiguard/plans/plan.json
```

Plan requirements:

- Every step maps to one or more AC IDs.
- Every expected file or module has a rationale.
- Risky changes are explicitly justified.
- Tests are included.
- Non-goals are protected.
- Rollback or revert strategy is described for high-risk changes.

### Phase 9 — Plan Review Gate

Command:

```bash
aiguard plan review --reviewers claude,codex
```

Plan review checks:

- Every AC is covered by at least one step.
- Every step maps to an AC or approved technical necessity.
- No unapproved risky file areas are included.
- Required tests exist in the plan.
- Explicit non-goals are protected.
- The plan does not overreach.

Output:

```text
.aiguard/plans/plan-review.md
.aiguard/plans/plan-review.json
```

### Phase 10 — Implementation Support

AIGuard may invoke an AI coding agent or let the user code manually.

Commands:

```bash
aiguard implement --agent claude
aiguard implement --agent codex
aiguard checkpoint
```

AIGuard should support external CLI adapters.

It should not require direct API calls.

Example:

```yaml
agents:
  claude_code:
    enabled: true
    command: claude
    modes:
      coding:
        args: ["-p"]
  codex:
    enabled: true
    command: codex
```

Implementation checkpoints:

After each logical feature or plan step, AIGuard captures:

- Changed files.
- Diff summary.
- Related ACs.
- Tests run.
- Scope drift check.
- Risk warnings.
- Recommended continue/revise/stop decision.

Output:

```text
.aiguard/implementation/checkpoint-001.md
.aiguard/implementation/checkpoint-001.json
```

### Phase 11 — Changed File Rationale

Command:

```bash
aiguard rationale update
```

AIGuard maintains:

```text
.aiguard/implementation/changed-files-rationale.md
```

Every changed file must have:

- Reason.
- Related AC.
- Related plan step.
- Risk level.
- Whether it was expected by the plan.

Example:

| File | Reason | Related AC | Plan Step | Risk | Status |
|---|---|---|---|---|---|
| billing/cancellation_service.go | Persist cancellation reason | AC2 | Step 2 | High | Expected |
| billing/refund_policy.go | No approved rationale | None | None | High | Flagged |

### Phase 12 — Final Verification

Command:

```bash
aiguard verify --base develop --reviewers claude,codex
```

Final verification uses the actual git diff:

```bash
git diff develop...HEAD
```

Inputs:

- Raw requirement.
- Digest.
- Acceptance criteria JSON.
- Tech guidance.
- Plan.
- Changed files rationale.
- Test strategy.
- Actual git diff.
- Test results.
- Snapshot map.
- Targeted actual files where needed.
- Deterministic check results.
- Independent AI reviewer outputs.

Final verification must produce:

```text
.aiguard/reviews/final-verdict.md
.aiguard/reviews/final-verdict.json
.aiguard/reviews/disagreement-report.md
```

Verdict values:

- APPROVE
- WARN
- REJECT
- BLOCKED

The report must include:

- Requirement coverage.
- Traceability matrix.
- Missing ACs.
- Test gaps.
- Scope drift.
- Risky files.
- Forbidden file changes.
- Reviewer disagreement.
- Recommended fix prompt.
- PR readiness summary.

## 9. AI Model Routing

AIGuard must support per-step model routing.

The user’s preferred routing:

- Snapshot updates: fast/cheap model, for example Haiku.
- Brainstorm, tech guidance, plan, code review: strongest reasoning model, for example Opus High.
- Coding: strong coding model, for example Sonnet Medium.
- Scrutiny/review: strongest reasoning model or independent reviewer.

Because model names change, AIGuard must not hardcode specific model names. It should support configurable profiles:

```yaml
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
```

CLI command templates must be configurable because vendor CLI flags may change:

```yaml
agent_commands:
  claude_code:
    command: claude
    prompt_arg: "-p"
    model_arg_template: "--model {{model_alias}}"
    effort_arg_template: "--effort {{effort}}"
```

If a CLI does not support model flags, AIGuard should still run and include the routing metadata in the prompt.

## 10. Deterministic Checks

AIGuard must use deterministic checks before AI review.

Checks include:

- Git working tree status.
- Base branch freshness.
- Snapshot freshness.
- Changed files.
- Diff size.
- Forbidden paths touched.
- High-risk paths touched.
- Dependency files changed.
- Migration files changed.
- Test files changed.
- Deleted files.
- Generated files.
- Large file changes.
- Public API files changed.
- Config files changed.
- Environment files changed.
- Every changed file has rationale.
- Every plan step maps to AC.
- Every AC maps to implementation or documented reason.
- Implementation touched files not listed in plan.
- Test commands pass/fail.

## 11. AI Reviews

AIGuard should support multiple reviewer types:

- Claude Code CLI reviewer.
- Codex CLI reviewer.
- Local LLM reviewer.
- API reviewer if configured.
- Manual reviewer notes.

Reviewer outputs must be structured.

AIGuard must not simply trust one reviewer. It must aggregate:

- Agreements.
- Disagreements.
- Findings supported by deterministic evidence.
- Findings unsupported by evidence.
- Highest severity issues.

If reviewers disagree, AIGuard should usually produce WARN or REJECT, depending on severity.

## 12. Severity Levels

Findings use:

- Critical
- High
- Medium
- Minor
- Info

Definitions:

### Critical

Code likely violates a must-have requirement, breaks existing business-critical behavior, touches forbidden paths, introduces security risk, or is unsafe to submit.

### High

Likely missing AC, risky behavior not tested, high-risk area modified without justification, or reviewer disagreement on core requirement.

### Medium

Potential requirement ambiguity, insufficient tests, minor scope drift, unclear rationale.

### Minor

Style, naming, small convention issues, documentation gaps.

### Info

Observations that do not block submission.

## 13. Gate Behavior

Gate statuses:

- NOT_STARTED
- IN_PROGRESS
- NEEDS_REVIEW
- PASS
- WARN
- FAIL
- BLOCKED
- APPROVED_BY_USER
- SKIPPED_WITH_REASON

AIGuard must prevent later steps from running if required prior gates are blocked, unless the user passes:

```bash
--force --reason "..."
```

Forced progression must be logged.

## 14. Audit Trail

AIGuard must maintain an audit trail.

Each action records:

- Timestamp.
- Command.
- User.
- Git branch.
- Git commit.
- Base branch.
- Snapshot commit.
- Inputs used.
- AI reviewer used.
- Model profile used.
- Output files.
- Gate status.

Audit log:

```text
.aiguard/logs/audit.jsonl
```

## 15. Privacy and Security

Requirements:

- Local-first by default.
- Do not upload code unless configured.
- Show warning before external API usage.
- Do not store secrets.
- Redact known secret patterns from AI prompts.
- Support `.aiguardignore`.
- Support forbidden file paths.
- Never include `.env` files in AI context.
- Never include private keys or tokens in prompts.
- Allow local-only deterministic mode.

Modes:

```bash
aiguard verify --local-only
aiguard verify --reviewers local
aiguard verify --reviewers claude,codex
```

## 16. Configuration

Default config:

```yaml
project:
  name: ""
  source_branch: develop
  snapshot_dir: .aiguard/snapshot

privacy:
  default_mode: local_first
  allow_external_ai: false
  redact_secrets: true

risk_paths:
  critical:
    - ".env"
    - "secrets/**"
    - "infra/prod/**"
  high:
    - "auth/**"
    - "billing/**"
    - "payments/**"
    - "permissions/**"
    - "migrations/**"
    - "database/**"

forbidden_paths:
  - ".env"
  - ".env.*"
  - "secrets/**"
  - "*.pem"
  - "*.key"

test_commands:
  - name: unit
    command: "go test ./..."
  - name: lint
    command: "golangci-lint run"

gates:
  require_user_approval:
    - digest
    - test_strategy
    - tech_guidance
    - plan
    - final_verdict
  block_on:
    - forbidden_path_changed
    - missing_source_acceptance_criteria
    - high_risk_unjustified_change
    - tests_failed
```

## 17. CLI Commands

Required commands:

```bash
aiguard init

aiguard snapshot update --base develop
aiguard snapshot status
aiguard snapshot show

aiguard req ingest --file ticket.md
aiguard req ingest --paste
aiguard req status

aiguard digest create
aiguard digest review
aiguard digest approve

aiguard clarify
aiguard clarify answer

aiguard tests strategy
aiguard tests review
aiguard tests approve

aiguard guidance create
aiguard guidance review
aiguard guidance approve

aiguard plan create
aiguard plan review
aiguard plan approve

aiguard implement --agent claude
aiguard checkpoint

aiguard rationale update
aiguard rationale check

aiguard verify --base develop
aiguard final

aiguard status
aiguard report
aiguard doctor
```

## 18. Success Metrics

### Personal/Developer Success

- Fewer bad AI-generated PRs.
- More confidence before submitting code.
- Reduced manual review time.
- Better traceability from ticket to diff.
- Fewer reviewer comments about missed requirements.
- Reduced scope drift.

### Product Metrics

- Percentage of reviews resulting in actionable findings.
- Number of missing ACs caught.
- Number of risky files flagged.
- Number of unrelated file changes flagged.
- Test gaps identified.
- Time saved per ticket.
- False positive rate.
- User override frequency.

## 19. Full Feature Roadmap

### v1 — Full Local CLI

- Project init.
- Snapshot update.
- Requirement ingestion.
- Digest creation.
- Acceptance criteria extraction.
- Clarification questions.
- Test strategy.
- Technical guidance.
- Guidance review.
- Plan creation.
- Plan review.
- Changed-file rationale.
- Final verification.
- Deterministic checks.
- Claude Code CLI adapter.
- Codex CLI adapter.
- Model routing profiles.
- Audit trail.
- Markdown and JSON reports.

### v2 — TUI

- Interactive gate dashboard.
- Approve/reject gates.
- View traceability matrix.
- View reviewer disagreements.
- Run next recommended command.
- Browse artifacts.

### v3 — GitHub/Jira Integrations

- Jira fetch.
- Jira comment with digest or verdict.
- GitHub PR review comments.
- GitHub check status.
- PR template generation.

### v4 — Team/Policy Mode

- Shared policy packs.
- Repo-specific rules.
- Team approval workflow.
- CODEOWNERS awareness.
- CI enforcement.

### v5 — Learning System

- Track model performance by repo and task type.
- Learn common scope drift patterns.
- Recommend best reviewer/coder per task.
- Auto-tune risk paths and context packs.

## 20. Build Priority

Even though the full feature set is specified, implementation should be sequenced to reduce risk:

1. Core artifact model and config.
2. Git diff analyzer and deterministic checks.
3. Requirement ingestion and AC extraction.
4. Final verification report.
5. Snapshot map.
6. Guidance and plan gates.
7. Multi-reviewer adapter layer.
8. Implementation checkpoints.
9. Integrations and TUI.

## 21. Final Product Principle

The AI may generate recommendations, but the harness owns:

- State.
- Gates.
- Evidence.
- Traceability.
- Deterministic checks.
- Reviewer disagreement.
- Final verdict.

AIGuard’s job is not to make developers trust AI blindly. Its job is to make AI-generated code auditable, reviewable, and safer to submit.
