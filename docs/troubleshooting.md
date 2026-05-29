# Troubleshooting Guide

Common errors and how to resolve them.

---

## "no .aiguard/ or .git/ found"

### Error

```
Error: no .aiguard/ or .git/ found; run 'aiguard init' to initialize this project
Run 'aiguard init' to initialize a workspace first.
```

### Cause

You ran an AIGuard command outside a git repo or without initializing `.aiguard/`.

### Solution

1. Ensure you're in a git repository:

```bash
$ cd /path/to/your/repo
$ git status   # If this works, you're in a git repo
```

2. Initialize AIGuard:

```bash
$ aiguard init
Initialized AIGuard workspace at ...
```

---

## "config.yaml not found"

### Error

```
Error: load config: open .aiguard/config.yaml: no such file or directory
```

### Cause

`.aiguard/config.yaml` doesn't exist (init didn't run or failed).

### Solution

```bash
$ aiguard init
# If init fails, check:
$ git status                    # Are you in a git repo?
$ ls -la .aiguard/             # Does .aiguard/ exist?
```

If `.aiguard/` exists but is empty:

```bash
$ aiguard init --force         # Re-init and overwrite
```

---

## "source_branch 'develop' does not resolve"

### Error

```
Error: load config: validate: source_branch 'develop' does not resolve
```

### Cause

Your configured `source_branch` doesn't exist locally or remotely.

### Solution

Check what branches exist:

```bash
$ git branch -a
main
feature/work
remotes/origin/main
remotes/origin/develop
```

If the branch exists remotely but not locally, fetch it:

```bash
$ git fetch origin develop
$ git branch -t develop origin/develop
```

Then update config:

```bash
$ aiguard init --source-branch develop --force
```

Or edit `.aiguard/config.yaml` directly:

```yaml
project:
  source_branch: develop
```

---

## "forbidden_path glob pattern is invalid"

### Error

```
Error: load config: validation failed:
- glob pattern ".env[" is invalid: syntax error in pattern
```

### Cause

Your glob pattern in `.aiguard/config.yaml` has invalid syntax.

### Solution

Check the pattern:

```bash
$ cat .aiguard/config.yaml | grep -A5 "forbidden_paths"
forbidden_paths:
  - ".env["        # ← Invalid! Missing closing ]
```

Fix it:

```yaml
forbidden_paths:
  - ".env*"        # Valid: matches .env, .env.local, .env.production, etc.
```

Common mistakes:

```yaml
# WRONG
forbidden_paths:
  - ".env[" ]      # Mismatched brackets

# RIGHT
forbidden_paths:
  - ".env*"
  - ".env.*"
  - "secrets/**"
  - "*.pem"
```

---

## "model profile not found"

### Error

```
Error: load config: validation failed:
- model profile "reasoning_high" not found in agents
- agent "claude_code" does not reference model profile "reasoning_high"
```

### Cause

Your `step_routing` references a profile that doesn't exist in `model_profiles`.

### Solution

Check `step_routing` and `model_profiles`:

```bash
$ grep -A10 "step_routing:" .aiguard/config.yaml | grep "reasoning_high"
  requirement_digest: reasoning_high

$ grep -A10 "model_profiles:" .aiguard/config.yaml | grep "reasoning_high"
# Missing!
```

Add the missing profile:

```yaml
model_profiles:
  reasoning_high:
    provider: claude_code
    model_alias: opus
    effort: high
```

---

## "command not found: claude"

### Error

```
Error: agent command not found: claude
```

### Cause

The `claude` CLI is not installed or not on your `$PATH`.

### Solution

1. Install Claude Code CLI:

```bash
# macOS
brew install claude

# Or build from source
go install github.com/anthropics/claude-code@latest
```

2. Verify it's on PATH:

```bash
$ which claude
/usr/local/bin/claude

$ claude --version
claude version 1.0.0
```

3. If not using this agent, disable it in config:

```yaml
agents:
  claude_code:
    enabled: false   # ← Set to false
```

4. Update step_routing to use a different agent:

```yaml
step_routing:
  final_review: codex_review  # or another agent
```

---

## "verify hangs waiting for input"

### Error

```
$ aiguard verify
# ...nothing happens, command seems stuck...
```

### Cause

