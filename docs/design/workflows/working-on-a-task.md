# Working on a task

## What

The developer says "work on TASK-whatever" or "what's the next thing to do?" or "continue the lexer." The agent picks a task (or the developer names one), loads its context, does the work, and records deliverables on completion.

## Why

This is the core execution loop. The task DAG expresses what needs to happen; this workflow is where it actually happens. The graph provides context (what requirements does this task address, what did predecessor tasks produce, what's the module's state) so the agent can work with full awareness of why the code it's writing exists.

## What happens in the graph

If the developer asks `what's next?`, the agent runs `arci task ready` to find tasks with all dependencies satisfied. It presents the options (or picks the highest-priority one) and loads context via `arci context TASK-X`.

The task transitions from `ready` to `in_progress`. The agent does the work: writing code, producing documents, whatever the task type calls for. Along the way it might discover blocked dependencies, in which case the task transitions to `blocked` with a reason.

On completion, the task transitions to `complete` with deliverables recorded: commits, files created or modified, documents produced, test results. If the task is a review task, it also produces defects (see the "reviewing work" workflow).

## Trigger patterns

`Work on TASK-X`, `what should I work on next?`, `pick a task`, `continue where I left off`, `what's ready?`

## Graph before

TASK-* nodes in `ready` or `in_progress` status. Related requirements, module context, and predecessor task deliverables.

## Graph after

The task in `complete` status with deliverables. Dependent tasks may transition from `pending` to `ready` once this task finishes.

## Agent interaction layer

### Skills

The `arci:task` skill builds this workflow and is the core execution skill in ARCI's repertoire. Preprocessing is the richest of any skill: it injects the full task context including the task's requirements, deliverables expected, dependencies and their deliverables, module domain context, and any previous session notes. The !`command` directives call `arci context TASK-X` to assemble this context package before the agent sees its instructions.

The skill body's instructions vary by `taskType`. An `implement-feature` task gets guidance on recording commits and modified files as deliverables. A `design-api` task gets guidance on producing spec documents. A `review-code` task gets guidance on evaluating against requirements (though review tasks typically run through `arci:review` in a subagent instead). Instructed commands let the agent record deliverables, update task status, and flag blockers during execution.

### Policies

The `session-context` policy injects the current task state at session start, so the agent knows what it's supposed to work on and where a prior session left off. For resumed sessions, this is the primary mechanism for re-establishing context after a compaction or session break.

The `task-completion-gate` policy fires on TaskCompleted and is the workflow's most important enforcement mechanism. It checks that the agent recorded deliverables, that required graph mutations have happened, and that verification results exist if applicable. The `session-completion-gate` policy provides a catch-all at Stop time, preventing the agent from ending a session with an in-progress task that the agent hasn't properly closed. Together these two policies ensure task execution produces the graph artifacts that downstream workflows depend on.

The `mutation-feedback` policy fires after each deliverable recording and status transition, keeping the agent's running context current. If the task has four expected deliverables and the agent recorded two, the feedback injection surfaces that state. The `baseline-protection` policy fires if the task involves editing files in a baselined module, requiring the agent to justify or unlock modifications. The `graph-integrity` and `cli-auto-approve` policies apply as usual.

### Task types

Any of the 27 task types can execute through this workflow. The task's `taskType` field determines what the `arci:task` skill expects in terms of deliverables and completion criteria. Architecture-phase types like `decompose-module` and `define-interface` produce documents and diagrams. Coding-phase types like `implement-feature` and `refactor` produce commits and source files. Verification-phase types like `review-code` and `execute-tests` produce review reports and test results, though these typically execute via subagents rather than through this workflow directly.

## Open questions

**How much context should the skill load?** `arci context TASK-X` includes task details, module info, related requirements, dependency status, and previous session notes. But loading everything into the agent's context window is expensive. Should there be a tiered approach: load the essentials, and let the agent pull more context as needed?

**Session boundaries.** The design says tasks execute in "atomic Claude Code sessions." But what if a task takes multiple sessions? The task's prose content file can hold progress notes, and the graph knows what deliverables the agent recorded. But the specific in-progress state (which file the agent was editing, what approach it was taking) needs persistence. Where does that state live? In the task's prose file? In a session state store?

**Task selection heuristics.** When multiple tasks are ready, how does the agent pick? Priority is one signal. Critical path relevance is another. The developer's recent focus area might be a third. Should the agent present options, or should there be a configurable selection strategy?

**Partial completion.** What if the agent finishes part of a task's scope but not all of it? Should it mark the task complete (accepting partial deliverables), split the task into done/remaining, or leave it in_progress? The design doesn't have a partial completion state.

**Context from predecessor tasks.** A design task's deliverable (say, an API spec document) is input to the coding task. How does the agent access predecessor deliverables? Through the deliverables array on the predecessor task? By reading the files directly? The context loader needs to bridge graph metadata and filesystem artifacts.
