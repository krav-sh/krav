# CLI error presentation

This document covers how the CLI presents errors to users, including formatted error output and health checks. For exit code semantics, see [exit codes](exit-codes.md).

## Error formatting

The CLI uses the console for human-readable error output. The CLI formats errors with context and suggestions when possible:

```text
error: configuration validation failed

  --> .krav/rules.yaml:15:3
   |
15 |   condition: tool.name =~ /rm/ && args contains "-rf"
   |              ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
   |
   = error: unknown operator 'contains', did you mean 'has'?
   = help: see https://krav.sh/docs/expressions for expression syntax
```

For validation commands, the CLI collects errors and reports them together rather than failing on the first error:

```text
error: found 3 configuration errors

.krav/rules.yaml:
  line 15: unknown operator 'contains' in condition
  line 28: duplicate rule ID 'block-rm'

~/.config/krav/config.yaml:
  line 5: unknown event type 'BeforeToolUse', did you mean 'PreToolUse'?
```

## Health checks with `krav doctor`

The doctor command performs full health checks:

```text
$ krav doctor

Installation      OK    krav 0.1.0 at /usr/local/bin/krav
Claude Code       OK    hooks configured in ~/.claude/settings.json
Configuration     OK    12 rules loaded from 3 sources
Rule validation   WARN  1 rule has warnings (use --verbose for details)
State store       OK    state.duckdb accessible, 45 entries
Extensions        OK    2 extensions loaded
Logs              OK    log directory writable

Overall: PASS with warnings (1)
```

Use `--verbose` to see details about warnings and `--fix` to attempt automatic repairs.

## See also

- [Exit codes](exit-codes.md): exit code conventions and semantics
- [Hook errors](../hooks/errors.md): hook troubleshooting workflows and diagnostic commands
- [Server errors](../server/errors.md): server troubleshooting and recovery
