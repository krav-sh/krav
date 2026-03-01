# Stage 0 skills replaced by Stage 1 CLI commands

The project shall replace every Stage 0 shell-based Claude Code skill with a Stage 1 CLI command that provides equivalent or superior capability before Stage 1 completion.

## Rationale

Shell-based skills are fragile approximations of graph operations. If Stage 1 ships without replacing them, the developer continues depending on brittle workarounds despite having a real graph engine, and the CLI commands lack validation against actual workflow needs.

## Verification criteria

For each shell-based skill in .claude/skills/ at Stage 0 completion, demonstrate a Stage 1 CLI command that produces equivalent or superior output for the same operation.
