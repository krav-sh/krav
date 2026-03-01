# Predicates

## Overview

Predicates define the semantic relationships between nodes in the knowledge graph. Each predicate maps to a DuckDB edge table with `src` and `dst` columns (plus optional metadata columns). Each predicate has a specific domain (source types), range (target types), cardinality, structural constraint, and behavior for suspect link propagation.

In the NDJSON serialization, each edge table has its own file (`derives_from.ndjson`, `verified_by.ndjson`, etc.). Each line contains `src` (source node ID) and `dst` (target node ID), plus any metadata columns. Directionality runs from the source node to the target node.

## Predicate classification

### Hierarchical

Predicates that form tree structures with single-parent constraints.

#### childOf

Expresses module hierarchy. A module is a child of another module.

| Aspect | Value |
|--------|-------|
| Domain | MOD |
| Range | MOD |
| Cardinality | 0..1 (single parent) |
| Structure | Tree |
| Inverse | (queried as "children of") |
| Suspect propagation | Phase constraint changes propagate to children |

In `child_of.ndjson`:

```json
{"src": "MOD-A4F8R2X1", "dst": "MOD-OAPSROOT"}
```

### Derivation

Predicates that form DAGs expressing formal transformation chains.

#### derivesFrom

Expresses formal transformation: concepts become needs, needs become requirements, parent requirements produce child requirements.

| Aspect | Value |
|--------|-------|
| Domain | NED, REQ |
| Range | CON (for NED), NED or REQ (for REQ) |
| Cardinality | 1..* (at least one source) |
| Structure | DAG (no cycles) |
| Inverse | (queried as "derives to") |
| Suspect propagation | Modification of source marks downstream `derivesFrom` edges as suspect |

In `derives_from.ndjson`:

```json
{"src": "NEED-B7G3M9K2", "dst": "CON-K7M3NP2Q"}
{"src": "REQ-C2H6N4P8", "dst": "NEED-B7G3M9K2"}
{"src": "REQ-CH1LDR3Q", "dst": "REQ-C2H6N4P8"}
```

### Ownership

Predicates that assign entities to their owning container.

#### Module

Expresses which module owns a node.

| Aspect | Value |
|--------|-------|
| Domain | NED, REQ, TC, TSK, DEF, BSL |
| Range | MOD |
| Cardinality | 1 (exactly one) |
| Structure | Single-value |
| Inverse | (queried as "owned by module") |
| Suspect propagation | None (organizational, not semantic) |

In `module.ndjson`:

```json
{"src": "NEED-B7G3M9K2", "dst": "MOD-A4F8R2X1"}
{"src": "REQ-C2H6N4P8", "dst": "MOD-A4F8R2X1"}
{"src": "TASK-E3K8S6V2", "dst": "MOD-A4F8R2X1"}
```

#### Stakeholder

Expresses which stakeholders have expectations captured by a need.

| Aspect | Value |
|--------|-------|
| Domain | NED |
| Range | STK |
| Cardinality | 1..* (at least one stakeholder) |
| Structure | Unrestricted |
| Inverse | (queried as "needs for stakeholder") |
| Suspect propagation | None (stakeholder identity is organizational, not semantic) |

In `stakeholder.ndjson`:

```json
{"src": "NEED-B7G3M9K2", "dst": "STK-H5N7P3Q9"}
{"src": "NEED-ERR0R002", "dst": "STK-H5N7P3Q9"}
{"src": "NEED-ERR0R002", "dst": "STK-1NT3GR8R"}
```

### Allocation

Predicates that express flow-down of obligations to architectural elements.

#### allocatesTo

Expresses requirement flow-down to child modules. When children must meet a parent module's requirement, `allocatesTo` distributes the obligation, potentially with budgets or partitions.

| Aspect | Value |
|--------|-------|
| Domain | REQ |
| Range | MOD |
| Cardinality | 0..* |
| Structure | Unrestricted |
| Metadata | Optional budget, partition, or other allocation parameters |
| Inverse | (queried as "allocated from") |
| Suspect propagation | Modification of the requirement marks `allocatesTo` edges as suspect |

