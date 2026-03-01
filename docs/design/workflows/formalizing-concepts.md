# Formalizing concepts into needs

## What

The developer says "the team decided the error handling approach; capture what stakeholders need" or "formalize this concept." The agent transforms a crystallized concept into structured stakeholder expectations. The agent walks through the project's active stakeholders, extracts expectations from the concept's content, and creates NEED-* nodes linked back to the concept via `derivesFrom`.

## Why

This is the first formal transformation in the INCOSE chain: concept to need. It's where unstructured thinking becomes structured, validatable stakeholder expectations. The transformation is important because it forces the question: who cares about this and what do they actually need? That question is easy to skip when jumping straight from idea to code.

## What happens in the graph

The source concept must be in `crystallized` status. The agent reads the concept's prose content to understand what the team decided.

For each active stakeholder in the project, the agent considers whether the concept implies expectations for that party. A concept about error handling might produce: "Users need clear error messages that identify the problem location" (linked to the CLI end user stakeholder), "Contributors need documented error types for consistent handling" (linked to the contributor stakeholder), "Integrators need machine-readable error output" (linked to the tool integrator stakeholder). Some needs serve multiple stakeholders and reference more than one STK-* node.

Each expectation becomes a NEED-* node with `derivesFrom` pointing at the source concept, `module` pointing at the relevant module, `stakeholder` referencing one or more STK-* nodes, a `statement` in stakeholder language, and an initial `priority` (MoSCoW).

The concept transitions from `crystallized` to `formalized`. The needs start in `draft` status. Validation (confirming these are real stakeholder expectations) is a separate step, though the developer might validate immediately if they're the primary stakeholder.

## Trigger patterns

`Formalize this concept`, `capture the needs for X`, `what do stakeholders need from this?`, `turn this into needs`, or as a natural continuation of crystallizing a concept.

## Graph before

At least one CON-* node in `crystallized` status with prose content. The module it `informs` should exist.

## Graph after

The CON-* node transitions to `formalized`. Multiple NEED-* nodes (typically 2-6) in `draft` status, each with `derivesFrom` edges to the concept, `module` ownership, stakeholder class, and a statement.

## Agent interaction layer

### Skills

The `krav:formalize` skill builds this workflow. Preprocessing loads the crystallized concept's prose content, its module context (the module the concept `informs` and that module's existing needs), and the project's active stakeholders (STK-* nodes). This gives the agent everything it needs to identify expectations without additional graph queries.

The skill's instructed commands handle creating NEED-* nodes via `krav need create`, linking them back to the source concept and to relevant stakeholders, and transitioning the concept to `formalized` status when the agent has captured all needs. The skill instructions guide the agent through each active stakeholder, using the stakeholder's description and concerns to identify relevant expectations from the concept.

### Policies

The `mutation-feedback` policy is particularly useful during formalization because the workflow produces multiple graph nodes in sequence. After the agent creates each need, the feedback injection tells it how many needs the agent has captured so far and which stakeholders it covers. This running tally helps the agent judge when formalization is complete without re-querying the graph.

The `prompt-context` policy fires if the developer mentions specific concepts or modules during the formalization conversation, injecting their current state. The `graph-integrity` and `cli-auto-approve` policies ensure all need creation goes through the CLI without friction.

### Task types

Formalization doesn't create tasks. The workflow produces NEED-* nodes, which become inputs to the `krav:derive` workflow for producing requirements. Tasks only appear later when requirements get decomposed into work.

## Open questions

**How interactive is formalization?** The design describes `krav concept formalize` as an "interactive process." But what does the agent do vs. what does the developer do? Does the agent propose needs and the developer approves/edits? Does the agent ask "what do users need from this?" for each stakeholder class? Does it draft all needs at once and present them for review?

**Which stakeholders are relevant?** Not every concept affects all stakeholders. An internal algorithm concept might only produce needs for contributors and maintainers. Should the agent filter stakeholders based on the concept type and module level, or always walk through every active stakeholder? The stakeholder's `concerns` field provides a signal (if the concept doesn't overlap with any of a stakeholder's concerns, the agent can skip them), but the agent may also discover unexpected relevance.

**How specific should needs be?** A concept about caching could produce one broad need ("users need fast response times") or multiple specific ones (`users need sub-second responses for cached data`, `users need stale data indicators`, `operators need cache hit rate visibility`). The INCOSE guidance is that needs should be individually validatable, but the agent needs heuristics for the right level of specificity.

**Validation as part of formalization or separate?** The design has needs starting in `draft` and moving to `validated` as a separate step. But if the developer is the primary stakeholder and they're sitting right there during formalization, should the agent offer to validate immediately?

**Concept-to-need cardinality.** A single concept might produce needs across multiple modules (if it's a cross-cutting concept). The `module` field on needs requires each need to belong to exactly one module. How does the agent handle a concept that `informs` the root module but produces needs on child modules?