An agent is waiting for stdin (prompt not piped correctly).

### Solution

Press Ctrl+C to interrupt:

```bash
^C
Error: context deadline exceeded
```

Check your agent config:

```bash
$ grep -A5 "prompt_arg:" .aiguard/config.yaml
prompt_arg: "-p"
```

If `prompt_arg` is empty, the agent expects stdin. Ensure it's configured correctly:

```yaml
agents:
  claude_code:
    prompt_arg: "-p"              # Use this flag for prompt file
    supports_stdin: true          # Or set true for stdin
```

---

## "ac_traceability: AC not covered by plan"

### Error

```
Verdict: REJECT
...
ac_traceability: AC2 (ticket_acceptance_criteria): "persist user credentials" not covered by any plan step
```

### Cause

An acceptance criterion from the original ticket is missing from your plan.

### Solution

1. Check what ACs are defined:

```bash
$ cat .aiguard/requirements/acceptance-criteria.json | jq '.[].id'
AC1
AC2
AC3
```

2. Check what your plan covers:

```bash
$ cat .aiguard/plans/plan.json | jq '.steps[].ac_ids'
["AC1"]
["AC3"]
# AC2 is missing!
```

3. Update the plan to cover AC2. Edit `.aiguard/plans/plan.json`:

```json
{
  "steps": [
    {"id":"S1","ac_ids":["AC1"],...},
    {"id":"S2","ac_ids":["AC2"],...},    // ← Add this step for AC2
    {"id":"S3","ac_ids":["AC3"],...}
  ]
}
```

Or regenerate the plan:

```bash
$ aiguard plan create --force --reason "Regenerate to cover all ACs"
```

---

## "test_presence warning but I added tests!"

### Error

```
Verdict: WARN
test_presence: behavior files changed without any test file changes: internal/handlers/user.go
```

You added `internal/handlers/user_test.go` but the warning still fires.

### Cause

The test file was not staged in git, so it's not in the diff.

### Solution

```bash
$ git status
modified: internal/handlers/user.go
untracked: internal/handlers/user_test.go   # ← Not staged!

$ git add internal/handlers/user_test.go
$ git commit -m "test: add user handler tests"

$ aiguard verify --local-only
Verdict: APPROVE    # Now both files are in the diff
```

Or check your base branch:

```bash
$ git diff --name-only origin/main...HEAD
internal/handlers/user.go
internal/handlers/user_test.go

# If this shows both files, the check should pass
# If not, ensure your current HEAD is the test commit:
$ git log --oneline -2
abc1234 test: add user handler tests
def5678 feat: add user handler

# If not, commit the test file:
$ git add internal/handlers/user_test.go && git commit -m "test: ..."
```

---

## "BLOCKED verdict on .env — can't auto-fix"

### Error

```
Verdict: BLOCKED
...
auto-fix skipped: BLOCKED verdicts require manual remediation.
```

### Cause

A forbidden file (`.env`) is in the diff. AIGuard intentionally skips auto-fix for BLOCKED.

### Solution

Remove the forbidden file manually:

```bash
$ git status
modified: internal/handlers/user.go
new file: .env                             # ← The problem

$ git rm --cached .env                     # Remove from staging
$ rm .env                                  # Delete locally
$ git add internal/handlers/user.go
$ git commit -m "fix: remove .env from diff"

$ aiguard verify --local-only
Verdict: APPROVE    # Now clean
```

Prevent future accidents:

```bash
$ echo ".env" >> .gitignore
$ git add .gitignore && git commit -m "chore: ignore .env"
```

---

## "rationale update: pick a mode"

### Error

```
Error: pick a mode: --file <path>, --ai, or --interactive (no stdin prompting by default to keep the command scriptable)
```

### Cause

You ran `rationale update` without specifying how to provide rationale.

### Solution

Use one of three modes:

```bash
# Option 1: From a YAML file
aiguard rationale update --file rationale.yaml

# Option 2: From your AI agent
aiguard rationale update --ai

# Option 3: Interactive prompts (local only)
aiguard rationale update --interactive
```

---

## "Multi-reviewer disagreement — can't decide"

### Error

No error, but you got:

```
Verdict: WARN
...
## Reviewer Disagreements

- claude_code: [High] Missing error handling
- codex: [Info] Error handling not critical
```

