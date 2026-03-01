# Verifications module type

## Overview

Verifications (VRF-*) provide evidence that the system satisfies requirements. Each verification has a method (inspection, demonstration, test, analysis) and links to the requirements it verifies via verifiedBy relationships.

Verifications exist at different levels corresponding to the module hierarchy. Component-level verifications verify component requirements; integration verifications verify interface requirements; system verifications verify system-level requirements.

## Purpose

Verifications serve multiple roles:

**Verification evidence**: verifications provide evidence that the system satisfies requirements. When verifications pass, the linked requirements can transition to "verified" status.

**Regression protection**: automated verifications (those using the test method) catch regressions when changes break previously verified requirements.

**Coverage tracking**: the verifiedBy relationship enables coverage analysis: which requirements have verification, which do not, which have gaps.

**Documentation**: verifications demonstrate expected behavior. They're executable specifications.

## Verification methods

From INCOSE, verifications use one of four verification methods:

| Method        | Description                                | When to use                              |
|---------------|--------------------------------------------|------------------------------------------|
| test          | Execute with defined inputs, check outputs | Behavioral requirements, APIs            |
| inspection    | Examine artifacts without execution        | Code standards, documentation presence   |
| demonstration | Operate system, observe behavior           | User-facing workflows, UI requirements   |
| analysis      | Models, calculations, simulations          | Performance projections, safety analysis |

## Lifecycle

Verifications progress through states:

```text
draft → ready → passing → failing → skipped → obsolete
```

| State    | Description                         |
|----------|-------------------------------------|
| draft    | Someone is developing the verification |
| ready    | Verification implemented, not yet executed |
| passing  | Last execution passed               |
| failing  | Last execution failed               |
| skipped  | Intentionally not run (with reason) |
| obsolete | No longer relevant                  |

Note: `passing` and `failing` reflect last execution. Verifications can flip between these states as code and verifications evolve.

## Verification levels

Verifications align with the module hierarchy:

| Level       | Module scope | Verifications                   |
|-------------|--------------|---------------------------------|
| Unit        | Component    | Function/method verifications   |
| Integration | Subsystem    | Component interaction verifications |
| System      | Root         | End-to-end verifications        |
| Acceptance  | Root         | Stakeholder validation          |

Unit-level verifications verify a component module's requirements. Integration verifications verify a subsystem's interface requirements. System-level requirements need system verifications.

## Storage model

The graph stores verification metadata in `graph.jsonlt` as JSON-LD compact form. Prose files have no frontmatter; `graph.jsonlt` is the single source of truth for all structured data.

```json
{"@context": "context.jsonld", "@id": "VRF-D9J5Q1R3", "@type": "Verification", "title": "Parser error latency benchmark", "module": {"@id": "MOD-A4F8R2X1"}, "description": "Verifies error reporting meets 50ms requirement", "method": "test", "level": "unit", "status": "passing", "implementation": "tests/parser/error_latency_test.ts"}
```

Fields:

- `@id`: Unique identifier (VRF-XXXXXXXX format)
- `@type`: Always "Verification"
- `title`: Human-readable title
- `module`: Module this verification belongs to (required)
- `description`: What this verification verifies (optional)
- `method`: Verification method (inspection, demonstration, test, analysis)
- `level`: Verification level (unit, integration, system, acceptance)
- `status`: Lifecycle state (draft, ready, passing, failing, skipped, obsolete)
- `implementation`: Path to verification code or procedure (optional)
- `lastRun`: ISO 8601 timestamp of last execution (optional)
- `lastResult`: Last execution result object (optional)
- `created`, `updated`: ISO 8601 timestamps
- `tags`: Array of strings (optional)

## Relationships

The verification's JSON-LD record embeds relationships using `{"@id": "..."}` values.

### Outgoing relationships

| Property | Target | Cardinality | Description                              |
|----------|--------|-------------|------------------------------------------|
| module   | MOD-*  | Single      | Module this verification belongs to      |
| verifies | REQ-*  | Multi       | Requirements this verification verifies  |

### Incoming relationships (queried via graph)

| Property   | Source | Description                               |
|------------|--------|-------------------------------------------|
| verifiedBy | REQ-*  | Requirements verified by this verification |

Note: verifiedBy on REQ points to VRF; verifies on VRF points to REQ. These are inverses.

Example with relationships:

```json
{"@context": "context.jsonld", "@id": "VRF-D9J5Q1R3", "@type": "Verification", "title": "Parser error latency benchmark", "module": {"@id": "MOD-A4F8R2X1"}, "method": "test", "level": "unit", "status": "passing", "implementation": "tests/parser/error_latency_test.ts", "verifies": [{"@id": "REQ-C2H6N4P8"}]}
```

## Verification types

### Automated tests

Most common. Code that executes and asserts outcomes:

