# Creating a baseline

## What

The developer says `baseline the architecture`, `create a release baseline`, `snapshot the current state`, or `lock down the design before we build`. The agent verifies preconditions (no uncommitted changes, no blocking defects, no unresolved suspect links), then creates a BSL-* node anchored to the current git commit that captures the graph state for a module subtree.

## Why

Baselines are decision records. They say "at this commit, the team declared this module's graph state to be authoritative." Everything after the baseline is change against a known reference. Baseline diffs show how the project evolved. Phase gate baselines mark completion of a phase. Release baselines mark shippable state.

With hierarchical phase constraints removed, baselines become the primary coordination mechanism. A hook policy can require that all child modules have approved verification baselines before the system creates a release baseline on the parent. This replaces the structural enforcement of C-PH1 with explicit, policy-driven synchronization.

## What happens in the graph

The agent runs precondition checks: git working tree is clean, no blocking defects on the target module, no unresolved suspect links in the module's subtree. If preconditions fail, the agent reports what needs fixing.

If preconditions pass, the agent creates a BSL-* node with `module` pointing at the scoped module, `commitSha` anchoring it to the git state, `status: draft` initially (pending review and approval), and statistics about the graph state (node counts, coverage percentages, defect status).

The baseline captures the state of the module's entire subtree (all descendants via `childOf`). For a release baseline on the root module, this is the entire graph.

Approval transitions the baseline to `status: approved`, making it an official reference point.

## Trigger patterns

`Create a baseline`, `baseline the architecture`, `snapshot this module`, `lock down the design`, `are we ready to baseline?`

## Graph before

A module with completed phase work, no blocking defects, clean git state.

## Graph after

A BSL-* node linked to the module, anchored to a commit, with captured statistics.

## Agent interaction layer

### Skills

The `krav:baseline` skill builds this workflow. It has `disable-model-invocation: true` in its frontmatter because baselines are deliberate milestones that shouldn't happen accidentally through agent initiative. Preprocessing loads the current module state (node counts, coverage percentages, defect status, suspect link count) so the developer can review the graph's health before confirming the baseline.

The skill's instructed commands run precondition checks via the CLI, present the module state summary for review, and create the BSL-* node on confirmation. The skill body emphasizes that baseline creation is a developer decision, not an agent optimization.

### Policies

The `graph-integrity` policy matters here because baseline creation must go through the CLI to ensure the commit anchor, statistics capture, and baseline scope computations are correct. The `cli-auto-approve` policy keeps the operation frictionless once the developer has confirmed.

The `mutation-feedback` policy fires after the agent creates the BSL-* node, injecting the baseline's statistics and anchor commit into the agent's context. This is useful for confirming what the baseline captured. Indirectly, creating a baseline activates the `baseline-protection` policy for future sessions: files in the baselined module's scope gain protection from unauthorized modification going forward.

### Task types

Baseline creation doesn't produce tasks. It's often a step within a `release` task (which aggregates verification and validation results and creates a release baseline as one of its deliverables), but the baseline operation itself is a graph snapshot, not a work item.

## Open questions

**Baseline approval workflow.** Who approves a baseline? The developer? A separate review step? For a solo developer using Krav, self-approval is fine. For a team, there might be a review process. The design mentions `status: approved` but not who does the approving or what criteria they use beyond the preconditions.

**What exactly does the baseline capture?** The design says baselines capture "graph state" but doesn't specify the serialization mechanism. Is it a copy of the relevant `graph.jsonlt` entries? A reference to the git commit (meaning the graph state is reconstructable from the commit)? The `commitSha` field suggests the latter, but that only works if the graph is fully committed to git at baseline time.

**Baseline detail level.** Can you baseline a single module without its subtree? The design scopes baselines to module subtrees, which means baselining a leaf component also baselines just that component. But can you baseline a parent without its children? That would serve the case of "system-level architecture has finished but subsystem architecture continues."

**Hook policy for cross-module coordination.** The index mentions hook policies that require child baselines before parent baselines. This is the replacement for C-PH1. But the policy language and enforcement mechanism aren't specified. What does the policy syntax look like? What phase-tag on child baselines does the system require? This needs design alongside the baseline workflow.

**Baseline frequency.** Should the developer create baselines at every phase transition? Only at major milestones? The design doesn't prescribe frequency. Phase advancement and baselining are separate operations, which is correct (you might advance without baselining, or baseline without advancing). But the typical patterns should get documented.
