# AIGuard

Local-first CLI that verifies AI-generated diffs against business requirements before you push.

You point it at your work branch, it answers: **does this `git diff` satisfy the ticket — without leaking secrets, touching forbidden paths, or skipping tests?**

```bash
go install github.com/mchau/aiguard/cmd/aiguard@latest
```

---

## The 90% workflow: `verify` only

Most users only need three commands. No gate ceremony, no AI subscription, no SaaS.

```bash
cd your-repo
aiguard init               # writes .aiguard/ — autodetects 'main' from origin/HEAD
git checkout -b feature
# ...you (or your AI) make changes...
aiguard verify --local-only
```

`verify --local-only` runs 11 deterministic checks against `git diff <base>...HEAD`:

| Check | Severity | What it catches |
|---|---|---|
| `forbidden_path` | BLOCKED | `.env`, `*.pem`, `secrets/**` in the diff |
| `secret_detection` | BLOCKED | AWS keys, GitHub PATs, bearer tokens, `password=`, private key blocks |
| `risk_path` | Critical/High/Medium | Changes to `auth/`, `billing/`, `migrations/`, etc. |
| `diff_size` | WARN | Oversized patches |
| `dependency_changes` | WARN | `go.mod`, `package.json`, `Gemfile.lock` touched |
| `migration_changes` | WARN | DB migration files touched |
| `test_presence` | WARN | Behavior file changed without a test file |
| `plan_drift` | WARN | Risk-path file changed without a plan step authorizing it |
| `snapshot_freshness` | Info | Codebase map stale vs HEAD |
| `ac_traceability` | High/Info | Source-provided ACs missing from the plan |
| `rationale_missing` | High | Changed file without a documented reason |

Exit code is non-zero on `BLOCKED` or `REJECT`. Reports land at `.aiguard/reviews/final-verdict.{md,json}`.

That's it. You can stop reading here if `verify --local-only` is all you need.

---

## When verify rejects: close the loop

```bash
aiguard verify --copy-fix    # copy the recommended fix prompt to clipboard
aiguard verify --auto-fix    # send the prompt to your configured agent, save the transcript
```

`--auto-fix` refuses to run on `BLOCKED` verdicts — forbidden paths and secrets need manual triage, not an AI suggestion.

---

## Multi-reviewer verification (opt-in)

If you have `claude` and `codex` CLIs installed, reduce same-model bias by running both:

```bash
aiguard verify --reviewers claude_code,codex
```

The harness aggregates findings, detects disagreements (coarse buckets: soft vs hard severity — Info-vs-Medium is not flagged), and writes `disagreement-report.md`. Deterministic `BLOCKED` is never overridden by reviewer output.

---

## The full gated workflow (opt-in, more ceremony)

If you want the harness to enforce the digest → guidance → plan → implementation pipeline:

```bash
aiguard snapshot update         # codebase map (one-time, refresh periodically)
aiguard req ingest --file ticket.md
aiguard digest create           # extract source ACs + AI inferences
aiguard digest approve --reason "...
aiguard guidance create         # tech guidance against approved digest
aiguard guidance approve --reason "..."
aiguard plan create             # implementation plan
aiguard plan approve --reason "..."
aiguard prompt --copy           # assemble prompt for your AI tool
# ...you implement...
aiguard rationale update --ai   # or --file rationale.yaml, or --interactive
aiguard checkpoint              # mid-implementation check
aiguard verify                  # final verdict
```

Each gate has `--force --reason "..."` if you need to skip; forced gates are recorded in the audit log.

---

## Principles

1. The `git diff` is the source of truth for review.
2. Source-provided acceptance criteria are authoritative — never dropped, never replaced by AI inference.
3. Deterministic checks run before any AI is involved.
4. AI reviewers cannot override `BLOCKED` deterministic findings.
5. Local-first and privacy-safe by default. `.env`, secrets, and forbidden paths are excluded from context packs even when file inclusion is on.

---

## Configuration

`.aiguard/config.yaml` is created by `init` with sensible defaults. Customize:

- `project.source_branch` — defaults to `main` (autodetected from `origin/HEAD`).
- `risk_paths.critical/high/medium` — globs flagged at corresponding severity.
- `forbidden_paths` — globs that always BLOCK.
- `agents.claude_code.command` / `agents.codex.command` — bin names; templates are config-driven, no vendor flags hardcoded.
- `step_routing` — which `model_profile` is used for each pipeline step.

See `examples/config.yaml` for a fuller annotated example.

---

## Documentation

- `docs/PRD.md` — product requirements
- `docs/SPEC.md` — technical specification
- `CLAUDE.md` — instructions used to build AIGuard with Claude Code

---

## Development

```bash
make build           # go build ./...
make test            # go test ./...
make lint            # go vet ./...
./scripts/smoke.sh   # end-to-end smoke against a real or fake agent
```
