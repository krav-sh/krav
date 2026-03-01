# Hook

The `arci hook` command group contains all hook-related commands: the Claude Code integration point (`apply`), policy management (`policy`), execution logs (`logs`), and aggregated metrics (`stats`).

## Apply

The apply subcommand is the primary integration point with Claude Code. It reads hook input from stdin, applies matching policies, and returns the result.

### Synopsis

```
arci hook apply --event <type>
```

### Description

The `--event` flag specifies the canonical hook event type (e.g., `pre_tool_call`, `post_tool_call`, `pre_prompt`). See [hook schema](../../hooks/hook-schema.md) for the complete list of event types and their mappings to Claude Code hook names.

The command reads JSON from stdin, applies the six-stage policy evaluation pipeline (see [execution model](../../hooks/execution-model.md)), writes any output JSON to stdout, and exits with the appropriate code.

In direct mode (the default), the command loads configuration, compiles policies, and applies them on every invocation. This incurs 50-200ms overhead but requires no daemon process. If the daemon is running and reachable, the command can optionally delegate to it for faster evaluation with cached configuration.

Configuration errors and policy evaluation failures never block Claude Code. Only explicit deny decisions from successfully-evaluated validation rules block operations. This fail-open behavior is a critical safety property.

```go
type ApplyOptions struct {
	// The canonical hook event type
	Event string
}
```

### Options

- `--event <type>` — The canonical hook event type (required). See [hook schema](../../hooks/hook-schema.md) for valid event types.

### Exit codes

| Code | Meaning |
|------|---------|
| 0    | Allow or warn — operation should proceed (possibly with output) |
| 2    | Deny — operation is blocked |
| 128  | Catastrophic error — something went seriously wrong |

See [exit-codes.md](../exit-codes.md) for full exit code reference.

---

## Policy

The policy subcommand group provides the most frequently used commands for working with policies. These commands cover listing, enabling/disabling, inspecting, testing, and validating policies.

### Synopsis

    arci hook policy <subcommand> [options]

### Description

Policy commands operate on the merged configuration — the result of combining all configuration sources in precedence order. Commands that accept a selector can target one or more policies by name or pattern. See [Selectors](../selectors.md) for the full selector syntax.

### Subcommands

#### list

Lists all policies in the merged configuration. For each policy, it displays the name, source file, priority, enabled status, event types, and rule count. The `arci hook policies` command is an alias.

```
$ arci hook policy list
NAME                        SOURCE                              PRIORITY  STATUS   EVENTS           RULES
security-baseline           ~/.config/arci/policies.d/    high      enabled  pre_tool_call    3
coding-standards-injection  .arci/policies.d/             medium    enabled  pre_tool_call    2
session-tracking            ~/.config/arci/policies.d/    low       enabled  pre_tool_call    2
```

#### enable

Enables one or more policies matching the selector.

```bash
# Enable all policies matching a pattern
arci hook policy enable 'security-*'
```

#### disable

Disables one or more policies matching the selector.

```bash
# Disable a specific policy
arci hook policy disable coding-standards-injection
```

#### explain

Shows detailed information about matching policies. For a single match, it displays the full configuration including all rules, source file, parsed condition expressions, and recent match history. For multiple matches, it shows summary information for each.

```
$ arci hook policy explain security-baseline
POLICY: security-baseline
SOURCE: ~/.config/arci/policies.d/security-baseline.yaml
PRIORITY: high
STATUS: enabled

MATCH:
  events: [pre_tool_call]
  tools: [Bash, Write, Edit]

CONDITIONS:
  - not-safe-mode: !session.safe_mode

PARAMETERS:
  - blockedCommands (from: specs, defaults: [])

VARIABLES:
  - command: tool_input.command ?? ""
  - is_blocked: params.blockedCommands.exists(b, command.contains(b))

RULES:
  NAME              TYPE       ACTION   MATCH
  blocked-commands  validate   deny     tools: [Bash]
  large-file-warn   validate   warn     tools: [Write]
  track-edits       effects    -        tools: [Write, Edit]

RECENT HISTORY:
  2024-01-15 14:32:01  pre_tool_call/Bash   blocked-commands  deny
  2024-01-15 14:30:45  pre_tool_call/Write  large-file-warn   warn
```

#### test

Dry-runs policies against sample input. Shows which matching policies and rules would fire and what actions would execute.

