# Logging

ARCI produces three distinct categories of output: command output, diagnostic traces, and hook event logs. These serve different audiences, use different mechanisms, and follow different conventions. This document describes each category, its purpose, and how it's configured.

## Output categories

**Command output** is what a command exists to produce. `arci config show` outputs the configuration. `arci policies list` outputs the policy list. This goes to stdout (or stderr for errors) and is part of the command's user-facing contract.

**Diagnostic traces** describe what ARCI does internally: loading configuration, watching files, evaluating expressions, encountering errors. The audience is developers debugging ARCI itself. These traces help answer questions like "why isn't the config loading?" or "what's happening during policy evaluation?"

**Hook event logs** describe what ARCI decided: which policies matched, what actions it took, why it allowed or denied a request. The audience includes users who want to understand why something got blocked and tooling that consumes the event stream (like the dashboard). These logs help answer questions like "why did Claude Code get blocked from running that command?"

The separation is deliberate. Command output is for users. Diagnostic traces are for debugging. Hook event logs are for auditing and understanding policy decisions.

## Output verbosity vs diagnostic logging

These are distinct concepts that mature CLI tools keep separate.

**Output verbosity** (controlled by flags like `-v`, `--verbose`, `--quiet`) determines how much detail a command includes in its normal output. `arci policies list` might show just policy names by default, but with `-v` it shows descriptions and source files. This is part of the command's user experience.

**Diagnostic logging** (controlled by `ARCI_DEBUG`) produces internal traces for debugging the tool itself. Structured, timestamped, with module paths. This is for developers, not normal users.

This distinction follows the pattern established by tools like terraform (`TF_LOG`), git (`GIT_TRACE`), cargo (`CARGO_LOG`), and the entire Rust ecosystem (`RUST_LOG`). These tools use environment variables for diagnostic tracing and flags for output verbosity, keeping the concerns cleanly separated.

The rationale: logging configuration is an operational concern, not app configuration. Config files are for how the tool should behave functionally. Diagnostic verbosity is for troubleshooting the tool itself. Separating them avoids situations where someone commits a config change that turns on debug logging and breaks downstream tooling that parses command output.

## The `arci hook apply` output contract

The `arci hook apply` command is the hook entry point invoked by Claude Code. Its stdout and stderr are part of the protocol contract. Neither output verbosity flags nor diagnostic logging apply to this command; it has a fixed contract.

Claude Code's hook contract defines the meaning of exit codes and output streams:

Exit code 0 indicates success. The policy response JSON goes to stdout. Claude Code shows stdout to the user in verbose mode and parses it for structured control fields like `decision`, `reason`, and `continue`.

Exit code 2 indicates a blocking error. The error message goes to stderr. Claude Code feeds stderr back to the AI to process automatically. Claude Code ignores JSON on stdout for exit code 2.

Other exit codes indicate non-blocking errors. Claude Code shows stderr to the user in verbose mode. Execution continues.

This creates a hard constraint: `arci hook apply` never emits anything except the protocol-defined output. No verbosity flags, no diagnostic traces to stderr. Doing so would corrupt the JSON response (stdout) or confuse Claude Code with spurious messages (stderr).

Diagnostics about `arci hook apply` execution go to the hook event log (described below), which the dashboard and `arci hook logs` command can surface.

### Fail-open and error handling

ARCI follows fail-open semantics: configuration errors, rule failures, and internal problems never block Claude Code from operating. Only explicit deny decisions from successfully evaluated policies block operations.

