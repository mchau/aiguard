# Quick Start — 5 Minutes to First Verdict

This guide walks you through AIGuard's essential workflow in a real repository.

## Installation

```bash
go install github.com/mchau/aiguard/cmd/aiguard@latest
# or build locally:
cd ~/aiguard && go build -o ~/bin/aiguard ./cmd/aiguard
```

## Step 1: Initialize (1 minute)

In your repository root:

```bash
$ aiguard init
Initialized AIGuard workspace at /home/alice/myrepo/.aiguard/config.yaml
Source branch: main
Next step: run 'aiguard snapshot update' to map the codebase.
```

That's it. AIGuard autodetected your default branch (`main`) from `origin/HEAD` and created `.aiguard/`:

```
.aiguard/
├── config.yaml           # Edit this to customize risk paths, agents, etc.
├── .gitignore            # Excludes logs/, reviews/, checkpoints/ from your repo diff
├── snapshot/             # Codebase map (refresh when major structure changes)
├── requirements/         # Acceptance criteria
├── guidance/             # Technical guidance
├── plans/                # Implementation plan
├── implementation/       # Rationale for each changed file
├── reviews/              # Verdict reports
├── tests/                # Test strategy
└── logs/                 # Audit trail (ignored by .gitignore)
```

## Step 2: Create a Feature Branch and Make Changes (1 minute)

```bash
$ git checkout -b feature/add-user-profile
$ # ...you or Claude Code make changes...
$ git add . && git commit -m "feat: add user profile endpoint"
```

## Step 3: Verify Against Local Checks (1 minute)

```bash
$ aiguard verify --local-only
Verdict: WARN
Summary: Verdict: WARN. Deterministic checks: 1 WARN, 2 Info

Recommended Fix:
Address these warnings:
- test_presence: behavior files changed without any test file changes: internal/handlers/user.go

Reports written to:
  /home/alice/myrepo/.aiguard/reviews/final-verdict.md
  /home/alice/myrepo/.aiguard/reviews/final-verdict.json
```

The 11 deterministic checks ran instantly, no AI involved:

- ✓ Forbidden paths (`.env`, `*.pem`, `secrets/**`) — clean
- ✓ Secret detection (AWS keys, GitHub PATs, etc.) — clean
- ✓ Risk paths (`auth/`, `billing/`, `migrations/`) — clean
- ⚠ Test presence — `internal/handlers/user.go` changed without a test

## Step 4: Act on the Verdict (1 minute)

Three options:

### Option A: Copy the Fix Prompt to Your AI Tool

```bash
$ aiguard verify --local-only --copy-fix
Fix prompt copied to clipboard. Paste it into your AI tool.
```

Open Claude Code, paste the prompt, implement the tests, and commit.

### Option B: Auto-Fix Loop (AI Reprompts)

```bash
$ aiguard verify --local-only --auto-fix
# Sends the recommended fix prompt to your configured agent,
# saves the transcript to .aiguard/reviews/auto-fix-transcript.md
```

Check the transcript, apply the agent's suggestions if they look right.

### Option C: Manual Fix

Read the verdict report at `.aiguard/reviews/final-verdict.md` and fix it yourself.

## Step 5: Re-Verify (1 minute)

After adding tests:

```bash
$ git add . && git commit -m "test: add user profile endpoint tests"
$ aiguard verify --local-only
Verdict: APPROVE
Summary: Verdict: APPROVE. Deterministic checks: 2 Info
```

Done. You can now push with confidence.

---

## When to Use Multi-Reviewer (Optional)

If you have both `claude` and `codex` CLIs installed and want AI to review the code:

```bash
$ aiguard verify --reviewers claude_code,codex
```

This runs both agents concurrently, aggregates findings, detects disagreements, and writes:

- `.aiguard/reviews/final-verdict.md` — merged verdict
- `.aiguard/reviews/disagreement-report.md` — where reviewers differed
- `.aiguard/reviews/{claude_code,codex}-review.md` — raw reviewer output

Deterministic `BLOCKED` verdicts (forbidden paths, secrets) are never overridden by AI.

---

## Pain Points & Gotchas

### 1. **Branch Auto-Detection Fails**

If `aiguard init` picks the wrong branch:

```bash
# Run init again with an explicit branch
aiguard init --source-branch develop --force
```

Check `.aiguard/config.yaml` to confirm.

### 2. **"Pick a Mode" Error on rationale**

```bash
$ aiguard rationale update
Error: pick a mode: --file <path>, --ai, or --interactive (no stdin prompting by default to keep the command scriptable)
```

Rationale requires an explicit mode to be scriptable:

```bash
# Option 1: from a YAML file
aiguard rationale update --file rationale.yaml

# Option 2: via your AI agent
aiguard rationale update --ai

# Option 3: interactive prompts (only for local work)
aiguard rationale update --interactive
```

### 3. **BLOCKED Verdicts Won't Auto-Fix**

`--auto-fix` intentionally skips `BLOCKED` verdicts because secrets and forbidden paths need manual triage:

```bash
Verdict: BLOCKED
...
auto-fix skipped: BLOCKED verdicts require manual remediation.
```

Remove the forbidden file, then re-verify.

### 4. **AC IDs Don't Match Plan**

The digest extracts ACs like `AC1`, `AC2`. If your AI-generated plan uses `ac-1` or `AC_001`, they'll be normalized to match automatically. If a source-provided AC is missing from the plan entirely, `plan create` will flag it as uncovered.

### 5. **Privacy Mode Excludes File Contents**

By default, `.env`, secrets, and forbidden paths are always excluded from AI context. Additionally, `privacy.include_actual_files_by_default` is **false** by default, so file contents are not sent to reviewers unless you override it:

```bash
aiguard verify --reviewers claude_code  # Files excluded (safe)
```

To include files (useful for code review):

```yaml
# .aiguard/config.yaml
privacy:
  include_actual_files_by_default: true  # Now send file contents to AI
```

Still redacted: `.env`, `*.pem`, `secrets/**`, anything matching `forbidden_paths`.

### 6. **Transient Files Appear in Diffs**

The first time you run the gated workflow (digest, guidance, plan), transient files like `logs/audit.jsonl` or `reviews/claude-review.md` show up in your git status. `aiguard init` now writes `.aiguard/.gitignore` to exclude them. Commit `.aiguard/.gitignore` to your repo once, and you're done.

### 7. **Test Presence Warns Incorrectly**

The `test_presence` check warns when behavior files change without test files. It excludes config/docs (`.yaml`, `.md`, `.json`) but not auxiliary files like `handlers/router.go`. If you touched a router file to wire a call, you'll get a warning. Add a `_test.go` file (even empty) to silence it, or update a real test.

### 8. **Multi-Reviewer Disagreement**

When two reviewers give different severity levels on the same AC:

- **Info vs Medium** — not flagged (same "soft" bucket; likely just verbose vs terse)
- **Info vs High** — flagged (different "hard" bucket; real signal)

Check `disagreement-report.md` to decide which reviewer to trust for that AC.

---

## Next Steps

- Read [`docs/commands.md`](commands.md) for the full command reference
- Read [`docs/configuration.md`](configuration.md) to customize risk paths and agents
- Read [`docs/usecases.md`](usecases.md) for specific workflow scenarios
