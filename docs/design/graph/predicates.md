# Predicates

## Overview

Predicates define the semantic relationships between nodes in the knowledge graph. Each predicate has a specific domain (source types), range (target types), cardinality, structural constraint, and behavior for suspect link propagation.

All relationships are stored as JSON-LD properties on the source node, with `{"@id": "TARGET-ID"}` values. Directionality is always from the node that "has" the relationship to the node it references.

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

```json
{"@id": "MOD-A4F8R2X1", "childOf": {"@id": "MOD-OAPSROOT"}}
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

```json
{"@id": "NED-B7G3M9K2", "derivesFrom": [{"@id": "CON-K7M3NP2Q"}]}
{"@id": "REQ-C2H6N4P8", "derivesFrom": [{"@id": "NED-B7G3M9K2"}]}
{"@id": "REQ-CH1LDR3Q", "derivesFrom": [{"@id": "REQ-C2H6N4P8"}]}
```

### Ownership

Predicates that assign entities to their owning container.

#### module

Expresses which module owns a node.

| Aspect | Value |
|--------|-------|
| Domain | NED, REQ, VRF, TSK, DEF |
| Range | MOD |
| Cardinality | 1 (exactly one) |
| Structure | Single-value |
| Inverse | (queried as "owned by module") |
| Suspect propagation | None (organizational, not semantic) |

```json
{"@id": "NED-B7G3M9K2", "module": {"@id": "MOD-A4F8R2X1"}}
{"@id": "REQ-C2H6N4P8", "module": {"@id": "MOD-A4F8R2X1"}}
{"@id": "TSK-E3K8S6V2", "module": {"@id": "MOD-A4F8R2X1"}}
```

### Allocation

Predicates that express flow-down of obligations to architectural elements.

#### allocatesTo

Expresses requirement flow-down to child modules. When a parent module's requirement must be met by children, `allocatesTo` distributes the obligation, potentially with budgets or partitions.

| Aspect | Value |
|--------|-------|
| Domain | REQ |
| Range | MOD |
| Cardinality | 0..* |
| Structure | Unrestricted |
| Metadata | Optional budget, partition, or other allocation parameters |
| Inverse | (queried as "allocated from") |
| Suspect propagation | Modification of the requirement marks `allocatesTo` edges as suspect |

```json
{"@id": "REQ-H4J7N2P5", "allocatesTo": [
  {"@id": "MOD-A4F8R2X1", "budget": "50ms"},
  {"@id": "MOD-B9G3M7K2", "budget": "30ms"}
]}
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

```json
{"@id": "TSK-E3K8S6V2", "dependsOn": [{"@id": "TSK-G5M2R8X4"}]}
```

### Verification

Predicates that connect obligations to evidence.

#### verifiedBy

Links requirements to the verifications that provide evidence of satisfaction.

| Aspect | Value |
|--------|-------|
| Domain | REQ |
| Range | VRF |
| Cardinality | 0..* |
| Structure | Unrestricted |
| Inverse | (queried as "verifies") |
| Suspect propagation | Modification of the requirement marks `verifiedBy` edges as suspect |

```json
{"@id": "REQ-C2H6N4P8", "verifiedBy": [{"@id": "VRF-D9J5Q1R3"}]}
```

### Quality

Predicates that connect defects to their context.

#### subject

Points at the node that has a problem. Carries the semantic "this node is defective in some way."

| Aspect | Value |
|--------|-------|
| Domain | DEF |
| Range | any node type |
| Cardinality | 0..1 |
| Structure | Single-value |
| Inverse | (queried as "defects for") |
| Suspect propagation | None (defect is an observation, not a dependency) |

```json
{"@id": "DEF-F1L4T7W5", "subject": {"@id": "REQ-3RR0R001"}}
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

```json
{"@id": "DEF-F1L4T7W5", "detectedBy": {"@id": "TSK-R3V13W01"}}
```

#### generates

Expresses that a defect creates a remediation task.

| Aspect | Value |
|--------|-------|
| Domain | DEF |
| Range | TSK |
| Cardinality | 0..1 |
| Structure | Single-value |
| Inverse | (queried as "generated from") |
| Suspect propagation | None |

```json
{"@id": "DEF-F1L4T7W5", "generates": {"@id": "TSK-F1X00001"}}
```

### Informal

Predicates that express non-formal relationships useful for navigation and context.

#### informs

Expresses that a concept informs a module. This is a bootstrap/documentation relationship, not part of the formal transformation chain. Unlike `derivesFrom`, `informs` does not establish traceability obligations.

| Aspect | Value |
|--------|-------|
| Domain | CON |
| Range | MOD |
| Cardinality | 0..* |
| Structure | Unrestricted |
| Inverse | (queried as "informed by") |
| Suspect propagation | None (informal, no traceability obligation) |

```json
{"@id": "CON-K7M3NP2Q", "informs": {"@id": "MOD-OAPSROOT"}}
```

## Domain/range matrix

Source types as rows, target types as columns, predicates in cells. Empty cells indicate no valid predicate between those types.

|  | → CON | → MOD | → NED | → REQ | → VRF | → TSK | → DEF | → BSL |
|---|---|---|---|---|---|---|---|---|
| **CON →** | | informs | | | | | | |
| **MOD →** | | childOf | | | | | | |
| **NED →** | derivesFrom | module | | | | | | |
| **REQ →** | | module, allocatesTo | derivesFrom | derivesFrom | verifiedBy | | | |
| **VRF →** | | module | | | | | | |
| **TSK →** | | module | | | | dependsOn | | |
| **DEF →** | | module, subject | subject | subject | subject | detectedBy, generates, subject | subject | |
| **BSL →** | | module | | | | | | |

## Directionality conventions

Relationships are stored on the **source** node as JSON-LD properties:

- "MOD-A has childOf MOD-B" → the `childOf` property is on MOD-A's record
- "REQ-A has derivesFrom NED-B" → the `derivesFrom` property is on REQ-A's record
- "DEF-A has subject REQ-B" → the `subject` property is on DEF-A's record

To query inverse relationships (e.g., "what are the children of MOD-B?"), traverse the graph to find all nodes with `childOf` pointing at MOD-B. See [query patterns](query-patterns.md) for canonical traversal patterns.

## Suspect propagation summary

Predicates that propagate suspect status when the source is modified:

| Predicate | Propagation direction | Trigger |
|-----------|----------------------|---------|
| `derivesFrom` | Source modified → downstream edges marked suspect | Source node's `statement`, `status`, or key properties change |
| `verifiedBy` | Requirement modified → verification edges marked suspect | Requirement's `statement` changes |
| `allocatesTo` | Requirement modified → allocation edges marked suspect | Requirement's `statement` or allocation parameters change |

Predicates that do **not** propagate suspect status: `childOf`, `module`, `dependsOn`, `subject`, `detectedBy`, `generates`, `informs`.

See [constraints](constraints.md) for the full suspect propagation rules and [lifecycle coordination](lifecycle-coordination.md) for how suspect links interact with lifecycle state.
