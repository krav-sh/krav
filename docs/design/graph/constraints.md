# Constraints

## Overview

Constraints are structural invariants that must always hold in the knowledge graph. They define the conditions under which the graph is well-formed. If any constraint fails, the graph is in an invalid state and the system must reject the operation that caused the failure. Constraints govern node type membership, edge table referential integrity, graph topology (trees, DAGs), and cross-entity consistency. At runtime, DuckDB foreign key constraints and app-level validation enforce these rules.

## Identifier constraints

**C-ID1**: identifier format. Every `id` must match the pattern `PREFIX-NANOID` where PREFIX is a variable-length uppercase prefix (2-4 characters) from the [node type taxonomy](schema.md#node-type-taxonomy) and `NANOID` is an 8-character Crockford Base32 string.

**C-ID2**: type-prefix consistency. The type prefix of `id` must be consistent with `type`. A node with `"id": "CON-K7M3NP2Q"` must have `"type": "Concept"`. The mapping is:

| Prefix | `type` |
|--------|---------|
| CON | Concept |
| MOD | Module |
| NEED | Need |
| REQ | Requirement |
| TC | `TestCase` |
| TASK | Task |
| DEF | Defect |
| BSL | Baseline |
| STK | Stakeholder |
| TP | TestPlan |
| DEV | Developer |
| AGT | Agent |

**C-ID3**: uniqueness. No two nodes may share the same `id`. Each vertex table enforces this as a primary key constraint.

## Referential integrity

**C-REF1**: edge targets must exist. Every `src` and `dst` value in an edge table must correspond to an existing node in the appropriate vertex table. DuckDB foreign key constraints enforce this at the database level.

**C-REF2**: predicate domain/range. Each edge table has valid source and target types defined in the [domain/range matrix](predicates.md#domainrange-matrix). A `child_of` row with a non-MOD source or non-MOD target is invalid.

## Structural constraints

**C-TREE1**: `child_of` forms a tree. Each MOD node has at most one row in the `child_of` edge table. The subgraph of `child_of` edges must have no cycles. Exactly one root module (a MOD with no `child_of` row) must exist.

**C-DAG1**: `derives_from` forms a DAG. The subgraph of all `derives_from` edges must have no cycles. Needs derive from concepts; requirements derive from needs or other requirements. A node cannot transitively derive from itself.

**C-DAG2**: `depends_on` forms a DAG. The subgraph of all `depends_on` edges must have no cycles. A task cannot transitively depend on itself.

**C-SINGLE1**: `module` is single-valued. Each node has at most one row in the `module` edge table. A need, requirement, test case, task, or defect belongs to exactly one module.

**C-SINGLE2**: `child_of` is single-valued. Each module has at most one parent.

**C-SINGLE3**: `subject` is single-valued. Each defect has at most one row in the `subject` edge table.

**C-SINGLE4**: `detected_by` is single-valued. At most one examination task detects each defect.

**C-MULTI1**: `stakeholder` must be present and supports multiple values. Each need must have at least one row in the `stakeholder` edge table. A need may reference multiple stakeholders when parties share the expectation.

**C-SINGLE5**: `generates` is single-valued. Each defect generates at most one remediation task.

## Phase constraints

**C-PH2**: phase advancement preconditions. A module's phase can only advance to the next phase if:

- All tasks for the current phase (where `module` is this MOD and `processPhase` matches the current phase) have `status = complete`
- No blocking defects exist (DEF nodes where `module` is this MOD and `severity` is critical or major and `status` is open or confirmed)
- All review tasks for the current phase have acceptable dispositions

**C-PH3**: task process phase constraint. A task's `processPhase` must be at or behind its module's `phase`. Users cannot create tasks for phases the module hasn't reached yet.

## Cross-entity consistency

**C-CROSS1**: derivation chain module alignment. When a requirement derives from a need, both must belong to the same module or the requirement must belong to a child module of the need's module. Derivation chains must not cross unrelated module boundaries.

**C-CROSS2**: verification module alignment. A verification's module must match or be a descendant of the module of the requirement it verifies. Component-level verifications verify component requirements; system-level verifications can verify system-level requirements.

**C-CROSS3**: allocation target validity. `allocatesTo` targets must be child modules of the module that owns the requirement. The system allocates requirements only to child modules, not to siblings or ancestors.

**C-CROSS4**: defect module consistency. A defect's `module` should match the module of its `subject` node (when `subject` has a value). If the subject is a requirement owned by MOD-A, the defect should also belong to MOD-A.

## Suspect link rules

**C-SUSPECT1**: suspect triggering. When someone modifies a node's `statement`, `status`, or key semantic properties, downstream `derives_from`, `verified_by`, and `allocates_to` edges become suspect. The rules are:

- Modifying a CON node marks `derives_from` edges where the CON is the `dst` (source of the derivation)
- Modifying a NEED node marks `derives_from` edges where the NEED is the `dst`
- Modifying a REQ node marks `derives_from` edges where the REQ is the `dst`, plus `verified_by` edges where the REQ is the `src`, plus `allocates_to` edges where the REQ is the `src`

**C-SUSPECT2**: suspect propagation is non-transitive by default. Modifying a CON marks the CON→NED edges as suspect but does not automatically mark the NED→REQ edges. A reviewer must examine each suspect link and decide whether to propagate further, clear the flag, or create a defect.

**C-SUSPECT3**: suspect links do not auto-generate defects. Suspect status is a flag for human review, not an automatic quality gate. See [lifecycle coordination](lifecycle-coordination.md) for the review workflow.

## Baseline integrity

**C-BSL1**: baseline commit validity. A baseline's `commitSha` must reference a valid git commit from which the system can reconstruct the graph state by reading the NDJSON files at that commit.

**C-BSL2**: baseline module scope. A baseline's `module` must be a valid MOD node. The baseline captures the state of that module's subtree (all descendants via `childOf`).

**C-BSL3**: baseline approval. Only baselines with `status = approved` serve as official reference points for phase gates and change auditing.
