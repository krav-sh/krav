# Vocabulary alignment

## Overview

The ARCI vocabulary aligns with established external ontologies as a design reference for semantic interoperability. The runtime data model is a property graph in DuckDB queried via SQL/PGQ, not an RDF triple store. The alignment declarations (`rdfs:subClassOf`, `rdfs:subPropertyOf`) documented here are design-time metadata that record the conceptual mapping between ARCI's node types, predicates, and properties and their counterparts in Dublin Core, PROV-O, and OSLC. They are not runtime-enforced constraints. An RDF-aware tool could reconstruct the triples from ARCI's NDJSON data using these mappings, but ARCI's own code never imports external namespaces at runtime.

The ARCI ontology defines terms in the `arci:` namespace. Rather than inventing terms where established vocabularies already define them, the schema reuses external vocabulary terms directly or declares formal alignment. This keeps ARCI's conceptual model interoperable with standards-aware tooling while preserving ARCI-specific semantics where the structure adds real constraints beyond what external vocabularies provide.

Three external vocabularies are relevant:

**Dublin Core Terms** (`dcterms:`): the standard metadata vocabulary for titles, descriptions, and timestamps. OSLC and most RDF vocabularies use Dublin Core for these properties rather than defining their own.

**W3C PROV-O** (`prov:`): the W3C provenance ontology, with three core classes (Entity, Activity, Agent) and properties for derivation, generation, and attribution. PROV-O maps well onto ARCI's derivation chains, task execution model, and agent provenance tracking. ARCI renamed the Module type from Entity to avoid collision with `prov:Entity`.

**OSLC** (`oslc_rm:`, `oslc_qm:`, `oslc_cm:`, `oslc_config:`): the OASIS Open Services for Lifecycle Collaboration family of specifications defines RDF vocabularies for requirements management, quality management, change management, and configuration management. OSLC's class and property definitions cover much of ARCI's domain, though ARCI adds lifecycle state machines, suspect link propagation, DAG enforcement, and the concept-to-need formal transformation chain that OSLC doesn't model.

## Alignment strategy

The strategy has three tiers based on whether ARCI adds semantics beyond what the external vocabulary defines.

### Use directly

When ARCI's property means exactly the same thing as an external property and adds no constraints, enforcement behavior, or additional semantics, the JSON-LD context maps the compact key directly to the external IRI. ARCI mints no `arci:` term.

This applies to Dublin Core metadata properties (`title`, `description`, `created`, `updated`), PROV-O provenance properties (`generatedBy`, `attributedTo`), and the OSLC CM task-to-requirement link (`implements`).

### Declare subclass or subproperty alignment

When ARCI defines a class or property that corresponds to an external term but adds real constraints (lifecycle states, suspect propagation, DAG enforcement, cardinality rules, domain/range restrictions), ARCI mints its own term in the `arci:` namespace and declares the formal alignment in the schema. Instance data uses the `arci:` term. The alignment declarations live in the T-Box (schema definition) and don't appear in instance data. An RDF-aware tool can discover the alignment through inference; ARCI's own code never needs to import external namespaces at runtime.

### Keep in ARCI namespace with no external alignment

When ARCI defines something that has no counterpart in any established vocabulary, the term lives in the `arci:` namespace with no alignment declaration. This covers ARCI's unique domain modeling: the Concept and Need types, the formal transformation chain, suspect propagation predicates, and structural enforcement semantics.

## Properties used directly

These properties use external IRIs directly in the JSON-LD context. ARCI mints no `arci:` term for them.

### Dublin Core metadata

| Compact key | External IRI | Notes |
|-------------|-------------|-------|
| `title` | `dcterms:title` | Human-readable title. Identical semantics. |
| `description` | `dcterms:description` | Extended description. Identical semantics. |
| `created` | `dcterms:created` | ISO 8601 creation timestamp. Identical semantics. |
| `updated` | `dcterms:modified` | ISO 8601 last-modified timestamp. Note the key rename: ARCI uses `updated`, Dublin Core uses `modified`. The JSON-LD context handles the mapping. |

### PROV-O provenance

| Compact key | External IRI | Notes |
|-------------|-------------|-------|
| `generatedBy` | `prov:wasGeneratedBy` | Links any node to the TASK-* that created it. Optional on all node types. |
| `attributedTo` | `prov:wasAttributedTo` | Links any node to the AGT-* or DEV-* responsible. Optional on all node types. |

ARCI adds no constraints on these properties; no suspect propagation, no enforcement behavior. They enable queries like "which task created this requirement?" and "which agent session wrote this defect?" without relying on git history reconstruction.

