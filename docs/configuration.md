# Configuration Reference

Comprehensive guide to customizing `.aiguard/config.yaml`.

## File Location

```
.aiguard/config.yaml
```

Created automatically by `aiguard init`. Edit it directly to customize behavior.

---

## Project Section

```yaml
project:
  name: ""                    # Optional: project name for documentation
  source_branch: main         # Default: autodetected from origin/HEAD, falls back to main
  default_compare_ref: main   # Used when --base is not specified
```

### source_branch

The authoritative branch to diff against. Typically `main`, `develop`, or `master`.

```bash
# Set explicitly if autodetection fails
aiguard init --source-branch develop --force
```

All `aiguard verify` commands use this as the default base unless overridden:

```bash
aiguard verify --base develop  # Diff develop...HEAD
aiguard verify --base main     # Diff main...HEAD (if main is not source_branch)
```

---

## Privacy Section

Controls what data is sent to AI reviewers.

```yaml
privacy:
  default_mode: local_first                      # Always exclude secrets locally
  allow_external_ai: false                       # Never upload code to cloud APIs
  redact_secrets: true                           # Strip known secret patterns
  include_actual_files_by_default: false         # Don't send file contents to AI
  max_file_bytes_for_context: 50000              # Max file size (50KB)
```

### include_actual_files_by_default

**false (recommended for security)**: File contents are not sent to AI reviewers. Only diff is sent.

```yaml
privacy:
  include_actual_files_by_default: false
```

AI reviewers see:
```diff
+ func CreateUser(name, email string) error {
+   db.Insert(...)
+   return nil
+ }
```

But NOT the surrounding context (imports, type definitions, etc.).

**true (useful for code review)**: File contents are included.

```yaml
privacy:
  include_actual_files_by_default: true
```

AI can now see the full file and make better suggestions. Still excludes:
- Files matching `forbidden_paths` (`.env`, `*.key`, `secrets/**`)
- Files larger than `max_file_bytes_for_context`
- Any file redacted by `redact_secrets`

### redact_secrets

Always enabled. Known secret patterns are stripped from context before sending to AI:

- AWS access keys (`AKIA...`)
- GitHub PATs (`ghp_...`)
- Slack tokens (`xox[baprs]-...`)
- Bearer tokens
- Hardcoded `password=`, `api_key=`, `secret=` assignments

These are redacted from **context packs** sent to AI, not from your actual code.

### max_file_bytes_for_context

Files larger than this are excluded from context packs (AI has limited context windows).

```yaml
privacy:
  max_file_bytes_for_context: 100000  # 100KB instead of 50KB
```

---

## Snapshot Section

Controls codebase mapping and update behavior.

```yaml
snapshot:
  include_tree: true                  # Generate folder tree
  include_component_notes: true       # Scan for comments marking components
  update_from_diff: true              # Incremental updates
  max_tree_depth: 8                   # Folder nesting depth
  stale_after_commits: 20             # Warning threshold
  ignore:
    - ".git/**"
    - "node_modules/**"
    - "vendor/**"
    - "dist/**"
    - "build/**"
    - ".aiguard/**"
```

### ignore

Glob patterns (same syntax as `.gitignore`). Files matching these are excluded from the snapshot.

```yaml
snapshot:
  ignore:
    - ".git/**"
    - "node_modules/**"
    - ".terraform/**"        # Add for Terraform projects
    - "venv/**"              # Add for Python projects
    - "target/**"            # Add for Rust projects
```

### stale_after_commits

Warning threshold. If `HEAD` is more than N commits ahead of the snapshot, a warning fires.

```yaml
snapshot:
  stale_after_commits: 50    # Less frequent warnings for fast-moving repos
```

Set to 0 to disable staleness warnings.

---

## Risk Paths Section

Glob patterns that trigger risk-level warnings. Three levels: `critical`, `high`, `medium`.

```yaml
risk_paths:
  critical:                             # Triggers BLOCKED or High
    - ".env"
    - ".env.*"
    - "secrets/**"
    - "infra/prod/**"
  
  high:                                 # Triggers High
    - "auth/**"
    - "billing/**"
    - "payments/**"
    - "permissions/**"
    - "migrations/**"
    - "database/**"
    - "db/migrate/**"
  
  medium:                               # Triggers Medium
    - "config/**"
    - "deployment/**"
```

### Customization Example

For a Python Django project:

```yaml
risk_paths:
  critical:
    - ".env"
    - ".env.*"
    - "secrets/"
    - "deploy/prod/"
  
  high:
    - "auth/"
    - "accounts/"
    - "payments/"
    - "migrations/"
    - "settings.py"          # Django settings
  
  medium:
    - "views/"
    - "api/"
    - "utils/"
```