### Cause

Two reviewers have different opinions on severity.

### Solution

Read both full reviews:

```bash
$ cat .aiguard/reviews/claude_code-review.md | grep -A5 "Missing error"
[HIGH] Missing error handling for duplicate profiles
Evidence: No try-catch in user creation path
Suggestion: Add validation

$ cat .aiguard/reviews/codex-review.md | grep -A5 "error handling"
[INFO] Error handling deferred to phase 2
Evidence: MVP scope doesn't require complete handling
Suggestion: Document known issues
```

Decide based on:

1. **Your ticket scope**: Does the ticket require error handling? → Choose the reviewer who aligns with scope
2. **Risk assessment**: Are the risks acceptable? → Balance risk vs schedule
3. **Team standards**: What does your code review process expect?

Then address the issue:

```bash
# Option A: Follow claude_code (add error handling)
$ # Edit code to add error handling
$ git add . && git commit -m "feat: add error handling for duplicate profiles"

# Option B: Follow codex (document limitation)
$ # Add a TODO/FIXME explaining the limitation
$ git add . && git commit -m "docs: note error handling in future phase"

# Re-verify
$ aiguard verify --reviewers claude_code,codex
Verdict: APPROVE   # Both satisfied
```

---

## "snapshot status: stale"

### Error

```
$ aiguard snapshot status
Snapshot at commit abc123 (version 1)
Status: stale (22 commits since snapshot)
Warning: snapshot is older than 20 commits; run 'aiguard snapshot update' to refresh
```

### Cause

Your codebase has evolved significantly since the snapshot. AI reviewers might use outdated information.

### Solution

Refresh the snapshot:

```bash
$ aiguard snapshot update --full
Snapshot updated.

$ git add .aiguard/snapshot/ && git commit -m "chore: refresh snapshot"

$ aiguard snapshot status
Status: fresh (0 commits since snapshot)
```

Disable the warning if snapshots are updated infrequently:

```yaml
# .aiguard/config.yaml
snapshot:
  stale_after_commits: 0   # Disable staleness warnings
```

---

## "clarification questions block plan creation"

### Error

```
Error: there are open blocking clarification questions; run 'aiguard clarify' to resolve them (or use --force --reason)
```

### Cause

The digest extracted a blocking question that hasn't been answered yet.

### Solution

Answer the questions:

```bash
$ aiguard clarify
Q1 (blocking): Should profiles be public or private?
Q2 (blocking): Do we need versioning?

$ aiguard clarify answer --id Q1 --text "Private by default"
Q1 answered.

$ aiguard clarify answer --id Q2 --assume
Q2 assumed.

$ aiguard plan create
# Now proceeds
```

Or force through if the question is not relevant:

```bash
$ aiguard plan create --force --reason "PM confirmed this is out of scope"
# Forced gates are recorded in audit.jsonl
```

---

## "verify --base doesn't find branch"

### Error

```
Error: get changed files: fatal: ambiguous argument 'develop'
```

### Cause

The branch you specified with `--base` doesn't exist.

### Solution

Check what branches exist:

```bash
$ git branch -a
main
feature/work
remotes/origin/main
remotes/origin/develop

$ git fetch origin develop        # Fetch if remote-only
$ git branch -t develop origin/develop

$ aiguard verify --base develop
```

Or use the default base (configured in config.yaml):

```bash
$ aiguard verify  # Uses config.project.source_branch
```

---

## "config validation: source branch doesn't resolve"

### Error

```
Error: load config: validate: source_branch 'develop' does not resolve
```

### Cause

The branch in config doesn't exist locally or on the remote.

### Solution

```bash
# Check what exists
$ git branch -a | grep develop

# If it exists remotely but not locally
$ git fetch origin develop && git branch -t develop origin/develop

# If it doesn't exist, pick a valid branch
$ aiguard init --source-branch main --force
```

---

## "permission denied: .aiguard/"

### Error

```
Error: create directory .aiguard/: permission denied
```

### Cause

You don't have write permissions in the current directory.

### Solution

```bash
# Check permissions
$ ls -ld .
drwxr-xr-x  user  group  .

# If not writable, change ownership or run with sudo
$ sudo aiguard init

# Or use a writable directory
$ cd /tmp && aiguard init
```

