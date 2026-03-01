# Hook event logging

This document covers the `arci hook apply` output contract, the hook event log schema, querying, and security considerations.

## The `arci hook apply` output contract

The `arci hook apply` command is the hook entry point invoked by Claude Code. Its stdout and stderr are part of the protocol contract. Neither output verbosity flags nor diagnostic logging apply to this command; it has a fixed contract.

Claude Code's hook contract defines the meaning of exit codes and output streams:

Exit code 0 indicates success. The policy response JSON goes to stdout. Claude Code shows stdout to the user in verbose mode and parses it for structured control fields like `decision`, `reason`, and `continue`.

Exit code 2 indicates a blocking error. The error message goes to stderr. Claude Code feeds stderr back to the AI to process automatically. Claude Code ignores JSON on stdout for exit code 2.

Other exit codes indicate non-blocking errors. Claude Code shows stderr to the user in verbose mode. Execution continues.

This creates a hard constraint: `arci hook apply` never emits anything except the protocol-defined output. No verbosity flags, no diagnostic traces to stderr. Doing so would corrupt the JSON response (stdout) or confuse Claude Code with spurious messages (stderr).

The hook event log (described below) captures diagnostics about `arci hook apply` execution, and the dashboard and `arci hook logs` command can surface them.

### Fail-open and error handling

ARCI follows fail-open semantics: configuration errors, rule failures, and internal problems never block Claude Code from operating. Only explicit deny decisions from successfully evaluated policies block operations.

When `arci hook apply` encounters an internal error (can't load config, can't parse input, can't open state store), it returns exit code 0 with a permissive response that allows the operation. ARCI writes diagnostics to the hook event log and state store, not to stderr.

The only time `arci hook apply` exits with code 2 is when a policy explicitly denies an action. That's a deliberate blocking decision, not an error condition.

## Hook event log schema

Hook events record policy decisions as structured data, always serialized as newline-delimited JSON (JSONL) using Hive-style partitioning for efficient analytical queries. This is not diagnostic logging; it's an audit trail of what ARCI decided and why.

### What gets logged

Each hook event record includes:

| Field | Description |
|-------|-------------|
| `timestamp` | ISO 8601 timestamp of the evaluation |
| `project_path` | Canonical absolute path to the project |
| `session_id` | Assistant-specific session identifier (null if unavailable) |
| `event_type` | Hook event type (`PreToolUse`, `PostToolUse`) |
| `tool_name` | Tool under evaluation (null for non-tool events) |
| `tool_input_command` | Command for shell/bash tools (null otherwise) |
| `tool_input_file_path` | File path for file operations (null otherwise) |
| `tool_input_structured` | JSON object for complex tool inputs like MCP arguments |
| `matched_policies` | List of policy IDs that matched |
| `final_decision` | The aggregate decision returned to Claude Code |
| `deny_reason` | Explanation when decision is deny (null otherwise) |
| `evaluation_duration_ms` | How long evaluation took |
| `error` | Error details if evaluation failed (with fail-open result) |

The schema stays as flat as practical for efficient querying. Common fields like `tool_input_command` and `tool_input_file_path` sit at the top level rather than nested, while inherently variable data (arbitrary MCP tool arguments) goes in `tool_input_structured` as a JSON column.

Claude Code provides `session_id` in all events. When the value is unavailable, the field defaults to null.

## File organization

Hook event logs use Hive-style partitioning for efficient analytical queries with tools like DuckDB. The base log directory follows platform conventions:

| Platform | Base log directory |
|----------|-------------------|
| Linux | `~/.local/state/arci/log/` |
| macOS | `~/Library/Logs/arci/` |
| Windows | `%LOCALAPPDATA%\ARCI\log\` |

Within the base directory, the system partitions logs by project:

```text
<base log dir>/
  project=a1b2c3d4e5f6/
    2025-01-30.jsonl
    2025-01-31.jsonl
    server.log
```

The directory structure encodes `project` as a partition key. DuckDB (and similar tools) can extract these from the path without scanning file contents, enabling efficient predicate pushdown (examples use Linux paths; substitute your platform's base directory):

```sql
SELECT * FROM read_json(
  '~/.local/state/arci/log/**/*.jsonl',
  hive_partitioning=true
)
WHERE project = 'a1b2c3d4e5f6'
  AND timestamp >= '2025-01-01'
```

Each file takes a date-based name (YYYY-MM-DD.jsonl), providing natural rotation. A project with no activity on a given day simply has no file for that date. This bounds growth without requiring explicit rotation logic.

The project directory name is a truncated SHA256 hash of the project's canonical absolute path (first 12 hex characters). Each record includes the full `project_path` for human readability.

### Log path computation

ARCI can compute the log path for `arci hook apply` before reading stdin:

```text
project     → SHA of canonical cwd
date        → current date (YYYY-MM-DD)
```

This means ARCI needs no stdin buffering or parsing just to determine where to write the log.

## Configuration

The ARCI config file configures hook event logging under `hooks.logging`:

```yaml
hooks:
  logging:
    enabled: true
    redact:
      - "**password**"
      - "**token**"
      - "**secret**"
```

The `enabled` field controls whether ARCI logs hook events at all. The `redact` list specifies patterns for field values that ARCI should redact.

## Querying with DuckDB

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

## Log cleanup

Date-based files make cleanup straightforward (example uses Linux path):

```bash
# Delete logs older than 30 days
find ~/.local/state/arci/log -name "*.jsonl" -mtime +30 -delete

# Or via arci (works on all platforms)
arci hook logs prune --older-than 30d
```

For archival, `arci hook logs compact` could convert older JSONL files to Parquet for better compression and query performance. This remains a potential future enhancement.

## Security considerations

Hook event logs may contain sensitive information from evaluated inputs: file paths, command arguments, environment variable names. The `redact` configuration controls what ARCI scrubs.

Log files inherit the permissions of the state directory. On Unix systems, `~/.local/state/arci/` should use mode 0700, restricting access to the owning user.

Diagnostic traces (`ARCI_DEBUG`) may contain even more sensitive information than hook event logs, including full config contents and expression evaluation details. Treat debug output as sensitive.

## See also

- [CLI logging](../cli/logging.md): output verbosity flags and diagnostic tracing
- [Server logging](../server/logging.md): server log location and event types
- [Hook error troubleshooting](errors.md): troubleshooting workflows and diagnostic commands
