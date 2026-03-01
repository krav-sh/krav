# Building a feature end-to-end

## What

The developer says `add offline support` or `build the caching layer` or `build search`. This is a high-level intent that spans the full transformation chain: exploring the design, capturing stakeholder needs, deriving requirements, decomposing into tasks, executing those tasks, writing tests, reviewing, and fixing defects. It is not a single workflow; it is an orchestrated sequence of other workflows.

## Why

This is what developers actually ask for most of the time. They don't say "create a concept node in exploring status"; they say "build the thing." The agent needs to translate a high-level feature request into a structured sequence of graph operations, making the underlying workflows invisible unless the developer wants to see them.

## What happens in the graph

Everything. This workflow touches every node type:

It starts with concept exploration if the feature has not yet gone through design. If a relevant crystallized concept already exists, it picks up from there. Concepts get formalized into needs. Needs get derived into requirements. Work gets decomposed into a task DAG. Tasks get executed. Test cases get created and built. Reviews happen and produce defects. Defects get triaged and fixed.

The agent manages the transitions between these sub-workflows, knowing when concept exploration finishes and formalization should start, when enough requirements exist to begin decomposition, and when the code is complete enough for verification.

## Trigger patterns

`Build X`, `add feature Y`, `create Z`, `I want the system to do W`.

## Graph before

Varies. Could be a fresh module with nothing, or a well-populated graph where the feature adds to existing structure.

## Graph after

A complete subgraph: concepts, needs, requirements, tasks (all or nearly all complete), test cases (passing), possibly defects (resolved). A baselineable state for the affected modules.

## Agent interaction layer

### Skills

The `krav:feature` skill builds this workflow as a composite orchestrator. It has `disable-model-invocation: true` in its frontmatter because it has large side effects across the full transformation chain. The skill doesn't do the work itself; it sequences other skills: `krav:explore` for concept development, `krav:formalize` for need extraction, `krav:derive` for requirement derivation, `krav:decompose` for task planning, `krav:task` (repeated) for execution, `krav:testcase` and `krav:verify` for verification. The open question of how exactly this composition works (skill references by name, subagent delegation, or inline instruction) appears in the skills index.

Preprocessing loads the module's current state to determine where in the chain to start. If crystallized concepts already exist, the orchestrator skips exploration. If requirements exist, it skips derivation. The skill body includes checkpoint logic for pausing at transitions that benefit from developer judgment.

### Agents

The `krav-reviewer` subagent handles the review step, bringing isolated context and a critical evaluation posture to deliverables the main agent produced. The `krav-verifier` subagent handles test execution. The orchestrator launches both at the appropriate point in the chain, with the `subagent-context` policy injecting module requirements and deliverables at launch.

### Policies

All five context injection policies are active across a feature build's lifecycle. The `session-context` policy sets the starting state. The `prompt-context` policy handles developer interjections mid-workflow. The `mutation-feedback` policy keeps the agent current as the graph grows (dozens of nodes might get created across the full chain). The `compaction-context` policy preserves graph awareness if the context window compacts during a long feature build, which is likely given the workflow's scope.

The enforcement policies layer in as the workflow progresses. The `task-completion-gate` and `session-completion-gate` policies apply during task execution steps. The `review-completion-gate` applies when subagents run. The `baseline-protection` policy fires if the feature touches baselined modules. The `cli-auto-approve` and `graph-integrity` policies are active throughout.

### Task types

A feature build produces tasks across multiple phases. The typical progression for a substantial feature runs through architecture-phase tasks (`decide-architecture`, `decompose-module`), design-phase tasks (`design-api`, `design-data-model`), coding-phase tasks (`implement-feature`, `implement-tests`, `write-documentation`), verification-phase tasks (`review-code`, `execute-tests`), and possibly validation-phase tasks (`accept-user`). The specific mix depends on the feature's complexity and how far the developer wants to go in a single session.

## Open questions

**How does the agent manage the sequence?** Does it run through the full chain in one session, or does it break into multiple sessions with the developer re-engaging between stages? The former is ambitious (context window limits). The latter requires the agent to know where it left off and what's next.

**When should the agent pause for developer input?** Some transitions benefit from human judgment: crystallizing a concept (the developer decides), validating needs (the developer confirms), approving requirements (the developer accepts). The agent shouldn't blast through the whole chain without checkpoints. But too many pauses make the workflow feel bureaucratic.

**How much of the chain can the agent skip?** If the developer has a clear, simple feature in mind, concept exploration or formal need derivation may not apply. `Add a --verbose flag` doesn't need an INCOSE-grade transformation chain. The agent needs to calibrate ceremony to feature complexity, the same question as the "ceremony spectrum" from the index, applied to a specific workflow.

**Scope creep detection.** During coding, the agent might discover that the feature is bigger than expected or requires changes to existing modules. Should it flag this and pause, or expand the scope automatically? The answer probably depends on how far the expansion goes.

**How does this interact with existing graph content?** If needs and requirements already exist for the module, the new feature's needs and requirements should coexist. The agent needs to check for conflicts and overlaps (is there already a requirement about this?) rather than creating duplicates.
