---
name: arci-formalize
description: >-
  Formalize crystallized concepts into stakeholder needs. Use when a concept
  has been explored and crystallized and its stakeholder expectations should
  be captured as formal NEED-* nodes in the knowledge graph.
---

# Formalize concepts into needs

Extract stakeholder expectations from crystallized concepts as formal needs.

## Context

!`CONCEPT_ID="$1"; jq -s --arg id "$CONCEPT_ID" '
  (.[] | select(."@id" == $id)) as $con |
  [.[] | select(."@type" == "Stakeholder" and .status == "active")] as $stk |
  {
    concept: $con,
    stakeholders: [$stk[] | {id: ."@id", title: .title, concerns: .concerns}]
  }
' .arci/graph.jsonlt 2>/dev/null || echo '{"error": "Provide a CON-* identifier."}'`

## Instructions

1. Read the concept's prose file for the full exploration content.
2. For each active stakeholder, consider: does this concept imply an expectation from their perspective?
3. For each expectation found, create a NEED-* node with: `derivesFrom` pointing to the concept, `stakeholder` pointing to the relevant STK-* nodes, a `statement` expressing the expectation from the stakeholder's perspective, and `status: "draft"`.
4. Write needs as stakeholder expectations, not solution descriptions. "The developer needs feedback within 500 ms" not "The system shall respond in 500 ms" (that's a requirement).
5. After extraction, transition the concept to `"formalized"` once all expectations have landed.