In `allocates_to.ndjson`:

```json
{"src": "REQ-H4J7N2P5", "dst": "MOD-A4F8R2X1", "budget": "50ms"}
{"src": "REQ-H4J7N2P5", "dst": "MOD-B9G3M7K2", "budget": "30ms"}
```

### Dependency

Predicates that express ordering constraints between work items.

#### dependsOn

Expresses task ordering. A task depends on other tasks completing before it can start.

| Aspect | Value |
|--------|-------|
| Domain | TSK |
| Range | TSK |
| Cardinality | 0..* |
| Structure | DAG (no cycles) |
| Inverse | (queried as "blocks") |
| Suspect propagation | None (dependency is structural, not semantic) |

In `depends_on.ndjson`:

```json
{"src": "TASK-E3K8S6V2", "dst": "TASK-G5M2R8X4"}
```

### Verification

Predicates that connect obligations to evidence.

#### verifiedBy

Links requirements to the verifications that provide evidence of satisfaction.

| Aspect | Value |
|--------|-------|
| Domain | REQ |
| Range | TC |
| Cardinality | 0..* |
| Structure | Unrestricted |
| Inverse | (queried as "verifies") |
| Suspect propagation | Modification of the requirement marks `verifiedBy` edges as suspect |

In `verified_by.ndjson`:

```json
{"src": "REQ-C2H6N4P8", "dst": "TC-D9J5Q1R3"}
```

### Quality

Predicates that connect defects to their context.

#### Subject

Points at the node that has a problem. Carries the semantic "this node is defective in some way."

| Aspect | Value |
|--------|-------|
| Domain | DEF |
| Range | any node type |
| Cardinality | 0..1 |
| Structure | Single-value |
| Inverse | (queried as "defects for") |
| Suspect propagation | None (defect is an observation, not a dependency) |

In `subject.ndjson`:

```json
{"src": "DEF-F1L4T7W5", "dst": "REQ-3RR0R001"}
```

#### detectedBy

Points at the examination task that found the defect.

| Aspect | Value |
|--------|-------|
| Domain | DEF |
| Range | TSK |
| Cardinality | 0..1 |
| Structure | Single-value |
| Inverse | (queried as "defects found by") |
| Suspect propagation | None |

In `detected_by.ndjson`:

```json
{"src": "DEF-F1L4T7W5", "dst": "TASK-R3V13W01"}
```

#### Generates

Expresses that a defect creates a remediation task.

| Aspect | Value |
|--------|-------|
| Domain | DEF |
| Range | TSK |
| Cardinality | 0..1 |
| Structure | Single-value |
| Inverse | (queried as "generated from") |
| Suspect propagation | None |

In `generates.ndjson`:

```json
{"src": "DEF-F1L4T7W5", "dst": "TASK-F1X00001"}
```

### Implementation

Predicates that link work items to the obligations they satisfy.

#### Builds

Links tasks to the requirements they exist to satisfy. Uses `oslc_cm:implementsRequirement` directly because ARCI adds no additional semantics (no suspect propagation, no DAG enforcement).

| Aspect | Value |
|--------|-------|
| Domain | TASK |
| Range | REQ |
| Cardinality | 0..* |
| Structure | Unrestricted |
| Inverse | (queried as "implemented by") |
| Suspect propagation | None (`verifiedBy` suspect propagation catches requirement changes) |

In `implements.ndjson`:

```json
{"src": "TASK-E3K8S6V2", "dst": "REQ-C2H6N4P8"}
```

### Provenance

Predicates that track who performed actions and agent hierarchy.

#### Operator

Links an agent to the developer who initiated the session.

| Aspect | Value |
|--------|-------|
| Domain | AGT |
| Range | DEV |
| Cardinality | 0..1 |
| Structure | Single-value |
| Inverse | (queried as "sessions for developer") |
| Suspect propagation | None |

