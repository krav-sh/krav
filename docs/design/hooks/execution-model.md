# Execution model

This document describes how ARCI evaluates policies when a hook event arrives. It covers the evaluation pipeline, priority cascading, result aggregation, and the data flow through the system.

## Overview

When Claude Code fires a hook (such as before executing a Bash command), ARCI receives the event and evaluates all matching policies to produce a response. The response determines whether the tool call proceeds, any mutations to apply, warnings to surface, and side effects to execute.

The evaluation pipeline has these stages: structural matching filters policies quickly using indexable criteria, condition evaluation applies CEL expressions with runtime context, rule evaluation runs the actual validations and mutations, effect execution handles side actions, and result aggregation combines outcomes from all matching policies into a single response.

## Enforcement state handling

Before the evaluation pipeline runs, the engine filters policies by their enforcement state. Each policy exists in one of three states: enabled, turned off, or audit. The policy cascade during loading determines the enforcement state, as described in [policy-loading.md](policy-loading.md).

The engine evaluates enabled policies normally. They take part fully in structural matching, condition evaluation, validation, mutation, and effects. Their actions have full force: deny blocks tool calls, the engine applies mutations, and it executes effects.

The engine skips turned-off policies entirely. It does not match, check, or log them. They don't appear in the audit trail. Turning off a policy is the same as removing it from the configuration, except that it remains visible in diagnostic commands.

The engine evaluates audit policies in a "dry-run" mode. It matches and evaluates the policy normally, but downgrades its actions. Deny actions become warnings (the tool call proceeds with a warning). The engine computes mutations but does not apply them to the event (the original event continues through the pipeline). It logs effects but does not execute them. Audit results appear in the audit trail with an `enforcement: audit` marker.

The enforcement state filter happens before Stage 1 (structural matching). The engine removes turned-off policies from consideration before any evaluation begins. Audit policies proceed through evaluation with their state tracked for action downgrading.

### Terminology note

This document uses "audit" for two distinct concepts. The validation action `audit` (appearing in `validate.action: audit`) describes what happens when a specific rule's validation fails: it logs silently without user notification. The enforcement state `audit` describes how the engine evaluates an entire policy: in dry-run mode with all actions downgraded. Context disambiguates since validation actions apply per-rule while enforcement states apply per-policy.

## Evaluation pipeline

### Stage 1: Structural matching

The first stage filters policies using structural criteria that the engine can index and check without CEL. This is fast; the engine can scan hundreds of policies in microseconds.

For each policy, the engine checks whether its `match` block matches the current event:

```text
event.type ∈ policy.match.events (or events is empty)
AND event.tool ∈ policy.match.tools (or tools is empty)
AND event.path matches policy.match.paths (include/exclude)
AND event.branch matches policy.match.branches (include/exclude)
```

The engine eliminates policies that don't match structurally. No CEL evaluation happens for these policies, and they don't appear in audit logs.

### Stage 2: Condition evaluation

For policies that pass structural matching, the engine evaluates their `conditions` in declaration order. Conditions are CEL expressions that must all return true.

```yaml
conditions:
  - name: not-safe-mode
    expression: '!session.safe_mode'
  - name: on-protected-branch
    expression: '$current_branch() in params.protectedBranches'
```

Evaluation short-circuits: if the first condition returns false, the engine does not check later conditions. This is both an optimization and a way to guard expensive expressions behind cheap checks.

If any condition returns false, the engine skips the policy for this event. Skipped policies don't contribute to the final decision and appear in audit logs only if the user enables verbose logging.

### Stage 3: Parameter resolution

Before evaluating rules, the engine resolves all parameters declared in the policy. Parameters can come from static values, named providers, inline providers, or environment variables.

```yaml
parameters:
  - name: blockedCommands
    from: specs
    defaults: []
```

Parameter resolution may involve I/O (file reads, HTTP calls) and the engine caches results where possible. If resolution fails and defaults do not exist, behavior depends on `config.failurePolicy`:

- `allow` (default): The engine skips the policy, tool call proceeds
- `deny`: The policy errors, tool call blocks

The engine makes resolved parameters available in all later expressions as `params.paramName`.

### Stage 4: Variable computation

The engine computes policy-level variables in declaration order. Variables can reference parameters, built-in functions, and earlier variables.

```yaml
variables:
  - name: command
    expression: 'tool_input.command ?? ""'
  - name: is_blocked
    expression: 'params.blockedCommands.exists(b, command.contains(b))'
```

The engine evaluates variables once per policy and caches them for use by all rules in that policy.

### Stage 5: Rule evaluation

For each rule in the policy, the engine:

1. Checks structural match (rule's `match` intersected with policy's `match`)
2. Evaluates rule conditions
3. Computes rule-local variables
4. Runs validation or mutation
5. Queues effects for later execution

The engine evaluates rules in declaration order within a policy.

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

Validation rules produce a result (pass or fail with action). Mutation rules produce a transformation to apply. The engine collects effects but does not execute them until stage 6.

