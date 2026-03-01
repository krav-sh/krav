---
name: arci-trace
description: >-
  Trace the derivation chain for any node to explain why it exists and what
  depends on it. Use for impact analysis, provenance questions, or when asked
  "why does this exist" or "what depends on this."
allowed-tools:
  - Read
  - Grep
  - Glob
  - Bash
---

# Trace requirements

Walk the derivation chain from any node to explain provenance and impact.

## Instructions

Given a node identifier (any type), trace in both directions:

Upstream (why does this exist?): follow `derivesFrom` edges back through the chain. A requirement derives from a need, which derives from a concept. Show the full chain with each node's title and statement.

Downstream (what depends on this?): find all nodes whose `derivesFrom`, `verifiedBy`, `implements`, or `allocatesTo` edges point at this node. Show what would change if this node changed.

To trace upstream from a node:

```bash
NODE_ID="$1"
jq -s --arg id "$NODE_ID" '
  def ancestors($nid):
    [.[] | select(."@id" == $nid)] as $nodes |
    if ($nodes | length) == 0 then []
    else
      $nodes[0] as $n |
      [$n] + ([$n.derivesFrom // [] | .[]."@id"] | map(ancestors(.)) | add // [])
    end;
  ancestors($id) | map({id: ."@id", type: ."@type", title: .title, statement: .statement})
' .arci/graph.jsonlt
```

Present the trace as a narrative: "REQ-X exists because NEED-Y captured the expectation that (statement), which came from CON-Z where the team explored (topic)."
