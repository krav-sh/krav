# Deriving requirements from needs

## What

The developer says "derive requirements from these needs" or "what requirements does this need produce?" Validated needs get transformed into verifiable design obligations. The agent produces "shall" statements with verification methods and acceptance criteria, linked to their source needs via `derivesFrom`.

## Why

This is the second formal transformation: need to requirement. It's where stakeholder language ("users need fast feedback") becomes system language ("the parser shall report the first syntax error within 50 ms"). The requirement must be verifiable: if you can't test it, it's not a requirement.

## What happens in the graph

The source need must be in `validated` status. The agent reads the need's statement and rationale, considers the module's context, and produces requirements that collectively satisfy the need.

Each requirement becomes a REQ-* node with `derivesFrom` pointing at the source need, `module` ownership (same module as the need, or a child module per C-CROSS1), a `statement` in "shall" language, a `verificationMethod` (test, inspection, demonstration, analysis), and `verificationCriteria` describing how to verify it.

A single need typically produces one to five requirements. "Users need fast feedback" might produce separate requirements for parsing latency, error message content, and error message format. Complex needs might need decomposition into child needs first.

Requirements start in `draft` status and progress through `proposed` then `approved` as they're reviewed.

## Trigger patterns

`Derive requirements from this need`, `what requirements does NEED-X produce?`, `turn these needs into requirements`, or as a continuation of formalization.

## Graph before

One or more NEED-* nodes in `validated` status with statements and module ownership.

## Graph after

Multiple REQ-* nodes in `draft` status, each with `derivesFrom` edges to source needs, module ownership, shall-statements, verification methods, and acceptance criteria.

## Agent interaction layer

### Skills

The `arci:derive` skill builds this workflow. Preprocessing loads the validated need's statement and rationale, the module's existing requirements (to avoid duplication and identify coverage gaps), and the module's domain context. This is one of the more heavily preprocessed skills because the agent needs substantial context to produce good shall-statements with appropriate verification methods and acceptance criteria.

The skill's instructed commands create REQ-* nodes via `arci requirement create`, link them to their source needs, and set verification methods. The skill instructions emphasize the quality bar for requirements: verifiable, unambiguous, atomic, feasible. This is where the skill body carries the most instructional weight of any graph-building skill.

### Policies

The `mutation-feedback` policy fires after the agent creates each requirement, injecting a running view of how many requirements exist for the source need and what verification methods the agent has assigned. This helps the agent judge coverage: if a need has produced three requirements all verified by test, but one aspect is better verified by inspection, the feedback surfaces that gap.

The `prompt-context` policy injects relevant state when the developer mentions specific needs, requirements, or modules during the derivation conversation. The `graph-integrity` and `cli-auto-approve` policies apply as usual.

### Task types

Derivation doesn't create tasks directly. Requirements are the output, and they become inputs to `arci:decompose` for task creation. The connection is that each requirement's `verificationMethod` influences what task types the decompose skill produces: a requirement verified by test produces `implement-tests` and `execute-tests` tasks, while one verified by inspection produces a `review-code` or `review-design` task.

## Open questions

**How does the agent produce good shall-statements?** The quality bar is high: requirements must be verifiable, unambiguous, atomic, and feasible. The agent needs to understand what "verifiable" means for each verification method. "The parser shall be fast" is not a requirement. "The parser shall report the first syntax error within 50 ms at p99" is. What skill instructions or reference material does the agent need to produce the latter consistently?

**Verification method selection.** Who decides whether a requirement uses test, inspection, demonstration, or analysis? The need doesn't specify this; it's a design decision made during derivation. Should the agent propose methods based on the requirement's nature, or ask the developer?

**How quantitative should requirements be?** Some needs produce requirements with obvious quantitative criteria (latency targets, coverage thresholds). Others are qualitative ("error messages shall include the source location"). The agent needs to push for testability without forcing artificial numbers on things that are better verified by inspection.

**Cross-need requirements.** A single requirement might satisfy multiple needs. "The parser shall produce structured error output" could satisfy both "users need clear error messages" and "integrators need machine-readable error output." Should the agent identify these overlaps and produce shared requirements with multiple `derivesFrom` edges? Or produce separate requirements and let the developer consolidate?

**When to stop deriving.** How does the agent know it's produced enough requirements to satisfy a need? The INCOSE answer is coverage analysis, but at derivation time there are no test cases yet. The heuristic might be: every testable aspect of the need statement should map to at least one requirement.
