# Concepts module type

## Overview

Concepts (CON-*) are explorations of how something could work. They capture design thinking, architectural options, decisions made and their rationale, research findings, and any other crystallized thinking that informs the project.

In INCOSE terms, concepts are "lifecycle concepts," the raw material from which the team derives needs through formal transformation. A concept represents exploration and understanding; a need represents a formalized stakeholder expectation extracted from that understanding.

## Purpose

Concepts serve these roles:

**Exploration space**: before committing to requirements, teams explore options. Concepts provide a place to think through alternatives, tradeoffs, and implications without the formality of requirements.

**Decision capture**: when the team makes architectural or design decisions, concepts record what the team decided, what alternatives the team considered, and why. This prevents relitigating settled questions.

**Research repository**: technical research, spike findings, proof-of-concept results, and external reference material can live as concepts.

**Transformation source**: when a concept crystallizes, it becomes the source for formal transformation into needs. The derivesFrom relationship maintains traceability from needs back to the concepts that informed them.

## Lifecycle

Concepts progress through states:

```text
draft → exploring → crystallized → formalized → superseded
```

| State        | Description                                            |
|--------------|--------------------------------------------------------|
| draft        | Initial capture, incomplete thinking                   |
| exploring    | Active exploration, the team evaluates options         |
| crystallized | Thinking complete, ready to formalize into needs       |
| formalized   | The team derived needs; concept is reference material  |
| superseded   | Replaced by newer understanding                        |

State transitions:

- `draft → exploring`: Work begins on fleshing out the concept
- `exploring → crystallized`: Exploration complete, understanding settled
- `crystallized → formalized`: the user ran `krav conceptformalize`
- `* → superseded`: New concept replaces this one (via supersedes relationship)

## Concept types

The `conceptType` field categorizes the nature of the exploration:

| Type          | Description                                             | Examples                                              |
|---------------|---------------------------------------------------------|-------------------------------------------------------|
| architectural | System structure, module boundaries, decomposition      | `Parser subsystem architecture`, `Service boundaries` |
| operational   | How stakeholders use, operate, or experience the system | `CLI user workflows`, `Error recovery experience`     |
| technical     | Internal mechanisms, algorithms, data structures        | `Tokenization algorithm`, `Graph storage format`      |
| interface     | Contracts between components, APIs, protocols           | `Parser-CLI interface`, `Plugin API design`           |
| process       | Workflows, procedures, ways of working                  | `Release process`, `Code review workflow`             |
| integration   | Connections with external systems and standards         | `INCOSE alignment`, `Git integration`                 |

## Storage model

Krav stores concept metadata in `graph.jsonlt` as JSON-LD compact form. Prose files have no frontmatter; `graph.jsonlt` is the single source of truth for all structured data.

```json
{"@context": "context.jsonld", "@id": "CON-C0NC3PT5", "@type": "Concept", "title": "Parser architecture", "status": "exploring", "conceptType": "architectural", "content": "concepts/20260103164500-C0NC3PT5-parser-architecture.md", "informs": {"@id": "MOD-P4RS3R01"}}
```

Fields:

- `@id`: Unique identifier (CON-XXXXXXXX format)
- `@type`: Always "Concept"
- `title`: Human-readable title
- `description`: Brief description (optional)
- `status`: Lifecycle state (draft, exploring, crystallized, formalized, superseded)
- `conceptType`: Type of exploration (architectural, operational, technical, interface, process, integration)
- `content`: Path to prose file relative to `.krav/`
- `created`, `updated`: ISO 8601 timestamps
- `tags`: Array of strings (optional)

## Prose structure

Concept prose files have flexible structure, but commonly include:

```markdown
# Parser architecture

## Context

What situation or problem prompted this exploration?

## Options considered

### Option A: Recursive descent

Description, pros, cons...

### Option B: PEG parser

Description, pros, cons...

## Decision

What was decided and why?

## Implications

What follows from this decision? What constraints does it impose?

## Open questions

What remains unresolved?
```

For research/spike concepts:

