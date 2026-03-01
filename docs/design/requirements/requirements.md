# Requirements module type

## Overview

Requirements (REQ-*) are design obligations—formal constraints on the system that, when satisfied, fulfill stakeholder needs. Unlike needs (stakeholder expectations), requirements are stated in terms of what the system shall do and must be verifiable.

In INCOSE terms, a requirement is "an agreed-to obligation." It's a contract: if the system meets this requirement, it satisfies (part of) a stakeholder need.

## Purpose

Requirements serve several roles:

**Design constraint**: Requirements constrain design and implementation choices. They're the "shall" statements that implementations must satisfy.

**Verification target**: Every requirement must be verifiable. Verifications (using test, inspection, analysis, or demonstration methods) provide evidence that requirements are met.

**Traceability anchor**: Requirements link upward to needs (why does this exist?) and downward to verifications (how is this verified?). This bidirectional traceability is essential for understanding and maintaining the system.

**Flow-down mechanism**: Parent module requirements allocate to child modules, creating derived requirements with budgets or partitions.

## Requirements vs needs

| Aspect      | Need                            | Requirement            |
|-------------|---------------------------------|------------------------|
| Perspective | Stakeholder                     | System                 |
| Language    | "Users need..."                 | "The system shall..."  |
| Precision   | Qualitative OK                  | Must be verifiable     |
| Owner       | Stakeholder                     | Design team            |
| Validation  | Did we capture the right thing? | Did we build it right? |

Example:

```
Need: "Users need quick feedback when parsing fails"
    ↓
Requirement: "The parser shall report the first syntax error within 50ms"
```

## Lifecycle

Requirements progress through states:

```
draft → proposed → approved → implemented → verified → obsolete
```

| State       | Description                           |
|-------------|---------------------------------------|
| draft       | Initial capture, being refined        |
| proposed    | Ready for review and approval         |
| approved    | Accepted as a design obligation       |
| implemented | Implementation claims to satisfy this |
| verified    | Tests confirm requirement is met      |
| obsolete    | No longer applicable                  |

State transitions:

- `draft → proposed`: Requirement refined and ready for review
- `proposed → approved`: Stakeholders/reviewers approve
- `approved → implemented`: Implementation complete
- `implemented → verified`: Tests pass
- `* → obsolete`: Requirement no longer relevant

## Requirement qualities

Good requirements are:

| Quality     | Description                 | Example                     |
|-------------|-----------------------------|-----------------------------|
| Verifiable  | Can be tested or measured   | "within 50ms" not "quickly" |
| Unambiguous | Single clear interpretation | "first error" not "errors"  |
| Atomic      | Tests one thing             | Split compound requirements |
| Traceable   | Links to need(s)            | derivesFrom relationships   |
| Feasible    | Technically achievable      | Within project constraints  |

## Storage model

Requirement metadata is stored in `graph.jsonlt` as JSON-LD compact form. There is no frontmatter in prose files—`graph.jsonlt` is the single source of truth for all structured data.

```json
{"@context": "context.jsonld", "@id": "REQ-C2H6N4P8", "@type": "Requirement", "title": "Parser error latency", "module": {"@id": "MOD-A4F8R2X1"}, "statement": "The parser shall report the first syntax error within 50ms", "status": "approved", "priority": "must", "verificationMethod": "test", "verificationCriteria": "Benchmark suite achieves p99 < 50ms", "derivesFrom": [{"@id": "NED-B7G3M9K2"}]}
```

Fields:

- `@id`: Unique identifier (REQ-XXXXXXXX format)
- `@type`: Always "Requirement"
- `title`: Human-readable title
- `module`: Module this requirement belongs to (required)
- `statement`: The requirement statement ("shall" language)
- `rationale`: Why this requirement exists (optional)
- `status`: Lifecycle state (draft, proposed, approved, implemented, verified, obsolete)
- `priority`: MoSCoW priority (must, should, could, wont)
- `verificationMethod`: inspection, demonstration, test, analysis
- `verificationCriteria`: How to verify this requirement
- `created`, `updated`: ISO 8601 timestamps
- `tags`: Array of strings (optional)

## Priority levels

Requirements use MoSCoW prioritization (inherited from needs or set directly):

| Priority | Description                          |
|----------|--------------------------------------|
| must     | Essential; system fails without this |
| should   | Important; significant value         |
| could    | Desirable; if time permits           |
| wont     | Explicitly out of scope              |

## Verification methods

From INCOSE, four verification methods:

| Method        | Description                                | Example                      |
|---------------|--------------------------------------------|------------------------------|
| inspection    | Examine artifacts without execution        | Code review, document review |
| demonstration | Operate the system and observe             | Run CLI, check output        |
| test          | Execute with defined inputs, check outputs | Automated test suite         |
| analysis      | Use models, calculations, simulations      | Performance modeling         |

Each requirement specifies its verification method and criteria.

## Relationships

Relationships are embedded in the requirement's JSON-LD record using `{"@id": "..."}` values.

