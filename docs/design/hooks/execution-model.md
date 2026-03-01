# Execution model

This document describes how arci evaluates policies when a hook event arrives. It covers the evaluation pipeline, priority cascading, result aggregation, and the data flow through the system.

## Overview

When Claude Code fires a hook (for example, before executing a Bash command), arci receives the event and evaluates all matching policies to produce a response. The response determines whether the tool call proceeds, any mutations to apply, warnings to surface, and side effects to execute.

The evaluation pipeline has several stages: structural matching filters policies quickly using indexable criteria, condition evaluation applies dynamic CEL expressions, rule evaluation runs the actual validations and mutations, effect execution handles side actions, and result aggregation combines outcomes from all matching policies into a single response.

## Enforcement state handling

Before the evaluation pipeline runs, policies are filtered by their enforcement state. Each policy exists in one of three states: enabled, disabled, or audit. The enforcement state is determined by the policy cascade during loading, as described in [policy-loading.md](policy-loading.md).

Enabled policies are evaluated normally. They participate fully in structural matching, condition evaluation, validation, mutation, and effects. Their actions have full force: deny blocks tool calls, mutations are applied, effects are executed.

Disabled policies are completely skipped. They are not matched, evaluated, or logged. They don't appear in the audit trail. Disabling a policy is equivalent to removing it from the configuration entirely, except that it remains visible in diagnostic commands.

Audit policies are evaluated in a "dry-run" mode. The policy is matched and evaluated normally, but its actions are downgraded. Deny actions become warnings (the tool call proceeds with a warning). Mutations are computed but not applied to the event (the original event continues through the pipeline). Effects are logged but not executed. Audit results appear in the audit trail with an `enforcement: audit` marker. This enables teams to test new policies in production without risk.

The enforcement state filter happens before Stage 1 (structural matching). Disabled policies are removed from consideration before any evaluation begins. Audit policies proceed through evaluation with their state tracked for action downgrading.

### Terminology note

This document uses "audit" for two distinct concepts. The validation action `audit` (appearing in `validate.action: audit`) describes what happens when a specific rule's validation fails: it logs silently without user notification. The enforcement state `audit` describes how an entire policy is evaluated: in dry-run mode with all actions downgraded. Context disambiguates since validation actions apply per-rule while enforcement states apply per-policy.

## Evaluation pipeline

### Stage 1: structural matching

The first stage filters policies using structural criteria that can be indexed and evaluated without CEL. This is fast—the engine can scan hundreds of policies in microseconds.

For each policy, the engine checks whether its `match` block is satisfied by the current event:

```
event.type ∈ policy.match.events (or events is empty)
AND event.tool ∈ policy.match.tools (or tools is empty)
AND event.path matches policy.match.paths (include/exclude)
AND event.branch matches policy.match.branches (include/exclude)
```

Policies that don't match structurally are eliminated. No CEL evaluation happens for these policies, and they don't appear in audit logs.

### Stage 2: condition evaluation

For policies that pass structural matching, the engine evaluates their `conditions` in declaration order. Conditions are CEL expressions that must all return true.

```yaml
conditions:
  - name: not-safe-mode
    expression: '!session.safe_mode'
  - name: on-protected-branch
    expression: '$current_branch() in params.protectedBranches'
```

Evaluation short-circuits: if the first condition returns false, subsequent conditions are not evaluated. This is both an optimization and a way to guard expensive expressions behind cheap checks.

If any condition returns false, the policy is skipped for this event. Skipped policies don't contribute to the final decision and appear in audit logs only if verbose logging is enabled.

### Stage 3: parameter resolution

Before evaluating rules, the engine resolves all parameters declared in the policy. Parameters can come from static values, named providers, inline providers, or environment variables.

```yaml
parameters:
  - name: blockedCommands
    from: specs
    defaults: []
```

Parameter resolution may involve I/O (file reads, HTTP calls) and is cached where possible. If resolution fails and no defaults are provided, behavior depends on `config.failurePolicy`:

- `allow` (default): The policy is skipped, tool call proceeds
- `deny`: The policy errors, tool call is blocked

Resolved parameters are available in all subsequent expressions as `params.paramName`.

### Stage 4: variable computation

Policy-level variables are computed in declaration order. Variables can reference parameters, built-in functions, and earlier variables.

