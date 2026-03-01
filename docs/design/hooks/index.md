# Hooks

The hook subsystem is arci's policy engine for intercepting and controlling Claude Code tool execution. Hooks evaluate policies against Claude Code hook events — tool calls, prompts, session lifecycle, and more — to enforce safety rules, inject context, track state, and automate workflows.

Hooks are one component of arci's architecture. For the overall system design, see [architecture.md](../architecture.md). For how arci integrates with Claude Code's hook system, see [claude-code-integration.md](claude-code-integration.md).

## How it works

When Claude Code fires a hook event (e.g., before executing a Bash command), arci evaluates all matching policies and returns a response: allow, warn, or deny. Policies can also mutate tool inputs, inject context into the conversation, and trigger side effects like state tracking or notifications.

## Documentation

- [Hook schema](hook-schema.md) — Normalized event types, tool names, input/output schemas
- [Policy model](policy-model.md) — Policy structure, rules, validation, mutation, and effects
- [Match schema](match-schema.md) — Structural matching semantics (events, tools, paths, branches)
- [Execution model](execution-model.md) — Evaluation pipeline, priority cascading, result aggregation
- [Policy loading](policy-loading.md) — Configuration cascade, manifest merging, two-phase loading
- [Expressions](expressions.md) — CEL expressions and Go template syntax for policy logic
- [Starlark scripting](starlark-scripting.md) — Embedded scripting for complex macros and effects
- [GritQL](gritql.md) — Structural code analysis for syntax-aware policy matching
- [Builtins](builtins.md) — Curated built-in policies for common safety and workflow patterns
- [Recipes](recipes.md) — Practical policy examples for common use cases
- [Claude Code integration](claude-code-integration.md) — Claude Code hook events, input/output schemas, configuration

## Examples

The [examples/](examples/) directory contains complete YAML configurations:

- [validating-policy.yaml](examples/validating-policy.yaml) — Validation policies with CEL expressions
- [mutating-policy.yaml](examples/mutating-policy.yaml) — Mutation policies with GritQL transforms
- [parameterized-policies.yaml](examples/parameterized-policies.yaml) — Policies with external parameters
- [policy-bindings.yaml](examples/policy-bindings.yaml) — Binding policies to scopes and params
- [param-providers.yaml](examples/param-providers.yaml) — HTTP, file, and environment param providers
- [param-resources.yaml](examples/param-resources.yaml) — Static parameter resources
- [specs-integration.yaml](examples/specs-integration.yaml) — Specs subsystem integration
