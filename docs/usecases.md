# Use Case Walkthroughs

Real scenarios and how to handle them with AIGuard.

---

## Scenario 1: "I Got a REJECT Verdict — How Do I Fix It?"

### Situation

You ran `aiguard verify --local-only` and got:

```
Verdict: REJECT
Summary: Verdict: REJECT. Deterministic checks: 2 High, 1 WARN, 2 Info

Fix these high-severity issues:
- risk_path: risk path changed (High): auth/login.go
- ac_traceability: AC2 (ticket_acceptance_criteria): "persist user credentials" not covered by any plan step

Address these warnings:
...

Recommended Fix:
Revise the current patch.

Fix these high-severity issues:
- risk_path: risk path changed (High): auth/login.go
- ac_traceability: AC2 (ticket_acceptance_criteria): "persist user credentials" not covered by any plan step

Do not modify files unrelated to the above issues.
```

### What This Means

- You changed `auth/login.go` (a high-risk path) without documenting why in the plan
- AC2 ("persist user credentials") is in the original ticket but not in your implementation plan
- You need to either cover both, or revert unrelated changes

### How to Fix It

**Option 1: Copy the fix prompt and ask your AI to revise**

```bash
$ aiguard verify --local-only --copy-fix
# Prompt is now in your clipboard

# Open Claude Code (or your AI tool)
$ claude -p  # Paste the prompt (Cmd+V)
# Ask: "Here's feedback on my patch. Please revise it to fix these issues."
# ...Claude suggests changes...
$ git diff  # review
$ git add . && git commit -m "fix: address AC2 and auth changes"
```

**Option 2: Auto-fix loop (let AI reprompt)**

```bash
$ aiguard verify --local-only --auto-fix
# Agent receives the fix prompt, transcript saved to .aiguard/reviews/auto-fix-transcript.md

$ cat .aiguard/reviews/auto-fix-transcript.md
# View the agent's suggestions

$ git diff  # review changes made by your AI
$ git add . && git commit -m "fix: auto-fix revision"
```

**Option 3: Manual fix**

```bash
# Read the verdict report
$ cat .aiguard/reviews/final-verdict.md

# Option A: Update the plan to cover auth/login.go
# Edit .aiguard/plans/plan.json and add auth/login.go to the expected_files of a step

# Option B: Revert auth/login.go if it was unintended
$ git checkout HEAD -- auth/login.go

# Option C: Add AC2 to your implementation plan
# Edit .aiguard/plans/plan.json, create a step for AC2 that includes your auth changes
```

**Option 4: Force-gate and document**

If you're confident the change is correct:

```bash
# Create/update the plan to authorize the auth/login.go change
# Then re-run verify

aiguard verify --local-only
# Should now be WARN or APPROVE
```

### Next Steps

```bash
$ aiguard verify --local-only
Verdict: APPROVE
Summary: Verdict: APPROVE. Deterministic checks: 2 Info

# Now safe to push
$ git push origin feature/auth-work
```

---

## Scenario 2: "I Got a BLOCKED Verdict — Secrets Detected"

### Situation

```
Verdict: BLOCKED
Summary: Verdict: BLOCKED. Deterministic checks: 3 BLOCKED

Blocked changes:
- forbidden_path: forbidden path (/home/alice/myrepo/.env) in diff
- secret_detection: AWS access key detected in diff
- secret_detection: bearer token detected in diff

Reports written to:
  .aiguard/reviews/final-verdict.md
```

### What This Means

- `.env` file was staged
- AWS credential and bearer token are visible in the diff
- **This is a hard stop.** Auto-fix is intentionally disabled for BLOCKED verdicts.

### How to Fix It

**These are unambiguous failures. Fix manually:**

```bash
# Unstage and remove the secrets
$ git reset HEAD .env
$ rm .env  # Never commit .env

# Check the diff for inline secrets
$ git diff
# Look for lines containing AKIA, Bearer, etc. and remove them

# Check your code for hardcoded credentials
$ git log --all -p | grep -i "password\|secret\|token" | head -20

# Commit the fixes
$ git add . && git commit -m "chore: remove secrets from diff"

# Re-verify
$ aiguard verify --local-only
# Should now be APPROVE or WARN, not BLOCKED
```

**Why auto-fix is blocked:**

AIGuard intentionally refuses to auto-fix BLOCKED verdicts because:
1. The agent can't be trusted to remove secrets correctly
2. Secrets in your diff indicate a deeper issue (committing `.env`? hardcoding creds?)
3. Human review is non-negotiable for security

### Prevention