```json
{"@context": "context.jsonld", "@id": "VRF-D9J5Q1R3", "@type": "Verification", "title": "Parser error latency benchmark", "module": {"@id": "MOD-A4F8R2X1"}, "method": "test", "level": "unit", "status": "passing", "implementation": "tests/parser/error_latency_test.ts"}
```

### Inspection checklists

For requirements verified by artifact examination:

```json
{"@context": "context.jsonld", "@id": "VRF-1NSP3CT1", "@type": "Verification", "title": "Documentation completeness", "module": {"@id": "MOD-OAPSROOT"}, "method": "inspection", "level": "system", "status": "passing", "checklist": ["API documentation exists for all public functions", "Error codes are documented with examples", "README includes quick start guide"]}
```

### Demonstrations

For requirements verified by operation and observation:

```json
{"@context": "context.jsonld", "@id": "VRF-D3M00001", "@type": "Verification", "title": "Module lifecycle demonstration", "module": {"@id": "MOD-B9G3M7K2"}, "method": "demonstration", "level": "acceptance", "status": "passing", "procedure": "Execute standard user workflow and verify outputs"}
```

### Analyses

For requirements verified through modeling or calculation:

```json
{"@context": "context.jsonld", "@id": "VRF-4N4LYS15", "@type": "Verification", "title": "System latency analysis", "module": {"@id": "MOD-OAPSROOT"}, "method": "analysis", "level": "system", "status": "passing", "approach": "Performance modeling based on component benchmarks", "conclusion": "System latency under load: 87ms (within 100ms budget)"}
```

## Coverage

Coverage tracks which requirements have verification:

```bash
Krav verificationcoverage                    # Overall coverage report
Krav verificationcoverage --module MOD-A4F8R2X1  # Module-specific
Krav verificationuntested                    # Requirements without verifications
Krav verificationgaps                        # Requirements with insufficient coverage
```

Coverage analysis considers:

- Requirements with no verifiedBy links
- Requirements with only failing verifications
- Requirements where verification method doesn't match requirement's verification method

## Implementation architecture

Verification capability follows the three-layer architecture (see CON-GR4PH4RC).

### Typed node

VerificationNode is an independent dataclass with typed fields. Graph stores VerificationNode directly, and IO serializes directly to/from VerificationNode (see CON-GR4PH4RC for the full architecture).

```python
@dataclass(frozen=True, slots=True)
class VerificationNode:
    """Verification module—verification evidence."""
    id: str
    title: str
    status: VerificationStatus             # Typed enum
    method: VerificationMethod             # Typed enum (from krav.core.requirement._types)
    level: VerificationLevel               # Typed enum
    implementation: str = ""               # Path to verification code
    last_run: datetime | None = None
    last_passed: datetime | None = None    # Timestamp of last successful run
    description: str = ""
    created: datetime | None = None
    updated: datetime | None = None
```

All fields use proper types. The IO layer creates VerificationNode directly from JSON-LD records, preserving all type-specific fields like `method`, `level`, and `implementation`.

### Core layer (`krav.core.verification`)

Pure functions and typed data structures:

```python
# Types (krav.core.verification._types)
class VerificationStatus(StrEnum):
    DRAFT = "draft"
    READY = "ready"
    PASSING = "passing"
    FAILING = "failing"
    SKIPPED = "skipped"
    OBSOLETE = "obsolete"

class VerificationLevel(StrEnum):
    UNIT = "unit"
    INTEGRATION = "integration"
    SYSTEM = "system"
    ACCEPTANCE = "acceptance"

# VerificationMethod is defined in krav.core.requirement._types
# and imported by VerificationNode
from krav.core.requirement._types import VerificationMethod

# Typed node
@dataclass(frozen=True, slots=True)
class VerificationNode(Node):
    method: VerificationMethod = VerificationMethod.TEST
    level: VerificationLevel = VerificationLevel.UNIT
    implementation: str = ""
    last_run: datetime | None = None
    last_passed: datetime | None = None  # Timestamp of last successful run

# Operations (pure functions)
def from_node(node: Node) -> VerificationNode: ...
def with_status(verification: VerificationNode, status: VerificationStatus) -> VerificationNode: ...
def with_result(verification: VerificationNode, passed: bool, run_time: datetime) -> VerificationNode: ...

# Queries (pure, take Graph)
def get(graph: Graph, verification_id: str) -> VerificationNode | None: ...
def list_all(graph: Graph) -> tuple[VerificationNode, ...]: ...
def list_by_module(graph: Graph, module_id: str) -> tuple[VerificationNode, ...]: ...
def list_by_status(graph: Graph, status: VerificationStatus) -> tuple[VerificationNode, ...]: ...
def verifies(graph: Graph, verification_id: str) -> frozenset[str]: ...  # Returns requirement IDs
def owning_module(graph: Graph, verification_id: str) -> str | None: ...
```

### Service layer (`krav.service.verification`)

Orchestrates core and IO:

```python
# Thin query wrappers
def get(store: GraphStore, verification_id: str) -> VerificationNode | None: ...
def list_all(store: GraphStore) -> tuple[VerificationNode, ...]: ...
def list_by_module(store: GraphStore, module_id: str) -> tuple[VerificationNode, ...]: ...

# Mutations
def create(
    store: GraphStore,
    module_id: str,
    title: str,
    method: VerificationMethod = VerificationMethod.TEST,
    level: VerificationLevel = VerificationLevel.UNIT,
    implementation: str = "",
    verifies: list[str] | None = None,
    ...
) -> VerificationNode: ...
def update(store: GraphStore, verification_id: str, **fields) -> VerificationNode: ...
def delete(store: GraphStore, verification_id: str) -> None: ...

# Workflows
def record_result(store: GraphStore, verification_id: str, passed: bool, details: str = "") -> VerificationNode: ...
def link_requirement(store: GraphStore, verification_id: str, req_id: str) -> VerificationNode: ...
def coverage_report(store: GraphStore, module_id: str | None = None) -> CoverageReport: ...
```

## CLI commands

```bash
# CRUD
Krav verificationcreate --module MOD-A4F8R2X1 --title "Error latency verification" \
  --method test --implementation "tests/parser/error_latency_test.ts"
Krav verificationshow VRF-D9J5Q1R3
Krav verificationlist
Krav verificationlist --module MOD-A4F8R2X1 --status failing
Krav verificationupdate VRF-D9J5Q1R3 --status passing
Krav verificationdelete VRF-D9J5Q1R3

# Relationships
Krav verificationlink VRF-D9J5Q1R3 --verifies REQ-C2H6N4P8
Krav verificationunlink VRF-D9J5Q1R3 --verifies REQ-C2H6N4P8

# Execution tracking
Krav verificationrecord VRF-D9J5Q1R3 --passed --duration 1250 --details "p99: 42ms"
Krav verificationrecord VRF-D9J5Q1R3 --failed --details "p99: 67ms (exceeds 50ms)"

# Coverage
Krav verificationcoverage
Krav verificationcoverage --module MOD-A4F8R2X1
Krav verificationuntested
```

## Examples

### Automated unit verification

```json
{"@context": "context.jsonld", "@id": "VRF-D9J5Q1R3", "@type": "Verification", "title": "Parser error latency benchmark", "module": {"@id": "MOD-A4F8R2X1"}, "method": "test", "level": "unit", "status": "passing", "implementation": "tests/parser/error_latency_test.ts", "verifies": [{"@id": "REQ-C2H6N4P8"}]}
```

### Integration verification

```json
{"@context": "context.jsonld", "@id": "VRF-1NT3GR01", "@type": "Verification", "title": "Parser-CLI integration", "module": {"@id": "MOD-A4F8R2X1"}, "method": "test", "level": "integration", "status": "passing", "implementation": "tests/integration/parser_cli_test.ts"}
```

### Inspection verification

```json
{"@context": "context.jsonld", "@id": "VRF-D0C5CH3K", "@type": "Verification", "title": "Documentation completeness", "module": {"@id": "MOD-OAPSROOT"}, "method": "inspection", "level": "system", "status": "passing", "checklist": ["README exists and is current", "API docs generated and published", "CONTRIBUTING.md present"]}
```

### Demonstration verification

```json
{"@context": "context.jsonld", "@id": "VRF-D3M0CL11", "@type": "Verification", "title": "CLI workflow demonstration", "module": {"@id": "MOD-B9G3M7K2"}, "method": "demonstration", "level": "acceptance", "status": "passing", "procedure": "Execute standard user workflow and verify outputs"}
```

## Relationship to tasks

Verification-phase tasks create and execute verifications:

```json
{"@context": "context.jsonld", "@id": "TSK-V3R1FY01", "@type": "Task", "title": "Implement parser verifications", "module": {"@id": "MOD-A4F8R2X1"}, "processPhase": "verification", "taskType": "verification-implementation"}
```

Tasks can record verification execution results as deliverables.

## Implementation status

| Layer | Status | Notes |
|-------|--------|-------|
| Core | Implemented | Typed node, operations, queries in `krav.core.verification` |
| IO | Implemented | JSON-LD serialization via `krav.io.graph` |
| Service | Partial | Basic CRUD and transitions; coverage report workflow not yet implemented |
| CLI | Implemented | CRUD, transition, link, record commands; coverage/untested stubs |
| Tests | Implemented | 46 integration tests in `tests/integration/commands/test_verification.py` |

## Summary

Verifications provide verification evidence:

- Linked to requirements via verifies/verifiedBy relationships
- Use one of four methods: test, inspection, demonstration, analysis
- Exist at unit, integration, system, and acceptance levels
- Track status from passing/failing execution
- Enable coverage analysis and gap identification
- Store metadata in graph.jsonlt with no frontmatter in prose files
- Implemented following three-layer architecture (core/io/service)

Verifications close the loop from requirements to evidence, proving that the system meets its obligations.
