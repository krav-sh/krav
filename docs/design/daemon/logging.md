# Daemon logging

This document covers daemon-specific logging: the daemon.log file, its contents, and behavior in foreground vs auto-spawned modes.

## Daemon log location

Each project can have its own daemon instance. Daemons write operational logs to `daemon.log` in their project's log directory (example uses Linux base path):

```text
~/.local/state/arci/log/
  project=a1b2c3d4e5f6/
    2025-01-30.jsonl
    daemon.log
```

## What gets logged

The daemon log captures lifecycle and operational events:

| Event type | Description |
|------------|-------------|
| `startup` | Daemon started, bound to socket |
| `shutdown` | Daemon stopping (idle timeout, signal, etc.) |
| `config_reload` | Configuration reloaded (file watcher trigger or explicit) |
| `config_error` | Configuration failed to load (with error details) |

These logs track operations, using logfmt or plain text format rather than JSONL:

```text
time=2025-01-15T10:30:00Z event=startup socket=/tmp/arci/a1b2c3d4/daemon.sock pid=12345
time=2025-01-15T10:35:00Z event=config_reload trigger=file_watcher files=policies.yaml
time=2025-01-15T11:30:00Z event=shutdown reason=idle_timeout uptime=3600s
```

Using `daemon.log` (not `.jsonl`) means analytical queries with `**/*.jsonl` cleanly get hook events only.

## Foreground vs. auto-spawned

When the daemon runs in the foreground (`arci daemon start`), operational output goes to stderr by default. The user can redirect as needed. The daemon determines the project directory automatically using the same logic as the CLI.

When the daemon is auto-spawned by `arci hook apply`, stderr isn't connected to anything useful. The daemon computes its log path from the project hash and writes there automatically. This requires no configuration; the daemon knows enough from its spawn arguments to determine the conventional location.

Diagnostic tracing (`ARCI_DEBUG=1`) carries over from the environment. When auto-spawned, the daemon inherits the environment from `arci hook apply`, so if the user has debug enabled, the daemon emits debug traces to its log file.

## Security note

The `daemon.log` files contain operational information that's generally less sensitive than hook event logs, but socket paths and PIDs could be useful to an attacker. The same directory permissions apply: on Unix systems, `~/.local/state/arci/` should be mode 0700, restricting access to the owning user.

## See also

- [Hook event logging](../hooks/logging.md): the `arci hook apply` output contract and event log schema
- [CLI logging](../cli/logging.md): output verbosity flags and diagnostic tracing
- [Daemon errors](errors.md): troubleshooting, recovery, and metrics
