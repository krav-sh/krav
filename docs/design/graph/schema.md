# Schema

## Overview

This document defines the ontology schema (T-Box) for the ARCI knowledge graph: the RDF classes, RDF properties, identifier scheme, and JSON-LD vocabulary that constitute the graph's formal structure. Node types are RDF classes in the `arci:` namespace, datatype properties capture per-node attributes, and object properties express semantic relationships between nodes.

## JSON-LD vocabulary

The ARCI schema is an RDF vocabulary expressed in JSON-LD compact form. A shared context document defines the namespace mappings, class aliases, and property definitions that let consumers interpret JSON-LD documents as RDF triples.

**Namespace**: `https://arci.dev/schema#`
**Prefix**: `arci`

The context document maps compact JSON keys to their full RDF IRIs and declares whether each property is a datatype property or an object property (via `@type: "@id"`):

```json
{
  "@context": {
    "@vocab": "https://arci.dev/schema#",
    "arci": "https://arci.dev/schema#",
    "dcterms": "http://purl.org/dc/terms/",
    "prov": "http://www.w3.org/ns/prov#",
    "oslc_rm": "http://open-services.net/ns/rm#",
    "oslc_qm": "http://open-services.net/ns/qm#",
    "oslc_cm": "http://open-services.net/ns/cm#",
    "oslc_config": "http://open-services.net/ns/config#",
    "rdfs": "http://www.w3.org/2000/01/rdf-schema#",

    "Module": "arci:Module",
    "Concept": "arci:Concept",
    "Need": "arci:Need",
    "Requirement": "arci:Requirement",
    "TestCase": "arci:TestCase",
    "Task": "arci:Task",
    "Defect": "arci:Defect",
    "Baseline": "arci:Baseline",
    "Stakeholder": "arci:Stakeholder",
    "TestPlan": "arci:TestPlan",
    "Developer": "arci:Developer",
    "Agent": "arci:Agent",

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
    "stakeholder": {"@id": "arci:stakeholder", "@type": "@id"},
    "operator": {"@id": "arci:operator", "@type": "@id"},
    "parentAgent": {"@id": "arci:parentAgent", "@type": "@id"},

    "implements": {"@id": "oslc_cm:implementsRequirement", "@type": "@id"},
    "generatedBy": {"@id": "prov:wasGeneratedBy", "@type": "@id"},
    "attributedTo": {"@id": "prov:wasAttributedTo", "@type": "@id"},

    "title": "dcterms:title",
    "description": "dcterms:description",
    "statement": "arci:statement",
    "status": "arci:status",
    "phase": "arci:phase",
    "processPhase": "arci:processPhase",
    "conceptType": "arci:conceptType",
    "priority": "arci:priority",
    "severity": "arci:severity",
    "category": "arci:category",
    "summary": "arci:summary",
    "sessionId": "arci:sessionId",
    "subagentId": "arci:subagentId",
    "created": {"@id": "dcterms:created", "@type": "xsd:dateTime"},
    "updated": {"@id": "dcterms:modified", "@type": "xsd:dateTime"},
    "tags": "arci:tags"
  }
}
```

## Node type taxonomy

The knowledge graph contains twelve node types, each with a variable-length prefix (2-4 characters):

