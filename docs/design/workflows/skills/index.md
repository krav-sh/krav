# Skills

This directory contains design specifications for the skills that Krav ships as part of its Claude Code plugin. Each skill encodes a workflow from the [workflow index](../index.md) as structured instructions that Claude Code follows during execution.

Skills live in the plugin's `claude/skills/` directory and are namespaced as `krav:skill-name` to avoid conflicts with project or personal skills (so the `init` skill gets invoked as `krav:init`). Each skill has YAML frontmatter (tool restrictions, invocation control, hooks, model overrides) and a markdown body with workflow instructions. Skills use `!`command`` preprocessing to inject graph context at load time and instructed commands for interactive graph queries during execution.

## Graph-building skills

These skills create new nodes and edges in the knowledge graph. They tend to involve formal transformations and creative decisions.

| Skill                       | Workflow                                                                  | What it does                                                                                                                                                                                                                           |
| --------------------------- | ------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| init             | [Starting a new project](../starting-a-project.md)                        | Bootstraps root module, module hierarchy, initial concepts and needs from a project description. Interactive: asks the developer about stakeholders, architectural boundaries, and initial scope.                                      |
| explore       | [Exploring a design question](../exploring-a-design-question.md)          | Creates and develops concepts through exploration. Freeform compared to other skills. Records decisions, alternatives considered, and rationale. Transitions concept through `draft` then `exploring` then `crystallized`.        |
| formalize   | [Formalizing concepts into needs](../formalizing-concepts.md)             | Reads a crystallized concept's prose content, walks the project's active stakeholders, extracts expectations as needs with `derivesFrom` edges back to the concept and `stakeholder` edges to relevant STK-* nodes. Preprocessing loads the concept, its module context, and active stakeholders. |
| derive         | [Deriving requirements from needs](../deriving-requirements.md)           | Transforms validated needs into verifiable requirements. Each requirement is more specific and constrained than its parent need, stated as a binding obligation. Preprocessing loads the need and existing requirements in the module. |
| allocate     | [Allocating requirements to child modules](../allocating-requirements.md) | Flows parent requirements down to child modules via `allocatesTo` edges with budgets. Preprocessing loads the parent module's requirements and child module structure.                                                                 |
| decompose   | [Decomposing work into tasks](../decomposing-work.md)                     | Generates a task DAG from a module's requirements. Uses decomposition templates for common patterns. Preprocessing loads module requirements, existing tasks, and available templates.                                                 |
| testcase     | [Creating test cases](../creating-test-cases.md)                          | Creates test case specifications linked to requirements via `verifiedBy`. Specifies verification method, acceptance criteria, and test procedure. Does not write the tests (that's a task).                                        |
| module-add | [Adding a module](../adding-a-module.md)                                  | Creates a new module node, establishes parent-child hierarchy, sets initial phase. Optionally flows down allocated requirements from the parent.                                                                                       |

## Graph-changing skills

These skills transition existing nodes through their lifecycles, record results, and resolve problems.

| Skill                         | Workflow                                                             | What it does                                                                                                                                                                                                                                                                    |
| ----------------------------- | -------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| task               | [Working on a task](../working-on-a-task.md)                         | The core execution skill. Preprocessing injects full task context (requirements, deliverables, dependencies, module domain context). Instructions guide the agent through coding, deliverable recording, and graph updates.                                             |
| feature         | [Building a feature end-to-end](../building-a-feature.md)            | Composite orchestrator that invokes other skills in sequence: explore then formalize then derive then decompose then task (repeated) then verify. Manages the end-to-end flow and decides when to delegate to subagents. `disable-model-invocation: true` since it has large side effects.  |
| quickfix       | [Implementing without ceremony](../implementing-without-ceremony.md) | Minimal-ceremony path. Determines appropriate ceremony level based on what's affected (baselined content? existing requirements? new area?). May create a lightweight task or skip the graph entirely.                                                           |
| review           | [Reviewing work](../reviewing-work.md)                               | Evaluates deliverables against requirements, produces defects for problems found. Runs inside the `krav-reviewer` subagent. Preprocessing loads the module's requirements and deliverable file list. Instructs the agent to create defects via CLI for each problem. |
| verify           | [Running verification](../running-verification.md)                   | Executes test cases and records results on TC-\* nodes. Runs inside the `krav-verifier` subagent. Preprocessing loads the module's test cases and their current results.                                                                                             |
| defect           | [Triaging and fixing defects](../triaging-defects.md)                | Walks through open defects for a module, helps with disposition (confirm, defer, reject), creates remediation tasks for confirmed defects, and verifies fixes.                                                                                                                  |
| suspect         | [Handling suspect links](../handling-suspect-links.md)               | Triages suspect links caused by upstream changes. For each suspect link, evaluates whether the downstream node needs updating, creates a defect, or clears the flag. Preprocessing loads the suspect link report.                                                               |
| advance         | [Advancing a module's phase](../advancing-a-phase.md)                | Checks phase advancement criteria (tasks complete, no blocking defects, verification coverage) and advances the module if met. If criteria aren't met, reports what's blocking.                                                                                                 |
| baseline       | [Creating a baseline](../creating-a-baseline.md)                     | Captures graph state as a named snapshot anchored to the current git commit. Preprocessing loads the current module state for the developer to review before confirming. `disable-model-invocation: true` since baselines are deliberate milestones.                            |
| restructure | [Restructuring modules](../restructuring-modules.md)                 | Reparents, splits, or merges modules. Handles requirement reallocation, task reassignment, and suspect link generation for affected edges. `disable-model-invocation: true` due to broad graph impact.                                                                          |

## Graph-reading skills

These skills query the graph without changing it. They synthesize useful answers from raw graph data.

| Skill | Workflow | What it does |
|-------|----------|-------------|
| status | [Checking project status](../checking-status.md) | Synthesizes current project state: module phases, task progress, defect counts, verification coverage, suspect links pending. Preprocessing loads the graph summary. |
| trace | [Tracing requirements](../tracing-requirements.md) | Walks the derivation chain from any node to explain why it exists and what depends on it. Useful for impact analysis and understanding provenance. |
| diff | [Comparing baselines](../comparing-baselines.md) | Produces a semantic diff between two baselines: nodes added, removed, changed; edges created, broken; status transitions. Runs inside the `krav-analyst` subagent for larger comparisons. |

## Open questions

A few skill design questions remain unresolved.

How skills compose in composite workflows like `feature` is the biggest open question. Does the orchestrator skill reference sub-skills by name? Does it delegate to subagents that have those skills preloaded? Does it just tell the agent to `now follow the formalization workflow`? The answer probably depends on context window budget and whether intermediate steps benefit from isolation.

Which skills should use `context: fork` to run in subagent isolation versus running in the main conversation context? The `review` and `verify` skills target subagent execution, but `formalize` and `derive` might also benefit from isolation to avoid polluting the main context with transformation details.

How thin should graph-reading skills be? `status`, `trace`, and `diff` might be simple enough that a well-formatted CLI output (via `--format agent`) plus a brief skill prompt is sufficient, or they might benefit from richer synthesis instructions.
