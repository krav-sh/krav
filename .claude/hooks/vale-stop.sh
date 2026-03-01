#!/usr/bin/env bash
# Stop hook: format and lint all modified markdown files before ending
set -euo pipefail

HOOK_DIR="$(cd "$(dirname "$0")" && pwd)"

# Consume stdin (hook input) but we don't need it
cat > /dev/null

# Find all modified .md files (staged + unstaged + untracked)
MD_FILES=$({ git diff --name-only HEAD -- '*.md' 2>/dev/null; \
             git diff --cached --name-only -- '*.md' 2>/dev/null; \
             git ls-files --others --exclude-standard -- '*.md' 2>/dev/null; \
           } | sort -u)

[[ -z "$MD_FILES" ]] && exit 0

# Format with rumdl
echo "$MD_FILES" | xargs -I{} bash -lc 'rumdl fmt "{}"' 2>/dev/null || true

# Lint with vale
VALE_OUTPUT=$(echo "$MD_FILES" | xargs bash -lc 'vale --output=JSON "$@"' _ 2>/dev/null \
  | python3 "$HOOK_DIR/vale_summary.py") || true

if [[ -n "$VALE_OUTPUT" && "$VALE_OUTPUT" != "No findings." ]]; then
  echo "$VALE_OUTPUT"
  echo ""
  echo ">>> ALL vale and rumdl warnings and errors MUST be resolved before continuing. <<<"
  echo ">>> You may edit rules in .vale/arci/ and add vocabulary to .vale/config/vocabularies/arci/. <<<"
  echo ">>> You MUST ask the user before modifying rumdl or vale project configs (.vale.ini, .rumdl.toml) or adding <!-- vale --> suppression comments. <<<"
fi

exit 0
