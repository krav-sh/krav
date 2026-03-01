# Hook

The `arci hook` command group contains all hook-related commands: the Claude Code integration point (`apply`), policy management (`list`, `enable`, `disable`, `explain`, `test`, `verify`), execution logs (`logs`), and aggregated metrics (`stats`).

## Apply

The apply subcommand is the primary integration point with Claude Code. It reads hook input from stdin, applies matching policies, and returns the result.

### Synopsis

```bash
arci hook apply --event <type>
```

### Description

The `--event` flag specifies the canonical hook event type, such as `pre_tool_call`, `post_tool_call`, or `pre_prompt`. See [hook schema](../../hooks/hook-schema.md) for the complete list of event types and their mappings to Claude Code hook names.

The command reads JSON from stdin, applies the six-stage policy evaluation pipeline (see [execution model](../../hooks/execution-model.md)), writes any output JSON to stdout, and exits with the appropriate code.

In direct mode (the default), the command loads configuration, compiles policies, and applies them on every invocation. This incurs 50 - 200 ms overhead but requires no daemon process. If the daemon is running and reachable, the command can optionally delegate to it for faster evaluation with cached configuration.

Configuration errors and policy evaluation failures never block Claude Code. Only explicit deny decisions from successfully evaluated validation rules block operations. This fail-open behavior is a critical safety property.

### Options

- `--event <type>`: the canonical hook event type (required). See [hook schema](../../hooks/hook-schema.md) for valid event types.

### Exit codes

| Code | Meaning |
|------|---------|
| 0    | Allow or warn; operation should proceed (possibly with output) |
| 2    | Deny; the operation cannot proceed |
| 128  | Catastrophic error; something went seriously wrong |

See [exit-codes.md](../exit-codes.md) for full exit code reference.

---

## List

Lists all policies in the merged configuration. For each policy, it displays the name, source file, priority, enabled status, event types, and rule count. The `arci hook policies` command is an alias.

### Synopsis

```bash
arci hook list
```

### Description

