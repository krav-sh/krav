---
name: krav-review
description: >-
  Review work deliverables against requirements and produce defects for
  problems found. Use when code review, design review, or any examination
  of deliverables is needed.
stage-classification: temporary
replacement-stage: 3
replacement: "Review subagent with `krav defect create` CLI command"
allowed-tools:
  - Read
  - Grep
  - Glob
  - Bash
---

# Review work

Examine deliverables against requirements and record defects for problems found.

## Module context

!`MOD_ID="$1"; jq -s --arg id "$MOD_ID" '
  (.[] | select(."@id" == $id)) as $mod |
  [.[] | select(."@type" == "Requirement" and .module."@id" == $id)] as $reqs |
  [.[] | select(."@type" == "Task" and .module."@id" == $id and .status == "complete")] as $tasks |
  {
    module: {id: $mod."@id", title: $mod.title, phase: $mod.phase},
    requirements: [$reqs[] | {id: ."@id", title: .title, statement: .statement}],
    completed_tasks: [$tasks[] | {id: ."@id", title: .title, deliverables: .deliverables}]
  }
' .krav/graph.jsonlt 2>/dev/null || echo '{"error": "Provide a MOD-* identifier."}'`

## Instructions

1. Read the module's requirements. These are the criteria for the review.
2. Read the deliverables from completed tasks. These are the artifacts under review.
3. For each deliverable, evaluate against the relevant requirements. Check for: correctness (does it satisfy the requirement?), completeness (does it cover the full requirement scope?), consistency (does it conflict with other deliverables or requirements?), and quality (code style, error handling, edge cases).
4. For each problem found, create a DEF-* node in `.krav/graph.jsonlt` with: `subject` pointing to the relevant REQ-* or TASK-*, `category` (missing, incorrect, ambiguous, inconsistent, etc.), `severity` (critical, major, minor, trivial), `statement` describing the problem, and `status: "open"`. Create a prose file at `.krav/defects/{timestamp}-{NANOID}-{slug}.md` with the full problem description, evidence from the reviewed artifact, the requirement or quality standard it violates, and suggested remediation approach.
5. Summarize findings at the end: how many defects found, severity breakdown, and assessment.

Be thorough but fair. Not every imperfection is a defect. Focus on problems that affect requirement satisfaction or system quality.

## Graph-editing conventions

| Pattern | Classification | Stage | Replacement |
|---------|---------------|-------|-------------|
| Module requirements and completed tasks context query | Temporary | 3 | `krav module review` CLI command |
