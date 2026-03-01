---
name: arci-feature
description: >-
  Build a feature end to end, from concept through verification. Use when
  the developer wants to design and implement a complete feature using the
  full transformation chain. This is a composite workflow that orchestrates
  other skills.
disable-model-invocation: true
stage-classification: temporary
replacement-stage: 3
replacement: "Production arci-feature skill backed by CLI commands and subagent orchestration"
---

# Build a feature end to end

Composite orchestrator that drives a feature from concept through verified code.

## Current state

!`jq -s '
  {
    modules: [.[] | select(."@type" == "Module") | {id: ."@id", title: .title, phase: .phase}],
    ready_tasks: [.[] | select(."@type" == "Task" and .status == "pending")] | length
  }
' .arci/graph.jsonlt 2>/dev/null || echo '{"error": "No graph data found."}'`

## Instructions

This skill drives the full transformation chain for a feature. Follow these phases in order, using the indicated skill for each step:

1. Explore the design space. Create a concept capturing the design thinking, alternatives, and rationale. (Follow the `arci-explore` workflow.)
2. Formalize into needs. Extract stakeholder expectations from the crystallized concept. (Follow the `arci-formalize` workflow.)
3. Derive requirements. Transform validated needs into verifiable obligations. (Follow the `arci-derive` workflow.)
4. Decompose into tasks. Generate a task DAG from the approved requirements. (Follow the `arci-decompose` workflow.)
5. Create test cases. Specify verification for the requirements. (Follow the `arci-testcase` workflow.)
6. Execute tasks. Work through the task DAG in dependency order. (Follow the `arci-task` workflow for each task.)
7. Run verification. Execute test cases and record results. (Follow the `arci-verify` workflow.)

At each phase, check with the developer before proceeding to the next. The developer may want to review intermediate results or adjust direction.

Not every feature needs every phase. If the developer says "skip formalization" or "just build it," respect their ceremony preference. The `arci-quickfix` skill handles the minimal-ceremony path.

## Graph-editing conventions

| Pattern | Classification | Stage | Replacement |
|---------|---------------|-------|-------------|
| Module and task count aggregation query | Temporary | 3 | `arci status` CLI command |
