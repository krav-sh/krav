# Server

The Krav server is a long-running process that owns the knowledge graph, caches compiled configuration, and serves the dashboard and REST API for a single project. It runs in the foreground via `krav server` and stops with Ctrl+C.

The server serves two distinct roles. For hook evaluation, it is a performance optimization: the CLI can evaluate policies directly without a running server, but the server provides sub-10 ms evaluation by keeping configuration pre-compiled and the DuckDB instance warm. For knowledge graph operations, the server is essential: it owns the in-memory DuckDB instance containing the hydrated graph, serializes mutations through a single process, and enforces invariants that direct CLI access cannot safely coordinate.

## Documentation

- [Discovery](discovery.md): how the CLI finds a running server
- [Error handling and recovery](errors.md): troubleshooting, automatic/manual recovery, metrics and observability
- [Logging](logging.md): log location, event types, and output modes

## Responsibilities

The server has six primary responsibilities.

Configuration management comes first. The server loads the merged configuration cascade for the project, compiles policy expressions, and watches configuration files for changes via `fsnotify`. When files change, the server reloads configuration atomically: in-flight requests complete with the old configuration, new requests get the new configuration. If reloading fails due to syntax or validation errors, the server logs the error and continues with the last known-good configuration.

The server owns the hook evaluation engine. It compiles policy expressions once and reuses them for each evaluation, avoiding the per-invocation cost of parsing. This drops evaluation overhead from 50 to 200 ms (direct mode) to single-digit milliseconds.

Knowledge graph ownership is the server's most critical role. The knowledge graph is a shared mutable resource. When multiple Claude Code subagents run tasks concurrently, or a human and an agent both issue graph mutations, the server serializes writes through its single DuckDB instance. Without the server, concurrent graph mutations would require external coordination. Read-only graph queries can work in direct mode by reading the NDJSON files (DuckDB's `read_json` function handles NDJSON natively), but any mutation requires the server. On startup, the server hydrates the graph from `.krav/graph/*.ndjson` into DuckDB tables and creates the SQL/PGQ property graph. On checkpoint or graceful shutdown, it dehydrates modified graph state back to NDJSON.

State store management handles the file-backed DuckDB database at `.krav/state.duckdb`, attached to the same DuckDB instance as the in-memory graph via the `ATTACH` mechanism, so queries can join across both domains without a separate connection pool.

Metrics accumulation tracks policy match counts, action executions, errors, and timing information in memory. The API exposes these metrics, and the dashboard displays them. Metrics reset on server restart; the project may add persistent metrics later but in-memory suffices for diagnostics.

The server serves the REST and WebSocket APIs that the CLI, dashboard, and MCP server use. It also serves the dashboard's static assets and rendered templates.

## Architecture

The server builds on Go's `net/http` with the `chi` router and `nhooyr.io/websocket`. Chi provides composable, lightweight HTTP routing. Go's goroutines and channels handle concurrent connections efficiently without a separate async runtime.

File watching uses the `fsnotify` package, which provides efficient cross-platform file system monitoring with support for Linux inotify, macOS FSEvents, and Windows ReadDirectoryChanges.

The dashboard uses Go's `html/template` with Sprig for server-side rendering and htmx for interactive updates without a JavaScript build system.

```mermaid
flowchart TB
    subgraph server["krav server"]
        subgraph http_server["HTTP server"]
            subgraph routes["Routes"]
                apply_ep["POST /apply"]
                health_ep["GET /health"]
                config_ep["GET /config/status"]
                reload_ep["POST /config/reload"]
                policies_ep["GET /policies"]
                graph_ep["POST /graph/*"]
                state_ep["GET /state"]
                metrics_ep["GET /metrics"]
                events_ep["WS /events"]
                dashboard_ep["GET /dashboard/*"]
            end
        end

        subgraph services["Services"]
            subgraph config_mgr["Config manager"]
                cache["Configuration cache"]
                watcher["File watcher\n(fsnotify)"]
            end

            subgraph duckdb_engine["DuckDB engine"]
                graph[("Knowledge graph\n(in-memory DuckDB + DuckPGQ)")]
                state[("State store\n(file-backed DuckDB)")]
                write_serializer["Write serializer"]
            end

            subgraph metrics_acc["Metrics accumulator"]
                policy_counts["Policy match counts"]
                action_stats["Action stats"]
                timing["Timing histograms"]
            end
        end
    end

    routes --> services
    watcher -.->|"hot reload"| cache
```

## Console modes

The server adapts its output to the terminal environment. When stderr connects to a TTY, the server renders a status TUI showing recent hook evaluations with their decisions, active policies and match counts, configuration reload events, and a rolling latency view. When stderr lacks a TTY (piped, redirected, or launched by a service manager), the server streams structured log lines instead.

The `--console` flag overrides auto-detection. `--console rich` forces the TUI. `--console plain` forces log output. Auto-detection is the default.

Bubble Tea (charmbracelet/bubbletea) powers the TUI, using an Elm-style architecture that fits naturally with the server's internal event stream. The TUI subscribes to the same event data that the WebSocket `/events` endpoint exposes to the dashboard, so the three visibility surfaces (TUI, dashboard, MCP tools) all consume the same underlying data through different renderers.

## Operating modes

Krav operates in two modes: direct execution and server-delegated execution. Both use the same evaluation engine; the difference is where configuration loading and state management happen.

