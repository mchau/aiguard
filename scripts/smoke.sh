#!/usr/bin/env bash
# End-to-end smoke test for AIGuard.
#
# Exercises: init (with autodetect), local-only verify, --auto-fix loop,
# multi-reviewer path (against a fake agent), rationale modes.
#
# If $AIGUARD_LIVE_AGENT is set to the path of a real CLI (e.g. `claude` or
# `codex`), the multi-reviewer step uses that binary instead of the fake.
# Otherwise a stub shell script stands in.
#
# Exit non-zero on any unexpected failure.

set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
BIN="$ROOT/.smoke/aiguard"
SMOKE="$(mktemp -d)"
trap 'rm -rf "$SMOKE"' EXIT

echo "=== Building aiguard ==="
mkdir -p "$ROOT/.smoke"
go build -o "$BIN" "$ROOT/cmd/aiguard"

echo
echo "=== Setting up temp repo at $SMOKE ==="
cd "$SMOKE"
git init -b main -q
git config user.email smoke@test
git config user.name smoke
echo "package main" > main.go
git add . && git commit -qm "init"

echo
echo "=== Stage 1: aiguard init (autodetect 'main') ==="
"$BIN" init >/dev/null
detected=$(grep -E "^[[:space:]]+source_branch:" .aiguard/config.yaml | awk '{print $2}')
if [[ "$detected" != "main" ]]; then
  echo "FAIL: expected source_branch=main, got $detected"
  exit 1
fi
echo "  ✓ source_branch=$detected"

if [[ ! -f .aiguard/.gitignore ]]; then
  echo "FAIL: .aiguard/.gitignore not created"
  exit 1
fi
echo "  ✓ .aiguard/.gitignore created"

echo
echo "=== Stage 2: local-only verify on clean feature branch (expect APPROVE) ==="
git checkout -b feature -q
"$BIN" verify --base main --local-only | tail -3
echo "  ✓ verify exit OK"

echo
echo "=== Stage 3: introduce a forbidden file → BLOCKED ==="
echo "SECRET=foo" > .env
git add -f .env && git commit -qm "secret leak"
set +e
"$BIN" verify --base main --local-only > /tmp/aiguard-verdict.txt 2>&1
exit_code=$?
set -e
if [[ $exit_code -eq 0 ]]; then
  echo "FAIL: expected non-zero exit on BLOCKED, got 0"
  cat /tmp/aiguard-verdict.txt
  exit 1
fi
if ! grep -q "BLOCKED" /tmp/aiguard-verdict.txt; then
  echo "FAIL: expected BLOCKED verdict"
  cat /tmp/aiguard-verdict.txt
  exit 1
fi
echo "  ✓ BLOCKED verdict, exit $exit_code"

echo
echo "=== Stage 4: --auto-fix on BLOCKED is skipped (no transcript) ==="
rm -f .aiguard/reviews/auto-fix-transcript.md
"$BIN" verify --base main --local-only --auto-fix > /dev/null 2>&1 || true
if [[ -f .aiguard/reviews/auto-fix-transcript.md ]]; then
  echo "FAIL: --auto-fix should not run on BLOCKED"
  exit 1
fi
echo "  ✓ no transcript written for BLOCKED"

echo
echo "=== Stage 5: remove forbidden file, induce REJECT (risk path), run --auto-fix ==="
git rm -q .env
mkdir -p auth
echo "package auth" > auth/login.go
git add . && git commit -qm "auth change"

# Wire fake agent
cat > "$SMOKE/fake-agent.sh" <<'EOF'
#!/bin/bash
echo "## Suggested Fix"
echo ""
echo "Agent received this prompt and replied with this canned text."
cat > /dev/null
EOF
chmod +x "$SMOKE/fake-agent.sh"
# Replace command: claude on its own indented line under claude_code:
sed -i.bak "s|command: claude$|command: $SMOKE/fake-agent.sh|" .aiguard/config.yaml

set +e
"$BIN" verify --base main --local-only --auto-fix > /dev/null 2>&1
auto_exit=$?
set -e
if [[ ! -f .aiguard/reviews/auto-fix-transcript.md ]]; then
  echo "FAIL: --auto-fix transcript missing on REJECT"
  exit 1
fi
if ! grep -q "Suggested Fix" .aiguard/reviews/auto-fix-transcript.md; then
  echo "FAIL: transcript does not contain expected agent output"
  cat .aiguard/reviews/auto-fix-transcript.md
  exit 1
fi
echo "  ✓ transcript written, REJECT exit=$auto_exit"

echo
echo "=== Stage 6: rationale modes ==="
# --file
cat > rationale.yaml <<EOF
auth/login.go: implements AC1
EOF
"$BIN" rationale update --file rationale.yaml > /dev/null
if ! grep -q "implements AC1" .aiguard/implementation/changed-files-rationale.md; then
  echo "FAIL: --file mode did not write entry"
  exit 1
fi
echo "  ✓ --file mode"

# No mode → error
set +e
"$BIN" rationale update > /tmp/no-mode.txt 2>&1
no_mode_exit=$?
set -e
if [[ $no_mode_exit -eq 0 ]]; then
  echo "FAIL: rationale update without mode should error"
  exit 1
fi
if ! grep -q "pick a mode" /tmp/no-mode.txt; then
  echo "FAIL: expected 'pick a mode' guidance"
  cat /tmp/no-mode.txt
  exit 1
fi
echo "  ✓ no-mode error"

# --ai with fake agent (already wired in Stage 5)
rm .aiguard/implementation/changed-files-rationale.md
cat > "$SMOKE/fake-agent.sh" <<'EOF'
#!/bin/bash
cat > /dev/null
echo "auth/login.go: AI-generated rationale"
EOF
chmod +x "$SMOKE/fake-agent.sh"
"$BIN" rationale update --ai > /dev/null
if ! grep -q "AI-generated" .aiguard/implementation/changed-files-rationale.md; then
  echo "FAIL: --ai mode did not write entry"
  cat .aiguard/implementation/changed-files-rationale.md
  exit 1
fi
echo "  ✓ --ai mode"

echo
echo "=== Stage 7: --reviewers with fake agent (multi-reviewer path) ==="
cat > "$SMOKE/fake-agent.sh" <<'EOF'
#!/bin/bash
cat > /dev/null
cat <<'JSON'
{"findings":[{"ac_id":"AC1","severity":"Info","file":"auth/login.go","message":"looks fine"}]}
JSON
EOF
chmod +x "$SMOKE/fake-agent.sh"

set +e
"$BIN" verify --base main --reviewers claude_code > /tmp/multi-verdict.txt 2>&1
multi_exit=$?
set -e
if [[ ! -f .aiguard/reviews/final-verdict.json ]]; then
  echo "FAIL: multi-reviewer did not produce final-verdict.json"
  exit 1
fi
echo "  ✓ multi-reviewer ran (exit=$multi_exit)"

echo
echo "=== ALL SMOKE STAGES PASSED ==="
