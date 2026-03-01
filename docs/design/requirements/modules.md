# Modules module type

## Overview

Modules (MOD-*) are architectural containers representing the things the team builds. A module could be a system, subsystem, component, module, or any identifiable element in the architecture. Modules form a hierarchy via parent-child relationships and serve as the organizing principle for needs, requirements, and work.

Unlike document-centric models where "specs" own requirements, ARCI organizes around the architectural elements themselves. A module owns its needs and requirements; the module is what the team builds, constrains, and verifies.

## Purpose

Modules serve multiple roles:

**Architectural decomposition**: the module hierarchy represents how the system breaks down into subsystems, components, and modules. This decomposition is the primary structuring mechanism for the project.

**Ownership**: needs and requirements belong to modules. "The parser shall tokenize input in under 10 ms" is a requirement owned by MOD-parser, not floating in a document.

**Phase tracking**: each module tracks its current lifecycle phase (architecture, design, coding, and so on). Phase-gated execution constrains work to the appropriate phase.

**Work scoping**: tasks relate to modules. "What's the plan for the parser?" becomes a query over tasks where module is MOD-parser or its descendants.

**Deliverable organization**: modules organize task outputs (architecture docs, API specs, code) by module in the filesystem.

## Hierarchy

Modules form a tree via childOf relationships:

```text
MOD-OAPSROOT (the project itself)
├── MOD-A4F8R2X1 (parser subsystem)
│   ├── MOD-L3X3R001 (lexer component)
│   └── MOD-T0K3N002 (tokenizer component)
├── MOD-B9G3M7K2 (CLI subsystem)
│   ├── MOD-C0MM4ND1 (command parser)
│   └── MOD-0UTPUT01 (output formatter)
└── MOD-K8G4R5X2 (knowledge graph subsystem)
```

The root module represents the project as a whole and owns project-wide needs and requirements.

### Hierarchy rules

- Every module except root has exactly one parent (single childOf relationship)
- Root module has no parent
- The hierarchy forbids cycles
- The system supports reparenting (with review of derived items)

## Lifecycle phase

Each module tracks its current phase:

```text
architecture → design → implementation → integration → verification → validation
```

| Phase          | Description                                    |
|----------------|------------------------------------------------|
| architecture   | Identifying components, boundaries, interfaces |
| design         | Defining APIs, data models, algorithms         |
| `implementation` | Writing code, building the thing               |
| integration    | Assembling components, resolving interfaces    |
| verification   | Testing against requirements                   |
| validation     | Confirming the product meets stakeholder needs |

### Phase constraints

**Hierarchical constraint**: child modules can be at or behind their parent's phase, never ahead.

```text
MOD-OAPSROOT.phase = implementation
  MOD-parser.phase = implementation  ✓ (at parent)
  MOD-cli.phase = design             ✓ (behind parent)
  MOD-cli.phase = verification       ✗ (ahead of parent, blocked)
```

**Task constraint**: the system only allows creating or executing tasks for the module's current phase or earlier phases.

### Phase advancement

```bash
arci moduleadvance MOD-A4F8R2X1 --to design
```

Advancement criteria:

- All tasks for the current phase are complete
- Verification tasks for the current phase have no blocking findings
- Parent module is at or ahead of the target phase

### Phase regression

```bash
arci moduleregress MOD-A4F8R2X1 --to architecture --reason "boundary unclear"
```

When a parent regresses:

- Children remain at their current phase
- The constraint blocks children from advancing past the new parent phase
- ARCI automatically creates a finding (FND-*) with the reason

## Storage model

ARCI stores module metadata in `graph.jsonlt` as JSON-LD compact form. Prose files have no frontmatter; `graph.jsonlt` is the single source of truth for all structured data.

```json
{"@context": "context.jsonld", "@id": "MOD-A4F8R2X1", "@type": "Module", "title": "Parser", "description": "Parses input into AST", "childOf": {"@id": "MOD-OAPSROOT"}, "phase": "implementation", "status": "active"}
```

Fields:

- `@id`: Unique identifier (MOD-XXXXXXXX format)
- `@type`: Always "Module"
- `title`: Human-readable title
- `description`: Brief description (optional)
- `childOf`: Parent module reference (null/absent for root)
- `phase`: Current lifecycle phase
- `status`: active, deprecated, archived
- `created`, `updated`: ISO 8601 timestamps
- `tags`: Array of strings (optional)

## Stakeholder classes

Modules at different levels serve different stakeholder classes:

| Level     | Stakeholders                                        | Need examples                             |
|-----------|-----------------------------------------------------|-------------------------------------------|
| Root      | OSS community, contributors, maintainers, ecosystem | Stability, compatibility, discoverability |
| Subsystem | Domain users, integrators                           | Functionality, performance, extensibility |
| Component | Developers, adjacent components                     | API clarity, error handling, testability  |

