---
name: arci-task
description: >-
  Work on a specific task from the knowledge graph. Use when starting work on
  a task, resuming a task, or when told to work on a specific TASK-* identifier.
---

# Work on a task

Execute a specific task in a focused session, guided by the task's requirements, context, and type-specific instructions.

## Task context

!`TASK_ID="$1"; jq -s --arg id "$TASK_ID" '
  ($id) as $tid |
  (.[] | select(."@id" == $tid)) as $task |
  [.[] | select(."@type" == "Requirement") | select(."@id" as $rid | $task.implements // [] | map(."@id") | index($rid))] as $reqs |
  (.[] | select(."@id" == ($task.module."@id" // ""))) as $mod |
  [.[] | select(."@type" == "Task" and .status == "complete") | select(."@id" as $did | $task.dependsOn // [] | map(."@id") | index($did))] as $deps |
  {
    task: $task,
    module: {id: $mod."@id", title: $mod.title, phase: $mod.phase},
    requirements: [$reqs[] | {id: ."@id", title: .title, statement: .statement}],
    completed_dependencies: [$deps[] | {id: ."@id", title: .title, deliverables: .deliverables}]
  }
' .arci/graph.jsonlt 2>/dev/null || echo '{"error": "Could not load task context. Provide a valid TASK-* identifier."}'`

## Instructions

You are working on the task described in the preceding context. Follow these steps:

1. Read the task's requirements and understand what needs to happen.
2. Check the completed dependencies for relevant deliverables and context.
3. Read the relevant design doc if this task covers a design-documented subsystem.
4. Implement the work, following the project's code organization and conventions.
5. When done, update the task in `.arci/graph.jsonlt`: set `status` to `"complete"`, add `completed` timestamp, and record `deliverables`.
6. Commit with a message referencing the task ID.

If you get blocked, update the task status to `"blocked"` and note the reason in the `summary` field.
