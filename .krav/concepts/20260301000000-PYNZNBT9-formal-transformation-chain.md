# Formal transformation chain

Work in Krav follows a formal transformation chain: concept → need → requirement → task → deliverable. Each step produces `derivesFrom` edges maintaining full traceability from stakeholder intent through to built artifacts.

## The chain

**Concepts** capture exploration and design thinking. A concept represents understanding: what options exist, what tradeoffs matter, what the team decided and why. Concepts have no obligations attached; they're the raw material from which formal expectations emerge.

**Needs** formalize stakeholder expectations. A need says "stakeholder X expects Y," a validated expectation that is not yet an engineering obligation. Different stakeholders can derive different needs from the same concept. The distinction between needs and requirements follows the INCOSE Needs/Requirements Manual: needs are expectations, requirements are obligations.

**Requirements** state verifiable obligations. A requirement says "the system shall do Z" in language precise enough to test against. Needs produce requirements, modules own them, and test cases verify them. The shift from need to requirement is the shift from "what the stakeholder expects" to "what the engineering team commits to delivering."

**Tasks** break requirements into executable work. Each task is an atomic unit in a dependency DAG, with a process phase (architecture, design, coding, etc.) that aligns with its module's lifecycle. Tasks produce deliverables: commits, files, test results, documents.

## Why the chain matters

The chain answers two questions at any point in the project. Forward: "why does this task exist?" Trace back through the `implements` and `derivesFrom` edges to find the stakeholder need and concept that justify the work. Backward: "is this need addressed?" Trace forward through `derivesFrom` and `implements` edges to the deliverables and see what the team built.

Without the chain, requirements appear from nowhere and tasks accumulate without justification. With it, every piece of work traces to a reason, and every stakeholder expectation traces to evidence.

## Skipping ceremony

Not every change needs the full chain. Quick fixes, trivial changes, and work where the transformation chain adds no value can skip ceremony. The `krav-quickfix` workflow exists for this purpose. But design decisions, architectural changes, and work that affects many modules should stay traceable. The judgment call is: "would someone later ask why this exists or what depends on it?" If yes, trace it.