```bash
# Add secrets to .gitignore
$ echo ".env" >> .gitignore
$ echo "*.key" >> .gitignore
$ git add .gitignore && git commit -m "chore: ignore secrets"

# Don't commit credentials. Use environment variables or secret management.
```

---

## Scenario 3: "Multi-Reviewer Disagreement — Who Do I Trust?"

### Situation

```
$ aiguard verify --reviewers claude_code,codex

Verdict: WARN
Summary: Verdict: WARN. Deterministic checks: 1 WARN...

Reports written to:
  .aiguard/reviews/final-verdict.md
  .aiguard/reviews/disagreement-report.md
  .aiguard/reviews/claude_code-review.md
  .aiguard/reviews/codex-review.md

$ cat .aiguard/reviews/disagreement-report.md

## Reviewer Disagreements

### AC1: "User can create a profile"

- **claude_code**: [High] Missing error handling for duplicate profile names
- **codex**: [Info] Error handling not critical for MVP

...
```

### What This Means

- claude_code found a high-severity issue; codex found it minor
- The harness promotes this to WARN (different severity buckets = signal, not noise)
- You need to decide which reviewer was right

### How to Decide

**Read both reviews fully:**

```bash
$ cat .aiguard/reviews/claude_code-review.md
# Review from claude_code
...
[HIGH] AC1: Missing error handling for duplicate profile names
Evidence: No try-catch around profile.create() in handlers/user.go
Suggestion: Add validation and error response

$ cat .aiguard/reviews/codex-review.md
# Review from codex
...
[INFO] AC1: Error handling deferred to phase 2
Evidence: MVP scope doesn't require complete error handling
Suggestion: Document known issues for future work
```

**Criteria for deciding:**

1. **Your ticket scope**: Does the ticket explicitly mention error handling? If yes, claude_code is right.
2. **Risk assessment**: Is the risk acceptable? Duplicate profiles are bad; deferring error handling might be okay if you document it.
3. **Your team's standards**: What does your code review process expect?

**Example resolution:**

```bash
# Option A: Claude_code is right — add error handling
$ # Edit handlers/user.go to add error handling
$ git add . && git commit -m "feat: add error handling for duplicate profiles"

# Option B: Codex is right — document the limitation
$ # Add a comment explaining why error handling is deferred
$ git add . && git commit -m "docs: note error handling deferred to phase 2"

# Option C: Meet in the middle — add minimal error handling
$ # Add basic error logging but not full handling
$ git add . && git commit -m "feat: log profile creation errors"

# Re-verify
$ aiguard verify --reviewers claude_code,codex
Verdict: APPROVE
# Both reviewers are now satisfied
```

### Pain Points

- **Info vs Medium** (same bucket) is not flagged as disagreement — only Info vs High is.
- If both reviewers agree with the same severity, no disagreement is recorded.
- AI reviewers can hallucinate issues. Cross-check findings against actual code.

---

## Scenario 4: "Test Presence Warning — But I Added Tests!"

### Situation

```
$ aiguard verify --local-only

Verdict: WARN
Summary: Verdict: WARN. Deterministic checks: 1 WARN, 2 Info

Address these warnings:
- test_presence: behavior files changed without any test file changes: internal/handlers/user.go
```

You added tests in `internal/handlers/user_test.go`, but the warning still fires.

### What This Means

The `test_presence` check looks for test files matching the pattern `*_test.go` or `.test.ts`, etc. It checks the **changed files list**, not whether test files exist on disk.

If you:
1. Modified `internal/handlers/user.go`
2. Added a new test file `internal/handlers/user_test.go`

Both are in the diff. The check should NOT warn. Let's diagnose.

### How to Fix It

**Check what files are in the diff:**

```bash
$ git diff --name-only origin/main...HEAD
internal/handlers/user.go
internal/handlers/user_test.go

$ aiguard verify --local-only
# Should be APPROVE (both behavior and test files changed)
```

**If test file is not showing up:**

```bash
# Check git status
$ git status
modified: internal/handlers/user.go
untracked: internal/handlers/user_test.go  # Not staged!

# Stage the test file
$ git add internal/handlers/user_test.go
$ git commit -m "test: add user handler tests"

# Re-verify
$ aiguard verify --local-only
Verdict: APPROVE
```

**If warning persists:**

```bash
# The check might be using a different base branch
$ aiguard verify --local-only --base develop
# If your source branch is develop, use that as base

# Or check your config
$ cat .aiguard/config.yaml | grep -A2 "source_branch"
source_branch: main
```

### Prevention

- Always commit tests in the same commit as behavior changes (or same pull request)
- Stage all changes before verify: `git add .`

---

## Scenario 5: "AC IDs Don't Match — Plan Drift on AC Traceability"

### Situation

Your digest has:

