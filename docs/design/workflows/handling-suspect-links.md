# Handling suspect links

## What

The developer says "are there any suspect links?" or suspect links surface during another workflow (a review, a status check, a phase advancement attempt). The agent reads suspect edges, examines what changed upstream, and for each one either clears the flag, updates the downstream node, or creates a defect.

## Why

Suspect links are the change impact analysis mechanism. When a concept changes after the agent has derived needs from it, or when a need changes after the agent has derived requirements, the downstream artifacts might be stale. The suspect flag forces explicit review rather than letting staleness propagate silently.

Without suspect links, the graph gives a false sense of completeness. All requirements might show as "approved" even though someone modified the need they derive from last week and the requirements no longer reflect it. Suspect links surface this kind of drift.

## What happens in the graph

Suspect flags get set automatically by constraint C-SUSPECT1 when upstream nodes change. The agent's job is to triage them.

For each suspect edge, the agent reads the upstream node's change (what changed and why), the downstream node's current state, and the relationship between them. Then it decides:

Clear: the upstream change doesn't affect the downstream node. The agent removes the flag. This is the common case for minor wording changes or metadata updates.

Update: the upstream change does affect the downstream node, but it's a straightforward adjustment. The agent updates the downstream node's statement, criteria, or other fields to align with the upstream change. The agent clears the flag after the update.

Defect: the upstream change reveals a real problem with the downstream node that needs more than a quick fix. The agent creates a DEF-* with category `inconsistent` or `incomplete`, and clears the suspect flag (the defect now tracks the problem).

Propagate: the upstream change is large enough that downstream edges from the affected node should also get suspect markings. C-SUSPECT2 says propagation is non-transitive by default; the reviewer explicitly chooses to propagate.

## Trigger patterns

`Are there suspect links?`, `check for stale traceability`, `review suspect edges`, or surfaced as blockers during phase advancement or baseline creation.

## Graph before

One or more edges marked suspect due to upstream modifications.

## Graph after

Suspect flags cleared, downstream nodes updated, or defects created. Possibly further propagation to deeper edges.

## Agent interaction layer

### Skills

The `krav:suspect` skill builds this workflow. Preprocessing loads the suspect link report: which edges carry flags, what changed upstream, and what the downstream nodes currently contain. This gives the agent (or analyst subagent) a complete triage queue with the context needed to evaluate each suspect edge.

The skill's instructed commands clear suspect flags, update downstream nodes, create defects for inconsistencies, and optionally propagate suspect flags to deeper edges. The skill instructions guide the triage decision for each edge: compare the upstream change against the downstream node's content and determine whether the downstream artifact is still valid.

### Agents

The `krav-analyst` subagent handles suspect link triage when the volume is large enough to benefit from a dedicated context window. For a handful of suspect links after a minor change, the main agent handles triage directly using the `krav:suspect` skill. For a major upstream modification that produces dozens of suspect flags, delegating to the analyst avoids consuming the main agent's context with graph traversal work.

The analyst recommends dispositions (clear, update, create defect, propagate) but doesn't act on them unilaterally. The main agent or developer makes the final call, which is important because suspect link triage often involves judgment about whether an upstream change matters semantically.

### Policies

The `session-context` policy surfaces pending suspect links at session start, so the developer sees them immediately rather than discovering them during a phase advancement attempt. The `mutation-feedback` policy fires after the agent clears each suspect flag or creates a defect, keeping the running count of pending suspect links current.

The `graph-integrity` and `cli-auto-approve` policies apply as usual. The `phase-gate-defense` policy is indirectly relevant: unresolved suspect links may block phase advancement, so triaging suspects is often a prerequisite for the `krav:advance` workflow.

### Task types

Suspect link triage doesn't create tasks directly, but defects created from suspect links flow into the `krav:defect` workflow, which creates `remediate-defect` tasks. The connection is indirect but important: a suspect link that reveals an inconsistent requirement produces a defect, which produces a remediation task, which produces a fix.

## Open questions

**How noisy are suspect links in practice?** If editing a concept's description marks all downstream derivesFrom edges suspect, even a typo fix triggers triage. The design says "key semantic properties" trigger suspect marking, but the boundary between semantic and cosmetic changes isn't defined. Should there be a distinction between "statement changed" (suspect) and "description typo fixed" (not suspect)?

**Batch triage.** After a major upstream change, there might be dozens of suspect links. The agent should present these in a way that allows batch disposition (clear all, review one by one, group by downstream type). What's the right UX for this?

**Who reviews?** Suspect link triage requires understanding why the upstream change happened and whether it matters downstream. The agent can identify what changed, but the developer needs to judge whether the downstream artifacts are still valid. How interactive should this be?

**Timing.** Should the developer triage suspect links immediately when they appear, or batch and review them periodically? Immediate triage interrupts the developer's flow. Batched triage risks accumulation. The design doesn't specify a preferred cadence.

**Suspect links across module boundaries.** When a requirement on a parent module changes and it has `allocatesTo` edges to child modules, the child modules' derived requirements become suspect. This crosses module boundaries, which might cross team boundaries in a larger project. The notification and triage workflow for cross-module suspects isn't specified.