## Relationships

The module's JSON-LD record embeds relationships using `{"@id": "..."}` values.

### Outgoing relationships

| Property   | Target | Cardinality | Description                                                  |
|------------|--------|-------------|--------------------------------------------------------------|
| childOf    | MOD-*  | Single      | This module's parent                                         |
| integrates | MOD-*  | Multi       | Peer modules this one integrates (for integration modules) |

### Incoming relationships (queried via graph)

| Property     | Source | Description                                            |
|--------------|--------|--------------------------------------------------------|
| childOf      | MOD-*  | Child modules                                         |
| allocatesTo  | REQ-*  | Requirements that allocate to this module               |
| module       | NED-*  | Needs owned by this module                             |
| module       | REQ-*  | Requirements owned by this module                      |
| module       | TSK-*  | Tasks for this module                                  |
| informs      | CON-*  | Concepts that inform this module                       |

Example with relationships:

```json
{"@context": "context.jsonld", "@id": "MOD-0BS3RV01", "@type": "Module", "title": "Observability", "childOf": {"@id": "MOD-OAPSROOT"}, "integrates": [{"@id": "MOD-A4F8R2X1"}, {"@id": "MOD-B9G3M7K2"}], "phase": "architecture", "status": "active"}
```

## Deliverable layout

ARCI organizes task deliverables by module:

```text
.arci/
  modules/
    MOD-A4F8R2X1/
      architecture.md       # From architecture tasks
      api-design.md         # From design tasks
      interface-spec.md     # From design tasks
    MOD-B9G3M7K2/
      user-guide.md         # From documentation tasks
```

## Special modules

### Root module

Every project has a root module representing the project as a whole:

```json
{"@context": "context.jsonld", "@id": "MOD-OAPSROOT", "@type": "Module", "title": "arci", "description": "Agentic Requirements Composition & Integration", "phase": "implementation", "status": "active"}
```

Root-level needs capture project-wide stakeholder expectations. Root-level requirements flow down to child modules.

### Integration modules

Some modules represent integrations between siblings rather than components:

```json
{"@context": "context.jsonld", "@id": "MOD-0BS3RV01", "@type": "Module", "title": "Observability", "childOf": {"@id": "MOD-OAPSROOT"}, "integrates": [{"@id": "MOD-A4F8R2X1"}, {"@id": "MOD-B9G3M7K2"}], "phase": "design", "status": "active"}
```

Sibling phases don't constrain integration modules, but their integration tasks may depend on sibling verification.

## Implementation architecture

Module capability follows the three-layer architecture (see CON-GR4PH4RC).

### Typed node

ModuleNode is an independent dataclass with typed fields. Graph stores ModuleNode directly, and IO serializes directly to/from ModuleNode (see CON-GR4PH4RC for the full architecture).

```python
@dataclass(frozen=True, slots=True)
class ModuleNode:
    """Module—architectural container."""
    id: str
    title: str
    status: ModuleStatus           # Typed enum
    phase: ModulePhase             # Typed enum
    description: str = ""
    created: datetime | None = None
    updated: datetime | None = None
```

All fields use proper types. The IO layer creates ModuleNode directly from JSON-LD records, preserving all type-specific fields.

### Core layer (`arci.core.module`)

Pure functions and typed data structures:

```python
# Types
class ModulePhase(StrEnum):
    ARCHITECTURE = "architecture"
    DESIGN = "design"
    IMPLEMENTATION = "implementation"
    INTEGRATION = "integration"
    VERIFICATION = "verification"
    VALIDATION = "validation"

class ModuleStatus(StrEnum):
    ACTIVE = "active"
    DEPRECATED = "deprecated"
    ARCHIVED = "archived"

# Typed node
@dataclass(frozen=True, slots=True)
class ModuleNode(Node):
    phase: ModulePhase = ModulePhase.ARCHITECTURE

# Operations (pure functions)
def from_node(node: Node) -> ModuleNode: ...
def with_phase(module: ModuleNode, phase: ModulePhase) -> ModuleNode: ...
def can_advance(graph: Graph, module_id: str, target: ModulePhase) -> tuple[bool, list[str]]: ...
def can_regress(module: ModuleNode, target: ModulePhase) -> bool: ...

# Queries (pure, take Graph)
def get(graph: Graph, module_id: str) -> ModuleNode | None: ...
def list_all(graph: Graph) -> tuple[ModuleNode, ...]: ...
def parent(graph: Graph, module_id: str) -> ModuleNode | None: ...
def children(graph: Graph, module_id: str) -> tuple[ModuleNode, ...]: ...
def ancestors(graph: Graph, module_id: str) -> tuple[ModuleNode, ...]: ...
def descendants(graph: Graph, module_id: str) -> tuple[ModuleNode, ...]: ...
def owned_needs(graph: Graph, module_id: str) -> frozenset[str]: ...
def owned_requirements(graph: Graph, module_id: str) -> frozenset[str]: ...
def owned_tasks(graph: Graph, module_id: str) -> frozenset[str]: ...
```