### Stage 6: Effect execution

After the engine evaluates all rules, it executes queued effects. Effects run after the engine decides on admission, so they can't influence whether the tool call proceeds.

Each effect has a `when` condition:

- `always` (default): Effect runs regardless of rule outcome
- `on_pass`: Effect runs only if the rule passed (or the engine skipped it)
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

The engine collects effects from all matching rules across all matching policies and executes them. Effect execution failures appear in logs but don't affect the tool call decision.

## Priority cascading

The engine groups policies by priority level: critical, high, medium, low. Within each level, policies follow config cascade order: universal, then project-specific.

```text
critical/universal → critical/project
high/universal → high/project
medium/universal → medium/project
low/universal → low/project
```

### Mutation visibility

The engine applies mutations from higher priority levels before lower priority levels run. Higher-priority policies can then transform requests in ways that affect lower-priority policy evaluation.

```text
1. Evaluate all critical-priority policies
2. Apply mutations from critical policies
3. Evaluate all high-priority policies (seeing mutated state)
4. Apply mutations from high policies
5. Continue through medium and low
```

Within a priority level, the engine applies mutations in config cascade order. If two policies at the same priority and cascade level mutate the same field, last-write-wins with a warning in the audit trail.

### Validation aggregation

The engine collects validation results from all priority levels. The most restrictive action wins:

```text
deny > warn > audit > allow
```

If any rule from any policy produces a `deny`, the result is deny. If no denies but some warns, the result is warn. Messages from all failures accumulate in the response.

## Data flow

The data flow through evaluation is unidirectional:

```text
parameters → variables → conditions → validate/mutate → effects → state
```

The engine resolves parameters first and they remain immutable during evaluation. Variables derive from parameters and the hook event. Conditions and validation expressions read variables but don't change them. Effects run last and can persist state for future evaluations.

This unidirectional flow makes evaluation predictable. A variable can't depend on an effect's outcome, and effects can't change variables that validations read.

## Result aggregation

After the engine evaluates all policies, it aggregates results into a single response:

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

The most restrictive action from any failing validation determines the final action:

1. If any validation failed with `action: deny`, the action is `deny`
2. Else if any validation failed with `action: warn`, the action is `warn`
3. Else the action is `allow`

### Message collection

The engine collects messages from all failing validations and categorizes them:

- `user` messages appear to the human operator
- `assistant` messages feed into Claude's context

A validation can specify which audience receives its message, or it can go to both by default.

### Mutation composition

Mutations accumulate across all matching policies. Later mutations (lower priority or later in cascade) see and can override earlier mutations.

The final mutated event is what Claude receives if the action is `allow` or `warn`. For `deny`, the engine still records mutations in the audit trail but does not apply them.

### Audit trail

The engine records every policy evaluation in the audit trail, regardless of outcome:

- Which policies matched structurally
- Which policies passed conditions
- Which rules fired
- What validations passed or failed
- What mutations the engine applied
- What effects executed

The audit trail is available through the dashboard, API, and log output.

## Fail-open semantics

ARCI operates as a guardrail, not a gate. Errors in the policy system should not block Claude from operating. Only explicit deny decisions block tool calls.

### Error handling

When errors occur during evaluation:

- Parameter resolution failure: The engine skips the policy (with `failurePolicy: allow`) or denies it (with `failurePolicy: deny`)
- Variable computation error: The engine skips the rule with warning
- Condition evaluation error: The engine treats the condition as false, skips the policy
- Validation expression error: The engine treats the validation as passed with warning
- Mutation expression error: The engine skips the mutation with warning
- Effect execution error: Logged, does not affect tool call

The principle is: uncertainty defaults to allowing the operation. Users must explicitly configure denial to block tool calls.

### Server unavailability

When the ARCI command-line tool can't reach the server:

- `server.on_unavailable: fallback`: the command-line tool falls back to direct execution
- `server.on_unavailable: start`: the command-line tool attempts to spawn the server
- `server.on_unavailable: fail`: the command-line tool fails (still doesn't block Claude; the assistant's hook system handles command-line tool failures)

Even if ARCI fails entirely, Claude continues operating. The assistant's native hook system treats hook script failures as non-blocking by default.

## Performance considerations

### Structural matching optimization

The engine maintains indexes for fast structural matching:

- Policies indexed by tool name
- Policies indexed by event type

When an event arrives, the engine uses these indexes to find candidate policies in O(1) rather than scanning all policies.

### Expression caching

The engine compiles CEL expressions once when configuration loads. It reuses the compiled representations for every evaluation. This avoids parsing overhead on the hot path.

### Parameter caching

Parameter providers can specify TTL for caching. The server caches resolved parameters and invalidates them based on TTL or explicit invalidation signals.

```yaml
parameters:
  - name: workflowContext
    from: specs
    ttl: 30s  # cache for 30 seconds
```

### Parallel evaluation

Within a priority level, policies without dependencies can run in parallel. The engine uses a work-stealing scheduler to increase throughput while respecting ordering constraints for mutations.

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
