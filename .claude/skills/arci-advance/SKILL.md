---
name: arci-advance
description: >-
  Advance a module to its next lifecycle phase. Use when checking whether
  phase advancement criteria are met and, if so, advancing the module.
stage-classification: temporary
replacement-stage: 1
replacement: "`arci module advance` CLI command"
---

# Advance a module's phase

Check criteria and move a module to its next lifecycle phase.

## Module state

!`MOD_ID="$1"; jq -s --arg id "$MOD_ID" '
  (.[] | select(."@id" == $id)) as $mod |
  [.[] | select(."@type" == "Task" and .module."@id" == $id)] as $tasks |
  [.[] | select(."@type" == "Defect" and .module."@id" == $id and (.status == "open" or .status == "confirmed") and (.severity == "critical" or .severity == "major"))] as $blockers |
  [.[] | select(."@type" == "TestCase" and .module."@id" == $id)] as $tcs |
  {
    module: {id: $mod."@id", title: $mod.title, phase: $mod.phase},
    tasks_by_status: ($tasks | group_by(.status) | map({(.[0].status): length}) | add),
    blocking_defects: [$blockers[] | {id: ."@id", title: .title, severity: .severity}],
    test_cases_by_result: ($tcs | group_by(.currentResult) | map({(.[0].currentResult // "unknown"): length}) | add)
  }
' .arci/graph.jsonlt 2>/dev/null || echo '{"error": "Provide a MOD-* identifier."}'`

## Instructions

The phase progression is: architecture → design → build → integration → verification → validation.

Check the criteria for the current phase:

Each phase requires that tasks for that phase are complete, no blocking defects (critical/major, open/confirmed) exist, and verification coverage meets the threshold for the phase.

If criteria pass, update the module's `phase` to the next value. If not, report what blocks advancement and what the team must do.

## Graph-editing conventions

| Pattern | Classification | Stage | Replacement |
|---------|---------------|-------|-------------|
| Module state context query (tasks, blockers, test results) | Temporary | 1 | `arci module show` CLI command |
