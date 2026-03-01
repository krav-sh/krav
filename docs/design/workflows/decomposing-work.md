# Decomposing work into tasks

## What

The developer says "what needs to happen to build the parser?" or "plan the work for this module" or "decompose the coding work." The agent reads the module's requirements, the current graph state, and the module's phase, then generates a task DAG. Tasks get `dependsOn` edges expressing ordering, `processPhase` labels indicating what kind of work they are, and module ownership.

## Why

Task decomposition is where the knowledge graph connects to actual work. Requirements describe what the system must do. Tasks describe what the developer (or agent) must do to make it so. Without decomposition, requirements are obligations with no path to fulfillment.

The DAG structure is important because it replaces plan documents. "What's the plan?" is a graph query, not a prose document that goes stale. Dependencies are explicit and queryable: "what's blocking this?" is a traversal, not a meeting.

## What happens in the graph

The agent examines the module's requirements, its current phase, and what tasks already exist. For each unaddressed requirement, it identifies what work the team needs. Work might span multiple phases: an architecture task to identify interfaces, a design task to define the API, coding tasks to build it, verification tasks to test it.

The agent also creates the prose content files for tasks that need detailed instructions (complex coding tasks, review tasks with specific checklists).

## Trigger patterns

`Plan the work for X`, `decompose the coding work`, `what needs to happen to build Y?`, `create tasks for this module`, `what's the next set of work?`

## Graph before

A module with requirements. Possibly some tasks already exist from earlier decomposition.

## Graph after

Multiple TASK-* nodes forming a DAG, with `dependsOn` edges, phase labels, and module ownership. Complex tasks have prose content files.

## Agent interaction layer

### Skills

The `krav:decompose` skill builds this workflow. Preprocessing is among the heaviest of any skill: it loads the module's requirements, existing tasks (so the agent doesn't create duplicates or ignore completed work), the module's current phase, and available decomposition templates. The templates provide reusable patterns for common requirement shapes, so the agent doesn't reinvent the same task chains every time.

The skill's instructed commands create TASK-* nodes, set `dependsOn` edges between them, assign `taskType` and `processPhase` labels, and generate prose content files for complex tasks. The skill instructions encode the logic for selecting task types based on requirement characteristics: a requirement with `verificationMethod: test` produces `implement-tests` and `execute-tests` tasks, a requirement about API behavior produces `design-api` then `implement-feature`, and so on.

### Policies

The `mutation-feedback` policy is critical here because decomposition creates many nodes in sequence and the agent needs to track the emerging DAG structure. After the agent creates each task, the feedback injection shows the current task count, dependency structure, and which requirements are now addressed by tasks.

The `session-context` policy provides the starting context for the module's phase and any active tasks, so the agent knows where decomposition left off if this is a continuation session. The `prompt-context`, `graph-integrity`, and `cli-auto-approve` policies apply as usual.

### Task types

This is the workflow that creates most task types. The decompose skill selects types based on what the requirements demand and what phase the module is in. A typical decomposition chain for a behavioral requirement runs `design-api` then `implement-feature` then `implement-tests` then `execute-tests` then `review-code`, with `dependsOn` edges encoding the ordering. Requirements about data storage add `design-data-model`. Performance requirements add `benchmark-performance`. Cross-module requirements add `integrate-components`. The full set of 27 task types across all six phases is available, though any single decomposition typically produces a subset from the phases relevant to the module's current work.

## Open questions

**How fine-grained should tasks be?** "Build the parser" is too coarse for an atomic Claude Code session. "Write line 47 of lexer.ts" is too fine. The design says tasks should be "atomic units of work with verifiable deliverables." But what's atomic? Is it a function, a module, a feature? The agent needs heuristics for the right level of detail, and those heuristics probably vary by phase (architecture tasks are broader, coding tasks are narrower).

**Should decomposition happen all at once or incrementally?** Generating the full task DAG from architecture through validation upfront means tasks for later phases are speculative; they change as earlier work reveals new information. Generating only the next phase's tasks is more accurate but requires re-decomposition at each phase boundary. Which should the agent default to?

**How does the agent handle existing tasks?** If some architecture tasks already exist and are complete, decomposition for the design phase should account for their deliverables. The agent needs to read existing task state, not just requirements.

**How do milestone tasks get created?** A release milestone is a validation-phase task that depends on verification tasks across modules. The developer might ask for this explicitly ("create a release milestone") or the agent might propose it when decomposing system-level work. Who initiates?

**Template interaction.** The templating system (when designed) provides reusable decomposition patterns. The open question is how templates and agent-driven decomposition interact: whether the agent uses templates as a starting point and customizes, whether the developer chooses a template explicitly, and whether the agent can compose multiple templates.
