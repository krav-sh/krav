# Error handling and diagnostics

This document describes how ARCI handles errors, what gets logged, and how users diagnose problems. Fail-open semantics mean errors should not block Claude Code, but users still need visibility into what went wrong.

## Error philosophy

ARCI treats errors as information rather than failures. When something goes wrong, the system should continue operating (fail-open), log useful diagnostic information, surface problems through appropriate channels (logs, dashboard, CLI output), and never leave users wondering "why didn't the rule fire?"

The challenge is balancing transparency with noise. Users need to know when rules are not working, but excessive warnings for benign conditions hurt usability.

### Fail-open semantics

This principle is non-negotiable: configuration errors, rule evaluation failures, action timeouts, and daemon unavailability never block Claude Code from operating. Only explicit deny decisions from successfully evaluated rules block operations. If ARCI encounters an internal error, it logs the problem and returns a permissive response.

The rationale is simple. ARCI is a guardrail, not a gate. A broken guardrail should not lock users out of their tools. Users trust Claude Code to help them work; ARCI adds safety checks but must not become a single point of failure that prevents work entirely.

Every error path must have a fallback. Configuration fails to load? Use empty rules (allow everything). Expression evaluation throws? Skip that rule (other rules still evaluate). Action handler times out? Log the timeout and continue. The daemon is unreachable? Fall back to direct evaluation.

## Error types and Go conventions

ARCI uses Go's built-in error handling conventions throughout the codebase. Simple errors use sentinel values defined with `errors.New()`, while errors that carry structured context use custom types that satisfy the `error` interface. Error wrapping with `fmt.Errorf()` and `%w` preserves the full error chain for inspection with `errors.Is()` and `errors.As()`.

### Defining error types

Each package defines its own error types. Sentinel errors cover simple cases, while custom struct types carry additional context fields.

The `Unwrap()` method enables error chain traversal, preserving the full error context for debugging while allowing each layer to add its own descriptive message.

### Error propagation with if err != nil

Go's explicit error propagation replaces implicit exception-based flow. The code checks each fallible call immediately and wraps errors with additional context using `fmt.Errorf()`:

```go
func loadConfig(path string) (*Configuration, error) {
 contents, err := os.ReadFile(path)
 if err != nil {
  return nil, &ReadError{Path: path, Err: err}
 }

 var parsed RawConfig
 if err := yaml.Unmarshal(contents, &parsed); err != nil {
  return nil, &ParseError{
   Path: path,
   Line: extractYAMLLine(err),
   Err:  err,
  }
 }

 if err := validateSchema(&parsed, path); err != nil {
  return nil, err
 }

 return materializeConfig(parsed)
}
```

### Zero values and pointer types for optional fields

Go uses zero values and pointer types to represent values that may or may not be present. A nil pointer indicates absence, while a zero-value string or int may be a valid value depending on context:

```go
type Rule struct {
 ID          *string  // Rules may be anonymous; nil means no ID
 Description *string  // Optional description
 Condition   string   // Required
 Priority    Priority
 // ...
}

// Checking optional fields is explicit
if rule.ID != nil {
 slog.Info("rule matched", "rule_id", *rule.ID)
} else {
 slog.Info("anonymous rule matched")
}
```

## Error categories

Errors fall into distinct categories based on where they occur and how the system should handle them.

### Configuration errors

Configuration errors occur during discovery, loading, parsing, and validation of configuration files. These errors are recoverable at the system level (ARCI continues with degraded configuration) but should surface prominently to users.

### Expression errors

Expression errors occur during condition parsing or evaluation. The compiler detects malformed expressions at compile time (when rules load); evaluation errors happen at runtime.

Expression errors in conditions cause the evaluator to skip the rule (fail-open) and log a warning.

### Action errors

Action errors occur during action execution. Shell commands may fail, scripts may error, and handlers may timeout.

The system logs action errors but does not block the hook response. If an action fails, the rule's result (allow/block) still applies.

### State store errors

State store errors involve SQLite operations. These are typically transient and recoverable.

State errors cause operations to proceed without state context. A rule that checks `session_get("counter")` receives nil if the state store is unavailable.

### Daemon errors

Daemon errors involve the optional daemon process.

## Logging architecture

ARCI uses the `log/slog` package from the Go standard library for structured logging. slog provides structured key-value fields for machine-parseable context, leveled logging, and configurable handlers for flexible output.

