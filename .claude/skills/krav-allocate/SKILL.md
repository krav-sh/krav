---
name: krav-allocate
description: >-
  Allocate parent module requirements to child modules. Use when flowing
  down requirements from a parent module to its children, optionally with
  budgets or partitions.
stage-classification: temporary
replacement-stage: 1
replacement: "`krav req allocate` CLI command"
---

# Allocate requirements to child modules

Flow parent requirements down to child modules via allocatesTo edges.

## Context

!`MOD_ID="$1"; jq -s --arg id "$MOD_ID" '
  (.[] | select(."@id" == $id)) as $mod |
  [.[] | select(."@type" == "Module" and .childOf."@id" == $id)] as $children |
  [.[] | select(."@type" == "Requirement" and .module."@id" == $id)] as $reqs |
  {
    module: $mod,
    children: [$children[] | {id: ."@id", title: .title, phase: .phase}],
    requirements: [$reqs[] | {id: ."@id", title: .title, statement: .statement, allocatesTo: .allocatesTo}]
  }
' .krav/graph.jsonlt 2>/dev/null || echo '{"error": "Provide a MOD-* identifier."}'`

## Instructions

1. Review the parent module's requirements and child module structure.
2. For each requirement that child modules must satisfy, add `allocatesTo` edges with the target module's `@id`. Include budget metadata where appropriate, like `{"@id": "MOD-CHILD01", "budget": "50ms"}`.
3. Not every requirement flows down. Some apply at the parent level only. Allocate only requirements that children are responsible for.
4. After allocation, child modules should derive their own requirements from the allocated parent requirements.

## Graph-editing conventions

| Pattern | Classification | Stage | Replacement |
|---------|---------------|-------|-------------|
| Module hierarchy and requirement context query | Temporary | 1 | `krav req allocate` CLI command |
