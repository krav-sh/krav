# Hook-based policy enforcement

arci intercepts Claude Code tool execution through declarative policies evaluated via hooks. Policies define what the agent can and cannot do, injecting graph context and enforcing development discipline without requiring the agent to remember rules on its own.

## The evaluation pipeline

Policy evaluation follows a six-stage pipeline that progressively narrows the set of applicable policies and computes a decision:

1. **Structural match** checks whether this policy applies to this event type (PreToolUse, PostToolUse, Stop, etc.).
2. **Conditions** run CEL expressions against the event payload.
3. **Parameters** resolve values from providers, environment, and inline definitions.
4. **Variables** compute derived values from parameters and event context.
5. **Rules** run rule expressions and produce per-rule decisions.
6. **Effects** execute side effects (state mutations, context injection, mutations).

Data flows forward through the pipeline with no cycles. Each stage's output feeds the next stage's input.

## Admission decisions

Policies produce one of three decisions:

- **Allow** lets the operation proceed. This is the default when no policies match or when all matching policies allow.
- **Deny** blocks the operation. The denial includes a `reason` that Claude Code surfaces to the agent, steering it toward the correct action.
- **Warn** lets the operation proceed but surfaces a warning. Used for advisory policies that flag concerns without blocking.

When policies match the same event, **deny-wins aggregation** applies: if any policy denies, the result is deny regardless of other policies. Security-critical policies can't be overridden by permissive ones.

## Fail-open semantics

Every error path in policy evaluation results in allow. Malformed policy files get skipped. CEL expression errors cause the condition to return false (policy doesn't match). Failed parameter resolution skips the policy. The system never blocks Claude Code because of arci's own errors.

This is a non-negotiable design constraint. A development tool that unpredictably blocks the agent is worse than no tool at all.

## Graph enforcement through hooks

The hook system enables these enforcement patterns:

- **Baseline protection** uses PreToolUse hooks to deny writes to files owned by baselined modules.
- **Task completion gates** use Stop hooks to block session end when the current task has no recorded deliverables.
- **Context injection** uses SessionStart and PostToolUse hooks to inject graph state into the agent's conversation.
- **Steering** provides the denial reason when a hook blocks an action, telling the agent what to do instead ("this module is baselined; create a defect or run `arci moduleunlock` with justification").
