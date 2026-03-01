# Daemon

The daemon command group manages the optional ARCI daemon process. The daemon is a performance optimization that keeps configuration cached in memory and provides the diagnostics API.

## Synopsis

    arci daemon <subcommand> [options]

## Description

The daemon is an optional long-running process that caches compiled configuration in memory, allowing the `arci hook apply` command to delegate evaluation for faster response times. It also exposes a diagnostics API used by the dashboard and other tooling. Running the daemon is not required; ARCI works in direct mode without it.

## Subcommands

### Start

Starts the daemon in the foreground. During development, running the daemon in the foreground in a terminal is the typical workflow.

**Flags:**

- `--port <port>`: HTTP port to listen on (default: `7680`).
- `--socket <path>`: Unix socket path.
- `--log-level <level>`: Logging verbosity level.

### Stop

Stops a running daemon by sending it a shutdown signal.

### Status

Shows whether the daemon is running, what port and socket it's listening on, its uptime, and summary statistics.

### Reload

Forces the daemon to reload all configuration. File watching normally handles this automatically, but you can trigger a manual reload if needed.

### Restart

Stops and then starts the daemon. Useful for applying configuration changes that require a full restart.

## See also

- [Daemon Design](../../daemon/index.md)
- [Daemon Auto-Spawn](../../daemon/auto-spawn.md)
