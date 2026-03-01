# Needs module type

## Overview

Needs (NED-*) are stakeholder expectations—what stakeholders need an module to do or be. They're expressed from the stakeholder's perspective and validated against stakeholder intent.

The key INCOSE distinction: a need is an "agreed-to expectation" while a requirement is an "agreed-to obligation." Needs express what stakeholders want; requirements express what the system must do to satisfy those needs.

## Purpose

Needs serve several roles:

**Stakeholder voice**: Needs capture expectations in stakeholder terms, not implementation terms. "Users need quick feedback when parsing" rather than "Parser shall complete in 10ms."

**Validation target**: Needs are validated—did we capture what stakeholders actually need? This is distinct from verification, which checks whether requirements are met.

**Derivation source**: Requirements derive from needs. A single need may produce multiple requirements; a requirement may satisfy multiple needs. The derivesFrom relationship maintains traceability.

**Scope anchor**: When scope questions arise ("do we need this feature?"), needs provide the answer by connecting back to stakeholder expectations.

## Needs vs requirements

| Aspect      | Need                                      | Requirement                 |
|-------------|-------------------------------------------|-----------------------------|
| Perspective | Stakeholder                               | System                      |
| Language    | "Users need...", "Contributors expect..." | "The system shall..."       |
| Validation  | Validated (right thing captured?)         | Verified (built correctly?) |
| Precision   | May be qualitative                        | Must be verifiable          |
| INCOSE term | Agreed-to expectation                     | Agreed-to obligation        |

Example transformation:

```
Need: "Users need quick feedback when parsing fails"
    ↓ derive
Requirement: "The parser shall report the first syntax error within 50ms"
Requirement: "Error messages shall include line number and column"
Requirement: "Error messages shall suggest corrections when possible"
```

## Lifecycle

Needs progress through states:

```
draft → proposed → validated → satisfied → obsolete
```

| State     | Description                                          |
|-----------|------------------------------------------------------|
| draft     | Initial capture, not yet reviewed                    |
| proposed  | Ready for stakeholder validation                     |
| validated | Stakeholders confirm this captures their expectation |
| satisfied | Derived requirements are verified; need is met       |
| obsolete  | No longer relevant (stakeholder needs changed)       |

State transitions:

- `draft → proposed`: Need is ready for validation
- `proposed → validated`: Stakeholder confirms this is what they need
- `validated → satisfied`: All derived requirements are verified
- `* → obsolete`: Need is no longer relevant

## Stakeholder classes

Needs are categorized by stakeholder:

| Stakeholder | Description                         | Example needs                            |
|-------------|-------------------------------------|------------------------------------------|
| user        | End users of the system             | Usability, performance, reliability      |
| contributor | People contributing to the project  | Onboarding, dev environment, testing     |
| maintainer  | People maintaining the project      | Automation, release process, monitoring  |
| integrator  | People integrating with the system  | API stability, documentation, versioning |
| operator    | People operating the system         | Deployment, configuration, observability |
| ecosystem   | Broader community, standards bodies | Interoperability, compliance, openness   |

## Storage model

Need metadata is stored in `graph.jsonlt` as JSON-LD compact form. There is no frontmatter in prose files—`graph.jsonlt` is the single source of truth for all structured data.

```json
{"@context": "context.jsonld", "@id": "NED-B7G3M9K2", "@type": "Need", "title": "Quick parsing feedback", "module": {"@id": "MOD-A4F8R2X1"}, "stakeholder": "user", "statement": "Users need quick feedback when parsing fails", "rationale": "Slow error reporting disrupts developer flow", "status": "validated", "priority": "must", "derivesFrom": [{"@id": "CON-K7M3NP2Q"}]}
```

Fields:

- `@id`: Unique identifier (NED-XXXXXXXX format)
- `@type`: Always "Need"
- `title`: Human-readable title
- `module`: Module this need belongs to (required)
- `stakeholder`: Stakeholder class (user, contributor, maintainer, integrator, operator, ecosystem)
- `statement`: The need statement in stakeholder terms
- `rationale`: Why this need exists (optional)
- `status`: Lifecycle state (draft, proposed, validated, satisfied, obsolete)
- `priority`: MoSCoW priority (must, should, could, wont)
- `validationEvidence`: Evidence of stakeholder validation (optional)
- `created`, `updated`: ISO 8601 timestamps
- `tags`: Array of strings (optional)

## Priority levels

Needs use MoSCoW prioritization:

| Priority | Description                                |
|----------|--------------------------------------------|
| must     | Essential; project fails without this      |
| should   | Important; significant value, not critical |
| could    | Desirable; nice to have if time permits    |
| wont     | Explicitly out of scope (for this release) |

## Relationships

