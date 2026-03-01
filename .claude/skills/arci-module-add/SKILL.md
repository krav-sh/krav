---
name: arci-module-add
description: >-
  Add a new module to the project. Use when introducing a new subsystem or
  component to the module hierarchy.
---

# Add a module

Create a new module node and establish its place in the hierarchy.

## Current hierarchy

!`jq -s '[.[] | select(."@type" == "Module") | {id: ."@id", title: .title, phase: .phase, parent: .childOf."@id"}]' .arci/graph.jsonlt 2>/dev/null || echo '[]'`

## Instructions

1. Determine the parent module. Every module except the root has a parent via `childOf`.
2. Create a MOD-* node with: `childOf` pointing to the parent, `phase: "architecture"` (new modules start in architecture phase), and `status: "active"`.
3. If the parent has requirements that should flow down, note them for later allocation (use the `arci-allocate` skill).
4. Consider what stakeholder concerns this module addresses and whether it warrants new needs.