### OSLC CM task-to-requirement link

| Compact key | External IRI | Notes |
|-------------|-------------|-------|
| `implements` | `oslc_cm:implementsRequirement` | Links a task to the requirements it exists to satisfy. Domain/range types align through inheritance: `arci:Task rdfs:subClassOf oslc_cm:Task` (subclass of `oslc_cm:ChangeRequest`, the defined domain) and `arci:Requirement rdfs:subClassOf oslc_rm:Requirement` (the defined range). |

ARCI adds no constraints on this property; no suspect propagation (`verifiedBy` suspect propagation catches requirement changes, not task links), no DAG enforcement, no special cardinality beyond 0..*. The decomposition transformation sets `implements` when creating tasks from requirements. The task template context's `requirements` array resolves from these edges.

### Properties remaining in ARCI namespace

Properties that have no external equivalent: `statement`, `status`, `phase`, `processPhase`, `conceptType`, `priority`, `severity`, `category`, `summary`, `tags`, `sessionId`, `subagentId`, and all other type-specific properties.

## Class alignments

Each ARCI class that corresponds to an external class declares `rdfs:subClassOf` in the schema definition.

| ARCI class | Alignment | External class | Rationale |
|-----------|-----------|---------------|-----------|
| `arci:Requirement` | `rdfs:subClassOf` | `oslc_rm:Requirement` | ARCI adds lifecycle states (draft → approved → satisfied → superseded), suspect propagation on `derivesFrom` and `verifiedBy`, and the need-to-requirement derivation chain that OSLC doesn't model. OSLC RM conflates needs and requirements into a single Requirement class; ARCI distinguishes them per INCOSE NRM. |
| `arci:TestCase` | `rdfs:subClassOf` | `oslc_qm:TestCase` | ARCI adds lifecycle states, verification method taxonomy, and the specification/execution decoupling. |
| `arci:TestPlan` | `rdfs:subClassOf` | `oslc_qm:TestPlan` | Both represent coordinated collections of test activity scoped to a milestone or phase gate. Full entity specification pending. |
| `arci:Defect` | `rdfs:subClassOf` | `oslc_cm:Defect` | ARCI adds defect category taxonomy (missing, incorrect, ambiguous, etc.), detection phase tracking, and the `subject`/`detectedBy`/`generates` predicate set. |
| `arci:Task` | `rdfs:subClassOf` | `oslc_cm:Task` | ARCI adds process phase tagging, DAG dependency enforcement, deliverable tracking, and lifecycle coordination with module phase gates. |
| `arci:Task` | `rdfs:subClassOf` | `prov:Activity` | Tasks are activities that produce deliverables and can generate/modify graph nodes. Dual superclass with `oslc_cm:Task`. |
| `arci:Module` | `rdfs:subClassOf` | `oslc_config:Component` | ARCI adds hierarchical decomposition via `childOf`, phase tracking, and module-scoped ownership of needs, requirements, tasks, and defects. |
| `arci:Baseline` | `rdfs:subClassOf` | `oslc_config:Baseline` | ARCI adds git commit anchoring, semantic diff, module subtree scoping, and phase gate integration. |
| `arci:Stakeholder` | `rdfs:subClassOf` | `prov:Agent` | Stakeholders are agents with concerns about the system. ARCI models them as abstract roles rather than concrete people. |
| `arci:Developer` | `rdfs:subClassOf` | `prov:Agent` | Human actors who initiate sessions, approve baselines, and make design decisions. Persistent identity across sessions. |
| `arci:Agent` | `rdfs:subClassOf` | `prov:Agent` | Claude Code sessions and subagents. Ephemeral; each invocation creates a new node. Distinguished from Developer by `sessionId` (required) and `subagentId` (nullable). |

Classes with no external counterpart:

| ARCI class | Why no alignment |
|-----------|-----------------|
| `arci:Concept` | OSLC starts at requirements. ARCI provides no standard vocabulary for pre-requirement design exploration. |
| `arci:Need` | OSLC RM conflates needs and requirements. INCOSE's need/requirement distinction is ARCI-specific in the RDF world. |

## Object property alignments

Each ARCI predicate that corresponds to an external property declares `rdfs:subPropertyOf` in the schema definition.