Policy commands operate on the merged configuration, which combines all configuration sources in precedence order. Commands that accept a selector can target one or more policies by name or pattern. See [Selectors](#selectors) for the full selector syntax.

```text
$ arci hook list
NAME                        SOURCE                              PRIORITY  STATUS   EVENTS           RULES
security-baseline           ~/.config/arci/policies.d/    high      enabled  pre_tool_call    3
coding-standards-injection  .arci/policies.d/             medium    enabled  pre_tool_call    2
session-tracking            ~/.config/arci/policies.d/    low       enabled  pre_tool_call    2
```

---

## Enable

Enables one or more policies matching the selector.

### Synopsis

```bash
arci hook enable <selector>
```

### Examples

```bash
# Enable all policies matching a pattern
arci hook enable 'security-*'
```

---

## Disable

Disables one or more policies matching the selector.

### Synopsis

```bash
arci hook disable <selector>
```

### Examples

```bash
# Disable a specific policy
arci hook disable coding-standards-injection
```

### Enable and turn-off state

Enable and turn-off state lives in a sidecar file at each configuration layer (`~/.config/arci/policies.json` or `.arci/policies.json`). This keeps policy definitions clean and makes enable/turn-off state a separate concern.

Open design questions:

> **Precedence:** if the user level turns off a policy but the project level enables it, which wins? Options include "higher precedence wins" (consistent with policy definition precedence) and "most restrictive wins" (any turn-off in the chain keeps it off).
>
> **File format:** minimal JSON with `enabled` and `disabled` lists to support explicit overrides, or just a `disabled` list with a default-enabled assumption?
>
> The `--scope` flag on enable and turn-off commands would control which layer's sidecar the command modifies.

---

## Explain

Shows detailed information about matching policies. For a single match, it displays the full configuration including all rules, source file, parsed condition expressions, and recent match history. When many policies match, it shows summary information for each.

### Synopsis

```bash
arci hook explain <selector>
```

### Examples

```text
$ arci hook explain security-baseline
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

---

## Test

Dry-runs policies against sample input. Shows which matching policies and rules would fire and what actions would execute.

### Synopsis

```bash
arci hook test <selector> [options]
```

### Options

- `--event <type>`: hook event type (defaults to `pre_tool_call`).
- `--input <json>`: JSON input, or `@filename` to read from a file.

When many policies match the selector, the command tests each one against the input.

### Examples

```bash
# Test a specific policy
arci hook test security-baseline --input '{"tool_name":"Bash","tool_input":{"command":"rm -rf /tmp"}}'

# Test from a file
arci hook test security-baseline --input @test-cases/dangerous-rm.json

# Test all policies matching a pattern
arci hook test 'security-*' --event pre_tool_call --input '{"tool_name":"Bash","tool_input":{"command":"git push --force"}}'
```

---

## Verify

Checks all policy and rule expressions, confirming that conditions parse correctly, that the system recognizes each action type, and that action parameters contain valid values. Useful for CI/CD pipelines and pre-commit hooks.

### Synopsis

```bash
arci hook verify
```

### Examples

```text
$ arci hook verify
Validating 12 policies...
  security-baseline: OK (3 rules)
  coding-standards-injection: OK (2 rules)
  rate-limiting: ERROR
    Rule "check-rate": invalid CEL expression at line 1, column 15:
      undeclared reference to 'attempt' (did you mean 'attempts'?)

Validation failed: 1 error in 12 policies
```

### Exit codes

- `0`: all policies are valid.
- `1`: one or more policies have errors.

---

## Logs

The logs subcommand provides access to hook execution logs. By default it tails recent events, streaming new events as they occur.

### Synopsis

```bash
arci hook logs [options]
```

### Description

This command reads from the hook execution logs that ARCI writes on every evaluation. It supports both streaming (tail-like) and bounded query modes. When the caller supplies `--since` or `--until` without `--follow`, the command performs a bounded search and exits after displaying results.

### Options

- `--follow`, `-f`: continue streaming after showing recent history.
- `--since <time>`: show events after this time.
- `--until <time>`: show events before this time.
- `--filter <pattern>`: pattern matching on event content.
- `--event <type>`: filter by event type.
- `--policy <name>`: filter by policy name.
- `--rule <name>`: filter by rule name.
- `--format <format>`: output format, either `text` (default) or `json`.

---

## Stats

The stats subcommand aggregates metrics from the project-level hook logs that ARCI writes on every evaluation. This works regardless of whether you use the daemon or direct execution mode.

### Synopsis

```bash
arci hook stats [options]
```

### Description

By default the command shows a summary including total invocations, block/allow counts, policy match frequencies, rule match frequencies, and action execution counts by type. The filtering options below scope the analysis.

You can combine filters to narrow results: `arci hook stats --policy 'security-*' --since 'last week'` shows statistics for security policies over the past week.

### Options

- `--since <time>`: filter to events after this time (such as `'last week'` or `'2024-01-15'`).
- `--until <time>`: filter to events before this time.
- `--policy <selector>`: filter to events involving policies matching the selector. Useful for answering "how often does this policy actually trigger?"
- `--rule <selector>`: filter by rule name within matched policies.
- `--event <type>`: filter by hook event type.
- `--format <format>`: output format, either `text` (default) for human-readable summary or `json` for structured output suitable for scripting.

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

## Selectors

Selectors identify one or more policies (and optionally rules within them) for commands to operate on. Any command that accepts a `<selector>` argument uses the syntax described here.

### Exact name matching

The simplest selector is an exact policy name:

```bash
arci hook explain security-baseline
```

This matches the single policy named `security-baseline`.

### Glob patterns

Selectors support glob patterns for matching more than one policy:

- `security-*` matches all policies whose names start with `security-`
- `*-injection` matches policies ending with `-injection`
- `*` matches all policies

```bash
# Enable all security policies
arci hook enable 'security-*'

# Disable all policies (use with care)
arci hook disable '*'
```

### Policy:rule syntax

To target a specific rule within a policy, use the `policy:rule` syntax:

- `security-baseline:blocked-commands` targets the `blocked-commands` rule within the `security-baseline` policy
- `security-baseline:*` targets all rules in the `security-baseline` policy

### Rule-level glob patterns

The rule part also supports glob patterns:

- `security-baseline:block-*` matches rules starting with `block-` in the `security-baseline` policy
- `*:track-*` matches rules starting with `track-` in any policy

### Selector examples

```bash
# Explain a specific rule
arci hook explain security-baseline:blocked-commands

# Test all rules in a policy
arci hook test security-baseline:*

# Test tracking rules across all policies
arci hook test '*:track-*' --input @test-input.json
```

Rule-level selectors are useful with `explain` and `test` commands when you want to focus on specific behavior within a policy.

---

## See also

- [Hook schema](../../hooks/hook-schema.md)
- [Execution model](../../hooks/execution-model.md)
- [Policy Model](../../hooks/policy-model.md)
- [Exit codes](../exit-codes.md)
