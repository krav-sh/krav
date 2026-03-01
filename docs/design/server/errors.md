# Server error handling and recovery

This document covers server-specific troubleshooting, error recovery (automatic and manual), and observability metrics.

## Troubleshooting

### The server does not start

Check for port conflicts. The server tries the base port (default 7680) and increments up to 20 times. If all ports are in use, it reports an error. Use `lsof -i :7680` (or the configured port range) to find what's occupying them.

Review configuration with `arci config validate`. The server does not start with invalid configuration.

Run with verbose logging: `ARCI_LOG_LEVEL=debug arci server`

### Configuration changes are not taking effect

The server watches for file changes but debounces them. Changes take effect within a few seconds.

Force a reload via the API: `curl -X POST http://localhost:7680/config/reload`

In direct execution mode (no server), configuration loads fresh on every invocation. If changes are not reflected, check that you are editing the correct file. Use `arci config list` to see which files the system loads.

### Stale lockfile

If `.arci/server.json` exists but the server isn't running, any `arci` command detects the stale PID and cleans up the file automatically. No manual intervention is necessary. See [discovery](discovery.md) for details.

## Error recovery

ARCI provides automatic recovery for transient failures and tools for manual recovery of persistent issues.

### Automatic recovery

The server automatically recovers from these failure modes.

Configuration reload failures fall back to the previously cached configuration. The server logs the error but evaluation continues with known-good rules.

State store connection failures trigger automatic reconnection with exponential backoff. Operations that need state proceed without it (returning nil for state lookups).

File watcher failures trigger watcher restart. If the server cannot restart the watcher, it continues without hot reload; manual reload via the API remains available.

### Manual recovery

Some situations require manual intervention.

Corrupted state store: `arci state clear --all` wipes the state database. For complete reset, delete the DuckDB file at `.arci/state.duckdb`.

Unresponsive server: press Ctrl+C in the terminal where the server runs for a graceful shutdown. If the server is truly hung, `kill <pid>` (reading the PID from `.arci/server.json`) forces termination. The next `arci` command cleans up the stale lockfile.

Extension conflicts: `arci extension sync` reinstalls extensions from the lockfile. For persistent issues, `arci extension remove <name>` and re-add.

## Metrics and observability

The server exposes metrics for monitoring and diagnostics.

### Error counters

Counters track error frequency by category: configuration load failures, rule evaluation failures, action execution failures, and state store failures. Each counter carries labels for error type and context (rule ID, action type, etc.).

### Error rates

Gauges track error rates over sliding windows: evaluation failures per evaluation, and action failures per action execution. These help distinguish between occasional transient errors and systemic problems.

### Prometheus integration

The server exposes a `/metrics` endpoint in Prometheus format:

```text
$ curl http://localhost:7680/metrics

# HELP arci_evaluations_total Total number of hook evaluations
# TYPE arci_evaluations_total counter
arci_evaluations_total 1523

# HELP arci_evaluation_errors_total Total evaluation errors by type
# TYPE arci_evaluation_errors_total counter
arci_evaluation_errors_total{error_type="expression_error"} 3
arci_evaluation_errors_total{error_type="state_error"} 1
```

## See also

- [Server logging](logging.md): log location and event types
- [CLI errors](../cli/errors.md): CLI error presentation and health checks
- [Hook errors](../hooks/errors.md): hook troubleshooting workflows and diagnostic commands