### Outgoing relationships

| Property    | Target | Cardinality | Description                              |
|-------------|--------|-------------|------------------------------------------|
| module      | MOD-*  | Single      | Module this requirement belongs to       |
| derivesFrom | NED-*  | Multi       | Need(s) this requirement satisfies       |
| derivesFrom | REQ-*  | Multi       | Parent requirement (for flow-down)       |
| allocatesTo | MOD-*  | Multi       | Child modules with derived requirements |

### Incoming relationships (queried via graph)

| Property   | Source | Description                        |
|------------|--------|------------------------------------|
| derivesFrom| REQ-*  | Child requirements (flow-down)     |
| verifiedBy | VRF-*  | Verifications that verify this requirement |

Example with relationships:

```json
{"@context": "context.jsonld", "@id": "REQ-C2H6N4P8", "@type": "Requirement", "title": "Parser error latency", "module": {"@id": "MOD-A4F8R2X1"}, "statement": "The parser shall report the first syntax error within 50ms", "status": "approved", "priority": "must", "verificationMethod": "test", "derivesFrom": [{"@id": "NED-B7G3M9K2"}], "verifiedBy": [{"@id": "VRF-D9J5Q1R3"}]}
```

## Flow-down

Parent module requirements allocate to child modules:

```bash
arci reqallocate REQ-H4J7N2P5 --to MOD-A4F8R2X1 --budget "50ms"
arci reqallocate REQ-H4J7N2P5 --to MOD-B9G3M7K2 --budget "30ms"
```

This creates:

- allocatesTo relationships from parent requirement to child modules (with budget metadata)
- Derived requirements on child modules
- derivesFrom relationships back to parent requirement

Example with allocation:

```json
{"@context": "context.jsonld", "@id": "REQ-H4J7N2P5", "@type": "Requirement", "title": "System latency", "module": {"@id": "MOD-OAPSROOT"}, "statement": "System latency shall be under 100ms", "status": "approved", "allocatesTo": [{"@id": "MOD-A4F8R2X1", "budget": "50ms"}, {"@id": "MOD-B9G3M7K2", "budget": "30ms"}]}
```

## Derivation

When a need is validated, derivation produces requirements:

```bash
arci needderive NED-B7G3M9K2
```

Or from a parent requirement (flow-down):

```bash
arci reqderive REQ-H4J7N2P5 --to MOD-A4F8R2X1
```

## Implementation architecture

Requirement functionality follows the three-layer architecture (see CON-GR4PH4RC).

### Typed node

RequirementNode is an independent dataclass with typed fields. Graph stores RequirementNode directly, and IO serializes directly to/from RequirementNode (see CON-GR4PH4RC for the full architecture).

```python
@dataclass(frozen=True, slots=True)
class RequirementNode:
    """Requirement module—design obligation."""
    id: str
    title: str
    status: RequirementStatus          # Typed enum
    priority: Priority                 # Typed enum
    verification_method: VerificationMethod  # Typed enum
    statement: str                     # Type-specific field
    rationale: str = ""                # Type-specific field
    verification_criteria: str = ""   # Type-specific field
    description: str = ""
    created: datetime | None = None
    updated: datetime | None = None
```

All fields use proper types. The IO layer creates RequirementNode directly from JSON-LD records, preserving all type-specific fields like `statement`, `verification_method`, and `verification_criteria`.

### Core layer (arci.core.requirement)

Pure functions and typed data structures:

```python
# Types
class RequirementStatus(StrEnum):
    DRAFT = "draft"
    PROPOSED = "proposed"
    APPROVED = "approved"
    IMPLEMENTED = "implemented"
    VERIFIED = "verified"
    OBSOLETE = "obsolete"

class VerificationMethod(StrEnum):
    INSPECTION = "inspection"
    DEMONSTRATION = "demonstration"
    TEST = "test"
    ANALYSIS = "analysis"

# Typed node
@dataclass(frozen=True, slots=True)
class RequirementNode(Node):
    statement: str = ""
    rationale: str = ""
    priority: Priority = Priority.SHOULD
    verification_method: VerificationMethod = VerificationMethod.TEST
    verification_criteria: str = ""

# Operations (pure functions)
def from_node(node: Node) -> RequirementNode: ...
def with_status(req: RequirementNode, status: RequirementStatus) -> RequirementNode: ...
def can_transition(req: RequirementNode, target: RequirementStatus) -> bool: ...

# Queries (pure, take Graph)
def get(graph: Graph, req_id: str) -> RequirementNode | None: ...
def list_all(graph: Graph) -> tuple[RequirementNode, ...]: ...
def list_by_module(graph: Graph, module_id: str) -> tuple[RequirementNode, ...]: ...
def owning_module(graph: Graph, req_id: str) -> str | None: ...
def derives_from_needs(graph: Graph, req_id: str) -> frozenset[str]: ...
def derives_from_requirements(graph: Graph, req_id: str) -> frozenset[str]: ...
def verified_by(graph: Graph, req_id: str) -> frozenset[str]: ...
def allocated_to(graph: Graph, req_id: str) -> frozenset[str]: ...
```

