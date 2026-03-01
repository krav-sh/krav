---
name: arci-explore
description: >-
  Explore a design question by creating and developing concepts. Use when
  investigating design alternatives, making architectural decisions, or
  thinking through a problem before formalizing it into needs and requirements.
stage-classification: temporary
replacement-stage: 3
replacement: "Production arci-explore skill backed by `arci concept` CLI subcommands (create, explore, crystallize)"
---

# Explore a design question

Create and develop concepts through exploration. Record decisions, alternatives considered, and rationale.

## Module context

!`jq -s '[.[] | select(."@type" == "Module") | {id: ."@id", title: .title, phase: .phase}]' .arci/graph.jsonlt 2>/dev/null || echo '[]'`

## Instructions

1. Identify which module this exploration belongs to. If unclear, ask the developer.
2. Create a CON-* node in `.arci/graph.jsonlt` with status `"draft"` and the appropriate `conceptType` (architectural, operational, technical, interface, process, integration).
3. Create a prose file at `.arci/concepts/{timestamp}-{NANOID}-{slug}.md` to capture the exploration.
4. As the exploration progresses, update the concept's status: `"draft"` → `"exploring"` → `"crystallized"`.
5. Record alternatives considered and the rationale for the chosen approach in the prose file.
6. If the concept informs a module's design, add an `informs` edge to that module.

A crystallized concept is ready for formalization into needs. Don't rush to crystallize, because the value of exploration is in considering alternatives.
