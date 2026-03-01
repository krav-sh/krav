# Dashboard

The dashboard command starts the diagnostics dashboard server in the foreground. The dashboard provides a web interface for policy testing, state inspection, and event monitoring.

## Synopsis

```text
arci dashboard [options]
```

## Description

The dashboard connects to the daemon as a client, so the daemon must be running for the dashboard to show live data.

## Options

- `--port <port>`: HTTP port for the dashboard server (default: `7681`).
- `--open`: automatically open a browser to the dashboard URL.

## Examples

```bash
# Start dashboard on default port
arci dashboard

# Start on a custom port and open browser
arci dashboard --port 8080 --open
```

## See also

- [Daemon](daemon.md)
