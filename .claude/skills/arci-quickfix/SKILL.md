---
name: arci-quickfix
description: >-
  Implement something with minimal or no graph ceremony. Use for trivial
  fixes, quick changes, and work where the full transformation chain adds
  no value. Also use when the developer says "just fix it" or "just do it."
stage-classification: permanent
rationale: "Methodology guidance about ceremony levels; encodes developer judgment that no tool can replace"
---

# Implement without ceremony

Determine the appropriate ceremony level and execute with minimal graph overhead.

## Graph context

!`jq -s '
  [.[] | select(."@type" == "Baseline" and .status == "approved")] as $baselines |
  [.[] | select(."@type" == "Module") | {id: ."@id", title: .title, phase: .phase}] as $mods |
  {
    baselines: [$baselines[] | {id: ."@id", scope: .scope, commitSha: .commitSha}],
    modules: $mods
  }
' .arci/graph.jsonlt 2>/dev/null || echo '{}'`

## Instructions

Assess how much ceremony the work needs based on these signals:

Does it modify baselined content? If yes, you cannot skip the graph. Create a defect, get the module unlocked, then proceed with at least a task.

Does it affect existing requirements? If yes, create a task linked to those requirements so traceability holds.

Is it a new area with no requirements yet? A lightweight task node (title, module, status) is enough. Skip needs and requirements for now.

Is it truly trivial (typo fix, config tweak, formatting)? Skip the graph entirely. Just make the change and commit.

When in doubt, create a TASK-* node. It's cheap and preserves a record of what happened and why. But don't create needs, requirements, and test cases for a one-line bug fix.

## Graph-editing conventions

| Pattern | Classification | Stage | Replacement |
|---------|---------------|-------|-------------|
| Baseline and module inventory query | Permanent | n/a | n/a |
