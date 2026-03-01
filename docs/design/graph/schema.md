# Schema

## Overview

This document defines the ontology schema (T-Box) for the arci knowledge graph: the node types, identifier scheme, properties, and JSON-LD vocabulary that constitute the graph's formal structure.

## JSON-LD vocabulary

The arci schema uses JSON-LD compact form with a shared context that defines the vocabulary namespace.

**Namespace**: `https://arci.dev/schema#`
**Prefix**: `arci`

The context file lives at `.arci/context.jsonld` and is referenced by every record in `graph.jsonlt`:

```json
{
  "@context": {
    "@vocab": "https://arci.dev/schema#",
    "arci": "https://arci.dev/schema#",

    "Module": "arci:Module",
    "Concept": "arci:Concept",
    "Need": "arci:Need",
    "Requirement": "arci:Requirement",
    "Verification": "arci:Verification",
    "Task": "arci:Task",
    "Defect": "arci:Defect",
    "Baseline": "arci:Baseline",

    "childOf": {"@id": "arci:childOf", "@type": "@id"},
    "derivesFrom": {"@id": "arci:derivesFrom", "@type": "@id"},
    "allocatesTo": {"@id": "arci:allocatesTo", "@type": "@id"},
    "dependsOn": {"@id": "arci:dependsOn", "@type": "@id"},
    "verifiedBy": {"@id": "arci:verifiedBy", "@type": "@id"},
    "subject": {"@id": "arci:subject", "@type": "@id"},
    "detectedBy": {"@id": "arci:detectedBy", "@type": "@id"},
    "generates": {"@id": "arci:generates", "@type": "@id"},
    "informs": {"@id": "arci:informs", "@type": "@id"},
    "module": {"@id": "arci:module", "@type": "@id"},

    "title": "arci:title",
    "description": "arci:description",
    "statement": "arci:statement",
    "status": "arci:status",
    "phase": "arci:phase",
    "processPhase": "arci:processPhase",
    "conceptType": "arci:conceptType",
    "priority": "arci:priority",
    "severity": "arci:severity",
    "category": "arci:category",
    "content": "arci:content",
    "created": {"@id": "arci:created", "@type": "xsd:dateTime"},
    "updated": {"@id": "arci:updated", "@type": "xsd:dateTime"},
    "tags": "arci:tags"
  }
}
```

## Node type taxonomy

The knowledge graph contains eight node types, each with a three-character prefix:

| Prefix | Type | `@type` | Semantic category | Role |
|--------|------|---------|-------------------|------|
| CON | Concept | Concept | Intent | Exploration, design decisions, crystallized thinking |
| MOD | Module | Module | Structure | Architectural container, owns needs and requirements |
| NED | Need | Need | Intent | Stakeholder expectation (validated) |
| REQ | Requirement | Requirement | Obligation | Design obligation (verified) |
| VRF | Verification | Verification | Evidence | Verification evidence for requirements |
| TSK | Task | Task | Execution | Atomic work unit in a DAG |
| DEF | Defect | Defect | Quality | Identified problem requiring action |
| BSL | Baseline | Baseline | Configuration | Named snapshot of graph state at a point in time |

## Identifier scheme

All identifiers follow the pattern `TYPE-NANOID`:

- **TYPE**: 3-character prefix from the table above
- **NANOID**: 8-character Crockford Base32 string

Crockford Base32 uses `0123456789ABCDEFGHJKMNPQRSTVWXYZ` (excludes I, L, O, U to avoid ambiguity). Identifiers are case-insensitive but conventionally uppercase.

```
CON-K7M3NP2Q   Concept
MOD-A4F8R2X1   Module
NED-B7G3M9K2   Need
REQ-C2H6N4P8   Requirement
VRF-D9J5Q1R3   Verification
TSK-E3K8S6V2   Task
DEF-F1L4T7W5   Defect
BSL-R3L3AS31   Baseline
```

The identifier is stored as the `@id` property in JSON-LD. The type prefix must be consistent with the `@type` value — a node with `@id: "CON-K7M3NP2Q"` must have `@type: "Concept"`.

## Common properties

All node types share these properties:

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `@id` | string | Yes | Unique identifier (TYPE-NANOID format) |
| `@type` | string | Yes | Node type (must match identifier prefix) |
| `@context` | string | Yes | Reference to `context.jsonld` |
| `title` | string | Yes | Human-readable title |
| `description` | string | No | Extended description |
| `status` | string | Yes | Lifecycle state (type-specific enum) |
| `created` | datetime | No | ISO 8601 creation timestamp |
| `updated` | datetime | No | ISO 8601 last-modified timestamp |
| `tags` | string[] | No | Freeform tags for filtering |
| `content` | string | No | Path to prose file (relative to `.arci/`) |

## Type-specific properties

Each node type has additional properties specific to its role. The summary below lists key fields; see entity-specific docs for complete definitions.

### Concept (CON-*)

| Property | Type | Description |
|----------|------|-------------|
| `conceptType` | enum | Category: architectural, operational, technical, interface, process, integration |

