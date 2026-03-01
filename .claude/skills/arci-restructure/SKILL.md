---
name: arci-restructure
description: >-
  Restructure the module hierarchy by reparenting, splitting, or merging
  modules. Use when the architectural decomposition needs to change.
disable-model-invocation: true
stage-classification: temporary
replacement-stage: 1
replacement: "`arci module` CLI subcommands (reparent, split, merge)"
---

# Restructure modules

Reparent, split, or merge modules in the hierarchy.

## Current hierarchy

!`jq -s '[.[] | select(."@type" == "Module") | {id: ."@id", title: .title, phase: .phase, parent: .childOf."@id"}]' .arci/graph.jsonlt 2>/dev/null || echo '[]'`

## Instructions

Module restructuring has broad graph impact. Proceed carefully:

1. Identify what needs to change: reparenting a module under a new parent, splitting one module into two, or merging two modules into one.
2. For reparenting: update the `childOf` edge. Check whether allocated requirements from the old parent still apply.
3. For splitting: create the new modules, reassign requirements and tasks, update all `module` edges on affected nodes.
4. For merging: pick the surviving module, move all nodes from the absorbed module, update hierarchy.
5. After any restructuring, expect suspect links on edges involving moved nodes. Triage them.
6. Check whether any baselines cover the affected modules. Restructuring baselined modules is a major change.

## Graph-editing conventions

| Pattern | Classification | Stage | Replacement |
|---------|---------------|-------|-------------|
| Module hierarchy context query | Temporary | 1 | `arci module list` CLI command |