Relationships are embedded in the need's JSON-LD record using `{"@id": "..."}` values.

### Outgoing relationships

| Property    | Target | Cardinality | Description                              |
|-------------|--------|-------------|------------------------------------------|
| module      | MOD-*  | Single      | Module this need belongs to              |
| derivesFrom | CON-*  | Multi       | Concept(s) this need was formalized from |

### Incoming relationships (queried via graph)

| Property    | Source | Description                          |
|-------------|--------|--------------------------------------|
| derivesFrom | REQ-*  | Requirements derived from this need  |
| derivesFrom | NED-*  | Child needs (for need decomposition) |

Example with relationships:

```json
{"@context": "context.jsonld", "@id": "NED-B7G3M9K2", "@type": "Need", "title": "Quick parsing feedback", "module": {"@id": "MOD-A4F8R2X1"}, "stakeholder": "user", "statement": "Users need quick feedback when parsing fails", "status": "validated", "priority": "must", "derivesFrom": [{"@id": "CON-K7M3NP2Q"}, {"@id": "CON-P3RF0RM1"}]}
```

## Derivation

When a need is validated, derivation produces requirements:

```bash
arci needderive NED-B7G3M9K2
```

This process:

1. Analyzes the need statement
2. Produces verifiable requirements that satisfy the need
3. Creates REQ-* records with derivesFrom relationships back to the need
4. Each requirement gets verification criteria

A single need typically produces 1-5 requirements. Complex needs may be decomposed into child needs first.

## Validation

Validation confirms that the need accurately captures stakeholder expectations:

```bash
arci needvalidate NED-B7G3M9K2 --evidence "User interviews Jan 2026"
```

Validation methods:

| Method      | Description                           |
|-------------|---------------------------------------|
| interview   | Direct stakeholder interviews         |
| survey      | Stakeholder surveys or questionnaires |
| observation | Observing stakeholder behavior        |
| prototype   | Stakeholder feedback on prototypes    |
| review      | Stakeholder review of need statements |

## Implementation architecture

Need functionality follows the three-layer architecture (see CON-GR4PH4RC).

### Typed node

NeedNode is an independent dataclass with typed fields. Graph stores NeedNode directly, and IO serializes directly to/from NeedNode (see CON-GR4PH4RC for the full architecture).

```python
@dataclass(frozen=True, slots=True)
class NeedNode:
    """Need module—stakeholder expectation."""
    id: str
    title: str
    status: NeedStatus             # Typed enum
    stakeholder: Stakeholder       # Typed enum
    priority: Priority             # Typed enum
    statement: str                 # Type-specific field
    rationale: str = ""            # Type-specific field
    validation_evidence: str = ""  # Type-specific field
    description: str = ""
    created: datetime | None = None
    updated: datetime | None = None
```

All fields use proper types. The IO layer creates NeedNode directly from JSON-LD records, preserving all type-specific fields like `statement`, `rationale`, and `validation_evidence`.

### Core layer (arci.core.need)

Pure functions and typed data structures:

```python
# Types
class NeedStatus(StrEnum):
    DRAFT = "draft"
    PROPOSED = "proposed"
    VALIDATED = "validated"
    SATISFIED = "satisfied"
    OBSOLETE = "obsolete"

class Stakeholder(StrEnum):
    USER = "user"
    CONTRIBUTOR = "contributor"
    MAINTAINER = "maintainer"
    INTEGRATOR = "integrator"
    OPERATOR = "operator"
    ECOSYSTEM = "ecosystem"

class Priority(StrEnum):
    MUST = "must"
    SHOULD = "should"
    COULD = "could"
    WONT = "wont"

# Typed node
@dataclass(frozen=True, slots=True)
class NeedNode(Node):
    stakeholder: Stakeholder = Stakeholder.USER
    statement: str = ""
    rationale: str = ""
    priority: Priority = Priority.SHOULD
    validation_evidence: str = ""

# Operations (pure functions)
def from_node(node: Node) -> NeedNode: ...
def with_status(need: NeedNode, status: NeedStatus) -> NeedNode: ...
def can_transition(need: NeedNode, target: NeedStatus) -> bool: ...

# Queries (pure, take Graph)
def get(graph: Graph, need_id: str) -> NeedNode | None: ...
def list_all(graph: Graph) -> tuple[NeedNode, ...]: ...
def list_by_module(graph: Graph, module_id: str) -> tuple[NeedNode, ...]: ...
def list_by_stakeholder(graph: Graph, stakeholder: Stakeholder) -> tuple[NeedNode, ...]: ...
def owning_module(graph: Graph, need_id: str) -> str | None: ...
def derives_from_concepts(graph: Graph, need_id: str) -> frozenset[str]: ...
def derived_requirements(graph: Graph, need_id: str) -> frozenset[str]: ...
```