When `arci hook apply` encounters an internal error (can't load config, can't parse input, can't open state store), it returns exit code 0 with a permissive response that allows the operation. ARCI writes diagnostics to the hook event log and state store, not to stderr.

The only time `arci hook apply` exits with code 2 is when a policy explicitly denies an action. That's a deliberate blocking decision, not an error condition.

## Output verbosity

Commands other than `arci hook apply` support output verbosity flags that control how much detail appears in their normal output.

### Flags

`-q` or `--quiet` suppresses non-essential output. The command produces only its primary result with no additional context.

`-v` or `--verbose` includes additional detail relevant to understanding the output. What "verbose" means is command-specific: `arci policies list -v` might show policy descriptions and source files; `arci config show -v` might show which files contributed to each setting.

These flags affect stdout content, not logging. They're part of the command's user experience.

### Machine-readable output

Many commands support `--json` for machine-readable output. When the user specifies `--json`, the command produces structured JSON to stdout suitable for parsing. Verbosity flags may still apply (controlling which fields the command includes), but the output format is always valid JSON.

In non-TTY contexts (piped output, CI), commands should behave consistently. The output format doesn't change based on TTY detection; only explicit flags like `--json` change the format.

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

### Implementation

ARCI initializes tracing early, before any other code runs. Because environment variables alone control tracing, there's no chicken-and-egg problem with config loading. The tracer is ready before ARCI reads any config, and config files have no influence over tracing behavior.

## Hook event logging

ARCI writes hook events as structured records of policy decisions in newline-delimited JSON (JSONL), using Hive-style partitioning for efficient analytical queries. This is not diagnostic logging; it's an audit trail of what ARCI decided and why.

### What gets logged

Each hook event record includes:

| Field | Description |
|-------|-------------|
| `timestamp` | ISO 8601 timestamp of the evaluation |
| `project_path` | Canonical absolute path to the project |
| `session_id` | Assistant-specific session identifier (null if unavailable) |
| `event_type` | Hook event type (such as `PreToolUse`, `PostToolUse`) |
| `tool_name` | Tool under invocation (null for non-tool events) |
| `tool_input_command` | Command for shell/bash tools (null otherwise) |
| `tool_input_file_path` | File path for file operations (null otherwise) |
| `tool_input_structured` | JSON object for complex tool inputs like MCP arguments |
| `matched_policies` | List of policy IDs that matched |
| `final_decision` | The aggregate decision returned to Claude Code |
| `deny_reason` | Explanation when decision is deny (null otherwise) |
| `evaluation_duration_ms` | How long evaluation took |
| `error` | Error details if evaluation failed (with fail-open result) |

The schema stays as flat as practical for efficient querying. Common fields like `tool_input_command` and `tool_input_file_path` sit at the top level rather than nested, while inherently variable data (arbitrary MCP tool arguments) goes in `tool_input_structured` as a JSON column.

Claude Code provides `session_id` in all events. When unavailable, the field is null.

### File organization

Hook event logs use Hive-style partitioning for efficient analytical queries with tools like DuckDB. The base log directory follows platform conventions:

| Platform | Base log directory |
|----------|-------------------|
| Linux | `~/.local/state/arci/log/` |
| macOS | `~/Library/Logs/arci/` |
| Windows | `%LOCALAPPDATA%\arci\log\` |

Within the base directory, ARCI partitions logs by project:

```text
<base log dir>/
  project=a1b2c3d4e5f6/
    2025-01-30.jsonl
    2025-01-31.jsonl
    daemon.log
```

The directory structure encodes `project` as a partition key. DuckDB (and similar tools) can extract these from the path without scanning file contents, enabling efficient predicate pushdown (examples use Linux paths; substitute the platform-appropriate base directory):

```sql
SELECT * FROM read_json(
  '~/.local/state/arci/log/**/*.jsonl',
  hive_partitioning=true
)
WHERE project = 'a1b2c3d4e5f6'
  AND timestamp >= '2025-01-01'
```

ARCI names files by date (YYYY-MM-DD.jsonl), providing natural rotation. A project with no activity on a given day simply has no file for that date. This bounds growth without requiring explicit rotation logic.

The project directory name is a truncated SHA256 hash of the project's canonical absolute path (first 12 hex characters). This handles path length limits and special characters while providing deterministic mapping. Each record includes the full `project_path` for human readability.

### Log path computation

ARCI can compute the log path for `arci hook apply` before reading stdin:

```text
project     → SHA of canonical cwd
date        → current date (YYYY-MM-DD)
```

This means ARCI needs no stdin buffering or parsing just to determine where to write the log.

### Configuration

Hook event logging uses the ARCI config file under `hooks.logging`:

```yaml
hooks:
  logging:
    enabled: true
    redact:
      - "**password**"
      - "**token**"
      - "**secret**"
```

The `enabled` field controls whether ARCI logs hook events at all. The `redact` list specifies patterns for field values that ARCI should scrub.

### DuckDB analytics queries

The Hive-partitioned structure makes analytical queries straightforward:

```sql
-- All denials in the last week
SELECT timestamp, tool_name, deny_reason
FROM read_json('~/.local/state/arci/log/**/*.jsonl', hive_partitioning=true)
WHERE final_decision = 'deny'
  AND timestamp >= current_date - interval '7 days'
ORDER BY timestamp DESC;

-- Policy hit frequency by project
SELECT project, matched_policies, count(*) as hits
FROM read_json('~/.local/state/arci/log/**/*.jsonl', hive_partitioning=true)
WHERE array_length(matched_policies) > 0
GROUP BY project, matched_policies
ORDER BY hits DESC;

-- Average evaluation time by project
SELECT project, avg(evaluation_duration_ms) as avg_ms
FROM read_json('~/.local/state/arci/log/**/*.jsonl', hive_partitioning=true)
GROUP BY project;
```

### Log cleanup

Date-based files make cleanup straightforward (example uses Linux path):

```bash
# Delete logs older than 30 days
find ~/.local/state/arci/log -name "*.jsonl" -mtime +30 -delete

# Or via arci (works on all platforms)
arci hook logs prune --older-than 30d
```

For archival, `arci hook logs compact` could convert older JSONL files to Parquet for better compression and query performance. This is a potential future enhancement.

## Daemon logging

Each project can have its own daemon instance. Daemons write operational logs to `daemon.log` in their project's log directory (example uses Linux base path):

```text
~/.local/state/arci/log/
  project=a1b2c3d4e5f6/
    2025-01-30.jsonl
    daemon.log
```

### What gets logged

The daemon log captures lifecycle and operational events:

| Event type | Description |
|------------|-------------|
| `startup` | Daemon started, bound to socket |
| `shutdown` | Daemon stopping (idle timeout, signal, etc.) |
| `config_reload` | Configuration reloaded (file watcher trigger or explicit) |
| `config_error` | Configuration failed to load (with error details) |

The daemon writes operational logs, not analytical data. The format is logfmt or plain text rather than JSONL:

```text
time=2025-01-15T10:30:00Z event=startup socket=/tmp/arci/a1b2c3d4/daemon.sock pid=12345
time=2025-01-15T10:35:00Z event=config_reload trigger=file_watcher files=policies.yaml
time=2025-01-15T11:30:00Z event=shutdown reason=idle_timeout uptime=3600s
```

Using `daemon.log` (not `.jsonl`) means analytical queries with `**/*.jsonl` cleanly get hook events only.

### Foreground compared to auto-spawned

When the daemon runs in the foreground (`arci daemon start`), operational output goes to stderr by default. The user can redirect as needed. The daemon determines the project directory automatically using the same logic as the CLI.

When `arci hook apply` auto-spawns the daemon, stderr isn't connected to anything useful. The daemon computes its log path from the project hash and writes there automatically. This requires no configuration; the daemon knows enough from its spawn arguments to determine the conventional location.

The daemon inherits diagnostic tracing (`ARCI_DEBUG=1`) from the environment. When auto-spawned, the daemon inherits the environment from `arci hook apply`, so if the user had debug enabled, the daemon emits debug traces to its log file.

## Dashboard integration

The dashboard reads hook event logs for display and filtering. JSONL files support efficient tailing for live updates. The Hive-partitioned structure allows the dashboard to scope queries by project without scanning irrelevant files.

For complex queries, the dashboard can use DuckDB's WASM build for in-browser analytics, or query via the daemon's API if the dashboard needs deeper aggregation.

## Security considerations

Hook event logs may contain sensitive information from evaluated inputs: file paths, command arguments, environment variable names. The `redact` configuration controls what the system scrubs.

Log files inherit the permissions of the state directory. On Unix systems, `~/.local/state/arci/` should be mode 0700, restricting access to the owning user.

Diagnostic traces (`ARCI_DEBUG`) may contain more sensitive information than hook event logs, including full config contents and expression evaluation details. Treat debug output as sensitive.

The `daemon.log` files contain operational information that's generally less sensitive, but socket paths and PIDs could be useful to an attacker. The same directory permissions apply.