| ARCI property | Alignment | External property | Rationale |
|--------------|-----------|------------------|-----------|
| `arci:derivesFrom` | `rdfs:subPropertyOf` | `prov:wasDerivedFrom` | Clean semantic match across all domain pairs (CON→NED, NED→REQ, REQ→REQ). ARCI adds DAG enforcement, suspect propagation, and domain/range constraints that PROV-O doesn't have. The single ARCI predicate maps to multiple OSLC RM predicates depending on context (`oslc_rm:elaboratedBy` for CON→NED, `oslc_rm:satisfiedBy` for NED→REQ, `oslc_rm:decomposedBy` for REQ→REQ), but `prov:wasDerivedFrom` captures the common semantic cleanly. |
| `arci:verifiedBy` | `rdfs:subPropertyOf` | `oslc_rm:validatedBy` | OSLC RM uses "validated" where INCOSE would say "verified" (checking the solution meets the requirement). The semantic intent is the same despite the terminology difference. ARCI adds suspect propagation behavior. |
| `arci:childOf` | `rdfs:subPropertyOf` | `dcterms:isPartOf` | Module hierarchy is a part-of relationship. ARCI adds single-parent tree enforcement and phase propagation constraints. OSLC Config lacks a hierarchical decomposition predicate. |

Predicates with no external counterpart:

| ARCI property | Why no alignment |
|--------------|-----------------|
| `arci:allocatesTo` | Requirement flow-down to child modules with optional metadata. No OSLC equivalent for this direction. |
| `arci:dependsOn` | Task ordering in a DAG. Generic enough that external alignment would be meaningless. |
| `arci:module` | Ownership assignment to architectural containers. Organizational, not semantic. |
| `arci:stakeholder` | Need-to-stakeholder attribution. OSLC doesn't model stakeholders as first-class resources. |
| `arci:subject` | Defect-to-any-node "this is the problem" link. ARCI-specific defect modeling. |
| `arci:detectedBy` | Defect-to-task "found by" link. ARCI-specific defect modeling. |
| `arci:generates` | Defect-to-task remediation link. ARCI-specific defect modeling. |
| `arci:informs` | Concept-to-module informal navigation link. No formal traceability obligation. |
| `arci:operator` | Links an Agent to the Developer who initiated the session. ARCI-specific provenance. |
| `arci:parentAgent` | Links a subagent to its spawning session agent. ARCI-specific agent hierarchy. |

## Closed gaps

The initial alignment analysis surfaced gaps where established vocabularies model something that ARCI did not. The team has resolved all of them.

**OSLC QM TestScript** (GAP-01): closed as deliberate simplification. TC-* nodes already have an `coding` field for the path to test code, and the design handles specification/coding decoupling by making the coding a task deliverable. A separate TestScript entity would duplicate existing information.

**OSLC QM TestResult/TestExecutionRecord** (GAP-02): closed, deferred to test plan spec. `currentResult` on TC-* nodes tracks the latest outcome for everyday development. Historical run data lives in git and task deliverables. The TestPlan entity (TP-*) addresses formal execution records when fully specified.

**No shared superclass for defects and tasks** (GAP-03): closed as not a gap. Defects and tasks have different lifecycles, fields, and blocking criteria. No competency question requires cross-type queries. The `generates` predicate bridges them when needed.

**Task-to-requirement linkage** (GAP-04): resolved by using `oslc_cm:implementsRequirement` directly. See the "OSLC CM task-to-requirement link" section preceding this one.

**No Agent entity** (GAP-05): resolved by adding Developer (DEV-*) and Agent (AGT-*) node types, both `rdfs:subClassOf prov:Agent`. See the preceding class alignments table.

**No provenance properties on nodes** (GAP-06): resolved by using `prov:wasGeneratedBy` and `prov:wasAttributedTo` directly. See the preceding "PROV-O provenance" section.

**ARCI-namespaced Dublin Core duplicates** (GAP-07): resolved by mapping `title`, `description`, `created`, and `updated` to their `dcterms:` IRIs in the JSON-LD context.

**Test plan entity not specified** (GAP-08): resolved by renaming TestCampaign (TCAM-*) to TestPlan (TP-*) with `rdfs:subClassOf oslc_qm:TestPlan`. Full entity specification still pending. The spec should address lifecycle states, fields, relationships, CLI commands, phase gate interaction, and test execution record modeling.

## JSON-LD context

The full context with external namespace imports, direct-use property mappings, and all class and predicate definitions. See [schema](schema.md) for the authoritative version.

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
    "TestPlan": "arci:TestPlan",
    "Task": "arci:Task",
    "Defect": "arci:Defect",
    "Baseline": "arci:Baseline",
    "Stakeholder": "arci:Stakeholder",
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

The class alignment declarations (`rdfs:subClassOf`) and property alignment declarations (`rdfs:subPropertyOf`) are T-Box statements in the schema definition, not part of the instance-level JSON-LD context. A separate schema ontology file or the schema validation layer would express them.
