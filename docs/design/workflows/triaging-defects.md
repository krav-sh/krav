# Triaging and fixing defects

## What

The developer says `look at the open defects`, `triage the review findings`, `fix DEF-X`, or `what defects are blocking advancement?` The agent reads open defects, helps the developer disposition them (confirm, reject, defer), generates remediation tasks for confirmed defects, executes fixes, and marks defects resolved. Re-examination after the fix transitions them to verified, then closed.

## Why

Defects are the corrective feedback loop. Without a structured triage process, the team tracks problems found in reviews or testing informally (or not at all), fixes happen ad hoc, and nobody records what went wrong, why, or whether the fix actually worked. The defect lifecycle (open then confirmed then resolved then verified then closed) ensures nothing falls through the cracks.

The categorization also matters for process improvement. If the graph consistently shows `ambiguous` defects against requirements from a particular module, the derivation step for that module needs attention. If `regression` defects spike after a refactor, the verification coverage was inadequate. These patterns only emerge from structured defect data.

## What happens in the graph

**Triage phase.** The agent presents open DEF-* nodes with their subject, category, severity, and the context of how the detection task found them (the `detectedBy` task). For each defect, the developer decides:

Confirm: the defect is real. It stays open with `status: confirmed` and the agent generates a remediation TASK-* node via `arci defect generate-task`.

Reject: the defect is not real (false positive, misunderstanding, working as intended). It transitions to `rejected` with a rationale on record.

Defer: the defect is real but not worth fixing now. It transitions to `deferred` with a target (milestone or version) and rationale.

**Remediation phase.** For confirmed defects, the agent works the generated remediation task like any other task (see "working on a task" workflow). The fix produces deliverables. When the task completes, the defect transitions to `resolved`.

**Verification phase.** After resolution, someone (agent or developer) re-examines the fix. If it's adequate, the defect transitions to `verified`, then `closed`. If the fix is inadequate, the defect reopens.

## Trigger patterns

`Show open defects`, `triage the review findings`, `fix DEF-X`, `what's blocking module advancement?`, `what defects are open for the parser?`

## Graph before

DEF-* nodes in `open` or `confirmed` status, linked to subjects and detection tasks.

## Graph after

Defects dispositioned (confirmed, rejected, deferred). Remediation tasks created and completed. Defects resolved, verified, closed.

## Agent interaction layer

### Skills

The `arci:defect` skill builds this workflow. Preprocessing loads the module's open defects with their subjects, categories, severities, and detection context (which review or verification task found them). This gives the agent a complete triage queue without additional graph queries.

The skill's instructed commands handle the full defect lifecycle: confirming, rejecting, and deferring defects via `arci defect disposition`; generating remediation tasks via `arci defect generate-task`; and recording resolution and verification via `arci defect resolve` and `arci defect verify`. The skill instructions walk the agent through presenting each defect in context and helping the developer make disposition decisions.

### Policies

The `session-context` policy surfaces open defects and their blocking status at session start, which is particularly important for triage sessions where the developer is coming back to address review findings. The `mutation-feedback` policy fires after each disposition change and remediation task creation, keeping the running defect count and blocking status current.

The `graph-integrity` and `cli-auto-approve` policies ensure all defect lifecycle operations go through the CLI. When the agent creates remediation tasks and transitions to executing them, the `task-completion-gate` and `session-completion-gate` policies from the task execution workflow also apply.

### Task types

The defect skill creates `remediate-defect` tasks for confirmed defects. This task type is always `remediate-defect` regardless of what the fix looks like in practice (it might be a refactor, a code change, a documentation update), because the completion criteria differ from other task types: the developer must address and verify the specific defect, not just achieve general improvement. The DEF-* node links to the remediation task via `generates`, closing the traceability loop.

## Open questions

**Who decides disposition?** Triage involves judgment calls: is this really a bug? Is it severe enough to block? The developer should make these calls, but the agent can help by presenting the defect in context (the requirement it targets, the code it affects, similar past defects). Should the agent propose dispositions, or just present the information?

**Remediation task scope.** `arci defect generate-task` creates a single remediation task per defect. But some defects require multiple steps to fix (design change, code changes, updated tests). Should the agent generate a task DAG for complex fixes, or is one task per defect the right default?

**Defect clustering.** Multiple defects from the same review might have a common root cause. Fixing one might fix the others. The agent should recognize this and link related defects, or at minimum warn when a remediation task addresses a scope that overlaps with other open defects. The graph doesn't currently have a "related to" relationship between defects.

**Re-verification after fixes.** When a defect fix changes code covered by existing test cases, those tests should run again to confirm no regressions. Should defect resolution automatically trigger the "running verification" workflow for affected test cases?

**Deferred defect management.** Deferred defects need to surface again at the right time. If a defect gets deferred to `v2.0`, what triggers it to reappear? A milestone task for v2.0 that depends on resolving deferred defects? A periodic "review deferred defects" workflow? This lifecycle isn't fully specified.
