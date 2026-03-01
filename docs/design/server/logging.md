# Server logging

This document covers server-specific logging: output destinations, event types, and behavior across console modes.

## Output destination

The server writes logs to stderr. In TUI mode (the default when stderr is a TTY), log events feed the status display rather than appearing as text lines. In plain mode (`--console plain`, or when stderr is not a TTY), structured log lines stream to stderr directly. Users can redirect stderr to a file or logging service as needed.

The server uses Go's `log/slog` package for structured logging. Log entries include timestamps, log levels, and structured key-value attributes.

## What gets logged

The server logs lifecycle and operational events:

| Event type | Description |
|------------|-------------|
| `startup` | Server started, bound to port |
| `shutdown` | Server stopping (signal received, completing in-flight requests) |
| `config_reload` | Configuration reloaded (file watcher trigger or API call) |
| `config_error` | Configuration failed to load (with error details) |
| `graph_mutation` | Knowledge graph write operation completed |
| `graph_error` | Knowledge graph operation failed |

In plain console mode, log output uses logfmt format:

```text
time=2026-03-01T10:30:00Z level=INFO event=startup port=7680 project=/home/user/myproject pid=48210
time=2026-03-01T10:35:00Z level=INFO event=config_reload trigger=file_watcher files=policies.yaml
time=2026-03-01T11:02:15Z level=WARN event=config_error file=.arci/policies/broken.yaml error="invalid expression on line 12"
```

## Log levels

Log levels are configurable via `ARCI_LOG_LEVEL` or the `--log-level` flag on `arci server`.

Debug logs detailed information about every evaluation, including expression matching and action execution. Info logs key events like startup, shutdown, and configuration reloads. Warn logs recoverable errors like failed config reloads or state store hiccups. Error logs failures that require attention.

## TUI vs. plain mode

In TUI mode, the server renders a live status display rather than streaming log lines. The TUI shows the same information (recent evaluations, config status, errors) in a compact, continuously updated format. The underlying event data is identical; only the rendering differs.

The `--console` flag controls which mode the server uses: `rich` forces the TUI, `plain` forces log output, and the default auto-detects based on whether stderr is a TTY. This means `arci server` in a terminal shows the TUI, while `arci server` launched by systemd, launchd, or piped through a log collector gets plain structured output automatically.

## Security note

Log output may contain file paths, policy names, and other project metadata. When redirecting logs to shared storage or a logging service, consider what information the logs expose. The same general principle applies as with all ARCI data: it describes your development workflow and deserves appropriate care.

## See also

- [Hook event logging](../hooks/logging.md): the `arci hook apply` output contract and event log schema
- [CLI logging](../cli/logging.md): output verbosity flags and diagnostic tracing
- [Server errors](errors.md): troubleshooting, recovery, and metrics