### Service layer (arci.service.requirement)

Orchestrates core and IO:

```python
# Thin query wrappers
def get(store: GraphStore, req_id: str) -> RequirementNode | None: ...
def list_all(store: GraphStore) -> tuple[RequirementNode, ...]: ...
def list_by_module(store: GraphStore, module_id: str) -> tuple[RequirementNode, ...]: ...

# Mutations
def create(
    store: GraphStore,
    module_id: str,
    statement: str,
    derives_from: list[str] | None = None,
    verification_method: VerificationMethod = VerificationMethod.TEST,
    ...
) -> RequirementNode: ...
def update(store: GraphStore, req_id: str, **fields) -> RequirementNode: ...
def delete(store: GraphStore, req_id: str) -> None: ...
def transition(store: GraphStore, req_id: str, target: RequirementStatus) -> RequirementNode: ...

# Workflows
def allocate(store: GraphStore, req_id: str, module_id: str, budget: str | None = None) -> RequirementNode: ...
def link_verification(store: GraphStore, req_id: str, verification_id: str) -> RequirementNode: ...
```

## CLI commands

```bash
# CRUD
arci reqcreate --module MOD-A4F8R2X1 \
  --statement "The parser shall report errors within 50ms" \
  --verification-method test
arci reqshow REQ-C2H6N4P8
arci reqlist
arci reqlist --module MOD-A4F8R2X1 --status approved
arci requpdate REQ-C2H6N4P8 --status implemented
arci reqdelete REQ-C2H6N4P8

# Relationships
arci reqlink REQ-C2H6N4P8 --derives-from NED-B7G3M9K2
arci reqlink REQ-C2H6N4P8 --verified-by VRF-D9J5Q1R3

# Flow-down
arci reqderive REQ-C2H6N4P8 --to MOD-L3X3R001
arci reqallocate REQ-H4J7N2P5 --to MOD-A4F8R2X1 --budget "50ms"

# Traceability
arci reqtrace REQ-C2H6N4P8  # Show full chain: concept → need → req → verifications

# Coverage
arci reqcoverage                    # Overall verification coverage
arci requnverified                  # Requirements without verifications
```

## Examples

### Functional requirement

```json
{"@context": "context.jsonld", "@id": "REQ-C2H6N4P8", "@type": "Requirement", "title": "Parser error latency", "module": {"@id": "MOD-A4F8R2X1"}, "statement": "The parser shall report the first syntax error within 50ms", "status": "approved", "priority": "must", "verificationMethod": "test", "verificationCriteria": "Benchmark suite achieves p99 < 50ms", "derivesFrom": [{"@id": "NED-B7G3M9K2"}]}
```

### Interface requirement

```json
{"@context": "context.jsonld", "@id": "REQ-1NT3RF01", "@type": "Requirement", "title": "Parser API signature", "module": {"@id": "MOD-A4F8R2X1"}, "statement": "The parser shall expose a parse() function accepting string input", "status": "implemented", "priority": "must", "verificationMethod": "inspection", "verificationCriteria": "API signature matches specification"}
```

### Quality requirement

```json
{"@context": "context.jsonld", "@id": "REQ-QU4L1TY1", "@type": "Requirement", "title": "Test coverage threshold", "module": {"@id": "MOD-OAPSROOT"}, "statement": "Test coverage shall exceed 80% for all modules", "status": "approved", "priority": "should", "verificationMethod": "analysis", "verificationCriteria": "Coverage report shows >80% for each module"}
```

### Allocated requirement

```json
{"@context": "context.jsonld", "@id": "REQ-L3X3R001", "@type": "Requirement", "title": "Lexer latency budget", "module": {"@id": "MOD-L3X3R001"}, "statement": "Lexer latency shall be under 30ms", "status": "approved", "priority": "must", "verificationMethod": "test", "derivesFrom": [{"@id": "REQ-H4J7N2P5"}]}
```

## Implementation status

| Layer | Status | Notes |
|-------|--------|-------|
| Core | Implemented | Typed node, operations, queries in `arci.core.requirement` |
| IO | Implemented | JSON-LD serialization via `arci.io.graph` |
| Service | Implemented | Full CRUD, transitions, derivation, allocation, verification linking |
| CLI | Implemented | All commands: create, show, list, update, delete, transition, link, derive, allocate, trace, coverage, unverified |

## Summary

Requirements are verifiable design obligations:

- Stated as "shall" statements the system must satisfy
- Derived from needs with derivesFrom traceability
- Verified by tests, inspections, demonstrations, or analyses
- Flow down from parent modules with budget allocations
- Progress from draft through approved to verified
- Store metadata in graph.jsonlt with no frontmatter in prose files
- Implemented following three-layer architecture (core/io/service)

Requirements are the contract between stakeholder expectations (needs) and implementation (code). Every requirement should trace back to a need and forward to verifications that provide evidence.