### Service layer (arci.service.need)

Orchestrates core and IO:

```python
# Thin query wrappers
def get(store: GraphStore, need_id: str) -> NeedNode | None: ...
def list_all(store: GraphStore) -> tuple[NeedNode, ...]: ...
def list_by_module(store: GraphStore, module_id: str) -> tuple[NeedNode, ...]: ...

# Mutations
def create(
    store: GraphStore,
    module_id: str,
    stakeholder: Stakeholder,
    statement: str,
    derives_from: list[str] | None = None,  # concept IDs
    ...
) -> NeedNode: ...
def update(store: GraphStore, need_id: str, **fields) -> NeedNode: ...
def delete(store: GraphStore, need_id: str) -> None: ...
def transition(store: GraphStore, need_id: str, target: NeedStatus) -> NeedNode: ...

# Workflows
def validate(store: GraphStore, need_id: str, evidence: str) -> NeedNode: ...
def derive(store: GraphStore, need_id: str, requirements: list[RequirementSpec]) -> list[RequirementNode]: ...
```

## CLI commands

```bash
# CRUD
arci needcreate --module MOD-A4F8R2X1 --stakeholder user \
  --statement "Users need quick feedback when parsing fails"
arci needshow NED-B7G3M9K2
arci needlist
arci needlist --module MOD-A4F8R2X1 --stakeholder user
arci needupdate NED-B7G3M9K2 --priority must
arci needdelete NED-B7G3M9K2

# Lifecycle
arci needtransition NED-B7G3M9K2 --to proposed
arci needvalidate NED-B7G3M9K2 --evidence "..."
arci needderive NED-B7G3M9K2

# Relationships
arci needlink NED-B7G3M9K2 --derives-from CON-K7M3NP2Q

# Traceability
arci needtrace NED-B7G3M9K2  # Show concept → need → requirements chain
```

## Examples

### User need

```json
{"@context": "context.jsonld", "@id": "NED-B7G3M9K2", "@type": "Need", "title": "Quick parsing feedback", "module": {"@id": "MOD-A4F8R2X1"}, "stakeholder": "user", "statement": "Users need quick feedback when parsing fails", "rationale": "Slow error reporting disrupts developer flow", "status": "validated", "priority": "must", "validationEvidence": "User interviews confirmed <3s threshold", "derivesFrom": [{"@id": "CON-K7M3NP2Q"}]}
```

### Contributor need

```json
{"@context": "context.jsonld", "@id": "NED-C0NTR1B1", "@type": "Need", "title": "Clear project setup guidance", "module": {"@id": "MOD-OAPSROOT"}, "stakeholder": "contributor", "statement": "Contributors need clear guidance on project setup", "rationale": "Reduces onboarding friction and first-contribution time", "status": "validated", "priority": "should"}
```

### Ecosystem need

```json
{"@context": "context.jsonld", "@id": "NED-3C0SYS01", "@type": "Need", "title": "Accurate package metadata", "module": {"@id": "MOD-OAPSROOT"}, "stakeholder": "ecosystem", "statement": "The project needs accurate package metadata", "rationale": "Enables discovery and dependency resolution", "status": "validated", "priority": "must"}
```

## Relationship to concepts and requirements

Needs sit between concepts and requirements in the formal transformation chain:

```
CON-* (exploration)
    ↓ formalize
NED-* (expectation, validated)
    ↓ derive
REQ-* (obligation, verified)
```

Each transformation produces traceability via derivesFrom relationships:

```
CON-K7M3NP2Q "Data model concept"
    ↑ derivesFrom
NED-B7G3M9K2 "Users need quick feedback"
    ↑ derivesFrom
REQ-C2H6N4P8 "Parser shall report errors within 50ms"
```

This chain answers "why does this requirement exist?" by tracing back through needs to concepts.

## Implementation status

| Layer | Status | Notes |
|-------|--------|-------|
| Core | Implemented | Typed node, operations, queries in `arci.core.need` |
| IO | Implemented | JSON-LD serialization via `arci.io.graph` |
| Service | Implemented | CRUD, transitions, validate, derivation management in `arci.service.need` |
| CLI | Implemented | 9 commands: create, show, list, update, delete, transition, validate, link, trace |

## Summary

Needs capture stakeholder expectations:

- Expressed in stakeholder terms, not implementation terms
- Validated against stakeholder intent
- Serve as derivation source for requirements
- Maintain traceability via derivesFrom relationships
- Progress from draft through validated to satisfied
- Store metadata in graph.jsonlt with no frontmatter in prose files
- Implemented following three-layer architecture (core/io/service)

Needs are the bridge between exploration (concepts) and obligation (requirements), ensuring that what we build traces back to what stakeholders actually need.
