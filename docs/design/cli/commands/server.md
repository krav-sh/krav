# Server

Starts the ARCI server for the current project.

## Synopsis

    arci server [options]

## Description

The server is a long-running process that caches compiled configuration, owns the knowledge graph, and serves the dashboard and REST API. It runs in the foreground and stops with Ctrl+C (SIGINT) or SIGTERM.

Hook evaluation works without a running server (the CLI evaluates policies directly), but the server provides faster evaluation and knowledge graph mutations require it. See the [server design](../../server/index.md) for details.

The server binds to localhost on the configured base port (default 7680). If the port is in use, it tries incrementing ports up to 20 times. The server writes the actual port to `.arci/server.json` for discovery by other arci commands.

When running in a TTY, the server displays a live TUI showing recent evaluations, active policies, and configuration status. When not in a TTY, it outputs structured logs to stderr.

## Flags

`--port <port>`: base port to try binding (default: `7680`). If unavailable, the server increments until it finds a free port.

`--console <mode>`: console output mode. `rich` forces the TUI, `plain` forces structured log output, default auto-detects based on TTY.

`--log-level <level>`: logging verbosity. Values: `error`, `warn`, `info`, `debug`. Default: `info`. Also settable via `ARCI_LOG_LEVEL`.

`--project-dir <path>`: project root directory. Default: auto-detected by walking up from cwd. Also settable via `ARCI_PROJECT_DIR`.

## Examples

Start the server with defaults:

    arci server

Start on a specific port with debug logging:

    arci server --port 8080 --log-level debug

Start with plain log output (useful for piping or service managers):

    arci server --console plain

## See also

- [Server design](../../server/index.md)
- [Server discovery](../../server/discovery.md)
- [MCP command](mcp.md)
