# Policy model

This document describes Krav's policy model.

## Design goals

The policy model favors clarity and self-containment over enterprise flexibility. A policy file should be readable top to bottom and understandable without cross-referencing other documents. When you need to know why a tool call failed, you should be able to find the answer in one place.

Policies are self-contained units that declare their own matching, parameters, and enforcement. Reuse comes from variables and macros for expression reuse, parameter providers for external data, and the config cascade for layered overrides.

## Policy structure

A policy is a YAML document with the following top-level fields:

```yaml
version: 1
name: security-baseline
metadata:
  description: Enterprise security baseline for AI coding assistants
  labels:
    owner: infosec
    category: security

config:
  failurePolicy: allow
  priority: high

match:
  events: [pre_tool_call]
  tools: [Bash, Write, Edit]
  paths:
    exclude: ["**/node_modules/**", "**/.git/**"]

conditions:
  - name: not-safe-mode
    expression: '!session.safe_mode'

parameters:
  - name: blockedCommands
    from: specs
    defaults: []

variables:
  - name: command
    expression: 'tool_input.command ?? ""'
  - name: is_blocked
    expression: 'params.blockedCommands.exists(b, command.contains(b))'

macros:
  - name: is_destructive
    expression: |
      command.matches("rm\\s+-rf") ||
      command.contains("chmod 777") ||
      command.contains("> /dev/sd")

rules:
  - name: blocked-commands
    match:
      tools: [Bash]
    validate:
      expression: '!is_blocked'
      message: "Command blocked by security policy"
      action: deny
```

### Version

Schema version for the policy format. Currently `1`. Future schema versions can evolve without requiring apiVersion/kind ceremony.

### Name

Unique identifier for the policy. Used in logging, metrics, and error messages. Names should be kebab-case and descriptive.

### Metadata

Optional metadata about the policy. The `description` field provides human-readable documentation. The `labels` field contains arbitrary key-value pairs for organization and filtering.

```yaml
metadata:
  description: |
    Blocks dangerous shell commands that could cause data loss or security issues.
    Applied to all Bash tool calls except in safe mode.
  labels:
    owner: infosec
    category: security
    compliance: soc2
```

### config

Policy-level configuration that affects evaluation behavior.

```yaml
config:
  failurePolicy: allow  # allow | deny
  priority: high        # critical | high | medium | low
  default_state: audit  # enabled | disabled | audit
```

The `failurePolicy` determines what happens when the policy itself errors (expression evaluation failure, parameter resolution failure, etc.). With `allow`, the engine skips the policy and the tool call proceeds. With `deny`, errors block the tool call. The default is `allow`, consistent with fail-open semantics.

The `priority` determines evaluation order. The engine groups policies by priority and runs them from highest to lowest: critical, high, medium, low. Within a priority level, the config cascade orders them (universal then project). Higher-priority policies can mutate requests before lower-priority policies see them. The default is `medium`.

The `default_state` declares the policy's preferred enforcement state when no manifest explicitly references it. This is useful for policies that should be opt-in (`default_state: disabled`) or policies under evaluation (`default_state: audit`). The precedence order is: manifest reference > policy self-declaration > layer `defaultBehavior`. If omitted, the policy inherits from the layer's `defaultBehavior` setting.

### Match

Structural matching that determines whether this policy applies to a given hook event. Match conditions use OR-within-arrays, AND-across-fields logic.

```yaml
match:
  events: [pre_tool_call, post_tool_call]
  tools: [Bash, Write, Edit]
  paths:
    include: ["src/**", "lib/**"]
    exclude: ["**/*.test.ts", "**/__tests__/**"]
  branches:
    include: ["main", "release/*"]
    exclude: ["experiments/*"]
```

All fields are optional. Omitting a field means "match all" for that dimension.

The `events` field specifies which hook events this policy handles. Common values include `pre_tool_call`, `post_tool_call`, and `stop`.

The `tools` field specifies which tool names this policy handles: `Bash`, `Write`, `Edit`, `Read`, `Glob`, `Grep`, etc.

The `paths` field filters by path when the hook event involves a file. Both `include` and `exclude` accept glob patterns. Exclude takes precedence over include.

The `branches` field filters by git branch. Useful for policies that should only apply on protected branches or only on feature branches.