| Prefix | Type | `@type` | Semantic category | Role |
|--------|------|---------|-------------------|------|
| CON | Concept | Concept | Intent | Exploration, design decisions, crystallized thinking |
| MOD | Module | Module | Structure | Architectural container, owns needs and requirements |
| NEED | Need | Need | Intent | Stakeholder expectation (validated) |
| REQ | Requirement | Requirement | Obligation | Design obligation (verified) |
| TC | Test case | `TestCase` | Evidence | Test specification for verifying requirements |
| TASK | Task | Task | Execution | Atomic work unit in a DAG |
| DEF | Defect | Defect | Quality | Identified problem requiring action |
| BSL | Baseline | Baseline | Configuration | Named snapshot of graph state at a point in time |
| STK | Stakeholder | Stakeholder | Intent | Named party with concerns about the system |
| TP | Test plan | TestPlan | Evidence | Test plan (not yet spec'd) |
| DEV | Developer | Developer | Provenance | Human actor who initiates sessions and makes decisions |
| AGT | Agent | Agent | Provenance | Claude Code session or subagent, ephemeral per invocation |

## Identifier scheme

All identifiers follow the pattern `PREFIX-NANOID`:

- **PREFIX**: variable-length uppercase prefix (2-4 characters) from the preceding table
- **`NANOID`**: 8-character Crockford Base32 nanoid

Crockford Base32 uses `0123456789ABCDEFGHJKMNPQRSTVWXYZ` (excludes the letters `I`, `L`, `O`, and `U` to avoid ambiguity). Identifiers are case-insensitive but conventionally uppercase.

```text
CON-K7M3NP2Q    Concept
MOD-A4F8R2X1    Module
NEED-B7G3M9K2   Need
REQ-C2H6N4P8    Requirement
TC-D9J5Q1R3     Test case
TASK-E3K8S6V2   Task
DEF-F1L4T7W5    Defect
BSL-R3L3AS31    Baseline
STK-H5N7P3Q9    Stakeholder
TP-G2H4J6K8     Test plan
DEV-J4R8T2W6    Developer
AGT-M5V9K3X7    Agent
```

The `@id` property in JSON-LD holds the identifier. The type prefix must be consistent with the `@type` value; a node with `@id: "CON-K7M3NP2Q"` must have `@type: "Concept"`.

## Common properties

All node types share these properties:

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `@id` | string | Yes | Unique identifier (TYPE-nanoid format) |
| `@type` | string | Yes | Node type (must match identifier prefix) |
| `@context` | string | Yes | Reference to the JSON-LD context document |
| `title` | string | Yes | Human-readable title (`dcterms:title`) |
| `description` | string | No | Extended description (`dcterms:description`) |
| `status` | string | Yes | Lifecycle state (type-specific enum) |
| `created` | datetime | No | ISO 8601 creation timestamp (`dcterms:created`) |
| `updated` | datetime | No | ISO 8601 last-modified timestamp (`dcterms:modified`) |
| `tags` | string[] | No | Freeform tags for filtering |
| `summary` | string | No | Inline prose for extended context beyond title and description |
| `generatedBy` | reference | No | Task that created this node (`prov:wasGeneratedBy`) |
| `attributedTo` | reference | No | Agent or developer responsible (`prov:wasAttributedTo`) |

## Prose files

Every node type can have an associated markdown file for extended prose that doesn't fit in structured fields. The system derives the path from the node's identifier rather than storing it on the node.

Prose files live in flat, type-specific directories under `.arci/`:

| Node type   | Directory              |
|-------------|------------------------|
| Concept     | `.arci/concepts/`      |
| Module      | `.arci/modules/`       |
| Stakeholder | `.arci/stakeholders/`  |
| Need        | `.arci/needs/`         |
| Requirement | `.arci/requirements/`  |
| Test case   | `.arci/test-cases/`    |
| Task        | `.arci/tasks/`         |
| Defect      | `.arci/defects/`       |
| Baseline    | `.arci/baselines/`     |
| Developer   | `.arci/developers/`    |
| Agent       | `.arci/agents/`        |

Filenames follow the pattern `{timestamp}-{NANOID}-{slug}.md`. The timestamp is `YYYYMMDDHHMMSS` and provides filesystem sort order. The nanoid is the same 8-character identifier from the node's `@id`. The slug is a human-readable kebab-case label. As an example, concept `CON-K7M3NP2Q` titled "Parser architecture" would have its prose at `.arci/concepts/20260103164500-K7M3NP2Q-parser-architecture.md`.

Resolution is mechanical: the node type determines the directory, and the nanoid matches the second segment of the filename when split by `-`. The timestamp and slug are filesystem conveniences for sorting and readability; the graph stores neither.

Not every node needs a prose file. Concepts almost always have one since exploration is their purpose. Tasks and defects often get by with the `summary` field for a paragraph or two of inline context, and only reach for a file when they need more room. Stakeholders, baselines, and other lightweight types rarely need files but the mechanism is there when they do.

The `summary` field and a prose file serve different purposes and can coexist on the same node. `summary` carries inline prose directly on the graph node, suited for a quick paragraph of context. The prose file is for extended content that would be unwieldy in a JSONLT line.

## Type-specific properties

Each node type has additional properties specific to its role. The summary below lists key fields; see entity-specific docs for complete definitions.

### Concept (CON-*)

| Property | Type | Description |
|----------|------|-------------|
| `conceptType` | enum | Category: architectural, operational, technical, interface, process, integration |

Status: draft → exploring → crystallized → superseded

See [Concepts](nodes/concepts.md) for full specification.

### Module (MOD-*)

| Property | Type | Description |
|----------|------|-------------|
| `phase` | enum | Current lifecycle phase: architecture, design, coding, integration, verification, validation |

Status: active → deprecated → archived

See [Modules](nodes/modules.md) for full specification.

### Stakeholder (STK-*)

| Property | Type | Description |
|----------|------|-------------|
| `concerns` | string | What this stakeholder cares about |

Status: active → archived

See [Stakeholders](nodes/stakeholders.md) for full specification.

### Need (NEED-*)

| Property | Type | Description |
|----------|------|-------------|
| `statement` | string | Stakeholder-perspective need statement |
| `rationale` | string | Why this need exists |

Status: draft → validated → addressed → superseded

See [Needs](nodes/needs.md) for full specification.

### Requirement (REQ-*)

| Property | Type | Description |
|----------|------|-------------|
| `statement` | string | Formal requirement statement (shall language) |
| `priority` | enum | Priority level |

Status: draft → approved → satisfied → superseded

See [Requirements](nodes/requirements.md) for full specification.

### Test case (TC-*)

| Property | Type | Description |
|----------|------|-------------|
| `method` | enum | Verification method: inspection, demonstration, test, analysis |
| `currentResult` | enum | Latest result: pass, fail, skip, unknown |
| `acceptanceCriteria` | string | Explicit pass/fail criteria |

Status: draft → specified → implemented → executable → obsolete

See [Test cases](nodes/test-cases.md) for full specification.

### Task (TASK-*)

| Property | Type | Description |
|----------|------|-------------|
| `processPhase` | enum | Lifecycle phase: architecture, design, coding, integration, verification, validation |
| `deliverables` | object[] | Array of deliverable records with `kind` discriminator |

Status: pending → active → complete → blocked → cancelled

See [Tasks](nodes/tasks.md) for full specification.

### Defect (DEF-*)

| Property | Type | Description |
|----------|------|-------------|
| `statement` | string | Full description of the problem |
| `severity` | enum | Impact level: critical, major, minor, trivial |
| `category` | enum | What kind of problem: missing, incorrect, ambiguous, inconsistent, non-verifiable, non-traceable, incomplete, superfluous, non-conformant, regression |
| `rationale` | string | For rejected/deferred: why this disposition |
| `deferralTarget` | string | For deferred: when to re-evaluate |
| `resolutionNotes` | string | How the team fixed the defect |
| `detectedInPhase` | enum | Module phase when the examination found the defect |

Status: open → confirmed → resolved → verified → closed (also: rejected, deferred)

See [Defects](nodes/defects.md) for full specification.

### Baseline (BSL-*)

| Property | Type | Description |
|----------|------|-------------|
| `scope` | string | What the baseline covers (subtree, module, etc.) |
| `commitSha` | string | Git commit SHA anchoring the graph state |
| `phase` | enum | Lifecycle phase this baseline captures |
| `approvedBy` | string | Who approved this baseline |
| `approvedAt` | datetime | When the approver approved it |
| `statistics` | object | Denormalized counts at baseline time |

Status: draft → proposed → approved → superseded

See [Baselines](nodes/baselines.md) for full specification.

### Developer (DEV-*)

| Property | Type | Description |
|----------|------|-------------|
| (no type-specific properties beyond common fields) | | |

Status: active → archived

See [Developers](nodes/developers.md) for full specification.

### Agent (AGT-*)

| Property | Type | Description |
|----------|------|-------------|
| `sessionId` | string | Claude Code session identifier (required) |
| `subagentId` | string | Subagent identifier within session (null for main session agent) |
| `startedAt` | datetime | When the session or subagent started |
| `endedAt` | datetime | When the session or subagent ended |

Status: active → closed

See [Agents](nodes/agents.md) for full specification.

## JSON-LD representation

Each node in the graph is a JSON-LD document in compact form. The `@context` key references the shared vocabulary, `@id` carries the node's identifier, `@type` names the RDF class, and remaining keys are RDF properties; either datatype properties (literal values) or object properties (references to other nodes via `{"@id": "..."}` values).

```json
{
  "@context": "https://arci.dev/schema#",
  "@id": "NEED-B7G3M9K2",
  "@type": "Need",
  "title": "Quick feedback",
  "status": "validated",
  "module": {"@id": "MOD-A4F8R2X1"},
  "derivesFrom": [{"@id": "CON-K7M3NP2Q"}]
}
```

Key representation conventions:

- Object property values use the `{"@id": "..."}` form to denote references to other RDF resources
- Multi-valued object properties use arrays: `[{"@id": "..."}, {"@id": "..."}]`
- Object properties with metadata use extended objects: `{"@id": "...", "budget": "50ms"}`

## Extensibility

### Adding a new node type

1. Choose a unique uppercase prefix (2-4 characters)
2. Add the `@type` mapping to the JSON-LD context
3. Define type-specific RDF properties and add them to the context
4. Define valid predicates (domain/range) in [predicates](predicates.md)
5. Add structural constraints in [constraints](constraints.md)
6. Create an entity-specific doc with lifecycle and field definitions

### Adding a new predicate

1. Add the predicate to the JSON-LD context with `"@type": "@id"` for object properties
2. Define domain, range, cardinality, and structural constraints in [predicates](predicates.md)
3. Add validation rules in [constraints](constraints.md)
4. Document suspect propagation behavior if applicable

### Adding a new property

1. Add the property to the JSON-LD context
2. Document it in the relevant entity-specific doc
3. Update the type-specific properties table in this document
