# Constraints

## Overview

Constraints are structural invariants that must always hold in the RDF graph. They define the conditions under which the graph is well-formed. If any constraint fails, the graph is in an invalid state and the system must reject the operation that caused the failure. Constraints govern RDF class membership, object property targets, graph topology (trees, DAGs), and cross-entity consistency.

## Identifier constraints

**C-ID1**: identifier format. Every `@id` must match the pattern `PREFIX-NANOID` where PREFIX is a variable-length uppercase prefix (2-4 characters) from the [node type taxonomy](schema.md#node-type-taxonomy) and `NANOID` is an 8-character Crockford Base32 string.

**C-ID2**: type-prefix consistency. The type prefix of `@id` must be consistent with `@type`. A node with `@id: "CON-K7M3NP2Q"` must have `@type: "Concept"`. The mapping is:

| Prefix | `@type` |
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

**C-ID3**: uniqueness. No two nodes may share the same `@id`.

## Referential integrity

**C-REF1**: predicate targets must exist. Every `{"@id": "TARGET-ID"}` value in a relationship property must correspond to an existing node in the graph. Dangling references are not permitted.

**C-REF2**: predicate domain/range. Each predicate has valid source and target types defined in the [domain/range matrix](predicates.md#domainrange-matrix). A `childOf` property on a non-MOD node, or a `childOf` value pointing at a non-MOD node, is invalid.

## Structural constraints

**C-TREE1**: `childOf` forms a tree. Each MOD node has at most one `childOf` edge. The subgraph of `childOf` edges must have no cycles. Exactly one root module (a MOD with no `childOf`) must exist.

**C-DAG1**: `derivesFrom` forms a DAG. The subgraph of all `derivesFrom` edges must have no cycles. Needs derive from concepts; requirements derive from needs or other requirements. A node cannot transitively derive from itself.

**C-DAG2**: `dependsOn` forms a DAG. The subgraph of all `dependsOn` edges must have no cycles. A task cannot transitively depend on itself.

**C-SINGLE1**: `module` is single-valued. Each node that has a `module` property has exactly one value. A need, requirement, verification, task, or defect belongs to exactly one module.

**C-SINGLE2**: `childOf` is single-valued. Each module has at most one parent.

**C-SINGLE3**: `subject` is single-valued. Each defect has at most one subject.

**C-SINGLE4**: `detectedBy` is single-valued. At most one examination task detects each defect.

**C-MULTI1**: `stakeholder` must be present and supports multiple values. Each need must reference at least one stakeholder via the `stakeholder` property. A need may reference multiple stakeholders when parties share the expectation.

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

**C-SUSPECT1**: suspect triggering. When someone modifies a node's `statement`, `status`, or key semantic properties, outgoing `derivesFrom`, `verifiedBy`, and `allocatesTo` edges from downstream nodes become suspect. The rules are:

- Modifying a CON node marks `derivesFrom` edges on NED nodes that derive from it
- Modifying a NED node marks `derivesFrom` edges on REQ nodes that derive from it
- Modifying a REQ node marks `derivesFrom` edges on child REQ nodes, `verifiedBy` edges, and `allocatesTo` edges

**C-SUSPECT2**: suspect propagation is non-transitive by default. Modifying a CON marks the CON→NED edges as suspect but does not automatically mark the NED→REQ edges. A reviewer must examine each suspect link and decide whether to propagate further, clear the flag, or create a defect.

**C-SUSPECT3**: suspect links do not auto-generate defects. Suspect status is a flag for human review, not an automatic quality gate. See [lifecycle coordination](lifecycle-coordination.md) for the review workflow.

## Baseline integrity

**C-BSL1**: baseline commit validity. A baseline's `commitSha` must reference a valid git commit from which the system can reconstruct the graph state.

**C-BSL2**: baseline module scope. A baseline's `module` must be a valid MOD node. The baseline captures the state of that module's subtree (all descendants via `childOf`).

**C-BSL3**: baseline approval. Only baselines with `status = approved` serve as official reference points for phase gates and change auditing.