Match evaluation is fast because it uses structural comparison and indexing. The engine can filter thousands of policies to a handful of candidates before evaluating any CEL expressions.

### Conditions

Runtime conditions that determine whether the policy applies, checked after structural matching passes. Conditions are CEL expressions that must all return true (AND logic with short-circuit evaluation).

```yaml
conditions:
  - name: not-safe-mode
    expression: '!session.safe_mode'
  - name: on-protected-branch
    expression: '$current_branch() in ["main", "release"]'
  - name: has-uncommitted-changes
    expression: '$git_is_dirty()'
```

Each condition has a `name` for debugging and an `expression` containing CEL code. Conditions can reference parameters, variables, and built-in functions.

If any condition returns false, the engine skips the policy for this event. This differs from a validation failure: skipped policies do not appear in audit logs or contribute to the final decision.

### parameters

Parameters bring external data into policy evaluation. A parameter can be a static value, a reference to a named provider, an inline provider definition, or an environment variable.

```yaml
parameters:
  # Static value
  - name: maxFileSize
    value: 50000

  # From named provider with fallback
  - name: workflowContext
    from: specs
    defaults:
      workflowPhase: none
      strictMode: false

  # Inline file provider
  - name: teamConfig
    from:
      file:
        path: .krav/team.json
    defaults: {}

  # From environment variable
  - name: environment
    from:
      env: KRAV_ENV
    default: development

  # HTTP provider
  - name: dynamicConfig
    from:
      http:
        endpoint: http://localhost:9100/config
        timeout: 100ms
    defaults:
      enabled: true
```

The engine resolves parameters before running any expressions. If resolution fails and no defaults exist, the behavior depends on `config.failurePolicy`. Resolved parameters are available in expressions as `params.parameterName`.

Provider types include `file` (reads JSON/YAML from disk), `http` (fetches from an endpoint), `env` (reads environment variable), and named providers defined in the Krav configuration.

### Variables

Variables hold values derived from parameters, the hook event, built-in functions, or other variables. They provide a way to factor out common expressions and make rules more readable.

```yaml
variables:
  - name: command
    expression: 'tool_input.command ?? ""'
  - name: file_path
    expression: 'tool_input.file_path ?? ""'
  - name: content_size
    expression: 'has(tool_input.content) ? size(tool_input.content) : 0'
  - name: is_shell
    expression: 'tool_name == "Bash"'
  - name: blocked_patterns
    expression: 'params.blockedCommands ?? []'
  - name: matches_blocked
    expression: 'blocked_patterns.exists(p, command.contains(p))'
```

The engine processes variables in declaration order, so later variables can reference earlier ones. Variables defined at the policy level are available to all rules. Rules can define their own variables that shadow policy-level variables.

### Macros

Macros are reusable expression fragments that you can call from CEL expressions. They work well for complex logic that appears in many rules or policies.

```yaml
macros:
  - name: is_destructive_command
    expression: |
      command.matches("rm\\s+-rf") ||
      command.contains("chmod 777") ||
      command.contains("> /dev/sd")

  - name: on_protected_branch
    expression: |
      $current_branch() in params.protectedBranches
```

Call macros with a `$` prefix: `$is_destructive_command()`. They can reference variables and parameters from the calling context.

### Rules

Rules are the core of a policy. Each rule defines a specific check, mutation, or side effect. A policy contains one or more rules.

```yaml
rules:
  - name: blocked-rm-rf
    match:
      tools: [Bash]
    validate:
      expression: '!command.matches("rm\\s+-rf")'
      message: "rm -rf is blocked"
      action: deny

  - name: large-file-warning
    match:
      tools: [Write]
    validate:
      expression: 'content_size <= params.maxFileSize'
      message: "File exceeds {{ params.maxFileSize }} bytes"
      action: warn

  - name: inject-coding-standards
    match:
      tools: [Write, Edit]
    conditions:
      - expression: 'file_path.endsWith(".py")'
    mutate:
      expression: |
        object.with({
          context: object.context + "\n\nFollow PEP 8 style guidelines."
        })

  - name: track-modifications
    match:
      tools: [Write, Edit]
    effects:
      - type: setState
        scope: session
        key: files_modified
        value: '{{ files_modified + 1 }}'
```

## Rule structure

Each rule has the following fields:

### Name

Required. Unique identifier within the policy. Used in logging and error messages.

### Match

