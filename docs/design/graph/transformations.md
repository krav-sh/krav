# Transformations

## Overview

Transformations are formal operations that create new nodes from existing ones, building the derivation chain that provides traceability from stakeholder intent to verified implementation. Each transformation has preconditions, graph operations, postconditions, and cardinality rules.

The transformation chain is the core workflow of the arci knowledge graph:

```
Concept (exploration)
    ↓ formalize
Need (expectation, validated)
    ↓ derive
Requirement (obligation, verified)
    ↓ flow-down
Requirement (child module obligation)
    ↓ decompose
Task (work, in DAG)
    ↓ execute
Deliverable (verified output)
```

Each step in this chain creates `derivesFrom` edges that maintain an unbroken traceability path. The chain is always connected — no requirement exists without tracing back to a need, and no need exists without tracing back to a concept.

## Formalize: concept → need

Formalization extracts stakeholder expectations from crystallized concepts. A concept captures exploration and design thinking; formalization transforms that thinking into structured, validatable need statements.

**Preconditions**:
- Source CON must exist and have `status = crystallized`
- Target MOD must exist (the module that will own the need)

**Graph operations**:
1. Create one or more NED nodes with `status = draft`
2. Set `derivesFrom` on each NED to reference the source CON
3. Set `module` on each NED to the target MOD

**Postconditions**:
- Each new NED has at least one `derivesFrom` edge to the source CON
- Each new NED has a `module` edge to a valid MOD
- The `derivesFrom` subgraph remains a DAG

**Cardinality**: One concept may produce multiple needs (1:N). Different stakeholder expectations from the same concept become separate needs. A need may also derive from multiple concepts (M:N) when a single expectation spans multiple areas of exploration.

**Example**:
```json
{"@id": "CON-K7M3NP2Q", "@type": "Concept", "status": "crystallized", "title": "Error reporting design"}

{"@id": "NED-ERR0R001", "@type": "Need", "status": "draft", "title": "Clear error messages",
 "module": {"@id": "MOD-A4F8R2X1"}, "derivesFrom": [{"@id": "CON-K7M3NP2Q"}],
 "statement": "Users need error messages that identify the location and nature of problems"}
{"@id": "NED-ERR0R002", "@type": "Need", "status": "draft", "title": "Structured error output",
 "module": {"@id": "MOD-A4F8R2X1"}, "derivesFrom": [{"@id": "CON-K7M3NP2Q"}],
 "statement": "Tool integrators need machine-readable error output for automated processing"}
```

## Derive: need → requirement

Derivation transforms a stakeholder expectation into a design obligation. The requirement is more specific, more constrained, and verifiable — stated in terms of what the system shall do rather than what the stakeholder needs.

