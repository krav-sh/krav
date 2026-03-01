---
name: arci-derive
description: >-
  Derive verifiable requirements from validated needs. Use when needs have
  been validated and should be transformed into formal REQ-* obligations
  in the knowledge graph.
stage-classification: temporary
replacement-stage: 3
replacement: "Production arci-derive skill backed by `arci req create` and `arci need derive` CLI commands"
---

# Derive requirements from needs

Transform validated needs into verifiable design obligations.

## Candidate picker

If the developer did not provide a NEED-* identifier, list validated needs and ask the developer to pick one:

!`jq -s '[.[] | select(."@type" == "Need" and .status == "validated") | {id: ."@id", title: .title, module: .module."@id", statement: .statement}]' .arci/graph.jsonlt 2>/dev/null || echo '[]'`

Present the candidates using the AskUserQuestion tool so the developer can select one. If the list is empty, there are no validated needs ready for derivation. Suggest the developer validate a need first.

## Context

After identifying the NEED-* identifier, load its context and trace the derivation chain back to originating concepts:

!`NEED_ID="$1"; jq -s --arg id "$NEED_ID" '
  (.[] | select(."@id" == $id)) as $need |
  (.[] | select(."@id" == ($need.module."@id" // ""))) as $mod |
  [.[] | select(."@type" == "Requirement" and .module."@id" == $mod."@id")] as $existing |
  [$need.derivesFrom[]."@id" // empty] as $con_ids |
  [.[] | select(."@id" == ($con_ids[] // empty))] as $concepts |
  {
    need: $need,
    module: {id: $mod."@id", title: $mod.title},
    originating_concepts: [$concepts[] | {id: ."@id", title: .title, conceptType: .conceptType}],
    existing_requirements: [$existing[] | {id: ."@id", title: .title, statement: .statement}]
  }
' .arci/graph.jsonlt 2>/dev/null || echo '{"error": "Provide a NEED-* identifier."}'`

## Instructions

1. Read the need's statement and rationale.
2. Read the originating concepts' prose files for full context on the design thinking and alternatives considered. The concept prose informs what the requirement should constrain, not just what the need says at face value.
3. Derive one or more requirements that, if satisfied, would address the need. Each requirement should be more specific and constrained than the need.
4. For each requirement, draft a REQ-* node with: `derivesFrom` pointing to the need, `module` matching the need's module, a `statement` using "shall" language ("The system shall"), and `priority`.
5. Check against existing requirements in the module to avoid duplication.
6. Each requirement must be verifiable. If you can't describe how to verify it, it's not specific enough.
7. Before writing to the graph, run the review loop (see below).
8. Incorporate review feedback, then present the final requirements to the developer for approval.
9. Write approved requirements to `graph.jsonlt` with `status: "approved"`. The developer's approval during this skill constitutes requirement approval: they review the obligation and confirm it should go into the build. In team workflows where a separate review board or architect must sign off, set `status: "draft"` instead and note that approval is pending.
10. For each requirement, create a prose file at `.arci/requirements/{timestamp}-{NANOID}-{slug}.md`. Include the full statement, rationale, verification approach, and the design context from originating concepts and needs that motivated this obligation. The prose file captures the reasoning chain so future readers understand not just what the requirement says but why it exists and how to verify it.

## Review loop

After drafting the requirements but before writing them to the graph, use the Agent tool to review them. Pass the agent the drafted requirements, the source need, the originating concepts' prose content, and the list of existing requirements in the module. Instruct the review agent:

"Review these drafted requirements against the source need and originating concepts. Check each of the following and report only problems found:

- Is each requirement more specific and constrained than the need it derives from, or is it just the need restated in 'shall' language?
- Is each requirement verifiable? Could you write a concrete test or inspection for it, or is it too vague?
- Do any drafted requirements overlap with each other or with existing requirements in the module?
- Did the draft miss context from the originating concepts that should have produced additional requirements or tighter constraints?
- Are there requirements that are really process preferences or doctrine rather than verifiable system obligations?

Do not create any graph nodes, tasks, or defects. Return only your critique."

## Graph-editing conventions

| Pattern | Classification | Stage | Replacement |
|---------|---------------|-------|-------------|
| Candidate picker: validated needs query | Temporary | 3 | `arci need list --status validated` CLI command |
| Need context and derivation chain query | Temporary | 3 | `arci need derive` CLI command |