Status: draft → exploring → crystallized → superseded

See [Concepts](../intent/concepts.md) for full specification.

### Module (MOD-*)

| Property | Type | Description |
|----------|------|-------------|
| `phase` | enum | Current lifecycle phase: architecture, design, implementation, integration, verification, validation |

Status: active → deprecated → archived

See [Modules](../requirements/modules.md) for full specification.

### Need (NED-*)

| Property | Type | Description |
|----------|------|-------------|
| `statement` | string | Stakeholder-perspective need statement |
| `stakeholder` | string | Who has this expectation |
| `rationale` | string | Why this need exists |

Status: draft → validated → addressed → superseded

See [Needs](../intent/needs.md) for full specification.

### Requirement (REQ-*)

| Property | Type | Description |
|----------|------|-------------|
| `statement` | string | Formal requirement statement (shall language) |
| `priority` | enum | Priority level |

Status: draft → approved → satisfied → superseded

See [Requirements](../requirements/requirements.md) for full specification.

### Verification (VRF-*)

| Property | Type | Description |
|----------|------|-------------|
| `method` | enum | Verification method: inspection, demonstration, test, analysis |

Status: planned → implementing → passing → failing → blocked

See [Verifications](../verification/verifications.md) for full specification.

### Task (TSK-*)

| Property | Type | Description |
|----------|------|-------------|
| `processPhase` | enum | Lifecycle phase: architecture, design, implementation, integration, verification, validation |
| `deliverables` | object[] | Array of deliverable records with `kind` discriminator |

Status: pending → active → complete → blocked → cancelled

See [Tasks](../execution/tasks.md) for full specification.

### Defect (DEF-*)

| Property | Type | Description |
|----------|------|-------------|
| `statement` | string | Full description of the problem |
| `severity` | enum | Impact level: critical, major, minor, trivial |
| `category` | enum | What kind of problem: missing, incorrect, ambiguous, inconsistent, non-verifiable, non-traceable, incomplete, superfluous, non-conformant, regression |
| `rationale` | string | For rejected/deferred: why this disposition |
| `deferralTarget` | string | For deferred: when to re-evaluate |
| `resolutionNotes` | string | How the defect was fixed |
| `detectedInPhase` | enum | Module phase when defect was found |

Status: open → confirmed → resolved → verified → closed (also: rejected, deferred)

See [Defects](../verification/defects.md) for full specification.

### Baseline (BSL-*)

| Property | Type | Description |
|----------|------|-------------|
| `scope` | string | What the baseline covers (subtree, module, etc.) |
| `commitSha` | string | Git commit SHA anchoring the graph state |
| `phase` | enum | Lifecycle phase this baseline captures |
| `approvedBy` | string | Who approved this baseline |
| `approvedAt` | datetime | When it was approved |
| `statistics` | object | Denormalized counts at baseline time |

Status: draft → proposed → approved → superseded

See [Baselines](../requirements/baselines.md) for full specification.

## Storage model

All nodes are stored in a single `graph.jsonlt` file using JSON-LD compact form. Each line is a complete node with its outgoing relationships embedded as properties:

```json
{"@context": "context.jsonld", "@id": "MOD-OAPSROOT", "@type": "Module", "title": "arci", "phase": "architecture", "status": "active"}
{"@context": "context.jsonld", "@id": "NED-B7G3M9K2", "@type": "Need", "title": "Quick feedback", "module": {"@id": "MOD-A4F8R2X1"}, "derivesFrom": [{"@id": "CON-K7M3NP2Q"}], "status": "validated"}
{"@context": "context.jsonld", "@id": "DEF-F1L4T7W5", "@type": "Defect", "title": "Error messages missing line numbers", "module": {"@id": "MOD-A4F8R2X1"}, "severity": "major", "status": "confirmed", "subject": {"@id": "REQ-3RR0R001"}}
```

Key storage properties:

- `@id` is the JSONLT key (last-wins resolution for updates)
- Relationships are properties with `{"@id": "..."}` values
- Multi-valued relationships use arrays: `[{"@id": "..."}, {"@id": "..."}]`
- Relationships with metadata use extended objects: `{"@id": "...", "budget": "50ms"}`
- JSONLT is append-only; updates add new lines rather than rewriting

## Extensibility

### Adding a new node type

1. Choose a unique 3-character prefix
2. Add the `@type` mapping to `context.jsonld`
3. Define type-specific properties and add them to the context
4. Define valid predicates (domain/range) in [predicates](predicates.md)
5. Add structural constraints in [constraints](constraints.md)
6. Create an entity-specific doc with lifecycle, fields, and CLI commands

### Adding a new predicate

1. Add the predicate to `context.jsonld` with `"@type": "@id"` for relationship predicates
2. Define domain, range, cardinality, and structural constraints in [predicates](predicates.md)
3. Add validation rules in [constraints](constraints.md)
4. Document suspect propagation behavior if applicable

### Adding a new property

1. Add the property to `context.jsonld`
2. Document it in the relevant entity-specific doc
3. Update the type-specific properties table in this document
