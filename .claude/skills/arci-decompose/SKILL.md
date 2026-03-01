---
name: arci-decompose
description: >-
  Decompose work into tasks by generating a task DAG from a module's
  requirements. Use when a module has approved requirements and needs
  an implementation plan broken into executable tasks.
stage-classification: temporary
replacement-stage: 3
replacement: "Production arci-decompose skill backed by `arci module decompose` CLI command"
---

# Decompose work into tasks

Generate a task DAG from requirements and module scope.

## Candidate picker

If the developer did not provide a MOD-* identifier, list modules that have approved requirements and ask them to pick one:

!`jq -s '
  [.[] | select(."@type" == "Requirement" and .status == "approved")] as $approved |
  [$approved[].module."@id"] | unique as $mod_ids |
  [.[] | select(."@id" == ($mod_ids[] // empty)) | {id: ."@id", title: .title, phase: .phase, approved_count: ([$approved[] | select(.module."@id" == ."@id")] | length)}]
' .arci/graph.jsonlt 2>/dev/null || echo '[]'`

Present the candidates using the AskUserQuestion tool so the developer can select one. If the list is empty, no modules have approved requirements ready for decomposition.

## Context

After identifying the MOD-* identifier, load its context:

!`MOD_ID="$1"; jq -s --arg id "$MOD_ID" '
  (.[] | select(."@id" == $id)) as $mod |
  [.[] | select(."@type" == "Requirement" and .module."@id" == $id and .status == "approved")] as $reqs |
  [.[] | select(."@type" == "Task" and .module."@id" == $id)] as $existing |
  {
    module: {id: $mod."@id", title: $mod.title, phase: $mod.phase},
    approved_requirements: [$reqs[] | {id: ."@id", title: .title, statement: .statement, priority: .priority}],
    existing_tasks: [$existing[] | {id: ."@id", title: .title, status: .status, processPhase: .processPhase}]
  }
' .arci/graph.jsonlt 2>/dev/null || echo '{"error": "Provide a MOD-* identifier."}'`

## Instructions

1. Review the module's approved requirements and any existing tasks.
2. For each requirement (or group of related requirements), draft TASK-* nodes with: `module`, `processPhase`, `taskType`, `status: "pending"`, `implements` edges to the requirements they satisfy, and `dependsOn` edges expressing ordering.
3. Follow the ISO/IEC/IEEE 15288 phase ordering: architecture → design → build → integration → verification → validation.
4. Tasks should be atomic enough for a single focused Claude Code session. If a task would take more than one session, break it down further.
5. Create verification tasks for requirements that have test cases.
6. Ensure `dependsOn` forms a DAG with no cycles.
7. Before writing to the graph, run the review loop (see below).
8. Incorporate review feedback, then present the final task DAG to the developer for approval.
9. Write approved tasks to `graph.jsonlt`.
10. For each task, create a prose file at `.arci/tasks/{timestamp}-{NANOID}-{slug}.md`. Include the task's scope, approach, relevant design context, key decisions or constraints from the requirements it covers, and any notes on integration with dependency deliverables. The prose file gives the agent working the task a richer starting point than the graph node's summary field alone.

## Review loop

After drafting the task DAG but before writing it to the graph, use the Agent tool to review it. Pass the agent the drafted tasks with their dependency edges, the approved requirements they cover, and any existing tasks in the module. Instruct the review agent:

"Review this drafted task DAG against the requirements it addresses. Check each of the following and report only problems found:

- Does every approved requirement have at least one task that covers it, or are there requirements with no coverage in the DAG?
- Are the dependency edges correct? Flag any task that depends on something it doesn't actually need to wait for, or any task that's missing a dependency it needs (like two tasks that modify the same file, which should run in sequence).
- Is each task atomic enough for a single Claude Code session, or are there tasks that need further breakdown?
- Are there tasks that are so small they should merge with an adjacent task?
- Does the DAG include verification tasks for requirements that have test cases?
- Is the phase ordering correct? Flag any task whose processPhase violates the architecture → design → build → integration → verification → validation sequence relative to its dependencies.

Do not create any graph nodes, tasks, or defects. Return only your critique."

## Graph-editing conventions

| Pattern | Classification | Stage | Replacement |
|---------|---------------|-------|-------------|
| Candidate picker: modules with approved requirements | Temporary | 3 | `arci module list --has-approved-reqs` CLI command |
| Module requirements and task inventory context query | Temporary | 3 | `arci module decompose` CLI command |
