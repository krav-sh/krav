# Daemon error handling and recovery

This document covers daemon-specific troubleshooting, both automatic and manual error recovery, and metrics and observability.

## Troubleshooting

### The daemon does not start

Check for port conflicts with `lsof -i :7680` (the default port). Another process may be using the port.

Review configuration with `arci config validate`. The daemon does not start with invalid configuration.

Check file permissions on the socket path and state directory.

Run the daemon in foreground with verbose logging: `ARCI_LOG_LEVEL=debug arci daemon start`

### Configuration changes are not taking effect

The daemon watches for file changes but has debouncing. Changes take effect within a few seconds.

Force a reload with `arci daemon reload`.

In direct execution mode (no daemon), configuration loads fresh on every invocation. If changes are not reflected, check that you are editing the correct file. Use `arci config list` to see which files the system loads.

Clear any cached state that might affect behavior: `arci state clear --session <id>`

## Error recovery

ARCI provides automatic recovery for transient failures and tools for manual recovery of persistent issues.

### Automatic recovery

The daemon automatically recovers from these failure modes.

Configuration reload failures fall back to the previously cached configuration. The daemon logs the error but evaluation continues with known-good rules.

State store connection failures trigger automatic reconnection with exponential backoff. Operations that need state proceed without it (returning nil for state lookups).

File watcher failures trigger watcher restart. If the daemon cannot restart the watcher, it continues without hot reload; manual reload remains available.

### Manual recovery

Some situations require manual intervention.

Corrupted state store: `arci state clear --all` wipes the state database. For complete reset, delete the SQLite file at `$XDG_STATE_HOME/arci/state.db`.

Stale daemon: `arci daemon stop && arci daemon start` performs a clean restart. The `--force` flag kills a daemon that is not responding to graceful shutdown.

Extension conflicts: `arci extension sync` reinstalls extensions from the lockfile. For persistent issues, `arci extension remove <name>` and re-add.

## Metrics and observability

The daemon exposes metrics for monitoring and alerting.

### Error counters

Counters track error frequency by category:

- `arci_config_errors_total` - configuration load failures
- `arci_evaluation_errors_total` - rule evaluation failures
- `arci_action_errors_total` - action execution failures
- `arci_state_errors_total` - state store failures

Each counter has labels for error type and context (rule ID, action type, etc.).

### Error rates

Gauges track error rates over sliding windows:

- `arci_evaluation_error_rate` - evaluation failures per evaluation
- `arci_action_failure_rate` - action failures per action execution

These help distinguish between occasional transient errors and systemic problems.

### Prometheus integration

The daemon exposes a `/metrics` endpoint in Prometheus format:

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

- [Daemon logging](logging.md): daemon.log location and event types
- [CLI errors](../cli/errors.md): CLI error presentation and health checks
- [Hook errors](../hooks/errors.md): hook troubleshooting workflows and diagnostic commands