### Log levels

Log levels follow standard severity conventions.

Error level is for failures that require attention: configuration errors that degrade features, state store corruption, daemon crashes. These are problems a user should investigate.

Warning level is for recoverable problems and skipped rules: expression evaluation failures that cause the evaluator to skip a rule, action timeouts, failed hot reloads that fall back to cached config. The system continues but something unexpected happened.

Info level is for key lifecycle events: configuration loaded, daemon started, session started. These provide a high-level audit trail without overwhelming detail.

Debug level is for detailed evaluation traces: rule matching decisions, expression evaluation steps, action execution details. Enable this when diagnosing why a rule did or did not fire.

slog does not have a built-in trace level. For detailed internal operations (individual expression AST nodes, state store queries, HTTP request/response bodies), ARCI uses debug level with a component group attribute to allow fine-grained filtering.

### Structured logging with slog

slog's structured fields make logs machine-parseable while remaining human-readable.

Using `logger.With()` creates a child logger that carries context attributes through the scope, equivalent to a tracing span. Grouping related attributes with `slog.Group()` provides additional structure for complex operations.

### Log output formats

The CLI supports multiple output formats via slog handlers.

Text format is human-readable, suitable for terminal output during development:

```text
2024-01-15T10:30:45.123Z INFO evaluating hook event event_type=PreToolUse tool=bash
2024-01-15T10:30:45.124Z DEBUG rule matched rule_id=block-rm-rf priority=Critical
```

JSON format is machine-parseable, suitable for log aggregation systems:

```json
{"time":"2024-01-15T10:30:45.123Z","level":"INFO","msg":"evaluating hook event","event_type":"PreToolUse","tool":"bash"}
```

The `ARCI_LOG_FORMAT` environment variable or `logging.format` configuration controls the format.

### Log destinations

The system can send logs to multiple destinations simultaneously using slog's handler interface and the `io.MultiWriter` pattern.

Stderr is the default for foreground processes. The daemon and CLI both write to stderr when running interactively.

File logging writes to a configurable path, with rotation based on size or time. The default location is `$XDG_STATE_HOME/arci/arci.log` for user-level logs and `.arci/logs/` for project-level logs.

The dashboard provides a live log view via WebSocket streaming from the daemon.

## Diagnostic output

Users need clear feedback when things go wrong. The CLI, daemon, and dashboard each present errors appropriately for their context.

### CLI error presentation

The CLI uses the console for human-readable error output. The CLI formats errors with context and suggestions when possible:

```text
error: configuration validation failed

  --> .arci/rules.yaml:15:3
   |
15 |   condition: tool.name =~ /rm/ && args contains "-rf"
   |              ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
   |
   = error: unknown operator 'contains', did you mean 'has'?
   = help: see https://arci.dev/docs/expressions for expression syntax
```

For validation commands, the CLI collects and reports errors together rather than failing on the first error:

```text
error: found 3 configuration errors

.arci/rules.yaml:
  line 15: unknown operator 'contains' in condition
  line 28: duplicate rule ID 'block-rm'

~/.config/arci/config.yaml:
  line 5: unknown event type 'BeforeToolUse', did you mean 'PreToolUse'?
```

### Exit codes

Exit codes provide machine-readable status for scripts and CI:

```go
package main

const (
 // ExitSuccess indicates the command was successful.
 ExitSuccess = iota // 0
 // ExitError indicates the command failed with a general error.
 ExitError // 1
 // ExitUsageError indicates the command failed due to invalid input.
 ExitUsageError // 2
 // ExitConfigError indicates the command failed due to invalid configuration.
 ExitConfigError // 3
)
```

For the evaluate command, exit codes have different semantics: exit code 0 means allow (proceed with possible output modifications), exit code 10 means block (operation denied), and exit code 128 means catastrophic failure (something went wrong internally). Catastrophic failures still result in fail-open behavior at the assistant level; the exit code signals to monitoring that the team should investigate.

### Dashboard diagnostics

The dashboard provides visual diagnostics for configuration and rule status.

The configuration panel shows each configuration source with its load status (loaded, error, not found). Errors display inline with the source path and expand to show full error details.

Rule validation status appears as badges: green checkmark for valid rules, yellow warning for rules with non-fatal issues, red X for rules that failed to compile. Hovering shows the specific issue.