In direct execution mode, `krav hook apply` loads configuration, compiles policy expressions, and evaluates policies on every invocation. This requires no setup beyond installing Krav and writing policies.

In server-delegated mode, `krav hook apply` sends requests to the running server, which maintains cached configuration and pre-compiled expressions. The server also handles graph mutations, state management, and metrics.

The CLI determines which mode to use by checking for a `.krav/server.json` lockfile in the project directory (see [discovery](discovery.md)). If the lockfile exists and the server process is alive, the CLI delegates. If not, the behavior depends on the operation: hook evaluation falls back to direct execution silently, while graph-mutating commands produce an error telling the user to start the server.

## API design

The server exposes an HTTP API on localhost. All endpoints use JSON for request and response bodies.

### `POST /apply`

The primary endpoint for hook evaluation. The CLI calls this for every hook invocation when it detects a running server.

The request body contains the hook input payload. The response body contains the JSON output to write to stdout (or null if no output) and the exit code the CLI should use.

This endpoint must be fast. It uses cached configuration, pre-compiled expressions, and pooled database connections. The handler deserializes the request, runs the evaluation engine, and returns the result.

### `POST /graph/*`

Graph mutation endpoints for knowledge graph operations. These serialize writes through the server's graph manager, ensuring transactional consistency and invariant enforcement. The spec system design docs define the graph API surface.

### `GET /health`

Health check endpoint. Returns HTTP 200 if the server is healthy, with a JSON body containing uptime, project root, port, and version.

### `GET /config/status`

Returns the status of the project's configuration: whether it is valid, when it was last reloaded, and any validation errors.

### `POST /config/reload`

Forces a configuration reload. File watching normally handles this automatically, but the API provides an escape hatch for edge cases.

### `GET /policies`

Returns the list of active policies for the project. Returns policy metadata including name, source file, priority, enabled status, event types, and rule count.

### `GET /state`

Returns state store entries. Accepts a `session` query parameter for filtering. Returns entries with their values and metadata.

### `GET /metrics`

Returns a snapshot of accumulated metrics. Includes policy match counts, action execution counts, timing percentiles, and error counts.

### `WS /events`

WebSocket endpoint for live event streaming. Clients connect and receive real-time events including hook invocations, policy matches, action executions, errors, and configuration reloads. Events are JSON objects with a `type` field and event-specific data. The dashboard and TUI both consume this stream.

### `GET /dashboard/*`

Serves the dashboard web interface. Returns HTML pages rendered with Go templates. Uses htmx for interactive updates via the WebSocket event stream.

## Lifecycle

The server runs in the foreground via `krav server`. It determines the project root using the same walk-up-the-tree logic as all other `krav` commands (looking for `.krav/` or other project markers), respecting `--project-dir` and `KRAV_PROJECT_DIR` overrides.

On startup, the server selects a port by trying the configured base port (default 7680) and incrementing until it finds a free port, up to a small scan limit. It then creates the in-memory DuckDB instance and loads the DuckPGQ extension, hydrates the knowledge graph from `.krav/graph/*.ndjson` into DuckDB vertex and edge tables, creates the SQL/PGQ property graph definition over those tables, attaches the file-backed state database at `.krav/state.duckdb`, writes `.krav/server.json` to the project directory (see [discovery](discovery.md)), initializes the HTTP server and registers routes, starts the file watcher for configuration directories, and begins accepting requests.

On shutdown (SIGTERM, SIGINT, or Ctrl+C), the server stops accepting new connections, waits for in-flight requests to complete (with a timeout), dehydrates the graph state back to sorted NDJSON files under `.krav/graph/`, stops the file watcher, closes the DuckDB instance, removes `.krav/server.json`, and exits. Dehydration also occurs on explicit save commands and baseline creation, not on every write.

The lockfile is the single artifact of a running server. If the server crashes without cleaning up, the discovery mechanism detects and handles the stale lockfile.

## Project scoping

Each server instance owns exactly one project. The server knows its project root because that directory is where it started (or the directory specified by `--project-dir`). Configuration loading, graph operations, and state management are all scoped to that project.

Running multiple Krav servers for different projects is straightforward: each binds its own port and writes its own `.krav/server.json`. The auto-detection scan handles port conflicts between projects; the second server simply takes the next available port.

## Process management

Krav takes a foreground-first approach: the server never forks, backgrounds, or detaches from the terminal. Users choose how to manage its lifecycle using external tools that are purpose-built for process supervision.

Running the server in a terminal, tmux session, or screen works well for development. The TUI makes this pane actively useful rather than a stream of logs to ignore.

For persistent installations, system service managers are the recommended approach. On Linux, systemd user services (`systemctl --user`) provide automatic restart, resource limits, and logging integration without requiring root. On macOS, launchd launch agents offer similar capabilities and run at user login.

Process supervisors like supervisord provide cross-platform management without root access. Container-based deployment works by running `krav server` as the container entrypoint.

## Error isolation

The server isolates errors to prevent cascading failures. Evaluation errors for one request don't affect other requests. The server logs individual rule or action failures but does not fail the evaluation as a whole.

The API returns appropriate HTTP status codes: 200 for success, 400 for bad requests, 500 for internal errors. The `/apply` endpoint always returns 200 with an appropriate `exit_code` in the body, even for errors, maintaining fail-open semantics for hook evaluation.
