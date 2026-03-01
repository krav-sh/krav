---
name: arci-init
description: >-
  Bootstrap a new project with a root module, module hierarchy, stakeholders,
  and initial concepts. Use when starting a new arci-managed project or when
  asked to initialize arci for an existing codebase.
disable-model-invocation: true
---

# Start a new project

Bootstrap the knowledge graph with the core structure for a project.

## Instructions

1. Ask the developer about the project: what it does, who the stakeholders are, and what the major architectural boundaries are.
2. Create the `.arci/` directory structure if it doesn't exist: `concepts/`, `modules/`, `needs/`, `requirements/`, `test-cases/`, `tasks/`, `defects/`, `baselines/`, `stakeholders/`, `developers/`, `agents/`, `policies/`.
3. Initialize `.arci/graph.jsonlt` with:
   - A root MOD-* node (phase: architecture, status: active)
   - Child MOD-* nodes for each major architectural boundary, with `childOf` edges to the root
   - STK-* nodes for each identified stakeholder, with `concerns`
   - DEV-* nodes for the developers working on the project
   - Initial CON-* nodes for key design decisions already made, with `informs` edges to relevant modules
4. Create prose files for concepts that need extended explanation.
5. Summarize the new graph contents and suggest next steps (typically: explore design questions, formalize concepts into needs).
