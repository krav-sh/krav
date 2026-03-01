# Exploring a design question

## What

The developer raises a design question mid-project: `how should error handling work?`, `what's the right approach for caching?`, `whether a plugin system is needed.` This creates or develops concept nodes without immediately producing needs or requirements. The developer might explore multiple options, weigh tradeoffs, and eventually crystallize a decision, or leave the concept in `exploring` status and come back later.

## Why

Not every design thought needs to immediately become a requirement. The INCOSE model recognizes this: concepts are the exploration phase where thinking happens before commitments happen. This workflow gives design thinking a home in the graph without forcing premature formalization.

The graph value is decision capture. Six months later, when someone asks `why did SQLite win over Postgres?`, the concept node has the answer: what options the team considered, what tradeoffs the team evaluated, and what the team decided.

## What happens in the graph

The agent creates a CON-* node with `conceptType` matching the domain (architectural, technical, operational, interface, process, integration). The concept starts in `draft` and moves to `exploring` as the developer and agent flesh it out.

The concept's prose content file gets populated with context, options considered, tradeoffs, and eventually a decision. The agent might create multiple concepts if the exploration branches (one concept per option considered, with `derivesFrom` edges connecting them).

The concept links to a module via `informs` if it's primarily about a specific subsystem. If it's cross-cutting, `informs` might point at the root module or the agent can omit it.

When the developer says "OK, go with option B," the concept transitions to `crystallized`. If it's time to formalize, this flows into the "formalizing concepts into needs" workflow. If not, the crystallized concept sits in the graph as a reference.

## Trigger patterns

`How should X work?`, `think about Y`, `what are the options for Z?`, `not sure whether to use A or B`, `the team needs to decide on the approach for X`.

## Graph before

An existing project with modules. Possibly related concepts from earlier exploration.

## Graph after

One or more CON-* nodes in `exploring` or `crystallized` status, with prose content files capturing the exploration. Possibly `derivesFrom` edges between concepts if one builds on another, and `informs` edges to modules.

## Agent interaction layer

### Skills

The `arci:explore` skill builds this workflow. Preprocessing loads existing concepts for the module the developer is working in, so the agent has context about what has already gone through exploration. This matters for linking new exploration to prior decisions: if a caching concept depends on an earlier data model concept, preprocessing surfaces that relationship.

The skill's instructed commands let the agent query for related concepts during the conversation, create new CON-* nodes, and transition concept status as the exploration progresses. The skill is freeform compared to transformation skills like `arci:formalize` (it gives the agent latitude to follow the developer's thinking rather than enforcing a rigid sequence).

### Policies

The `prompt-context` policy is the most useful here. When the developer mentions a module name or concept in their message, the hook injects that node's current state as additional context. The agent doesn't have to explicitly query the graph to recall a prior decision about a related concept before engaging with the current question.

The `mutation-feedback` policy fires after concept creation and status transitions, keeping the agent's context current as the exploration produces graph nodes. The `graph-integrity` and `cli-auto-approve` policies operate in the background as usual, ensuring graph mutations go through the CLI without permission friction.

### Task types

Exploration itself doesn't create tasks, but its outcomes often lead to them. A crystallized concept might produce a `decide-architecture` task if the decision needs formal evaluation and recording. If the exploration reveals uncertainty that needs hands-on investigation, a `spike` task captures that as time-boxed work. These tasks aren't created by the `arci:explore` skill directly; they emerge when the developer decides to act on what the exploration revealed, typically through `arci:decompose`.

## Open questions

**How does the agent distinguish exploration from coding requests?** "How should error handling work?" is exploration. "Add error handling" is a coding request. The same words might mean different things depending on context. Does the agent ask for clarification, or does it use signals like the module's current phase and whether relevant concepts already exist?

**When does exploration end?** The developer might explore indefinitely. Should the agent prompt for crystallization after a certain point? Should a concept in `exploring` status for weeks show up as a warning somewhere?

**Multiple options as separate concepts or one concept with sections?** The design supports both: one concept with an "options considered" section in prose, or multiple concepts with `derivesFrom` edges. The former is simpler. The latter gives each option its own lifecycle. Which should the agent default to?

**How much should the agent contribute to exploration?** Should the agent just record what the developer says, or should it actively suggest options, identify tradeoffs, and surface relevant technical considerations? The latter is more useful but risks steering the decision.

**Relationship to existing concepts.** If the project already has a crystallized concept about the data model, and the developer starts exploring caching (which affects the data model), should the agent automatically link the new concept to the existing one via `derivesFrom`? Or wait for the developer to make that connection?