**Flags:**

- `--event <type>` — Hook event type (defaults to `pre_tool_call`).
- `--input <json>` — JSON input, or `@filename` to read from a file.

When multiple policies match the selector, all are tested against the input.

```bash
# Test a specific policy
arci hook policy test security-baseline --input '{"tool_name":"Bash","tool_input":{"command":"rm -rf /tmp"}}'

# Test from a file
arci hook policy test security-baseline --input @test-cases/dangerous-rm.json

# Test all policies matching a pattern
arci hook policy test 'security-*' --event pre_tool_call --input '{"tool_name":"Bash","tool_input":{"command":"git push --force"}}'
```

#### validate

Validates all policy and rule expressions, checking that conditions parse correctly, that action types are known, and that action parameters are valid. Useful for CI/CD pipelines and pre-commit hooks.

```
$ arci hook policy validate
Validating 12 policies...
  security-baseline: OK (3 rules)
  coding-standards-injection: OK (2 rules)
  rate-limiting: ERROR
    Rule "check-rate": invalid CEL expression at line 1, column 15:
      undeclared reference to 'attempt' (did you mean 'attempts'?)

Validation failed: 1 error in 12 policies
```

**Exit codes:**

- `0` — All policies are valid.
- `1` — One or more policies have errors.

### Enable/disable state

> **TODO**: Finalize the sidecar file design. Open questions:
>
> The current thinking is that enable/disable state lives in a sidecar file at each configuration layer (e.g., `~/.config/arci/policies.json` or `.arci/policies.json`). This keeps policy definitions clean and makes enable/disable state a separate concern.
>
> Precedence: If a policy is disabled at the user level but enabled at the project level, which wins? Options are "higher precedence wins" (consistent with policy definition precedence) or "most restrictive wins" (any disable in the chain keeps it disabled).
>
> File format: Minimal JSON with `enabled` and `disabled` lists to support explicit overrides? Or just a `disabled` list with default-enabled assumption?
>
> The `--scope` flag on enable/disable commands would control which layer's sidecar is modified.

---

## Logs

The logs subcommand provides access to hook execution logs. By default it tails recent events, streaming new events as they occur.

### Synopsis

    arci hook logs [options]

### Description

This command reads from the hook execution logs that arci writes on every evaluation. It supports both streaming (tail-like) and bounded query modes. When `--since` or `--until` are provided without `--follow`, the command performs a bounded search and exits after displaying results.

### Options

- `--follow`, `-f` — Continue streaming after showing recent history.
- `--since <time>` — Show events after this time.
- `--until <time>` — Show events before this time.
- `--filter <pattern>` — Pattern matching on event content.
- `--event <type>` — Filter by event type.
- `--policy <name>` — Filter by policy name.
- `--rule <name>` — Filter by rule name.
- `--format <format>` — Output format: `text` (default) or `json`.

---

## Stats

The stats subcommand aggregates metrics from the project-level hook logs that arci writes on every evaluation. This works regardless of whether you're using the daemon or direct execution mode.

### Synopsis

    arci hook stats [options]

### Description

By default the command shows a summary including total invocations, block/allow counts, policy and rule match frequencies, and action execution counts by type. Several filtering options scope the analysis.

Filters can be combined to narrow results. For example, `arci hook stats --policy 'security-*' --since 'last week'` shows statistics for security policies over the past week.

### Options

- `--since <time>` — Filter to events after this time (e.g., `'last week'`, `'2024-01-15'`).
- `--until <time>` — Filter to events before this time.
- `--policy <selector>` — Filter to events involving policies matching the selector. Useful for answering "how often does this policy actually trigger?"
- `--rule <selector>` — Filter by rule name within matched policies.
- `--event <type>` — Filter by hook event type.
- `--format <format>` — Output format: `text` (default) for human-readable summary, `json` for structured output suitable for scripting.

### Examples

```bash
# Show overall summary
arci hook stats

# Security policy stats for the past week
arci hook stats --policy 'security-*' --since 'last week'

# JSON output for scripting
arci hook stats --format json --since '2024-01-01'
```

---

## See also

- [Hook schema](../../hooks/hook-schema.md)
- [Execution model](../../hooks/execution-model.md)
- [Selectors](../selectors.md)
- [Policy Model](../../hooks/policy-model.md)
- [Exit codes](../exit-codes.md)
