# Doctor

The doctor command performs a full health check of your Krav installation. It verifies that components are correctly installed and configured, surfacing any issues that might cause problems.

## Synopsis

```text
Krav doctor [options]
```

## Description

The command runs a series of checks covering the full Krav stack:

- **installation**: Krav is properly installed and on PATH
- **claude-code**: hooks are correctly wired up with Claude Code
- **config**: all config files parse without errors and pass schema validation
- **policies**: all policy and rule expressions compile, and the system recognizes all action types
- **state**: DuckDB state database is accessible and not corrupted
- **extensions**: installed extensions load without errors
- **logs**: can write to the project-level log location

Each check reports pass, warning, or fail status with details about any issues found.

## Options

- `--fix`: Attempt to automatically repair common issues where possible, such as recreating a corrupted state store or re-running `krav install` for misconfigured Claude Code integration. Prompts for confirmation before making changes.
- `--yes`: Skip confirmation prompts when used with `--fix`.
- `--check <name>`: Run only specific checks. Repeat this flag to select more than one check. Valid names: `installation`, `claude-code`, `config`, `policies`, `state`, `extensions`, `logs`.

## Exit codes

| Code | Meaning |
|------|---------|
| 0    | All checks pass |
| 1    | Warnings but no failures |
| 2    | One or more checks failed |

## Examples

```bash
# Run all health checks
Krav doctor

# Run only config and policy checks
Krav doctor --check config --check policies

# Auto-repair issues without prompting
Krav doctor --fix --yes
```

## See also

- [Installation](../../installation.md)