The rule tester provides step-by-step evaluation traces, showing each rule's condition, whether it matched, and why. This is invaluable for debugging "why didn't the rule fire?" questions.

## Troubleshooting workflows

Common troubleshooting scenarios and how to approach them.

### The rule does not match

Start with `arci hook policy test <rule-selector> --input @sample.json` to see if the rule matches against known input. If it does not match, the test command shows which part of the condition evaluated to false.

Confirm that the rule has `enabled: true` with `arci hook policy explain <rule-selector>`. The output shows enabled status and source file.

Verify the event type filter. A rule with `events: [PostToolUse]` does not match `PreToolUse` hooks.

Check priority and terminal rules. A higher-priority terminal rule may stop evaluation before your rule runs. Use `arci hook logs --rule <rule-selector>` to see if the evaluator reaches the rule.

Enable debug logging with `ARCI_LOG_LEVEL=debug` to see expression evaluation details.

### The rule matches when it should not

Use `arci hook policy test <rule-selector> --input @sample.json` with input that should not match. The test output shows the evaluation trace.

Check for overly broad conditions. A condition like `tool.name =~ /rm/` matches "transform" as well as "rm."

Review rule precedence. Lower-precedence rules may be overriding your rule's decision.

Use the dashboard rule tester for interactive exploration of complex conditions.

### The action does not execute

Check that the action type is compatible with the hook type. Some actions only make sense for certain events.

Review timeout configuration. Shell commands have a default timeout; the system may stop long-running commands before completion.

Check action handler output with debug logging. The system logs invalid output from an action handler as a warning.

For shell actions, verify the command path is correct and executable. The shell action runs in the project directory by default.

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

File watcher failures trigger watcher restart. If the system cannot restart the watcher, the daemon continues without hot reload; manual reload is still available.

### Manual recovery

Some situations require manual intervention.

Corrupted state store: `arci state clear --all` wipes the state database. For complete reset, delete the SQLite file at `$XDG_STATE_HOME/arci/state.db`.

Stale daemon: `arci daemon stop && arci daemon start` performs a clean restart. The `--force` flag kills a daemon that is not responding to graceful shutdown.

Extension conflicts: `arci extension sync` reinstalls extensions from the lockfile. For persistent issues, `arci extension remove <name>` and re-add.

## Debugging tools

These tools help diagnose problems.

### `arci doctor`

The doctor command performs full health checks:

```text
$ arci doctor

Installation      OK    arci 0.1.0 at /usr/local/bin/arci
Claude Code       OK    hooks configured in ~/.claude/settings.json
Configuration     OK    12 rules loaded from 3 sources
Rule validation   WARN  1 rule has warnings (use --verbose for details)
State store       OK    state.db accessible, 45 entries
Extensions        OK    2 extensions loaded
Logs              OK    log directory writable

Overall: PASS with warnings (1)
```

Use `--verbose` to see details about warnings and `--fix` to attempt automatic repairs.

### `arci hook policy explain`

The explain command shows everything about a rule:

```text
$ arci hook policy explain block-rm-rf

Rule: block-rm-rf
Source: ~/.config/arci/rules.yaml:15
Priority: critical
Enabled: true
Events: PreToolUse

Condition:
  tool.name == "bash" && input.command =~ /\brm\b.*-rf/

Result: block
Message: "Recursive force delete is not allowed. Please confirm this action."

Actions:
  - log: { level: "warn", message: "Attempted rm -rf" }

Match history (last 7 days):
  - 2024-01-14 15:30:22: matched (blocked)
  - 2024-01-12 09:15:01: matched (blocked)
```

### Debug logging

Enable detailed logging for specific components:

```bash
# All debug output
ARCI_LOG_LEVEL=debug arci run --event PreToolUse

# Filter to specific packages (requires custom slog handler)
ARCI_LOG_FILTER=runner=debug arci run --event PreToolUse
```

The `ARCI_LOG_LEVEL` environment variable sets the global log level. `ARCI_LOG_FILTER` allows fine-grained control over which packages emit debug output; a custom slog handler inspects logger group names to provide this, since the standard library does not include per-package filtering.

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

---

Good error handling is essential for user confidence. When something goes wrong, users should be able to understand why and fix it quickly. The fail-open design ensures that error handling complexity never blocks users from their work, while thorough logging and diagnostics ensure problems do not go unnoticed.
