# Implementing without ceremony

## What

The developer says `fix this bug`, `add a --verbose flag`, `rename this function`, or `update the README`. Small, well-understood changes that don't need concept exploration, need formalization, requirement derivation, or formal verification. The developer wants the agent to write code, not manage a knowledge graph.

## Why

This is the most common interaction in a mature project, and it's where ARCI risks being an obstacle rather than an aid. If every three-line change requires graph ceremony, developers turn off ARCI or work around it. The system needs a graceful path for low-ceremony work that still leaves a useful trace without demanding a full transformation chain.

## What happens in the graph

This is the central design question. A few options, from least to most ceremony:

**Nothing.** The agent writes code and ARCI isn't involved. The hooks might fire (PreToolUse on file writes) but no graph nodes get created. The change exists only in git history. This is the simplest option but means the graph drifts from reality: work happened that the graph doesn't know about.

**Task only.** The agent creates a lightweight TASK-* node with minimal metadata (title, module, phase: coding, status: complete, deliverables: the commit). No concepts, needs, or requirements. The graph records that work happened on this module. This is a reasonable middle ground.

**Task linked to existing requirements.** If the change addresses an existing requirement (a bug fix for a requirement that the system should already satisfy, a small enhancement covered by an existing need), the task links to those nodes. The graph accurately reflects what the change was for. This requires the agent to identify the relevant requirement, which isn't always obvious for small changes.

## Trigger patterns

`Fix this bug`, `add a flag for X`, `rename Y to Z`, `update the docs`, `clean up the error messages`, `refactor the parser`, `bump the version`.

## Graph before

Whatever the current state is. The developer isn't thinking about the graph.

## Graph after

Depends on the ceremony level chosen. Might stay unchanged, might have a single TASK-* node, might have a task linked to existing requirements.

## Agent interaction layer

### Skills

The `arci:quickfix` skill builds this workflow. Unlike most skills, its primary job is triage rather than execution: it determines the appropriate ceremony level based on what the developer is changing and the current graph state. If the affected module has a baseline, the skill escalates toward task creation. If the change addresses an existing defect, it links to the remediation task. If neither applies, it may create a lightweight task record or skip the graph entirely.

Preprocessing loads the module's baseline status, open defects, and existing requirements for the area the developer is changing. This gives the agent the signals it needs to make the ceremony decision. The skill's instructed commands cover both the lightweight path (create a minimal task, record deliverables) and the escalation path (defer to `arci:task` or `arci:defect` for more structured execution).

### Policies

The `baseline-protection` policy is the most important policy for this workflow. It distinguishes `quick change to new code` from `quick change to baselined code`: the former can proceed with minimal ceremony, while the latter triggers an enforcement response that pushes the agent toward creating a task or defect record. The escalating enforcement model (warn on first attempt, block on subsequent) gives the agent a chance to course-correct.

The `session-context` policy provides baseline status and active defect state at session start, giving the quick-fix skill the inputs it needs for its ceremony triage. The `mutation-feedback` policy fires if the agent does create graph nodes, keeping context current. The `graph-integrity` and `cli-auto-approve` policies apply when graph operations happen.

### Task types

When the quick-fix skill decides a task record warrants creation, the typical types are `remediate-defect` (if fixing a known defect), `implement-feature` (for small enhancements), or `refactor` (for cleanup work). The agent creates the task with minimal metadata and immediately marks it complete with the deliverables recorded. The ceremony is deliberately light: one task node, the commit as a deliverable, done.

## Open questions

**Where's the line?** What makes a change "small enough" to skip ceremony? Line count? Estimated time? Whether it touches existing requirements? Whether the affected module has a baseline? This might not have a universal answer; it could be a per-project policy configured via hooks.

**Should hooks enforce minimum ceremony?** A hook policy could require that all file modifications in a baselined module produce at least a task record. This prevents graph drift for stable modules while leaving un-baselined modules informal. Is this the right default?

**Defect-driven quick fixes.** If the developer is fixing a known defect (DEF-*), the fix should link to the defect's remediation task. But if they're fixing something that isn't tracked as a defect ("oh, this error message is wrong"), should the agent create a defect first and then fix it? That's the right thing for traceability but feels heavy for a typo fix.

**Retrospective graph updates.** Can the developer do work informally and then retroactively "explain" it to the graph? Something like `that last commit was for REQ-C2H6N4P8` after the fact, creating a task record with the deliverable. This is friendlier than requiring graph involvement upfront.

**How does the agent decide?** If the developer says `add a --verbose flag`, how does the agent decide whether this needs the full chain or just a quick task? Signals might include: does a need or requirement already exist for verbosity? Does the module have a baseline? Has the developer configured a ceremony policy? The agent's default behavior here shapes the entire user experience of ARCI.

**Interaction with verification.** If a quick fix changes behavior covered by existing test cases, those tests should run again and the agent should record the results. But if the change is documentation or cosmetic, no verification applies. The agent needs to distinguish.
