---
name: arci-diff
description: >-
  Compare two baselines to show what changed between them. Use when asked
  about changes between milestones, what was added or removed, or for
  release notes.
allowed-tools:
  - Read
  - Grep
  - Glob
  - Bash
---

# Compare baselines

Produce a semantic diff between two graph states.

## Instructions

Given two baseline identifiers (or one baseline and the current state):

1. Read the baseline nodes to get their `commitSha` values.
2. Use `git show {sha}:.arci/graph.jsonlt` to read the graph at each baseline's commit.
3. Compare the two graph states:
   - Added nodes (present in later, absent in earlier)
   - Removed nodes (present in earlier, absent in later)
   - Modified nodes (same `@id`, different field values)
   - Created or broken edges
   - Status transitions
4. Present the diff as a structured summary, grouped by node type.

If comparing against the current state (no second baseline), use the working tree's `graph.jsonlt` as the later state.
