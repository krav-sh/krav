---
name: krav-baseline
description: >-
  Create a baseline to capture graph state at a decision point. Use when
  a milestone is reached and the current state should be frozen as a named
  reference point.
disable-model-invocation: true
stage-classification: temporary
replacement-stage: 1
replacement: "`krav baseline create` CLI command"
---

# Create a baseline

Capture graph state as a named snapshot anchored to the current git commit.

## Current state

!`MOD_ID="$1"; jq -s --arg id "$MOD_ID" '
  (.[] | select(."@id" == $id)) as $mod |
  [.[] | select(."@type" == "Baseline" and .module."@id" == $id)] as $existing |
  {
    module: {id: $mod."@id", title: $mod.title, phase: $mod.phase},
    existing_baselines: [$existing[] | {id: ."@id", title: .title, phase: .phase, commitSha: .commitSha, status: .status}]
  }
' .krav/graph.jsonlt 2>/dev/null || echo '{"error": "Provide a MOD-* identifier."}'`

## Instructions

1. Show the developer the current module state and any existing baselines.
2. Confirm they want to baseline at this point.
3. Create a BSL-* node with: `module`, `phase` (current module phase), `commitSha` (current HEAD), `scope`, `approvedBy`, `status: "draft"`.
4. The developer reviews and approves the baseline, transitioning to `"approved"`.
5. Once approved, the baseline protects its scope from uncontrolled modification. Changes to baselined content require a defect and module phase regression.

## Graph-editing conventions

| Pattern | Classification | Stage | Replacement |
|---------|---------------|-------|-------------|
| Module state and baseline inventory query | Temporary | 1 | `krav baseline create` CLI command |
