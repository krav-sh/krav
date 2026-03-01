#!/usr/bin/env bash
# PostToolUse hook: format and lint markdown files after editing
set -euo pipefail

HOOK_DIR="$(cd "$(dirname "$0")" && pwd)"

INPUT=$(cat)
FILE=$(echo "$INPUT" | jq -r '.tool_input.file_path // .tool_input.notebook_path // empty')

# Only process markdown files
[[ -z "$FILE" || "$FILE" != *.md ]] && exit 0

# Format with rumdl (login shell for PATH)
bash -lc "rumdl fmt \"$FILE\"" 2>/dev/null || true

# Lint with vale, pipe JSON through summary script
VALE_OUTPUT=$(bash -lc "vale --output=JSON \"$FILE\"" 2>/dev/null \
  | python3 "$HOOK_DIR/vale_summary.py") || true

if [[ -n "$VALE_OUTPUT" && "$VALE_OUTPUT" != "No findings." ]]; then
  echo "$VALE_OUTPUT"
  echo ""
  echo ">>> ALL vale and rumdl warnings and errors MUST be resolved before continuing. <<<"
  echo ">>> You may edit rules in .vale/arci/ and add vocabulary to .vale/config/vocabularies/arci/. <<<"
  echo ">>> You MUST ask the user before modifying rumdl or vale project configs (.vale.ini, .rumdl.toml) or adding <!-- vale --> suppression comments. <<<"
fi

exit 0