### Service layer (`arci.service.module`)

Orchestrates core and IO:

```python
# Thin query wrappers
def get(store: GraphStore, module_id: str) -> ModuleNode | None: ...
def list_all(store: GraphStore) -> tuple[ModuleNode, ...]: ...
def children(store: GraphStore, module_id: str) -> tuple[ModuleNode, ...]: ...

# Mutations
def create(store: GraphStore, title: str, parent_id: str | None = None, ...) -> ModuleNode: ...
def update(store: GraphStore, module_id: str, **fields) -> ModuleNode: ...
def delete(store: GraphStore, module_id: str) -> None: ...  # Must have no children
def reparent(store: GraphStore, module_id: str, new_parent_id: str) -> ModuleNode: ...

# Phase management
def advance(store: GraphStore, module_id: str, target: ModulePhase) -> ModuleNode: ...
def regress(store: GraphStore, module_id: str, target: ModulePhase, reason: str) -> tuple[ModuleNode, FindingNode]: ...
```

## CLI commands

```bash
# CRUD
arci modulecreate --title "Parser" --parent MOD-OAPSROOT
arci moduleshow MOD-A4F8R2X1
arci modulelist
arci modulelist --parent MOD-OAPSROOT --phase implementation
arci moduleupdate MOD-A4F8R2X1 --title "Parser v2"
arci moduledelete MOD-A4F8R2X1  # Must have no children

# Hierarchy
arci modulechildren MOD-OAPSROOT
arci moduletree MOD-OAPSROOT
arci modulereparent MOD-A4F8R2X1 --to MOD-B9G3M7K2

# Phase management
arci modulephase MOD-A4F8R2X1
arci moduleadvance MOD-A4F8R2X1 --to design
arci moduleregress MOD-A4F8R2X1 --to architecture --reason "..."

# Work scoping
arci moduledecompose MOD-A4F8R2X1 --template full-feature
arci moduletasks MOD-A4F8R2X1
arci moduletasks MOD-A4F8R2X1 --include-descendants

# Context
arci contextmodule MOD-A4F8R2X1
```

## Examples

### Root module

```json
{"@context": "context.jsonld", "@id": "MOD-OAPSROOT", "@type": "Module", "title": "arci", "description": "Agentic Requirements Composition & Integration", "phase": "implementation", "status": "active"}
```

### Subsystem module

```json
{"@context": "context.jsonld", "@id": "MOD-A4F8R2X1", "@type": "Module", "title": "Parser", "description": "Parses arci commands and configuration", "childOf": {"@id": "MOD-OAPSROOT"}, "phase": "design", "status": "active", "tags": ["core"]}
```

### Component module

```json
{"@context": "context.jsonld", "@id": "MOD-L3X3R001", "@type": "Module", "title": "Lexer", "description": "Tokenizes input stream", "childOf": {"@id": "MOD-A4F8R2X1"}, "phase": "architecture", "status": "active"}
```

## Implementation status

| Layer | Status | Notes |
|-------|--------|-------|
| Core | Implemented | Typed node, operations, queries in `arci.core.module` |
| IO | Implemented | JSON-LD serialization via `arci.io.graph` |
| Service | Implemented | CRUD, hierarchy, and phase transitions in `arci.service.module` |
| CLI | Implemented | Commands in `arci.cli._commands._module` with output formatting in `arci.cli._output` |

### CLI details

The CLI includes:

- **Output formatting** (`arci.cli._output`): Protocol-based formatters (TableFormatter, JSONFormatter, AgentFormatter) with NodeDisplay adapter pattern
- **Graph context** (`arci.cli._commands._graph`): GraphContext provider for store access
- **Module commands** (`arci.cli._commands._module`): Full CRUD, hierarchy, and phase management

The CLI registers stub commands (decompose, tasks, context) but does not yet provide them.

## Summary

Modules are architectural containers that:

- Form a hierarchy representing system decomposition
- Own needs and requirements
- Track lifecycle phase with hierarchical constraints
- Scope tasks and deliverables
- Serve as the primary organizing principle for the project
- Store metadata in graph.jsonlt with no frontmatter in prose files
- Implemented following three-layer architecture (core/io/service)

The module hierarchy replaces document-centric organization: the thing the team builds is the organizing principle, not documents describing it.