In `operator.ndjson`:

```json
{"src": "AGT-M5V9K3X7", "dst": "DEV-J4R8T2W6"}
```

#### parentAgent

Links a subagent to the session agent that spawned it. Only valid on agents where `subagentId` is non-null.

| Aspect | Value |
|--------|-------|
| Domain | AGT |
| Range | AGT |
| Cardinality | 0..1 |
| Structure | Single-value |
| Inverse | (queried as "subagents of") |
| Suspect propagation | None |

In `parent_agent.ndjson`:

```json
{"src": "AGT-SUB4G3NT", "dst": "AGT-M5V9K3X7"}
```

### Informal

Predicates that express non-formal relationships useful for navigation and context.

#### Informs

Expresses that a concept informs a module. This is a bootstrap/documentation relationship, not part of the formal transformation chain. Unlike `derivesFrom`, `informs` does not establish traceability obligations.

| Aspect | Value |
|--------|-------|
| Domain | CON |
| Range | MOD |
| Cardinality | 0..* |
| Structure | Unrestricted |
| Inverse | (queried as "informed by") |
| Suspect propagation | None (informal, no traceability obligation) |

In `informs.ndjson`:

```json
{"src": "CON-K7M3NP2Q", "dst": "MOD-OAPSROOT"}
```

## Domain/range matrix

Source types as rows, target types as columns, predicates in cells. Empty cells indicate no valid predicate between those types.

|  | → CON | → MOD | → NED | → REQ | → TC | → TSK | → DEF | → BSL | → STK | → DEV | → AGT |
|---|---|---|---|---|---|---|---|---|---|---|---|
| **CON →** | | informs | | | | | | | | | |
| **MOD →** | | childOf | | | | | | | | | |
| **NED →** | derivesFrom | module | | | | | | | stakeholder | | |
| **REQ →** | | module, allocatesTo | derivesFrom | derivesFrom | verifiedBy | | | | | | |
| **TC →** | | module | | | | | | | | | |
| **TSK →** | | module | | builds | | dependsOn | | | | | |
| **DEF →** | | module, subject | subject | subject | subject | detectedBy, generates, subject | subject | | | | |
| **BSL →** | | module | | | | | | | | | |
| **DEV →** | | | | | | | | | | | |
| **AGT →** | | | | | | | | | | operator | parentAgent |

## Directionality conventions

Edge tables use `src` and `dst` columns. The **source** node is always the node that "has" the relationship:

- "MOD-A is a child of MOD-B" → `child_of` row: `src = MOD-A, dst = MOD-B`
- "REQ-A derives from NEED-B" → `derives_from` row: `src = REQ-A, dst = NEED-B`
- "DEF-A has subject REQ-B" → `subject` row: `src = DEF-A, dst = REQ-B`

To query inverse relationships ("what are the children of MOD-B?"), query the `child_of` edge table for rows where `dst = MOD-B`. SQL/PGQ pattern matching handles both directions naturally. See [query patterns](query-patterns.md) for canonical traversal patterns.

## Suspect propagation summary

Predicates that propagate suspect status when someone modifies the source:

| Predicate | Propagation direction | Trigger |
|-----------|----------------------|---------|
| `derivesFrom` | Source modified → downstream edges marked suspect | Source node's `statement`, `status`, or key properties change |
| `verifiedBy` | Requirement modified → verification edges marked suspect | Requirement's `statement` changes |
| `allocatesTo` | Requirement modified → allocation edges marked suspect | Requirement's `statement` or allocation parameters change |

Predicates that do not propagate suspect status: `childOf`, `module`, `dependsOn`, `subject`, `detectedBy`, `generates`, `informs`, `implements`, `operator`, `parentAgent`.

See [constraints](constraints.md) for the full suspect propagation rules and [lifecycle coordination](lifecycle-coordination.md) for how suspect links interact with lifecycle state.