```markdown
# Tokenization performance spike

## Goal

What were we trying to learn?

## Approach

What did we try?

## Findings

What did we learn?

## Recommendations

What should we do based on this?
```

## Relationships

The concept's JSON-LD record embeds relationships using `{"@id": "..."}` values.

### Outgoing relationships

| Property     | Target | Cardinality | Description                            |
|--------------|--------|-------------|----------------------------------------|
| derivesFrom  | CON-*  | Multi       | This concept builds on other concepts  |
| supersedes   | CON-*  | Single      | This concept replaces an older one     |
| informs      | MOD-*  | Single      | Module this concept informs (informal) |

### Incoming relationships (queried via graph)

| Property     | Source | Description                           |
|--------------|--------|---------------------------------------|
| derivesFrom  | NED-*  | Needs derived from this concept       |
| derivesFrom  | CON-*  | Other concepts that build on this one |
| supersedes   | CON-*  | Concept that replaced this one        |

Example with relationships:

```json
{"@context": "context.jsonld", "@id": "CON-N3W3R001", "@type": "Concept", "title": "Refined architecture", "status": "draft", "conceptType": "architectural", "derivesFrom": [{"@id": "CON-0LD3R001"}, {"@id": "CON-SP1K3456"}], "supersedes": {"@id": "CON-D3PR3C8D"}, "informs": {"@id": "MOD-P4RS3R01"}}
```

## Formalization

When a concept crystallizes, formalization extracts stakeholder expectations as needs:

```bash
Krav conceptformalize CON-C0NC3PT5
```

This interactive process:

1. Identifies stakeholders affected by the concept
2. For each stakeholder, extracts expectations
3. Creates NED-* records with derivesFrom relationships back to the concept
4. Transitions the concept to `formalized` state

A single concept may produce multiple needs for different stakeholders or different aspects of the same stakeholder's expectations.

## Implementation architecture

Concept features follow the three-layer architecture (see CON-GR4PH4RC).

### Typed node

ConceptNode is an independent dataclass with typed fields. Graph stores ConceptNode directly, and IO serializes directly to/from ConceptNode (see CON-GR4PH4RC for the full architecture).

```python
@dataclass(frozen=True, slots=True)
class ConceptNode:
    """Concept module—exploration and crystallized thinking."""
    id: str
    title: str
    status: ConceptStatus          # Typed enum, not str
    concept_type: ConceptType      # Typed enum
    content: str = ""              # Path to prose file
    description: str = ""
    created: datetime | None = None
    updated: datetime | None = None
```

All fields use proper types. The IO layer creates ConceptNode directly from JSON-LD records, preserving all type-specific fields.

### Core layer (`krav.core.concept`)

Pure functions and typed data structures:

```python
# Types
class ConceptStatus(StrEnum):
    DRAFT = "draft"
    EXPLORING = "exploring"
    CRYSTALLIZED = "crystallized"
    FORMALIZED = "formalized"
    SUPERSEDED = "superseded"

class ConceptType(StrEnum):
    ARCHITECTURAL = "architectural"
    OPERATIONAL = "operational"
    TECHNICAL = "technical"
    INTERFACE = "interface"
    PROCESS = "process"
    INTEGRATION = "integration"

# Typed node
@dataclass(frozen=True, slots=True)
class ConceptNode(Node):
    concept_type: ConceptType = ConceptType.ARCHITECTURAL
    content: str = ""

# Operations (pure functions)
def from_node(node: Node) -> ConceptNode: ...
def with_status(concept: ConceptNode, status: ConceptStatus) -> ConceptNode: ...
def can_transition(concept: ConceptNode, target: ConceptStatus) -> bool: ...

# Queries (pure, take Graph)
def get(graph: Graph, concept_id: str) -> ConceptNode | None: ...
def list_all(graph: Graph) -> tuple[ConceptNode, ...]: ...
def list_by_status(graph: Graph, status: ConceptStatus) -> tuple[ConceptNode, ...]: ...
def informs(graph: Graph, concept_id: str) -> str | None: ...  # Returns module ID
def derived_needs(graph: Graph, concept_id: str) -> frozenset[str]: ...  # Returns need IDs
```