```json
[
  {"id":"AC1","text":"User can create profile",...},
  {"id":"AC2","text":"Profile persists to database",...}
]
```

Your plan has:

```json
{
  "steps":[
    {"id":"S1","ac_ids":["ac-1"],...},
    {"id":"S2","ac_ids":["AC_002"],..."
  ]
}
```

Verdict:

```
ac_traceability: AC1 (ticket_acceptance_criteria): "User can create profile" not covered by any plan step
ac_traceability: AC2 (ticket_acceptance_criteria): "Profile persists to database" not covered by any plan step
```

### What This Means

The AC IDs (`AC1`, `ac-1`, `AC_002`) don't match because:
- Digest uses uppercase: `AC1`, `AC2`
- Plan uses lowercase with dashes: `ac-1`
- Plan uses leading zeros: `AC_002`

AIGuard **normalizes** these automatically, but the normalization requires all parts to match. `AC_002` normalizes to `ac2`, not `ac1`.

### How to Fix It

**Make all AC IDs consistent:**

```bash
# Edit your plan.json
{
  "steps":[
    {"id":"S1","ac_ids":["AC1"],...},     # ← changed from ac-1
    {"id":"S2","ac_ids":["AC2"],...}      # ← changed from AC_002
  ]
}

$ aiguard verify --local-only
# ac_traceability should now pass
Verdict: APPROVE
```

**Or let the AI regenerate the plan:**

```bash
$ aiguard plan create --force --reason "Regenerate plan with consistent AC IDs"
# This will re-run the plan generation with correct ID matching
```

### Prevention

- When creating a plan manually, copy AC IDs from `acceptance-criteria.json` exactly
- When asking AI to generate a plan, include in the prompt: "Use AC IDs exactly as listed: AC1, AC2, AC3" (no variations)

---

## Scenario 6: "Privacy Mode — File Contents Excluded from AI Review"

### Situation

```bash
$ aiguard verify --reviewers claude_code
# AI reviewer runs but says:
# "I cannot see the actual code in internal/handlers/user.go; I'm reviewing only the diff."
```

Your AI reviewer can't make detailed suggestions because file contents are excluded by default.

### What This Means

By default, AIGuard:

```yaml
# .aiguard/config.yaml
privacy:
  include_actual_files_by_default: false  # ← Files are excluded
```