```yaml
variables:
  - name: command
    expression: 'tool_input.command ?? ""'
  - name: is_blocked
    expression: 'params.blockedCommands.exists(b, command.contains(b))'
```

Variables are evaluated once per policy and cached for use by all rules in that policy.

### Stage 5: rule evaluation

For each rule in the policy, the engine:

1. Checks structural match (rule's `match` intersected with policy's `match`)
2. Evaluates rule conditions
3. Computes rule-local variables
4. Executes validate or mutate
5. Queues effects for later execution

Rules are evaluated in declaration order within a policy.

```yaml
rules:
  - name: blocked-commands
    match:
      tools: [Bash]
    conditions:
      - expression: 'is_shell'
    variables:
      - name: matched_pattern
        expression: 'params.blockedCommands.filter(b, command.contains(b)).first()'
    validate:
      expression: '!is_blocked'
      message: 'Blocked: {{ matched_pattern }}'
      action: deny
```

Validation rules produce a result (pass or fail with action). Mutation rules produce a transformation to apply. Effects are collected but not executed until stage 6.

### Stage 6: effect execution

After all rules have been evaluated, queued effects are executed. Effects run after the admission decision is determined, so they can't influence whether the tool call proceeds.

Each effect has a `when` condition:

- `always` (default): Effect runs regardless of rule outcome
- `on_pass`: Effect runs only if the rule passed (or was skipped)
- `on_fail`: Effect runs only if the rule's validation failed

```yaml
effects:
  - type: setState
    scope: session
    key: attempts
    value: '{{ attempts + 1 }}'
    when: always

  - type: notify
    title: "Blocked"
    message: "{{ rule.message }}"
    when: on_fail
```

Effects from all matching rules across all matching policies are collected and executed. Effect execution failures are logged but don't affect the tool call decision.

## Priority cascading

Policies are grouped by priority level: critical, high, medium, low. Within each level, policies are ordered by config cascade: universal, then project-specific.

```
critical/universal → critical/project
high/universal → high/project
medium/universal → medium/project
low/universal → low/project
```

### Mutation visibility

Mutations from higher priority levels are applied before lower priority levels evaluate. This allows high-priority policies to transform requests in ways that affect lower-priority policy evaluation.

```
1. Evaluate all critical-priority policies
2. Apply mutations from critical policies
3. Evaluate all high-priority policies (seeing mutated state)
4. Apply mutations from high policies
5. Continue through medium and low
```

Within a priority level, mutations are applied in config cascade order. If two policies at the same priority and cascade level mutate the same field, last-write-wins with a warning in the audit trail.

### Validation aggregation

Validation results from all priority levels are collected. The most restrictive action wins:

```
deny > warn > audit > allow
```

If any rule from any policy produces a `deny`, the overall result is deny. If no denies but some warns, the overall result is warn. Messages from all failures accumulate in the response.

## Data flow

The data flow through evaluation is unidirectional:

```
parameters → variables → conditions → validate/mutate → effects → state
```

Parameters are resolved first and are immutable during evaluation. Variables are computed from parameters and the hook event. Conditions and validation expressions read variables but don't modify them. Effects run last and can persist state for future evaluations.

This unidirectional flow makes evaluation predictable. There are no cycles—a variable can't depend on an effect's outcome, and effects can't modify variables that validations read.

## Result aggregation

After all policies have been evaluated, the engine aggregates results into a single response:

```yaml
action: deny | warn | allow
messages:
  user:
    - "Command blocked by security policy: rm -rf is not allowed"
    - "Warning: Large file detected (52KB exceeds 50KB limit)"
  assistant:
    - "The user's security policy blocks rm -rf commands"
mutations:
  context: "...injected coding standards..."
  tool_input:
    command: "rm -i file.txt"  # -rf replaced with -i
audit:
  - policy: security-baseline
    rule: blocked-rm-rf
    result: fail
    action: deny
  - policy: coding-standards
    rule: python-injection
    result: pass
    mutation: applied
```

### Action determination

The overall action is the most restrictive action from any failing validation:

1. If any validation failed with `action: deny`, overall action is `deny`
2. Else if any validation failed with `action: warn`, overall action is `warn`
3. Else overall action is `allow`

### Message collection

Messages are collected from all failing validations and categorized:

