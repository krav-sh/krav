# Competency questions

## Overview

Competency questions define what the knowledge graph must be able to answer. They serve as the requirements specification for the ontology itself — if the graph cannot answer a competency question, the schema or predicates need to change.

Each question maps to one or more [query patterns](query-patterns.md) that realize it as a graph traversal.

## Traceability

**CQ-T1**: Why does this requirement exist?
> Follow `derivesFrom` chain from REQ → NED → CON. The full ancestry reveals the stakeholder need and the concept that originated it.

**CQ-T2**: What is the full derivation chain from a concept to its requirements?
> Follow `derivesFrom` descendants from CON through NED to REQ. Shows how exploration became formalized obligations.

**CQ-T3**: What requirements derive from this need?
> Find all REQ nodes where `derivesFrom` includes the given NED. Direct derivation, not transitive.

**CQ-T4**: What needs does this concept inform?
> Find all NED nodes where `derivesFrom` includes the given CON.

**CQ-T5**: What is the complete traceability chain for this node?
> Bidirectional traversal: ancestors (via `derivesFrom`) and descendants (via `derivesFrom` inverse), plus `verifiedBy`, `module`, and task connections. The full web of relationships for any node.

## Coverage

**CQ-C1**: Which requirements lack verification?
> Find REQ nodes with no `verifiedBy` edges. These are unverified requirements.

**CQ-C2**: What is the verification coverage for this module?
> For a module's requirement set: count of requirements with at least one passing VRF vs total requirements. Can scope to subtree for hierarchical coverage.

**CQ-C3**: Which verifications are failing?
> Find VRF nodes with `status = failing`. Cross-reference their verified requirements to assess impact.

**CQ-C4**: What requirements are verified but the verification is not passing?
> Find REQ nodes where `verifiedBy` targets exist but all have `status != passing`.

## Impact analysis

**CQ-I1**: What is affected if this need changes?
> Follow `derivesFrom` descendants from NED to find all derived REQ nodes. Then follow `verifiedBy` to find affected VRF nodes. Follow `module` to find affected modules. The transitive closure shows full downstream impact.

**CQ-I2**: What becomes suspect if this node is modified?
> Suspect propagation follows `derivesFrom` (downstream), `verifiedBy` (requirements to verifications), and `allocatesTo` (parent to child module requirements). Any edge crossing a modified node becomes suspect.

**CQ-I3**: What defects exist for this node?
> Find DEF nodes where `subject` points at the given node.

**CQ-I4**: How many open defects block this module's advancement?
> Find DEF nodes where `module` is the given MOD and `severity` is critical or major and `status` is not closed/rejected/deferred.

## Progress

**CQ-P1**: What tasks are ready to start?
> Find TSK nodes where `status` is not complete and all `dependsOn` targets have `status = complete`. The "ready set" of the task DAG.

**CQ-P2**: What is the critical path for this milestone?
> Find the longest path through incomplete tasks in the `dependsOn` DAG that leads to the milestone task.

**CQ-P3**: What tasks are blocking this task?
> Transitive closure of `dependsOn` from the given task, filtered to incomplete tasks.

**CQ-P4**: What is the overall progress for this module?
> For a module: ratio of complete tasks to total tasks, optionally broken down by process phase.

**CQ-P5**: What tasks belong to this module?
> Find TSK nodes where `module` is the given MOD or any descendant MOD in the `childOf` subtree.

## Phase management

**CQ-PH1**: Can this module advance to the next phase?
> Check: all tasks for the current phase are complete, no blocking defects (critical/major with status open/confirmed), all review tasks have acceptable dispositions, parent module is at or ahead of the target phase.

**CQ-PH2**: Which modules are behind their parent's phase?
> For each MOD with a `childOf` parent, compare phases. Children behind their parent may need attention.

**CQ-PH3**: What phase is this module in?
> Direct property lookup: `phase` on the MOD node.

**CQ-PH4**: What tasks exist for each phase of this module?
> Find TSK nodes where `module` is the given MOD, grouped by `processPhase`.

## Structural integrity

**CQ-S1**: Are there orphaned nodes?
> Find nodes with no incoming or outgoing `derivesFrom`, `module`, `childOf`, or other structural edges (excluding root modules). Orphans may indicate incomplete modeling.

**CQ-S2**: Are there dangling references?
> Find predicate targets (values of `derivesFrom`, `module`, `childOf`, `dependsOn`, etc.) that don't correspond to existing nodes.

**CQ-S3**: Are there cycles in the derivation graph?
> Check the `derivesFrom` subgraph for cycles. Derivation must be a DAG.

**CQ-S4**: Are there cycles in the task dependency graph?
> Check the `dependsOn` subgraph for cycles. Task dependencies must be a DAG.

**CQ-S5**: Does any module have multiple parents?
> Check that every `childOf` relationship produces a tree, not a DAG. Each module has at most one parent.

## Audit

**CQ-A1**: What changed between two baselines?
> Semantic diff: reconstruct graph state at each baseline's commit SHA, compare nodes and relationships. Report additions, modifications, removals, and status changes.

**CQ-A2**: What examination found this defect?
> Follow `detectedBy` from DEF to TSK. The task is the examination activity.

**CQ-A3**: What defects were found during this review?
> Find DEF nodes where `detectedBy` points at the given review task.

**CQ-A4**: What is the defect density for this module?
> Count of DEF nodes where `module` is the given MOD, optionally broken down by category, severity, or phase detected.

**CQ-A5**: What deferred defects target this baseline or phase?
> Find DEF nodes where `status = deferred` and `deferralTarget` matches the given baseline or phase.

## Cross-cutting

**CQ-X1**: What is the full context for this node?
> Aggregate: the node's properties, its ancestry (`derivesFrom` chain), its module (`module`), its verifications (`verifiedBy`), its defects (`subject` incoming), its tasks, and its suspect links. The complete picture for any entity.

**CQ-X2**: What are the suspect links in this module's subtree?
> Find all edges marked suspect within the `childOf` subtree of the given module.

**CQ-X3**: How does this concept connect to deliverables?
> Trace: CON → NED (via `derivesFrom`) → REQ (via `derivesFrom`) → TSK (via decomposition) → deliverables (via task `deliverables` array). Shows the path from idea to implemented artifact.
