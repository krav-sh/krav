# Constraints

## Overview

Constraints are structural rules that must always hold in the knowledge graph. They define the graph invariants enforced by validation — if any constraint is violated, the graph is in an invalid state and the operation that caused the violation must be rejected.

Constraints are checked by the core validation layer as pure functions that take graph state as input and return error lists. The service layer orchestrates validation before persistence.

## Identifier constraints

**C-ID1**: Identifier format. Every `@id` must match the pattern `TYPE-NANOID` where TYPE is a 3-character prefix from the [node type taxonomy](schema.md#node-type-taxonomy) and NANOID is an 8-character Crockford Base32 string.

**C-ID2**: Type-prefix consistency. The type prefix of `@id` must be consistent with `@type`. A node with `@id: "CON-K7M3NP2Q"` must have `@type: "Concept"`. The mapping is:

| Prefix | `@type` |
|--------|---------|
| CON | Concept |
| MOD | Module |
| NED | Need |
| REQ | Requirement |
| VRF | Verification |
| TSK | Task |
| DEF | Defect |
| BSL | Baseline |

**C-ID3**: Uniqueness. No two nodes may share the same `@id`.

## Referential integrity

**C-REF1**: Predicate targets must exist. Every `{"@id": "TARGET-ID"}` value in a relationship property must correspond to an existing node in the graph. Dangling references are not permitted.

**C-REF2**: Predicate domain/range. Each predicate has valid source and target types defined in the [domain/range matrix](predicates.md#domainrange-matrix). A `childOf` property on a non-MOD node, or a `childOf` value pointing at a non-MOD node, is invalid.

## Structural constraints

**C-TREE1**: `childOf` forms a tree. Each MOD node has at most one `childOf` edge. The subgraph of `childOf` edges must have no cycles. There is exactly one root module (a MOD with no `childOf`).

**C-DAG1**: `derivesFrom` forms a DAG. The subgraph of all `derivesFrom` edges must have no cycles. Needs derive from concepts; requirements derive from needs or other requirements. A node cannot transitively derive from itself.

**C-DAG2**: `dependsOn` forms a DAG. The subgraph of all `dependsOn` edges must have no cycles. A task cannot transitively depend on itself.

**C-SINGLE1**: `module` is single-valued. Each node that has a `module` property has exactly one value. A need, requirement, verification, task, or defect belongs to exactly one module.

**C-SINGLE2**: `childOf` is single-valued. Each module has at most one parent.

**C-SINGLE3**: `subject` is single-valued. Each defect has at most one subject.

**C-SINGLE4**: `detectedBy` is single-valued. Each defect was detected by at most one examination task.

**C-SINGLE5**: `generates` is single-valued. Each defect generates at most one remediation task.

## Phase hierarchy constraint

**C-PH1**: Child module phase constraint. A child module's `phase` must be at or behind its parent module's `phase` in the phase ordering: architecture < design < implementation < integration < verification < validation. If `MOD-PARENT.phase = implementation`, then for any MOD where `childOf = MOD-PARENT`, that MOD's `phase` must be ≤ implementation.

**C-PH2**: Phase advancement preconditions. A module's phase can only advance to the next phase if:
- All tasks for the current phase (where `module` is this MOD and `processPhase` matches the current phase) have `status = complete`
- No blocking defects exist (DEF nodes where `module` is this MOD and `severity` is critical or major and `status` is open or confirmed)
- The parent module (if any) is at or ahead of the target phase
- All review tasks for the current phase have acceptable dispositions

**C-PH3**: Task process phase constraint. A task's `processPhase` must be at or behind its module's `phase`. Tasks cannot be created for phases the module hasn't reached yet.

## Cross-entity consistency

**C-CROSS1**: Derivation chain module alignment. When a requirement derives from a need, both must belong to the same module or the requirement must belong to a child module of the need's module. This ensures derivation chains don't cross unrelated module boundaries.

**C-CROSS2**: Verification module alignment. A verification's module must match or be a descendant of the module of the requirement it verifies. Component-level verifications verify component requirements; system-level verifications can verify system-level requirements.

**C-CROSS3**: Allocation target validity. `allocatesTo` targets must be child modules of the module that owns the requirement. Requirements can only be allocated to child modules, not to siblings or ancestors.

**C-CROSS4**: Defect module consistency. A defect's `module` should match the module of its `subject` node (when `subject` is set). If the subject is a requirement owned by MOD-A, the defect should also be owned by MOD-A.

## Suspect link rules

**C-SUSPECT1**: Suspect triggering. When a node's `statement`, `status`, or key semantic properties are modified, outgoing `derivesFrom`, `verifiedBy`, and `allocatesTo` edges from downstream nodes become suspect. Specifically:

- Modifying a CON node marks `derivesFrom` edges on NED nodes that derive from it
- Modifying a NED node marks `derivesFrom` edges on REQ nodes that derive from it
- Modifying a REQ node marks `derivesFrom` edges on child REQ nodes, `verifiedBy` edges, and `allocatesTo` edges

**C-SUSPECT2**: Suspect propagation is non-transitive by default. Modifying a CON marks the CON→NED edges as suspect but does not automatically mark the NED→REQ edges. A reviewer must examine each suspect link and decide whether to propagate further, clear the flag, or create a defect.

**C-SUSPECT3**: Suspect links do not auto-generate defects. Suspect status is a flag for human review, not an automatic quality gate. See [lifecycle coordination](lifecycle-coordination.md) for the review workflow.

## Baseline integrity

**C-BSL1**: Baseline commit validity. A baseline's `commitSha` must reference a valid git commit that contains a `graph.jsonlt` file. The graph state must be reconstructable from that commit.

**C-BSL2**: Baseline module scope. A baseline's `module` must be a valid MOD node. The baseline captures the state of that module's subtree (all descendants via `childOf`).

**C-BSL3**: Baseline approval. Only baselines with `status = approved` are considered official reference points for phase gates and change auditing.

## Validation approach

Constraints are validated at two levels:

**Per-operation validation**: When a node is created or updated, the constraints relevant to that operation are checked immediately. This includes identifier format (C-ID1, C-ID2), referential integrity (C-REF1, C-REF2), single-value constraints (C-SINGLE1–5), and domain/range checks.

**Graph-wide validation**: Structural constraints that require global analysis (cycle detection for C-DAG1 and C-DAG2, tree validation for C-TREE1, phase hierarchy for C-PH1) are checked when the operation affects the relevant subgraph. A new `dependsOn` edge triggers cycle detection on the task dependency subgraph.

All validation functions are pure — they take graph state as input and return error lists. No side effects, no I/O.
