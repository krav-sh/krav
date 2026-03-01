---
name: arci-derive
description: >-
  Derive verifiable requirements from validated needs. Use when needs have
  been validated and should be transformed into formal REQ-* obligations
  in the knowledge graph.
---

# Derive requirements from needs

Transform validated needs into verifiable design obligations.

## Context

!`NEED_ID="$1"; jq -s --arg id "$NEED_ID" '
  (.[] | select(."@id" == $id)) as $need |
  (.[] | select(."@id" == ($need.module."@id" // ""))) as $mod |
  [.[] | select(."@type" == "Requirement" and .module."@id" == $mod."@id")] as $existing |
  {
    need: $need,
    module: {id: $mod."@id", title: $mod.title},
    existing_requirements: [$existing[] | {id: ."@id", title: .title, statement: .statement}]
  }
' .arci/graph.jsonlt 2>/dev/null || echo '{"error": "Provide a NEED-* identifier."}'`

## Instructions

1. Read the need's statement and rationale.
2. Derive one or more requirements that, if satisfied, would address the need. Each requirement should be more specific and constrained than the need.
3. For each requirement, create a REQ-* node with: `derivesFrom` pointing to the need, `module` matching the need's module, a `statement` using "shall" language ("The system shall"), `priority`, and `status: "draft"`.
4. Check against existing requirements in the module to avoid duplication.
5. Each requirement must be verifiable. If you can't describe how to verify it, it's not specific enough.
