# Competency questions

## Overview

Competency questions define what the knowledge graph must be able to answer. They are a standard ontology engineering technique for RDF knowledge graphs: they serve as the requirements specification for the ontology itself. If the graph cannot answer a competency question, the schema or predicates need to change.

Each question maps to one or more [query patterns](query-patterns.md) that realize it as a graph traversal.

## Traceability

**CQ-T1**: why does this requirement exist?
> Follow `derivesFrom` chain from REQ → NED → CON. The full ancestry reveals the stakeholder need and the concept that originated it.

**CQ-T2**: what is the full derivation chain from a concept to its requirements?
> Follow `derivesFrom` descendants from CON through NED to REQ. Shows how exploration became formalized obligations.

**CQ-T3**: what requirements derive from this need?
> Find all REQ nodes where `derivesFrom` includes the given NED. Direct derivation, not transitive.

**CQ-T4**: what needs does this concept inform?
> Find all NED nodes where `derivesFrom` includes the given CON.

**CQ-T5**: what is the complete traceability chain for this node?
> Bidirectional traversal: ancestors (via `derivesFrom`) and descendants (via `derivesFrom` inverse), plus `verifiedBy`, `module`, and task connections. The full web of relationships for any node.

**CQ-T6**: which tasks satisfy this requirement?
> Find TASK nodes where `implements` includes the given REQ.

**CQ-T7**: which requirements does this task address?
> Direct property lookup: `implements` on the TASK node.

## Coverage

**CQ-C1**: which requirements lack verification?
> Find REQ nodes with no `verifiedBy` edges. These requirements lack verification.

**CQ-C2**: what is the verification coverage for this module?
> For a module's requirement set: count of requirements with at least one passing VRF vs total requirements. Can scope to subtree for hierarchical coverage.

**CQ-C3**: which verifications are failing?
> Find VRF nodes with `status = failing`. Cross-reference their verified requirements to assess impact.

**CQ-C4**: what requirements have verifications that do not pass?
> Find REQ nodes where `verifiedBy` targets exist but none have `status = passing`.

**CQ-C5**: which requirements have no coding tasks?
> Find REQ nodes with no incoming `implements` edges from any TASK.

## Impact analysis

**CQ-I1**: what does a change to this need affect?
> Follow `derivesFrom` descendants from NED to find all derived REQ nodes. Then follow `verifiedBy` to find affected VRF nodes. Follow `module` to find affected modules. The transitive closure shows full downstream impact.

**CQ-I2**: what becomes suspect if someone modifies this node?
> Suspect propagation follows `derivesFrom` (downstream), `verifiedBy` (requirements to verifications), and `allocatesTo` (parent to child module requirements). Any edge crossing a modified node becomes suspect.

**CQ-I3**: what defects exist for this node?
> Find DEF nodes where `subject` points at the given node.

**CQ-I4**: how many open defects block this module's advancement?
> Find DEF nodes where `module` is the given MOD and `severity` is critical or major and `status` is not closed/rejected/deferred.

## Progress

**CQ-P1**: what tasks are ready to start?
> Find TSK nodes where `status` is not complete and all `dependsOn` targets have `status = complete`. The "ready set" of the task DAG.

**CQ-P2**: what is the critical path for this milestone?
> Find the longest path through incomplete tasks in the `dependsOn` DAG that leads to the milestone task.

**CQ-P3**: what tasks are blocking this task?
> Transitive closure of `dependsOn` from the given task, filtered to incomplete tasks.

**CQ-P4**: what is the progress for this module?
> For a module: ratio of complete tasks to total tasks, optionally broken down by process phase.

**CQ-P5**: what tasks belong to this module?
> Find TSK nodes where `module` is the given MOD or any descendant MOD in the `childOf` subtree.

## Phase management

**CQ-PH1**: can this module advance to the next phase?
> Check: all tasks for the current phase are complete, no blocking defects (critical/major with status open/confirmed), all review tasks have acceptable dispositions.

**CQ-PH2**: which modules are at an earlier phase than their siblings?
> For each set of sibling MODs (same `childOf` parent), compare phases. Modules at earlier phases than their siblings may need attention or may simply be on independent timelines.

**CQ-PH3**: what phase is this module in?
> Direct property lookup: `phase` on the MOD node.

**CQ-PH4**: what tasks exist for each phase of this module?
> Find TSK nodes where `module` is the given MOD, grouped by `processPhase`.

## Structural integrity

**CQ-S1**: are there orphaned nodes?
> Find nodes with no incoming or outgoing `derivesFrom`, `module`, `childOf`, or other structural edges (excluding root modules). Orphans may indicate incomplete modeling.

**CQ-S2**: are there dangling references?
> Find predicate targets (values of `derivesFrom`, `module`, `childOf`, `dependsOn`, etc.) that don't correspond to existing nodes.

**CQ-S3**: are there cycles in the derivation graph?
> Check the `derivesFrom` subgraph for cycles. Derivation must be a DAG.

**CQ-S4**: are there cycles in the task dependency graph?
> Check the `dependsOn` subgraph for cycles. Task dependencies must be a DAG.

**CQ-S5**: does any module have multiple parents?
> Check that every `childOf` relationship produces a tree, not a DAG. Each module has at most one parent.

## Audit

**CQ-A1**: what changed between two baselines?
> Semantic diff: reconstruct graph state at each baseline's commit SHA, compare nodes and relationships. Report additions, modifications, removals, and status changes.

**CQ-A2**: what examination found this defect?
> Follow `detectedBy` from DEF to TSK. The task is the examination activity.

**CQ-A3**: what defects did this review find?
> Find DEF nodes where `detectedBy` points at the given review task.

**CQ-A4**: what is the defect density for this module?
> Count of DEF nodes where `module` is the given MOD, optionally broken down by category, severity, or phase detected.

**CQ-A5**: what deferred defects target this baseline or phase?
> Find DEF nodes where `status = deferred` and `deferralTarget` matches the given baseline or phase.

**CQ-A6**: who created this node?
> Follow `attributedTo` from the node to its AGT-* or DEV-* creator.

**CQ-A7**: what did this agent produce?
> Find all nodes where `attributedTo` points at the given AGT-* or DEV-*.

**CQ-A8**: which task created this node?
> Follow `generatedBy` from the node to its TASK-* creator.

## Cross-cutting

**CQ-X1**: what is the full context for this node?
> Aggregate: the node's properties, its ancestry (`derivesFrom` chain), its module (`module`), its verifications (`verifiedBy`), its defects (`subject` incoming), its tasks, and its suspect links. The complete picture for any entity.

**CQ-X2**: what are the suspect links in this module's subtree?
> Find all edges marked suspect within the `childOf` subtree of the given module.

**CQ-X3**: how does this concept connect to deliverables?
> Trace: CON → NED (via `derivesFrom`) → REQ (via `derivesFrom`) → TASK (via `implements`) → deliverables (via task `deliverables` array). Shows the path from idea to implemented artifact.
