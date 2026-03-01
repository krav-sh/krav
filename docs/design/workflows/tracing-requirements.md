# Tracing requirements

## What

The developer says `why does this requirement exist?`, `what's the traceability for the parser?`, `trace REQ-X back to its source`, or `what needs does this requirement satisfy?` The agent walks the derivation chain, from requirements back to needs back to concepts, or forward from needs to requirements to test cases, and presents the full story.

## Why

Traceability answers the "why" question that specifications alone can't. "The parser shall report errors within 50 ms" is a requirement. "Because users need fast feedback when their config file has a syntax error, which comes from the concept that the tool should feel responsive and never leave users wondering if it's hung" is the traceability story. This story matters when deciding whether to change, relax, or remove a requirement. Without traceability, changing a requirement is a gamble: you don't know what stakeholder expectation you might be violating.

Forward traceability matters too. "What happens if this need changes?" requires knowing which requirements derive from it, which test cases verify those requirements, and which tasks address them. Suspect links help here, but a full trace shows the complete impact chain.

## What happens in the graph

Nothing. The agent traverses graph edges and presents the chain.

Backward trace (from requirement to source): follow `derivesFrom` edges from REQ-* to NEED-* to CON-*. At each hop, load the node's statement, rationale, and status.

Forward trace (from need to code): follow `derivesFrom` edges forward from NEED-* to REQ-*, then `verifiedBy` edges to TC-*, and check for TASK-* nodes that address each requirement.

Impact analysis (what does changing this node affect): follow all outgoing edges from the node, recursively. This is the suspect link propagation path, but presented as a preview rather than triggered by an actual change.

Key CLI commands: `krav reqtrace REQ-X` (full chain), `krav taskancestors TASK-X`, `krav taskdescendants TASK-X`.

## Trigger patterns

`Why does REQ-X exist?`, `trace this requirement`, `what's the traceability for module X?`, `what would change if NEED-Y changed?`, `show the derivation chain`.

## Graph before

A populated graph with traceability edges.

## Graph after

Unchanged.

## Agent interaction layer

### Skills

The `krav:trace` skill builds this workflow. Preprocessing loads the target node and its immediate neighbors to give the agent a starting point, but the real work happens through instructed commands that walk the derivation chain in either direction. The skill's CLI commands (`krav reqtrace`, `krav taskancestors`, `krav taskdescendants`) provide the raw traversal data, and the skill body tells the agent how to synthesize chain data into a narrative explanation of why a node exists and what depends on it.

The skill is lightweight in terms of instructions. The challenge isn't guiding a complex transformation but presenting graph traversal results in readable form, especially for chains that cross module boundaries or branch at allocation points.

### Agents

The `krav-analyst` subagent handles deep or complex traces that would consume too much of the main agent's context window. A full impact analysis from a need through all derived requirements, their test cases, and dependent tasks can involve dozens of nodes. For simple traces ("why does this requirement exist?"), the main agent handles it directly. For full impact analysis ("what would change if this need changed?"), delegating to the analyst keeps the main context clean.

### Policies

The `prompt-context` policy is the most relevant here. When the developer mentions a node ID or module name, the hook injects that node's current state, giving the agent a starting point for the trace without an explicit query. The `session-context` policy provides module context at session start.

No enforcement or mutation policies apply since tracing only reads data.

### Task types

This workflow doesn't create or execute tasks. It provides the information needed to make decisions about tasks, requirements, and other graph content.

## Open questions

**Visualization.** Traceability chains are naturally tree-shaped (or DAG-shaped if requirements derive from multiple needs). Text output works for short chains but gets unwieldy for complex graphs. Should the agent produce a diagram (PlantUML, Mermaid) for complex traces? Should there be a dashboard view?

**Broken traces.** What if the chain has gaps? A requirement with no `derivesFrom` edge to a need is an orphan: it exists but nobody knows why. The agent should flag these during tracing. But should tracing be an audit tool (find all broken traces) or a per-node tool (trace this specific item)?

**Depth control.** "Trace REQ-X" might mean "show the immediate source" or "show the full chain back to the original concept." The agent needs to handle both, defaulting to the full chain but allowing scoped queries.

**Cross-module traces.** When a system-level requirement flows down via `allocatesTo` to child modules, the trace crosses module boundaries. The agent should present this with explicit boundary markers: "REQ-X on the root module allocates to REQ-Y on the parser and REQ-Z on the CLI, each derived from NEED-A."
