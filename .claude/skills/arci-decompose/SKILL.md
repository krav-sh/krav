---
name: arci-decompose
description: >-
  Decompose work into tasks by generating a task DAG from a module's
  requirements. Use when a module has approved requirements and needs
  an implementation plan broken into executable tasks.
---

# Decompose work into tasks

Generate a task DAG from requirements and module scope.

## Context

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
2. For each requirement (or group of related requirements), create TASK-* nodes with: `module`, `processPhase`, `taskType`, `status: "pending"`, `implements` edges to the requirements they satisfy, and `dependsOn` edges expressing ordering.
3. Follow the ISO/IEC/IEEE 15288 phase ordering: architecture → design → build → integration → verification → validation.
4. Tasks should be atomic enough for a single focused Claude Code session. If a task would take more than one session, break it down further.
5. Create verification tasks for requirements that need test cases.
6. Ensure `dependsOn` forms a DAG with no cycles.
