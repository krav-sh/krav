---
name: arci-formalize
description: >-
  Formalize crystallized concepts into stakeholder needs. Use when a concept
  has been explored and crystallized and its stakeholder expectations should
  be captured as formal NEED-* nodes in the knowledge graph.
stage-classification: temporary
replacement-stage: 3
replacement: "Production arci-formalize skill backed by `arci need create` and `arci concept formalize` CLI commands"
---

# Formalize concepts into needs

Extract stakeholder expectations from crystallized concepts as formal needs.

## Candidate picker

If the developer did not provide a CON-* identifier, list crystallized concepts and ask them to pick one:

!`jq -s '[.[] | select(."@type" == "Concept" and .status == "crystallized") | {id: ."@id", title: .title, module: .module."@id"}]' .arci/graph.jsonlt 2>/dev/null || echo '[]'`

Present the candidates using the AskUserQuestion tool so the developer can select one. If the list is empty, there are no crystallized concepts ready for formalization. Suggest the developer explore or crystallize a concept first.

## Context

After identifying the CON-* identifier, load its context:

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
3. For each expectation found, draft a NEED-* node with: `derivesFrom` pointing to the concept, `stakeholder` pointing to the relevant STK-* nodes, a `statement` expressing the expectation from the stakeholder's perspective, and `status: "draft"`.
4. Write needs as stakeholder expectations, not solution descriptions. "The developer needs feedback within 500 ms" not "The system shall respond in 500 ms" (that's a requirement).
5. Before writing to the graph, run the review loop (see below).
6. Incorporate review feedback, then present the final needs to the developer for approval.
7. Write approved needs to `graph.jsonlt` and transition the concept to `"formalized"` once all expectations have landed.

## Review loop

After drafting the needs but before writing them to the graph, use the Agent tool to review them. Pass the agent the drafted needs, the concept node, the concept's prose file content, and the active stakeholders. Instruct the review agent:

"Review these drafted needs against the source concept. Check each of the following and report only problems found:

- Is each need written as a stakeholder expectation, not a solution description or disguised requirement?
- Does every need trace to something specific in the concept's prose? Flag any that seem invented rather than extracted.
- Are there expectations the concept implies that are missing from the draft?
- Do any drafted needs overlap enough to merge?
- Is the stakeholder assignment correct for each need? Would a different stakeholder be more appropriate?

Do not create any graph nodes, tasks, or defects. Return only your critique."