---

## "agent timeout"

### Error

```
Error: agent claude_code timed out after 3600s
```

### Cause

The agent took too long and was killed.

### Solution

1. Increase the timeout in config:

```yaml
agents:
  claude_code:
    timeout_seconds: 7200   # 2 hours instead of 1
```

2. Or reduce the prompt size (exclude more files from context):

```yaml
privacy:
  max_file_bytes_for_context: 25000   # 25KB instead of 50KB
```

3. Or check what the agent was doing:

```bash
$ cat .aiguard/reviews/claude-raw.txt   # Raw agent output
```

---

## "secret detected but it's not a real secret"

### Error

```
Verdict: BLOCKED
...
secret_detection: secret pattern detected
```

You know this isn't a real secret (it's a comment, a test fixture, etc.).

### Cause

Your diff contains a string that matches a secret pattern (like `password=`).

### Solution

1. **Best**: Remove the false positive from the diff

```bash
# Example: test fixture with fake credentials
$ git diff | grep "password"
-password=test123   # Remove this line

# Edit the file and re-commit
```

2. **Temporary**: Force through with documentation

```bash
$ aiguard verify --local-only   # Fails with BLOCKED

# This is not a real secret, but we're forced to document it
$ aiguard verify --local-only --force --reason "Test fixture with mock password; not a real secret"
# Note: --force is not a flag for verify; instead, acknowledge and document manually
```

3. **Prevent false positives**: Review the secret patterns

```bash
$ grep -r "password\|secret\|token" .aiguard/implementation/ | grep -v ".md"
# Identify false positives and remove them

$ git add . && git commit -m "fix: remove test fixtures with fake secrets"
```

---

## "gates.json permission denied"

### Error

```
Error: set gate: open .aiguard/gates.json: permission denied
```

### Cause

`.aiguard/gates.json` exists but is not writable.

### Solution

```bash
$ ls -l .aiguard/gates.json
-r--r--r--  user  group  gates.json   # ← read-only

$ chmod 644 .aiguard/gates.json        # Make writable
```

---

## "doctor reports missing agents"

### Error

```
$ aiguard doctor
⚠ claude command not found (optional)
⚠ codex command not found (optional)
```

### Cause

Optional agents are not installed.

### Solution

Option A: Install them

```bash
# Claude Code
go install github.com/anthropics/claude-code@latest

# Codex
npm install -g codex-cli
```

Option B: Disable them in config

```yaml
agents:
  claude_code:
    enabled: false
  codex:
    enabled: false
```

These are warnings, not failures. You can run AIGuard without them.

---

## "invalid yaml in config.yaml"

### Error

```
Error: load config: yaml: line 10: mapping values are not allowed in this context
```

### Cause

YAML syntax error in `.aiguard/config.yaml`.

### Solution

Common mistakes:

```yaml
# WRONG: tabs instead of spaces
privacy:
	include_actual_files_by_default: false   # ← tab here

# RIGHT: 2 spaces
privacy:
  include_actual_files_by_default: false
```

```yaml
# WRONG: missing colon
risk_paths:
  high
    - "auth/**"

# RIGHT: colon after high
risk_paths:
  high:
    - "auth/**"
```

Validate the file:

```bash
$ python -m yaml .aiguard/config.yaml
# or
$ go run cmd/validate_yaml.go .aiguard/config.yaml
```

---

## "still stuck?"

If your error isn't listed, try:

1. **Run doctor**:
   ```bash
   $ aiguard doctor
   ```

2. **Check the audit log**:
   ```bash
   $ tail -20 .aiguard/logs/audit.jsonl | jq '.'
   ```

3. **Verbose output**:
   ```bash
   $ aiguard verify --local-only 2>&1 | head -50
   ```

4. **Read the reports**:
   ```bash
   $ cat .aiguard/reviews/final-verdict.md
   ```

5. **Report the issue**:
   ```
   https://github.com/mchau/aiguard/issues
   ```

Include:

- AIGuard version: `aiguard version`
- Config (redacted): `cat .aiguard/config.yaml`
- Error output: `aiguard <command> 2>&1`
- Repo state: `git log --oneline -5` and `git status`

