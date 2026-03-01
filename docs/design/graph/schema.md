# Schema

## Overview

This document defines the ontology schema (T-Box) for the Krav knowledge graph: the node types, properties, identifier scheme, and vocabulary that constitute the graph's formal structure. Node types map to DuckDB vertex tables, predicates map to edge tables, and the SQL/PGQ property graph is a view layer over these relational tables.

## Storage model

The knowledge graph runtime is an in-memory DuckDB instance with the DuckPGQ community extension (SQL/PGQ, part of the SQL:2023 standard). Each node type maps to a DuckDB vertex table (`concepts`, `modules`, `needs`, `requirements`, `test_cases`, `tasks`, `defects`, `baselines`, `stakeholders`, `test_plans`, `developers`, `agents`). Each predicate maps to an edge table (`child_of`, `derives_from`, `allocates_to`, `depends_on`, `verified_by`, `module`, `stakeholder`, `subject`, `detected_by`, `generates`, `informs`, `implements`, `operator`, `parent_agent`). DuckPGQ creates a SQL/PGQ property graph over these tables, making the same data queryable with both standard SQL (filtering, aggregation, analytics) and SQL/PGQ pattern matching syntax (graph traversals, shortest paths, variable-length paths).

On-disk serialization uses per-table NDJSON files under `.krav/graph/`. Vertex tables serialize to files named after the table (`concepts.ndjson`, `needs.ndjson`, `requirements.ndjson`, `tasks.ndjson`, etc.). Edge tables serialize the same way (`derives_from.ndjson`, `verified_by.ndjson`, etc.). Each file sorts deterministically (vertex tables by `id`, edge tables by all columns) for stable git diffs. The `HydrateFromDir` / `DehydrateToDir` pattern loads NDJSON files into DuckDB tables on server startup and writes them back on checkpoint or shutdown.

The NDJSON format uses plain JSON keys (`id`, `type`, `title`, `status`) rather than JSON-LD keys (`@id`, `@type`, `@context`). Object property values use plain ID strings (`"module": "MOD-A4F8R2X1"`) rather than JSON-LD reference objects (`"module": {"@id": "MOD-A4F8R2X1"}`). Multi-valued references use arrays of ID strings.

The DuckPGQ spike at `experiments/duckpgq/` validates this architecture with sub-50 ms hydration, sub-5 ms dehydration, and sub-3 ms graph pattern queries at small scale.

## Vocabulary

The Krav schema defines a vocabulary in the `krav:` namespace (`https://krav.sh/schema#`) that aligns with external ontologies (Dublin Core, PROV-O, OSLC) as a design reference for semantic interoperability. The runtime data model is relational tables queried via SQL/PGQ, not RDF triples. The RDF vocabulary alignment serves as design-time metadata documenting the conceptual mapping between Krav's property graph and established engineering ontologies. See [Vocabulary alignment](vocabulary-alignment.md) for the complete mapping.

**Namespace**: `https://krav.sh/schema#`
**Prefix**: `krav`

The vocabulary maps compact property names to their canonical IRIs in the Krav, Dublin Core, PROV-O, and OSLC namespaces:

```json
{
  "@context": {
    "@vocab": "https://krav.sh/schema#",
    "krav": "https://krav.sh/schema#",
    "dcterms": "http://purl.org/dc/terms/",
    "prov": "http://www.w3.org/ns/prov#",
    "oslc_rm": "http://open-services.net/ns/rm#",
    "oslc_qm": "http://open-services.net/ns/qm#",
    "oslc_cm": "http://open-services.net/ns/cm#",
    "oslc_config": "http://open-services.net/ns/config#",
    "rdfs": "http://www.w3.org/2000/01/rdf-schema#",

    "Module": "krav:Module",
    "Concept": "krav:Concept",
    "Need": "krav:Need",
    "Requirement": "krav:Requirement",
    "TestCase": "krav:TestCase",
    "Task": "krav:Task",
    "Defect": "krav:Defect",
    "Baseline": "krav:Baseline",
    "Stakeholder": "krav:Stakeholder",
    "TestPlan": "krav:TestPlan",
    "Developer": "krav:Developer",
    "Agent": "krav:Agent",

    "childOf": {"@id": "krav:childOf", "@type": "@id"},
    "derivesFrom": {"@id": "krav:derivesFrom", "@type": "@id"},
    "allocatesTo": {"@id": "krav:allocatesTo", "@type": "@id"},
    "dependsOn": {"@id": "krav:dependsOn", "@type": "@id"},
    "verifiedBy": {"@id": "krav:verifiedBy", "@type": "@id"},
    "subject": {"@id": "krav:subject", "@type": "@id"},
    "detectedBy": {"@id": "krav:detectedBy", "@type": "@id"},
    "generates": {"@id": "krav:generates", "@type": "@id"},
    "informs": {"@id": "krav:informs", "@type": "@id"},
    "module": {"@id": "krav:module", "@type": "@id"},
    "stakeholder": {"@id": "krav:stakeholder", "@type": "@id"},
    "operator": {"@id": "krav:operator", "@type": "@id"},
    "parentAgent": {"@id": "krav:parentAgent", "@type": "@id"},

    "implements": {"@id": "oslc_cm:implementsRequirement", "@type": "@id"},
    "generatedBy": {"@id": "prov:wasGeneratedBy", "@type": "@id"},
    "attributedTo": {"@id": "prov:wasAttributedTo", "@type": "@id"},

    "title": "dcterms:title",
    "description": "dcterms:description",
    "statement": "krav:statement",
    "status": "krav:status",
    "phase": "krav:phase",
    "processPhase": "krav:processPhase",
    "conceptType": "krav:conceptType",
    "priority": "krav:priority",
    "severity": "krav:severity",
    "category": "krav:category",
    "summary": "krav:summary",
    "sessionId": "krav:sessionId",
    "subagentId": "krav:subagentId",
    "created": {"@id": "dcterms:created", "@type": "xsd:dateTime"},
    "updated": {"@id": "dcterms:modified", "@type": "xsd:dateTime"},
    "tags": "krav:tags"
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

The `id` field holds the identifier. The type prefix must be consistent with the `type` value; a node with `"id": "CON-K7M3NP2Q"` must have `"type": "Concept"`.

## Common properties

All node types share these properties:

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `id` | string | Yes | Unique identifier (TYPE-nanoid format) |
| `type` | string | Yes | Node type (must match identifier prefix) |
| `title` | string | Yes | Human-readable title (maps to `dcterms:title`) |
| `description` | string | No | Extended description (maps to `dcterms:description`) |
| `status` | string | Yes | Lifecycle state (type-specific enum) |
| `created` | datetime | No | ISO 8601 creation timestamp (maps to `dcterms:created`) |
| `updated` | datetime | No | ISO 8601 last-modified timestamp (maps to `dcterms:modified`) |
| `tags` | string[] | No | Freeform tags for filtering |
| `summary` | string | No | Inline prose for extended context beyond title and description |
| `generated_by` | string | No | ID of the task that created this node (maps to `prov:wasGeneratedBy`) |
| `attributed_to` | string | No | ID of the agent or developer responsible (maps to `prov:wasAttributedTo`) |

## Prose files

Every node type can have an associated markdown file for extended prose that doesn't fit in structured fields. The system derives the path from the node's identifier rather than storing it on the node.

Prose files live in flat, type-specific directories under `.krav/`:

| Node type   | Directory              |
|-------------|------------------------|
| Concept     | `.krav/concepts/`      |
| Module      | `.krav/modules/`       |
| Stakeholder | `.krav/stakeholders/`  |
| Need        | `.krav/needs/`         |
| Requirement | `.krav/requirements/`  |
| Test case   | `.krav/test-cases/`    |
| Task        | `.krav/tasks/`         |
| Defect      | `.krav/defects/`       |
| Baseline    | `.krav/baselines/`     |
| Developer   | `.krav/developers/`    |
| Agent       | `.krav/agents/`        |

Filenames follow the pattern `{timestamp}-{NANOID}-{slug}.md`. The timestamp is `YYYYMMDDHHMMSS` and provides filesystem sort order. The nanoid is the same 8-character identifier from the node's `@id`. The slug is a human-readable kebab-case label. As an example, concept `CON-K7M3NP2Q` titled "Parser architecture" would have its prose at `.krav/concepts/20260103164500-K7M3NP2Q-parser-architecture.md`.

Resolution is mechanical: the node type determines the directory, and the nanoid matches the second segment of the filename when split by `-`. The timestamp and slug are filesystem conveniences for sorting and readability; the graph stores neither.

Not every node needs a prose file. Concepts almost always have one since exploration is their purpose. Tasks and defects often get by with the `summary` field for a paragraph or two of inline context, and only reach for a file when they need more room. Stakeholders, baselines, and other lightweight types rarely need files but the mechanism is there when they do.

The `summary` field and a prose file serve different purposes and can coexist on the same node. `summary` carries inline prose directly on the graph node, suited for a quick paragraph of context. The prose file is for extended content that would be unwieldy in an NDJSON record.

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

## NDJSON representation

Each node serializes as a single JSON object in the corresponding per-table NDJSON file. The `id` field carries the node's identifier, `type` names the node type, and remaining fields are properties. References to other nodes use plain ID strings rather than JSON-LD reference objects.

A need in `needs.ndjson`:

```json
{"id": "NEED-B7G3M9K2", "type": "Need", "title": "Quick feedback", "status": "validated"}
```

The relationship between this need and its module and source concept lives in edge table files. In `module.ndjson`:

```json
{"src": "NEED-B7G3M9K2", "dst": "MOD-A4F8R2X1"}
```

In `derives_from.ndjson`:

```json
{"src": "NEED-B7G3M9K2", "dst": "CON-K7M3NP2Q"}
```

Key representation conventions:

- Vertex tables contain node properties; edge tables contain relationships
- Edge table rows have `src` and `dst` columns identifying the source and target nodes
- Edge tables with metadata include additional columns: `{"src": "REQ-H4J7N2P5", "dst": "MOD-A4F8R2X1", "budget": "50ms"}`
- Files sort deterministically for stable git diffs

## Extensibility

### Adding a new node type

1. Choose a unique uppercase prefix (2-4 characters)
2. Define a DuckDB vertex table with the node's columns
3. Register the table in the SQL/PGQ property graph definition
4. Add the corresponding NDJSON file to the hydrate/dehydrate cycle
5. Define type-specific properties and document them
6. Define valid predicates (domain/range) in [predicates](predicates.md)
7. Add structural constraints in [constraints](constraints.md)
8. Create an entity-specific doc with lifecycle and field definitions
9. Update the vocabulary alignment if the type maps to an external ontology class

### Adding a new predicate

1. Define a DuckDB edge table with `src` and `dst` columns (plus any metadata columns)
2. Register the edge table in the SQL/PGQ property graph definition
3. Add the corresponding NDJSON file to the hydrate/dehydrate cycle
4. Define domain, range, cardinality, and structural constraints in [predicates](predicates.md)
5. Add validation rules in [constraints](constraints.md)
6. Document suspect propagation behavior if applicable

### Adding a new property

1. Add the column to the relevant DuckDB vertex table
2. Document it in the relevant entity-specific doc
3. Update the type-specific properties table in this document
