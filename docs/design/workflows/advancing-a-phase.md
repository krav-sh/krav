# Advancing a module's phase

## What

The developer says "advance the parser to design" or "is the parser ready for the next phase?" or "what's blocking the parser from advancing?" The agent checks the module-scoped advancement criteria, reports what's blocking if the criteria aren't met, and performs the advancement if they are.

## Why

Phase advancement is a deliberate ceremony. It asserts that the module's current-phase work is complete, reviewed, and free of blocking problems. The criteria are concrete: all phase tasks complete, no blocking defects, review dispositions acceptable. This prevents the module from moving forward with unfinished or unreviewed work. The system checks criteria per-module, not globally.

With module-scoped phase gating (C-PH1 removed), each module advances independently. The parser can move to `implementation` while the CLI is still in design. Task dependencies and baselines handle cross-module synchronization, not hierarchical phase constraints.

## What happens in the graph

The agent runs `krav moduleadvance MOD-X --to <target>` (or the equivalent check). The command evaluates:

All tasks for the current phase with `module = MOD-X` and `processPhase = current_phase` must have `status = complete`.

No DEF-* nodes where `module = MOD-X` and `severity` is critical or major and `status` is open or confirmed.

All review tasks for the current phase must have acceptable dispositions.

If any criterion fails, the agent reports what's blocking. `3 tasks incomplete`, `1 major defect open`, `design review not accepted`. The developer can then address the blockers and try again.

If all criteria pass, the module's `phase` field advances. The agent might suggest next steps: `parser is now in design phase, do you want to decompose design tasks?`

## Trigger patterns

`Advance X to Y phase`, `are we ready for the next phase?`, `what's blocking advancement?`, `move the parser forward`.

## Graph before

A module with current-phase work completed (or not; the check reveals the state).

## Graph after

If successful: the module's `phase` field advances by one step. If blocked: no change, but the developer has a clear list of what to resolve.

## Agent interaction layer

### Skills

The `krav:advance` skill runs this workflow. Preprocessing loads the module's current phase, task completion status, open defects, review dispositions, and verification coverage. This gives the agent a complete picture of whether the module satisfies advancement criteria before the developer even asks.

The skill's instructed commands run `krav moduleadvance` to attempt advancement and interpret the results. If criteria aren't met, the skill instructions guide the agent through reporting specific blockers and suggesting next steps (complete these tasks, resolve these defects, run these reviews). If advancement succeeds, the instructions suggest phase-appropriate follow-up work.

### Policies

The `phase-gate-defense` policy is the defining policy for this workflow. It fires on PreToolUse when the agent runs `krav moduleadvance` and independently validates that the module satisfies advancement preconditions, providing defense in depth alongside the CLI's own precondition checks. The denial message reports exactly what's blocking: open defects, incomplete tasks, insufficient verification coverage.

The `session-context` policy surfaces the module's phase and blocking status at session start. The `mutation-feedback` policy fires after a successful advancement, injecting the module's new phase and suggesting next steps. The `graph-integrity` and `cli-auto-approve` policies apply as usual.

### Task types

Advancement itself doesn't create tasks, but the `krav:advance` skill may create review and verification tasks as preconditions if they don't already exist. If a module is trying to advance from `implementation` to `integration` but has no `review-code` task, the skill creates one. If verification coverage is insufficient, it may create `execute-tests` tasks. These precondition tasks are verification-phase types: `review-code`, `review-design`, `review-architecture`, and `execute-tests`.

## Open questions

**Skipping phases.** Can a module advance from architecture to `implementation`, skipping design? The phase ordering follows a fixed sequence (architecture → design → `implementation` → integration → verification → validation), and the design implies single-step advancement. But for simple modules where "design" is trivial, requiring an empty design phase (with zero tasks) before advancing might feel like busywork. Should advancement allow skipping to a non-adjacent phase if all intermediate phases have no tasks?

**Implicit advancement.** If the developer starts creating `implementation` tasks for a module that's in architecture phase, should the agent automatically advance the module? Or should it block (per C-PH3) and require explicit advancement? The constraint says nobody can create tasks for phases the module hasn't reached. That means the developer must advance the module before creating tasks, which means thinking about phase transitions as a separate step.

**What happens after advancement?** The agent knows the module just moved to design phase. Should it suggest decomposing design tasks? Loading the design phase's context? Or wait for the developer to ask? Proactive suggestions are helpful for flow but annoying if the developer has a different plan.

**Advancement notifications.** When a module advances, downstream modules that depend on it (via task dependencies) might have tasks that become unblocked. Should the agent surface these? `Parser advanced to implementation, the following tasks in the CLI module are now ready.`