Optional structural matching that narrows the policy's match. A rule's match intersects with the policy's match, and both must pass for the rule to run. Syntax is identical to policy-level match.

```yaml
rules:
  - name: bash-specific-check
    match:
      tools: [Bash]  # narrows policy's tool list
    validate:
      # ...
```

### Conditions

Optional runtime conditions for the rule. The engine checks these after the rule's structural match passes. If any condition returns false, the engine skips the rule (not a failure, just not applicable).

```yaml
rules:
  - name: review-phase-restrictions
    conditions:
      - expression: 'params.workflowPhase == "review"'
    validate:
      expression: '!command.contains("rm ")'
      message: "File deletion blocked during review"
      action: deny
```

The distinction between conditions and validation is important: conditions determine whether the rule is relevant, validation determines whether it passes. The engine silently skips irrelevant rules. A rule that is relevant but fails validation is a violation.

### Variables

Optional rule-local variables. These can shadow policy-level variables and are only visible within this rule.

```yaml
rules:
  - name: check-file-size
    variables:
      - name: size_limit
        expression: 'params.limits[file_extension] ?? params.defaultLimit'
    validate:
      expression: 'content_size <= size_limit'
      message: "File exceeds limit for {{ file_extension }} files"
      action: warn
```

### Validation

Validation rules check a condition and produce a result if it fails. Each rule can have at most one `validate` block.

```yaml
validate:
  expression: '!command.contains("--force")'
  message: "Force flag is not allowed"
  action: deny  # deny | warn | audit
```

The `expression` is a CEL expression that must return true for the validation to pass. If it returns false, the validation fails with the specified action.

The `message` is a human-readable explanation shown when validation fails. It can include template expressions like `{{ variable }}` for variable content.

The `action` determines what happens on failure: `deny` blocks the tool call, `warn` allows it with a warning, `audit` logs silently without user notification.

### Mutate

Mutation rules transform the hook event. The engine applies mutations before validations, so validation rules see the mutated state.

```yaml
mutate:
  expression: |
    object.with({
      tool_input: object.tool_input.with({
        command: object.tool_input.command + " --dry-run"
      })
    })
```

The `expression` is a CEL expression that receives `object` (the current event state) and returns the modified state. Mutations use CEL's immutable update syntax.

### Effects

Effects are side actions that run after validation/mutation. A rule can have many effects.

```yaml
effects:
  - type: setState
    scope: session
    key: command_count
    value: '{{ command_count + 1 }}'
    when: always  # always | on_pass | on_fail

  - type: notify
    title: "Command blocked"
    message: "{{ rule.message }}"
    when: on_fail

  - type: log
    level: info
    message: "Processed {{ tool_name }} call"
    when: on_pass
```

The `when` field controls when the effect runs: `always` (default), `on_pass` (rule passed or the engine skipped it), `on_fail` (validation failed).

Effect types include:

`setState` persists a value to the session or project state store. The `scope` is either `session` (isolated per conversation) or `project` (persists across sessions). The `key` and `value` can include template expressions.

`notify` sends a notification to the user via OS notification system.

`log` writes to the Krav log at the specified level.

Future versions may introduce effect types for webhooks, metrics, and other integrations.

### Rule type constraints

A rule must have at least one of `validate`, `mutate`, or `effects`. It cannot have both `validate` and `mutate` (these are mutually exclusive action types), but it can combine either with `effects`.

```yaml
# Valid: validation only
rules:
  - name: check-something
    validate: { ... }

# Valid: mutation only
rules:
  - name: transform-something
    mutate: { ... }

# Valid: effects only
rules:
  - name: track-something
    effects: [{ ... }]

# Valid: validation with effects
rules:
  - name: check-and-track
    validate: { ... }
    effects: [{ ... }]

# Valid: mutation with effects
rules:
  - name: transform-and-track
    mutate: { ... }
    effects: [{ ... }]

# Invalid: validation and mutation together
rules:
  - name: not-allowed
    validate: { ... }
    mutate: { ... }  # Error: cannot combine validate and mutate
```

## Expression context

All CEL expressions have access to the following context:

### Hook event data

```text
tool_name         # string: normalized tool name ("Bash", "Write", etc.)
tool_input        # map: tool-specific input fields
event_type        # string: "pre_tool_call", "post_tool_call", etc.
session_id        # string: current session identifier
timestamp         # timestamp: when the event occurred
```

