# CLI logging

This document covers output verbosity flags and diagnostic tracing for the CLI.

## Output verbosity

Commands other than `arci hook apply` support output verbosity flags that control how much detail appears in their normal output.

### Flags

`-q` or `--quiet` suppresses non-essential output. The command produces only its primary result with no additional context.

`-v` or `--verbose` includes additional detail relevant to understanding the output. What "verbose" means is command-specific: `arci policies list -v` might show policy descriptions and source files; `arci config show -v` might show which files contributed to each setting.

These flags affect stdout content, not logging. They're part of the command's user experience.

### Machine-readable output

Many commands support `--json` for machine-readable output. When the user specifies `--json`, the command produces structured JSON to stdout suitable for parsing. Verbosity flags may still apply (controlling which fields the output includes), but the output format is always valid JSON.

In non-TTY contexts (piped output, CI), commands should behave consistently. The output format doesn't change based on TTY detection, only explicit flags like `--json` change the format.

## Diagnostic tracing

Diagnostic traces are for debugging ARCI internals. They're controlled entirely through environment variables, never through config files or command-line flags on normal commands.

### Environment variable

`ARCI_DEBUG` enables diagnostic tracing:

| Value | Effect |
|-------|--------|
| `1` or `true` | Enable debug tracing to stderr |
| (unset or empty) | Tracing off |

When enabled, ARCI emits structured log output covering config file discovery, loading, parsing, and merging; policy compilation and validation; expression evaluation; Claude Code protocol handling; state store operations; daemon lifecycle (startup, shutdown, config reload); and file watcher events.

### Trace output

Diagnostic traces use a human-readable format when stderr is a TTY and logfmt when piped or redirected:

```text
time=2024-01-15T10:30:00Z level=debug msg="loading config" path=/Users/tony/.config/arci/config.yaml
time=2024-01-15T10:30:00Z level=debug msg="merged user config" policies=3 bindings=2
time=2024-01-15T10:30:01Z level=debug msg="config loaded" total_policies=5 duration_ms=42
```

### Early initialization

ARCI initializes tracing early, before any other code runs. Because environment variables alone control tracing, there's no chicken-and-egg problem with config loading. The tracer starts before ARCI reads any config, and config files have no influence over tracing behavior.

## See also

- [Hook event logging](../hooks/logging.md): the `arci hook apply` output contract and event log schema
- [Server logging](../server/logging.md): server log location and event types