- `user` messages are shown to the human operator
- `assistant` messages are injected into Claude's context

A validation can specify which audience receives its message, or it can go to both by default.

### Mutation composition

Mutations accumulate across all matching policies. Later mutations (lower priority or later in cascade) see and can override earlier mutations.

The final mutated event is what Claude receives if the action is `allow` or `warn`. For `deny`, mutations are still recorded in the audit trail but not applied.

### Audit trail

Every policy evaluation is recorded in the audit trail, regardless of outcome:

- Which policies matched structurally
- Which policies passed conditions
- Which rules fired
- What validations passed or failed
- What mutations were applied
- What effects executed

The audit trail is available through the dashboard, API, and log output.

## Fail-open semantics

arci is designed as a guardrail, not a gate. Errors in the policy system should not block Claude from operating. Only explicit deny decisions block tool calls.

### Error handling

When errors occur during evaluation:

- Parameter resolution failure: Policy skipped (with `failurePolicy: allow`) or denied (with `failurePolicy: deny`)
- Variable computation error: Rule skipped with warning
- Condition evaluation error: Condition treated as false, policy skipped
- Validation expression error: Validation treated as passed with warning
- Mutation expression error: Mutation skipped with warning
- Effect execution error: Logged, does not affect tool call

The principle is: uncertainty defaults to allowing the operation. Users must explicitly configure denial to block tool calls.

### Daemon unavailability

When the CLI can't reach the daemon:

- `daemon.on_unavailable: fallback` — CLI falls back to direct execution
- `daemon.on_unavailable: start` — CLI attempts to spawn the daemon
- `daemon.on_unavailable: fail` — CLI fails (still doesn't block Claude; the assistant's hook system handles CLI failures)

Even if arci fails entirely, Claude continues operating. The assistant's native hook system treats hook script failures as non-blocking by default.

## Performance considerations

### Structural matching optimization

The engine maintains indexes for fast structural matching:

- Policies indexed by tool name
- Policies indexed by event type

When an event arrives, the engine uses these indexes to find candidate policies in O(1) rather than scanning all policies.

### Expression caching

CEL expressions are compiled once when configuration loads. The compiled representations are reused for every evaluation. This avoids parsing overhead on the hot path.

### Parameter caching

Parameter providers can specify TTL for caching. The daemon caches resolved parameters and invalidates them based on TTL or explicit invalidation signals.

```yaml
parameters:
  - name: workflowContext
    from: specs
    ttl: 30s  # cache for 30 seconds
```

### Parallel evaluation

Within a priority level, policies without dependencies can evaluate in parallel. The engine uses a work-stealing scheduler to maximize throughput while respecting ordering constraints for mutations.

## Evaluation flow

The following diagram illustrates the end-to-end hook evaluation pipeline, from event receipt through policy evaluation to the final decision returned to Claude Code.

```plantuml
@startuml
!include https://raw.githubusercontent.com/plantuml-stdlib/C4-PlantUML/master/C4_Sequence.puml

title Hook Evaluation Pipeline

Person(claude, "Claude Code", "Fires hook events")
System(cli, "arci CLI", "Hook entry point (arci hook apply)")
System(engine, "Evaluation Engine", "Policy evaluation pipeline")
SystemDb(state, "State Store", "Session and project state")

Rel(claude, cli, "Hook event", "stdin JSON")
Rel(cli, engine, "Parsed event + loaded policies")
Rel(engine, engine, "1. Structural matching", "Filter by tool, event type, paths, branches")
Rel(engine, engine, "2. Condition evaluation", "CEL expressions, short-circuit on false")
Rel(engine, state, "3. Parameter resolution", "Resolve with caching and TTL")
Rel(state, engine, "Resolved parameters")
Rel(engine, engine, "4. Variable computation", "Derive from params + event context")
Rel(engine, engine, "5. Rule evaluation", "Validate and mutate per priority level, apply mutations between levels")
Rel(engine, state, "6. Effect execution", "setState, notify (post-decision, does not influence outcome)")
Rel(engine, engine, "7. Result aggregation", "deny > warn > audit > allow, collect messages and mutations")
Rel(engine, cli, "Aggregated policy response")
Rel(cli, claude, "Decision + mutations", "stdout JSON")

@enduml
```
