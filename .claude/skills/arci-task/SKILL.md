---
name: arci-task
description: >-
  Work on a specific task from the knowledge graph. Use when starting work on
  a task, resuming a task, or when told to work on a specific TASK-* identifier.
stage-classification: temporary
replacement-stage: 3
replacement: "Production arci-task skill backed by `arci task` CLI commands"
---

# Work on a task

Execute a specific task in a focused session, guided by the task's requirements, context, and type-specific instructions.

## Candidate picker

If the developer did not provide a TASK-* identifier, list tasks that are ready to work on (all dependencies complete) and ask the developer to pick one:

!`jq -s '
  [.[] | select(."@type" == "Task" and .status == "complete") | ."@id"] as $done |
  [.[] | select(."@type" == "Task" and (.status == "pending" or .status == "blocked")) |
    select((.dependsOn // []) | map(."@id") | all(. as $d | $done | index($d) != null)) |
    {id: ."@id", title: .title, processPhase: .processPhase, module: .module."@id"}]
' .arci/graph.jsonlt 2>/dev/null || echo '[]'`

Present the candidates using the AskUserQuestion tool so the developer can select one. If the list is empty, no tasks are ready (either all are complete, or remaining tasks have incomplete dependencies).

## Task context

After identifying the TASK-* identifier, load its full context:

!`TASK_ID="$1"; jq -s --arg id "$TASK_ID" '
  ($id) as $tid |
  (.[] | select(."@id" == $tid)) as $task |
  [.[] | select(."@type" == "Requirement") | select(."@id" as $rid | $task.implements // [] | map(."@id") | index($rid))] as $reqs |
  (.[] | select(."@id" == ($task.module."@id" // ""))) as $mod |
  [.[] | select(."@type" == "Task" and .status == "complete") | select(."@id" as $did | $task.dependsOn // [] | map(."@id") | index($did))] as $deps |
  [$reqs[].derivesFrom[]."@id" // empty] as $need_ids |
  [.[] | select(."@id" == ($need_ids[] // empty))] as $needs |
  [$needs[].derivesFrom[]."@id" // empty] as $con_ids |
  [.[] | select(."@id" == ($con_ids[] // empty))] as $concepts |
  {
    task: $task,
    module: {id: $mod."@id", title: $mod.title, phase: $mod.phase},
    requirements: [$reqs[] | {id: ."@id", title: .title, statement: .statement}],
    originating_needs: [$needs[] | {id: ."@id", title: .title, statement: .statement}],
    originating_concepts: [$concepts[] | {id: ."@id", title: .title, conceptType: .conceptType}],
    completed_dependencies: [$deps[] | {id: ."@id", title: .title, deliverables: .deliverables}]
  }
' .arci/graph.jsonlt 2>/dev/null || echo '{"error": "Could not load task context. Provide a valid TASK-* identifier."}'`

## Exploration phase

Before starting work, use the Agent tool to run parallel exploration agents that investigate the problem space. Launch these agents concurrently:

**Design context agent**: "Read the design docs relevant to this task's module and requirements. Identify the specific design decisions, constraints, and patterns that apply. Report what the code must conform to and any open questions in the design." Pass it the task context, module ID, and the list of originating concepts so it knows which design docs to read.

**Codebase context agent**: "Examine the existing code in this task's module scope. Identify relevant files, existing patterns, and interfaces the new code must conform to. Report the current state and what needs to integrate with existing code." Pass it the task context and module ID.

**Dependency context agent**: "Review the deliverables from this task's completed dependencies. Identify what they produced, what interfaces or artifacts are now available, and any constraints or conventions they established that this task should follow. Report what's available to build on." Pass it the completed dependencies list. Skip this agent if there are no completed dependencies.

Synthesize the findings from all exploration agents into a brief plan before proceeding. If the exploration surfaces conflicts between design docs and existing code, or gaps in the requirements, flag them to the developer before starting work.

## Instructions

You are working on the task described in the preceding context, informed by the exploration findings.

1. Present the plan from the exploration phase to the developer for approval before writing code.
2. Do the work, following the project's code organization and conventions.
3. When done, run the review loop (see below) on your deliverables.
4. Incorporate review feedback, then update the task in `.arci/graph.jsonlt`: set `status` to `"complete"`, add `completed` timestamp, and record `deliverables`.
5. Commit with a message referencing the task ID.

If you get blocked, update the task status to `"blocked"` and note the reason in the `summary` field.

## Review loop

After finishing the task but before marking it complete, use the Agent tool to review the deliverables. Pass the agent the task's requirements, the plan, and the files you created or modified. Instruct the review agent:

"Review these deliverables against the task's requirements and plan. Check each of the following and report only problems found:

- Does the code satisfy every requirement statement the task links to, or are there gaps?
- Does the code follow the patterns and conventions identified during the exploration phase?
- Are there design doc constraints that the code violates or doesn't account for?
- Does the code integrate correctly with the interfaces and artifacts from completed dependencies?
- Are there obvious defects, edge cases, or missing error handling?

Do not create any graph nodes, tasks, or defects. Return only your critique."

## Graph-editing conventions

| Pattern | Classification | Stage | Replacement |
|---------|---------------|-------|-------------|
| Candidate picker: ready tasks with dependency closure | Temporary | 3 | `arci task list --status ready` CLI command |
| Task context with full derivation chain query | Temporary | 3 | `arci task show --context` CLI command |