**Preconditions**:
- Source NED must exist and have `status = validated`
- Target MOD must exist (same as need's module, or a child module)

**Graph operations**:
1. Create one or more REQ nodes with `status = draft`
2. Set `derivesFrom` on each REQ to reference the source NED
3. Set `module` on each REQ to the target MOD

**Postconditions**:
- Each new REQ has at least one `derivesFrom` edge to the source NED
- Each new REQ has a `module` edge to a valid MOD
- The `derivesFrom` subgraph remains a DAG
- The requirement's module is the same as or a child of the need's module (constraint C-CROSS1)

**Cardinality**: One need may produce multiple requirements (1:N). A need like "users need fast feedback" may produce separate requirements for parsing latency, rendering latency, and startup time. A requirement may also derive from multiple needs (M:N) when a single obligation satisfies several stakeholder expectations.

**Example**:
```json
{"@id": "NED-ERR0R001", "@type": "Need", "status": "validated", "title": "Clear error messages"}

{"@id": "REQ-3RR0R001", "@type": "Requirement", "status": "draft",
 "module": {"@id": "MOD-A4F8R2X1"}, "derivesFrom": [{"@id": "NED-ERR0R001"}],
 "statement": "The parser shall include line number and column position in all error messages"}
{"@id": "REQ-3RR0R002", "@type": "Requirement", "status": "draft",
 "module": {"@id": "MOD-A4F8R2X1"}, "derivesFrom": [{"@id": "NED-ERR0R001"}],
 "statement": "The parser shall provide a source context snippet of ±2 lines around each error"}
```

## Flow-down: parent requirement → child modules

Flow-down allocates a parent module's requirement to child modules. Each child module then has derived requirements that collectively satisfy the parent requirement. Flow-down may include budgets or partitions.

**Preconditions**:
- Source REQ must exist
- Source REQ's module must have child modules
- Target MOD(s) must be children of the source REQ's module (constraint C-CROSS3)

**Graph operations**:
1. Set `allocatesTo` on the source REQ to reference target child MODs, optionally with allocation metadata (budget, partition)
2. Create derived REQ nodes in each target child MOD with `derivesFrom` pointing at the parent REQ
3. Set `module` on each derived REQ to the target child MOD

**Postconditions**:
- The source REQ has `allocatesTo` edges to child modules
- Each child module has derived REQ nodes with `derivesFrom` edges to the parent REQ
- The `derivesFrom` subgraph remains a DAG
- Allocation budgets sum appropriately (when applicable)

**Cardinality**: One parent requirement may allocate to multiple child modules (1:N). Each child module may have one or more derived requirements from the allocation.

**Example**:
```json
{"@id": "REQ-P3RF0RM1", "@type": "Requirement", "status": "approved",
 "module": {"@id": "MOD-OAPSROOT"},
 "statement": "System shall respond within 100ms at p99",
 "allocatesTo": [
   {"@id": "MOD-A4F8R2X1", "budget": "50ms"},
   {"@id": "MOD-B9G3M7K2", "budget": "30ms"}
 ]}

{"@id": "REQ-P3RFPRS1", "@type": "Requirement", "status": "draft",
 "module": {"@id": "MOD-A4F8R2X1"}, "derivesFrom": [{"@id": "REQ-P3RF0RM1"}],
 "statement": "The parser shall complete processing within 50ms at p99"}
{"@id": "REQ-P3RFRND1", "@type": "Requirement", "status": "draft",
 "module": {"@id": "MOD-B9G3M7K2"}, "derivesFrom": [{"@id": "REQ-P3RF0RM1"}],
 "statement": "The renderer shall complete output within 30ms at p99"}
```

## Decompose: requirement/need/module → tasks

Decomposition generates work. Given a target (module, need, or requirement), decomposition produces a task DAG that, when executed, will satisfy the target.

**Preconditions**:
- Target entity (MOD, NED, or REQ) must exist
- Target module must be at or advancing to the appropriate phase

**Graph operations**:
1. Create TSK nodes with `status = pending`
2. Set `module` on each TSK to the target module
3. Set `processPhase` on each TSK based on the type of work
4. Set `dependsOn` edges between tasks to express ordering
5. Link tasks back to their motivating entities as appropriate

**Postconditions**:
- All new TSK nodes have `module` and `processPhase` set
- The `dependsOn` subgraph remains a DAG
- Tasks are ordered by phase: architecture tasks before design, design before implementation, etc.

**Cardinality**: One target may produce many tasks (1:N). The task DAG may include tasks across multiple process phases.

**Decomposition considerations**:
- **Process phases**: Architecture tasks before design tasks before implementation tasks
- **Module hierarchy**: Parent module work may gate child module work
- **Requirement dependencies**: Some requirements must be satisfied before others
- **Templates**: Decomposition can use [templates](../execution/templating.md) for common patterns

## Verify: requirement → verification

Verification creates evidence that a requirement is satisfied.

**Preconditions**:
- Target REQ must exist
- Verification method must be appropriate for the requirement

**Graph operations**:
1. Create VRF node with `status = planned`
2. Set `module` on VRF to the requirement's module (or descendant, per C-CROSS2)
3. Set `verifiedBy` on the target REQ to reference the new VRF

**Postconditions**:
- The REQ has a `verifiedBy` edge to the new VRF
- The VRF has a `module` edge to a valid MOD

**Cardinality**: One requirement may be verified by multiple verifications (1:N, different methods or aspects). One verification may verify multiple requirements (M:N).

## Remediate: defect → task

Remediation creates a task to fix a confirmed defect.

**Preconditions**:
- Source DEF must exist and have `status = confirmed`

**Graph operations**:
1. Create TSK node with `status = pending`
2. Set `module` on TSK to the defect's module
3. Set `generates` on the DEF to reference the new TSK

**Postconditions**:
- The DEF has a `generates` edge to the new TSK
- The TSK has a `module` edge matching the DEF's module

**Cardinality**: One defect produces one remediation task (1:1). If a fix addresses multiple defects, the task is linked from the primary defect and the others are resolved when the task completes.

## Transformation invariants

**TI-1**: Traceability chain connectivity. The chain CON → NED → REQ must be connected. Every REQ traces back to at least one NED via `derivesFrom`, and every NED traces back to at least one CON via `derivesFrom`.

**TI-2**: Module ownership consistency. Transformations that produce nodes in child modules (flow-down, decomposition) must respect the module hierarchy. Derived entities belong to the same module as or a descendant of their source entity.

**TI-3**: Phase consistency. Transformations cannot produce tasks for phases the target module hasn't reached (constraint C-PH3).

**TI-4**: DAG preservation. No transformation may introduce a cycle in `derivesFrom` or `dependsOn`.