This means:
- Diff is sent to AI (so they see what changed)
- Actual file contents are NOT sent (so they can't see context)
- Secrets, forbidden paths, and .env are always excluded regardless

For code review, this is too restrictive. For security-sensitive diffs, it's appropriate.

### How to Fix It

**Option A: Enable file inclusion globally**

```yaml
# .aiguard/config.yaml
privacy:
  include_actual_files_by_default: true  # Now files are included
```

Then re-verify:

```bash
$ aiguard verify --reviewers claude_code
# AI reviewer now sees file contents; better context for suggestions
```

**Option B: Only for local review (not CI)**

Keep the default (false) for safety, and only enable locally:

```bash
# Local machine — enable file inclusion
$ aiguard verify --reviewers claude_code
# (modify config temporarily, or use a script)

# CI pipeline — use defaults (files excluded, safe)
$ aiguard verify --local-only
```

### Pain Points

- Files are still redacted if they match `forbidden_paths` (`.env`, `*.key`, etc.)
- Large files (>50KB default) are excluded to save tokens
- If AI suggests changes without context, it might be because file contents weren't included

---

## Scenario 7: "Rationale Mode for CI/CD Pipeline"

### Situation

You're integrating AIGuard into CI. The `rationale update` command breaks because it tries to read stdin:

```
# In CI script:
aiguard rationale update --interactive
# ERROR: hangs waiting for stdin (CI has no TTY)
```

### What This Means

`--interactive` requires manual input via stdin. In CI/CD pipelines, there's no interactive terminal, so the command hangs.

### How to Fix It

**Option A: Use --file with a generated YAML**

```bash
# In your CI/deployment pipeline:
# Generate rationale from code changes
cat > rationale.yaml <<EOF
$(git diff --name-only origin/main...HEAD | while read f; do
  echo "$f: $(git log --oneline -1 -- $f | cut -d' ' -f2-)"
done)
EOF

aiguard rationale update --file rationale.yaml
```

**Option B: Use --ai with your agent**

```bash
# Delegate to AI — works in CI if your agent is available
aiguard rationale update --ai
```

**Option C: Skip rationale in CI**

If rationale is optional:

```bash
# Only run rationale in local development
if [[ -t 0 ]]; then  # if stdin is a TTY
  aiguard rationale update --interactive
else
  echo "Skipping interactive rationale in CI"
fi
```

### Typical CI Script

```bash
#!/usr/bin/env bash
set -euo pipefail

cd /repo
aiguard init
aiguard snapshot update --full --no-ai

# Ingest the ticket (from env var in CI)
echo "$TICKET_MD" | aiguard req ingest --paste

aiguard digest create --local-only
aiguard verify --local-only

# Optional: generate rationale from commit messages
git log --oneline --reverse $BASE..HEAD | \
  awk '{print $1}' | \
  xargs -I {} git diff {}^..{} --name-only | \
  sort -u | \
  while read f; do
    echo "$f: $(git log -1 --format=%s -- $f || echo 'see commit')"
  done > /tmp/rationale.yaml

aiguard rationale update --file /tmp/rationale.yaml || true
aiguard verify --local-only

echo "✓ All checks passed"
```

---

## Scenario 8: "Clarification Questions Block the Pipeline"

### Situation

```bash
$ aiguard guidance create
Error: there are open blocking clarification questions; run 'aiguard clarify' to resolve them (or use --force --reason)
```

You haven't answered a blocking question yet, so downstream commands are blocked.

### What This Means

The digest extracted a question marked as "blocking" (user uncertain, not definitional). Examples:

- "Should we use microservices or monolith?"
- "What's the SLA for profile queries?"
- "Do we need caching?"

These must be answered before plan creation to ensure the plan aligns with requirements.

### How to Fix It

**Option A: Answer the question**

```bash
$ aiguard clarify
Q1 (blocking): Should profiles be public or private by default?
Q2 (blocking): Do we need versioning?

$ aiguard clarify answer --id Q1 --text "Private by default, user-configurable"
Q1 answered.

$ aiguard clarify answer --id Q2 --assume
Q2 assumed (not fully answered, but documented).

$ aiguard guidance create
# Now proceeds
```

**Option B: Force through (not recommended)**

```bash
$ aiguard guidance create --force --reason "PM approved verbally, no time for formal answers"
# Gate is recorded as FORCED in the audit log
# This is noted in the report for later review
```

### Pain Points

- Questions are extracted heuristically (things marked uncertain in AC text)
- False positives can occur; use `--force --reason` if the question is truly not relevant
- Answering a question doesn't change the verdict; it's purely documentational

---

## Scenario 9: "First Run: Branch Auto-Detection Picks Wrong Branch"

### Situation

You run `aiguard init` and it picks the wrong source branch:

```bash
$ aiguard init
Initialized AIGuard workspace at /home/alice/myrepo/.aiguard/config.yaml
Source branch: master  # ← Expected 'main', got 'master'
```

### What This Means

Your repo has:
- No remote (no `origin/HEAD` to read)
- Both `main` and `master` exist locally
- AIGuard picked `master` (fallback heuristic)

But your actual source branch is `main`.

### How to Fix It

**Option A: Re-init with explicit branch**

```bash
$ aiguard init --source-branch main --force
Initialized AIGuard workspace at /home/alice/myrepo/.aiguard/config.yaml
Source branch: main
```

**Option B: Edit config.yaml manually**

```yaml
# .aiguard/config.yaml
project:
  source_branch: main
  default_compare_ref: main
```

### Prevention

- Set up a remote and ensure `origin/HEAD` points to the right branch:
  ```bash
  git remote add origin https://github.com/user/repo.git
  git symbolic-ref refs/remotes/origin/HEAD refs/remotes/origin/main
  aiguard init  # Now correctly detects 'main'
  ```

---

## Scenario 10: "Snapshot Stale — Too Many Commits Since Last Update"

### Situation

```bash
$ aiguard snapshot status
Snapshot at commit abc1234 (version 1)
Status: stale (22 commits since snapshot)
Warning: snapshot is older than 20 commits; run 'aiguard snapshot update' to refresh
```

Your codebase has evolved since the snapshot. The heuristic warning triggers after 20 commits (configurable).

### What This Means

If your codebase structure or component organization changed significantly, AI reviewers and context packing might use outdated information.

### How to Fix It

```bash
$ aiguard snapshot update --full
Snapshot updated.

$ git add .aiguard/snapshot/*.json .aiguard/snapshot/snapshot.md
$ git commit -m "chore: refresh snapshot"

$ aiguard verify --local-only
# Now uses fresh snapshot
```

### Pain Points

- Snapshot is optional for `verify --local-only` (deterministic checks don't use it)
- Only matters if you're using AI reviewers (`--reviewers`) or the gated workflow
- For fast-moving codebases, raise the threshold in config:

```yaml
# .aiguard/config.yaml
snapshot:
  stale_after_commits: 50  # ← Less frequent warnings
```

