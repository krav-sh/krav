---
name: arci-suspect
description: >-
  Handle suspect links caused by upstream changes. Use when nodes have been
  modified and downstream traceability links need review to determine if
  dependent nodes need updating.
stage-classification: temporary
replacement-stage: 2
replacement: "Post-mutation hook policy that marks downstream edges as suspect when a node's content fields change"
---

# Handle suspect links

Triage downstream impacts of upstream changes.

## Suspect links

!`jq -s '
  [.[] | to_entries | .[] | select(.value | type == "object" and .suspect == true) | {node: .key, edge: .value}] // [] |
  if length == 0 then "No suspect links found." else . end
' .arci/graph.jsonlt 2>/dev/null || echo '"No graph data found."'`

## Instructions

For each suspect link:

1. Read the upstream node (the source that changed) and understand what changed.
2. Read the downstream node and evaluate whether the change affects it.
3. Decide: clear the suspect flag (change doesn't affect the downstream node), update the downstream node to reflect the change, or create a defect if the change reveals a problem.
4. Clear the `suspect` flag on the edge after triaging.

Suspect links propagate through `derivesFrom`, `verifiedBy`, and `allocatesTo` edges. Other edge types don't carry suspect flags.

## Graph-editing conventions

| Pattern | Classification | Stage | Replacement |
|---------|---------------|-------|-------------|
| Suspect link detection query | Temporary | 2 | Post-mutation hook policy with `arci suspect list` CLI command |