### Service layer (`krav.service.concept`)

Orchestrates core and IO:

```python
# Thin query wrappers
def get(store: GraphStore, concept_id: str) -> ConceptNode | None: ...
def list_all(store: GraphStore) -> tuple[ConceptNode, ...]: ...

# Mutations
def create(store: GraphStore, title: str, concept_type: ConceptType, ...) -> ConceptNode: ...
def update(store: GraphStore, concept_id: str, **fields) -> ConceptNode: ...
def delete(store: GraphStore, concept_id: str) -> None: ...
def transition(store: GraphStore, concept_id: str, target: ConceptStatus) -> ConceptNode: ...

# Workflows
def formalize(store: GraphStore, concept_id: str, needs: list[NeedSpec]) -> list[NeedNode]: ...
```

## CLI commands

```bash
# CRUD
Krav conceptcreate --title "Parser architecture" --type architectural
Krav conceptshow CON-C0NC3PT5
Krav conceptlist
Krav conceptlist --status exploring --type architectural
Krav conceptupdate CON-C0NC3PT5 --status crystallized
Krav conceptdelete CON-C0NC3PT5

# Lifecycle
Krav concepttransition CON-C0NC3PT5 --to crystallized
Krav conceptformalize CON-C0NC3PT5

# Relationships
Krav conceptlink CON-C0NC3PT5 --supersedes CON-0LD3R123
Krav conceptlink CON-C0NC3PT5 --informs MOD-P4RS3R01
```

## Examples

### Architectural concept

```json
{"@context": "context.jsonld", "@id": "CON-K7M3NP2Q", "@type": "Concept", "title": "krav data model", "status": "crystallized", "conceptType": "architectural", "content": "concepts/20260103133127-K7M3NP2Q-data-model.md", "informs": {"@id": "MOD-OAPSROOT"}}
```

### Technical spike

```json
{"@context": "context.jsonld", "@id": "CON-SP1K3456", "@type": "Concept", "title": "JSONLT performance characteristics", "status": "formalized", "conceptType": "technical", "content": "concepts/20260102100000-SP1K3456-jsonlt-perf.md"}
```

### Process concept with derivation

```json
{"@context": "context.jsonld", "@id": "CON-PR0C3SS1", "@type": "Concept", "title": "Phase-gated execution model", "status": "exploring", "conceptType": "process", "derivesFrom": [{"@id": "CON-K7M3NP2Q"}], "content": "concepts/20260103140000-PR0C3SS1-phase-gates.md"}
```

## Relationship to modules

Concepts can have an informal `informs` relationship to modules, indicating which module the concept is primarily about. This is distinct from the formal derivation chain (CON → NED → REQ).

```text
CON-parser-arch (informs: MOD-parser)
    ↓ formalize
NED-parser-performance (module: MOD-parser)
NED-parser-extensibility (module: MOD-parser)
```

The `informs` relationship is for navigation and documentation. The `module` field on needs is the formal ownership relationship.

## Implementation status

| Layer | Status | Notes |
|-------|--------|-------|
| Core | Implemented | Typed node, operations, queries in `krav.core.concept` |
| IO | Implemented | JSON-LD serialization via `krav.io.graph` |
| Service | Implemented | CRUD, transitions, supersede, link_informs in `krav.service.concept` |
| CLI | Implemented | All commands in `krav.cli._commands._concept` |
| Tests | Implemented | 38 integration tests in `tests/integration/commands/test_concept.py` |

## Summary

Concepts capture exploration and crystallized thinking:

- Progress from draft through exploring to crystallized
- Typed by nature of exploration (architectural, technical, process, etc.)
- Stored in graph.jsonlt with prose files having no frontmatter
- Formalize into needs via `krav conceptformalize`
- Maintain traceability via derivesFrom relationships
- Can inform modules via informal `informs` relationship
- Implemented following three-layer architecture (core/io/service)
