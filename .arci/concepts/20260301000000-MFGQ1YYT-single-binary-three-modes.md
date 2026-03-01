# Single binary, three modes

arci ships as a single Go binary that serves all roles: command-line tool, daemon, and dashboard. No separately versioned components, no coordination problems between processes at different versions, and no external runtime dependencies for core operation.

## The three modes

**Command-line tool** is the default mode. The developer or agent invokes `arci` directly for graph management, hook evaluation, diagnostics, and administration. In direct execution mode, the command-line tool loads configuration, compiles policies, and evaluates locally. This is self-contained but pays startup costs on every invocation.

**Daemon** is a long-running process started with `arci daemon start`. The daemon maintains a config cache, compiled policies, connection pooling, and a REST API. When a daemon is available, the command-line tool delegates to it instead of evaluating locally, avoiding per-invocation startup costs. The daemon also provides WebSocket event streaming and hot-reload on config changes.

**Dashboard** is a web-based diagnostics interface started with `arci dashboard` or served by the daemon. It provides live event streaming, policy testing, state browsing, coverage reports, and graph browsing. It reads from the daemon's cached state for consistency.

## Why a single binary

**No version skew** because the command-line tool, daemon, and dashboard are the same binary, always at the same version. No "the command-line tool is at v0.3 but the daemon is at v0.2" problem.

**Simple distribution** means one binary to install, one binary to update. No dependency management beyond the binary itself. Platform-native packaging (Homebrew, go install, GitHub releases) delivers everything.

**Shared code** means the domain logic (graph engine, policy engine, config loader) follows the same code path regardless of mode. The daemon wraps it with caching and an HTTP server. The dashboard wraps it with a web UI. But the core evaluation and graph operations are identical.

## Mode detection

The binary determines its mode from the subcommand: `arci daemon start` runs the daemon, `arci dashboard` serves the dashboard, everything else runs as the command-line tool. Within command-line mode, the presence of a running daemon triggers delegation: the tool checks for a daemon, and if available, forwards requests to the daemon's API rather than loading config and evaluating locally.