### Severity Levels

- **critical** → BLOCKED (changes must be reverted or thoroughly justified)
- **high** → High severity (changes allowed but flag in `plan_drift` check if plan doesn't cover them)
- **medium** → Medium severity (informational warning)

---

## Forbidden Paths Section

Files that **never** appear in a diff. Found = immediate BLOCKED verdict.

```yaml
forbidden_paths:
  - ".env"
  - ".env.*"
  - "secrets/**"
  - "*.pem"
  - "*.key"
```

### When to Use

Use this for files that should **never** be committed:

```yaml
forbidden_paths:
  - ".env"
  - ".env.local"
  - "*.pem"
  - "*.key"
  - "*.p8"
  - "oauth_creds.json"
  - "private_key.txt"
```

**Do not** use for files that might legitimately change (like `config.yaml`). Use `risk_paths.high` instead.

---

## Test Commands Section

Shell commands to run tests. AIGuard doesn't enforce test passage, but can run them and report results.

```yaml
test_commands:
  - name: unit
    command: "go test ./..."
    timeout_seconds: 600
  
  - name: integration
    command: "make test-integration"
    timeout_seconds: 1200
```

### Example for Python

```yaml
test_commands:
  - name: unit
    command: "pytest tests/unit --tb=short"
    timeout_seconds: 300
  
  - name: integration
    command: "pytest tests/integration -v"
    timeout_seconds: 600
  
  - name: lint
    command: "pylint --errors-only src/"
    timeout_seconds: 120
```

### Timeout

Each command has a timeout in seconds. If the command takes longer, AIGuard kills it and reports TIMEOUT.

```yaml
test_commands:
  - name: slow_e2e
    command: "npm run test:e2e"
    timeout_seconds: 3600  # 1 hour for slow E2E tests
```

---

## Model Profiles Section

Predefined profiles combining a provider, model, and effort level.

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
```

### Customization

Add your own profiles or modify the defaults:

```yaml
model_profiles:
  cheap_fast:
    provider: claude_code
    model_alias: haiku
    effort: low
  
  quality_review:
    provider: claude_code
    model_alias: opus
    effort: high
  
  code_gen:
    provider: claude_code
    model_alias: sonnet
    effort: medium
  
  codex_review:
    provider: codex
    model_alias: gpt-4
    effort: high
```

### Fields

- **provider**: Agent name from `agents` section (e.g., `claude_code`, `codex`)
- **model_alias**: Model name (passed as `--model <value>` to the agent)
- **effort**: Effort level (passed as `--effort <value>` to the agent; "low", "medium", "high")

---

## Step Routing Section

Maps workflow steps to model profiles.

```yaml
step_routing:
  snapshot_update: cheap_fast           # Fast cheap model for snapshots
  requirement_digest: reasoning_high    # Strongest model for requirements
  brainstorm: reasoning_high
  test_strategy: reasoning_high
  tech_guidance: reasoning_high
  guidance_review: reasoning_high
  plan_create: reasoning_high
  plan_review: reasoning_high
  implementation: coding_balanced       # Balanced for coding
  checkpoint_review: reasoning_high
  final_review: reasoning_high
```

### Customization

Use a cheaper model for some steps:

```yaml
step_routing:
  snapshot_update: cheap_fast
  requirement_digest: reasoning_high
  tech_guidance: reasoning_high
  plan_create: reasoning_high
  implementation: cheap_fast            # Use cheaper model for coding
  final_review: reasoning_high          # Use best model for final review
```

Or use Codex for some steps:

```yaml
model_profiles:
  codex_review:
    provider: codex
    model_alias: gpt-4
    effort: high

step_routing:
  final_review: codex_review            # Use Codex for final review
```

---

## Agents Section

Configures how to invoke external agents (claude, codex, custom tools).

```yaml
agents:
  claude_code:
    enabled: true
    command: claude
    prompt_arg: "-p"
    model_arg_template: "--model {{.ModelAlias}}"
    effort_arg_template: "--effort {{.Effort}}"
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

### Fields

- **enabled**: Set to `false` to disable an agent without removing config
- **command**: CLI command to invoke (must be on PATH)
- **prompt_arg**: Flag to pass prompt to stdin or as a positional argument
  - `"-p"` → `claude -p < prompt.txt`
  - `""` → prompt passed via stdin
- **model_arg_template**: Template for model selection
  - `"--model {{.ModelAlias}}"` → rendered as `--model haiku`
  - `""` → no model arg (use agent's default)
- **effort_arg_template**: Template for effort selection
  - `"--effort {{.Effort}}"` → rendered as `--effort high`
  - `""` → no effort arg
- **supports_stdin**: If true, prompt is passed via stdin; if false, via `prompt_arg`
- **timeout_seconds**: Max runtime per agent call

### Custom Agent Example

```yaml
agents:
  my_custom_agent:
    enabled: true
    command: /usr/local/bin/my-ai-tool
    prompt_arg: "--prompt"               # Pass prompt as --prompt /tmp/prompt.txt
    model_arg_template: "--model {{.ModelAlias}}"
    effort_arg_template: ""
    supports_stdin: false                # Don't use stdin
    timeout_seconds: 1800
```

### Adding a New Agent

1. Install the agent binary on PATH (e.g., `which codex` should work)
2. Add it to `agents`:

```yaml
agents:
  my_new_agent:
    enabled: true
    command: my_new_agent
    prompt_arg: "-p"
    model_arg_template: ""
    effort_arg_template: ""
    supports_stdin: true
    timeout_seconds: 3600
```

3. Create a model profile for it:

```yaml
model_profiles:
  my_agent_profile:
    provider: my_new_agent
    model_alias: default
    effort: medium
```

4. Route steps to it:

```yaml
step_routing:
  final_review: my_agent_profile
```

---

## Example Configurations

### Minimal (Deterministic Checks Only)

```yaml
project:
  source_branch: main

privacy:
  default_mode: local_first
  include_actual_files_by_default: false

risk_paths:
  high:
    - "auth/**"
    - "migrations/**"

forbidden_paths:
  - ".env"
  - "*.key"
```

### Full (All Features)

```yaml
project:
  name: MyApp
  source_branch: main

privacy:
  include_actual_files_by_default: true

snapshot:
  stale_after_commits: 30

risk_paths:
  critical:
    - ".env"
    - "secrets/**"
  high:
    - "auth/**"
    - "billing/**"
    - "migrations/**"
  medium:
    - "config/**"

forbidden_paths:
  - ".env*"
  - "secrets/**"
  - "*.pem"
  - "*.key"

test_commands:
  - name: unit
    command: "go test ./..."
    timeout_seconds: 600
  - name: integration
    command: "make test-integration"
    timeout_seconds: 1200

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
  codex_review:
    provider: codex
    model_alias: gpt-4
    effort: high

step_routing:
  snapshot_update: cheap_fast
  requirement_digest: reasoning_high
  test_strategy: reasoning_high
  tech_guidance: reasoning_high
  guidance_review: reasoning_high
  plan_create: reasoning_high
  plan_review: reasoning_high
  implementation: coding_balanced
  checkpoint_review: reasoning_high
  final_review: codex_review

agents:
  claude_code:
    enabled: true
    command: claude
    prompt_arg: "-p"
    model_arg_template: "--model {{.ModelAlias}}"
    effort_arg_template: "--effort {{.Effort}}"
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

### Python Django Project

```yaml
project:
  source_branch: main

risk_paths:
  high:
    - "auth/"
    - "accounts/"
    - "payments/"
    - "migrations/"
    - "settings.py"

forbidden_paths:
  - ".env"
  - ".env.*"
  - "*.key"
  - "credentials.json"

snapshot:
  ignore:
    - "venv/**"
    - ".git/**"
    - "__pycache__/**"
    - "*.pyc"
    - "node_modules/**"

test_commands:
  - name: unit
    command: "python manage.py test"
    timeout_seconds: 600
  - name: lint
    command: "pylint django_app/ --fail-under 8.0"
    timeout_seconds: 300
```

### Go Monorepo

```yaml
project:
  source_branch: develop

risk_paths:
  high:
    - "services/auth/**"
    - "services/billing/**"
    - "infra/**"
    - "schema/migrations/**"

snapshot:
  ignore:
    - ".git/**"
    - "vendor/**"
    - "bin/**"
    - "dist/**"
    - ".vscode/**"

test_commands:
  - name: unit
    command: "go test ./..."
    timeout_seconds: 600
  - name: integration
    command: "go test -tags=integration ./..."
    timeout_seconds: 1200
  - name: lint
    command: "golangci-lint run"
    timeout_seconds: 300
```

---

## Validation

AIGuard validates config on load. If there are errors:

```bash
$ aiguard verify
Error: load config: validation failed:
- model profile "reasoning_high" not found in agent "claude_code"
- glob pattern ".env[" is invalid

Run 'aiguard doctor' to diagnose configuration issues.
```

Fix the issues and re-run.

```bash
$ aiguard doctor
Checking AIGuard health...

✓ config.yaml loads and validates
✓ risk_paths globs valid
✓ forbidden_paths globs valid
✓ model_profiles all referenced agents exist
✓ step_routing all profiles exist
```