### Parameters

```text
params.paramName  # value from parameter definition
```

### Variables

```text
variableName      # value from variable definition (policy or rule level)
```

### Built-in functions

```text
$file_exists(path)           # bool: check if file exists
$current_branch()            # string: current git branch
$git_is_dirty()              # bool: check for uncommitted changes
$is_staged(path)             # bool: check if file is staged
$env(name)                   # string: environment variable
$session_get(key, default)   # any: session state value
$project_get(key, default)   # any: project state value
$matches_glob(path, pattern) # bool: glob pattern matching
```

### Macros

```text
$macroName()                 # result of macro expression
```

## Example policies

### Simple validation policy

```yaml
version: 1
name: no-rm-rf

config:
  priority: critical

match:
  tools: [Bash]

rules:
  - name: block-rm-rf
    validate:
      expression: '!tool_input.command.matches("rm\\s+-rf\\s+/")'
      message: "Catastrophic deletion blocked"
      action: deny
```

### Parameterized policy with conditions

```yaml
version: 1
name: protected-branch-restrictions

parameters:
  - name: protectedBranches
    from: specs
    defaults: ["main"]

variables:
  - name: current_branch
    expression: '$current_branch()'
  - name: on_protected
    expression: 'current_branch in params.protectedBranches'

match:
  tools: [Bash, Write, Edit]

conditions:
  - name: on-protected-branch
    expression: 'on_protected'

rules:
  - name: no-force-push
    match:
      tools: [Bash]
    conditions:
      - expression: 'tool_input.command.startsWith("git push")'
    validate:
      expression: '!tool_input.command.contains("--force")'
      message: "Force push to {{ current_branch }} is not allowed"
      action: deny

  - name: no-direct-commits
    match:
      tools: [Bash]
    conditions:
      - expression: 'tool_input.command.startsWith("git commit")'
    validate:
      expression: 'false'  # always fails when conditions pass
      message: "Direct commits to {{ current_branch }} are not allowed"
      action: deny
```

### Mutation policy

```yaml
version: 1
name: coding-standards-injection

match:
  events: [pre_tool_call]
  tools: [Write, Edit]

variables:
  - name: file_path
    expression: 'tool_input.file_path ?? ""'
  - name: is_python
    expression: 'file_path.endsWith(".py")'
  - name: is_typescript
    expression: 'file_path.endsWith(".ts") || file_path.endsWith(".tsx")'

rules:
  - name: python-standards
    conditions:
      - expression: 'is_python'
    mutate:
      expression: |
        object.with({
          context: object.context + "\n\nPython guidelines: Use type hints, follow PEP 8, prefer pathlib over os.path."
        })

  - name: typescript-standards
    conditions:
      - expression: 'is_typescript'
    mutate:
      expression: |
        object.with({
          context: object.context + "\n\nTypeScript guidelines: Enable strict mode, avoid any, use interfaces for objects."
        })
```

### Effects-only policy

```yaml
version: 1
name: session-tracking

match:
  events: [pre_tool_call]

variables:
  - name: tool_count
    expression: '$session_get("tool_count", 0)'

rules:
  - name: increment-counter
    effects:
      - type: setState
        scope: session
        key: tool_count
        value: '{{ tool_count + 1 }}'

  - name: notify-on-stop
    match:
      events: [stop]
    effects:
      - type: notify
        title: "Session complete"
        message: "Processed {{ tool_count }} tool calls"
```

### Combined validation and effects

```yaml
version: 1
name: rate-limiting

parameters:
  - name: maxAttempts
    value: 3

variables:
  - name: attempts
    expression: '$session_get("dangerous_command_attempts", 0)'
  - name: command
    expression: 'tool_input.command ?? ""'

match:
  tools: [Bash]

conditions:
  - expression: '$is_destructive_command()'

macros:
  - name: is_destructive_command
    expression: |
      command.matches("rm\\s+-rf") ||
      command.contains("chmod 777")

rules:
  - name: rate-limit-dangerous-commands
    validate:
      expression: 'attempts < params.maxAttempts'
      message: "Too many dangerous command attempts ({{ attempts }}/{{ params.maxAttempts }})"
      action: deny
    effects:
      - type: setState
        scope: session
        key: dangerous_command_attempts
        value: '{{ attempts + 1 }}'
        when: always
```
