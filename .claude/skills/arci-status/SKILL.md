---
name: arci-status
description: >-
  Check project status by synthesizing current state from the knowledge graph.
  Use when asked about project progress, module phases, task status, defect
  counts, verification coverage, or suspect links.
allowed-tools:
  - Read
  - Grep
  - Glob
  - Bash
---

# Check project status

Synthesize the current state of the project from the knowledge graph.

## Context

!`jq -s '
  {
    modules: [.[] | select(."@type" == "Module") | {id: ."@id", title: .title, phase: .phase, status: .status}],
    tasks: (
      [.[] | select(."@type" == "Task")] |
      group_by(.status) | map({(.[0].status): length}) | add
    ),
    defects: (
      [.[] | select(."@type" == "Defect")] |
      group_by(.status) | map({(.[0].status): length}) | add
    ),
    test_cases: (
      [.[] | select(."@type" == "TestCase")] |
      group_by(.currentResult) | map({(.[0].currentResult // "unknown"): length}) | add
    )
  }
' .arci/graph.jsonlt 2>/dev/null || echo '{"error": "No graph data found. Run arci init or create graph.jsonlt first."}'`

## Instructions

Using the preceding graph data, provide a narrative summary of the project state. Cover:

1. Which modules exist and what phase each is in.
2. Task progress: how many are complete, in progress, ready, and blocked.
3. Open defects, especially any blocking ones (critical/major severity with open/confirmed status).
4. Verification coverage: how many test cases are passing vs failing vs not yet run.
5. Any suspect links pending review.

Keep the summary concise. Highlight blockers and anything that needs attention. If the developer asks about a specific module, focus on that module's subtree.
