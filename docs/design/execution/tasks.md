# Tasks module type

## Overview

Tasks (TSK-*) are atomic units of work with verifiable deliverables. Unlike hierarchical plan structures, tasks exist in a flat DAG where edges express dependencies. The "plan" for any scope is simply the subgraph of tasks reachable from a target.

Tasks have a process phase attribute indicating where they fit in the ISO/IEC/IEEE 15288 lifecycle, enabling phase-gated execution.

## Purpose

Tasks serve several roles:

**Work decomposition**: Complex work breaks down into manageable, atomic tasks that Claude Code can execute in focused sessions.

**Dependency management**: dependsOn relationships express ordering. Tasks can't start until dependencies complete.

**Phase organization**: The phase attribute organizes work by lifecycle stage and enables phase gates.

**Progress tracking**: Task status shows what's done, what's in progress, what's blocked.

**Deliverable tracking**: Each task's `deliverables` array records what it produced.

## Process phases

Tasks belong to one of six phases from ISO/IEC/IEEE 15288:

| Phase          | Purpose                                     | Example task types                                 |
|----------------|---------------------------------------------|----------------------------------------------------|
| architecture   | Identify components, boundaries, interfaces | module-decomposition, interface-identification     |
| design         | Define APIs, data models, algorithms        | api-design, data-model, algorithm-design           |
| implementation | Build the thing                             | feature-implementation, refactoring, documentation |
| integration    | Assemble components, resolve interfaces     | component-integration, deployment-integration      |
| verification   | Test against requirements                   | test-implementation, code-review, security-audit   |
| validation     | Confirm stakeholder needs are met           | user-acceptance, stakeholder-demo                  |

See CON-K7M3NP2Q for detailed task types per phase.

## Lifecycle

Tasks progress through states:

```
pending → ready → in_progress → blocked → complete → cancelled
```

| State       | Description                               |
|-------------|-------------------------------------------|
| pending     | Not yet ready (dependencies incomplete)   |
| ready       | All dependencies complete, can be started |
| in_progress | Currently being worked on                 |
| blocked     | Started but waiting on something          |
| complete    | Finished with deliverables                |
| cancelled   | Will not be done (with reason)            |

State transitions:

- `pending → ready`: All dependsOn tasks complete
- `ready → in_progress`: Work begins (session started)
- `in_progress → blocked`: Waiting on external factor
- `in_progress → complete`: Deliverables produced
- `* → cancelled`: Task cancelled with reason

## DAG structure

Tasks form a directed acyclic graph via dependsOn relationships:

```
TSK-arch-parser
    ↓
TSK-design-parser-api
    ↓
TSK-impl-lexer ─────────────────┐
    ↓                           │
TSK-impl-tokenizer              │
    ↓                           ↓
TSK-test-parser ←───────── TSK-review-parser
    ↓                           │
    └───────────┬───────────────┘
                ↓
         TSK-release-1.0
```

### DAG queries

```bash
arci taskancestors TSK-R7V3W9Y1      # What does this depend on?
arci taskdescendants TSK-G5M2R8X4    # What depends on this?
arci taskblocking TSK-R7V3W9Y1       # Incomplete ancestors
arci taskready                        # Tasks with no incomplete dependencies
arci taskcritical-path TSK-R7V3W9Y1  # Longest path to target
```

### "Plans" as queries

There are no plan containers. Work organization emerges from queries:

```bash
# "What's the plan for the parser?"
arci tasklist --module MOD-A4F8R2X1 --include-descendants

# "What's in release 1.0?"
arci taskancestors TSK-R7V3W9Y1

# "What's blocking release?"
arci taskblocking TSK-R7V3W9Y1

# "What's ready to work on?"
arci taskready
```

## Storage model

Task metadata is stored in `graph.jsonlt` as JSON-LD compact form. There is no frontmatter in prose files—`graph.jsonlt` is the single source of truth for all structured data.

```json
{"@context": "context.jsonld", "@id": "TSK-E3K8S6V2", "@type": "Task", "title": "Implement lexer tokenization", "module": {"@id": "MOD-A4F8R2X1"}, "processPhase": "implementation", "taskType": "feature-implementation", "status": "complete", "priority": "high", "content": "tasks/20260103165000-E3K8S6V2-implement-lexer.md", "dependsOn": [{"@id": "TSK-G5M2R8X4"}]}
```

Fields:

- `@id`: Unique identifier (TSK-XXXXXXXX format)
- `@type`: Always "Task"
- `title`: Human-readable title
- `module`: Module this task is for (required)
- `description`: Brief description (optional)
- `processPhase`: Process phase (architecture, design, implementation, integration, verification, validation)
- `taskType`: Type within phase (e.g., feature-implementation, code-review)
- `status`: Lifecycle state (pending, ready, in_progress, blocked, complete, cancelled)
- `priority`: high, medium, low
- `assignee`: Who's working on this (optional)
- `content`: Path to prose file relative to `.arci/` (optional)
- `started`, `completed`: ISO 8601 timestamps (optional)
- `created`, `updated`: ISO 8601 timestamps
- `tags`: Array of strings (optional)
- `deliverables`: Array of deliverable objects (optional)

## Deliverables

Tasks track outputs via a `deliverables` array with `kind` discriminator:

| Kind         | Fields                               | Example                            |
|--------------|--------------------------------------|------------------------------------|
| document     | path, type                           | API spec, architecture doc         |
| diagram      | path, type                           | Sequence diagram, module hierarchy |
| commit       | sha, message                         | Git commit                         |
| file         | path, action                         | Source file created/modified       |
| test-results | passed, failed, skipped, report_path | Test execution                     |
| findings     | ids                                  | Findings extracted from review     |
| artifact     | type, name, version                  | npm package, docker image          |
| external     | type, url                            | Pull request, issue                |

Example with deliverables:

```json
{"@context": "context.jsonld", "@id": "TSK-E3K8S6V2", "@type": "Task", "title": "Implement lexer", "module": {"@id": "MOD-A4F8R2X1"}, "processPhase": "implementation", "status": "complete", "deliverables": [{"kind": "commit", "sha": "a1b2c3d4e5f6", "message": "Implement lexer tokenization"}, {"kind": "file", "path": "src/parser/lexer.ts", "action": "created"}]}
```

## Prose content

Complex tasks have prose files with context, instructions, and progress:

```markdown
# Implement lexer tokenization

## Context

The lexer is the first stage of the parser pipeline. It transforms
raw input into a stream of tokens.

## Requirements to satisfy

- REQ-C2H6N4P8: Error reporting within 50ms
- REQ-T0K3N001: All token types recognized

## Approach

Use a state machine approach with...

## Progress

- [x] Define token types
- [x] Implement state machine
- [ ] Add error recovery
- [ ] Performance optimization

## Notes

Discovered edge case with Unicode identifiers...
```

## Relationships

Relationships are embedded in the task's JSON-LD record using `{"@id": "..."}` values.

### Outgoing relationships

| Property    | Target | Cardinality | Description                     |
|-------------|--------|-------------|---------------------------------|
| module      | MOD-*  | Single      | Module this task is for         |
| dependsOn   | TSK-*  | Multi       | Tasks this depends on           |
| addressedBy | FND-*  | Multi       | Findings resolved by this task  |

### Incoming relationships (queried via graph)

| Property  | Source | Description                     |
|-----------|--------|---------------------------------|
| dependsOn | TSK-*  | Tasks that depend on this       |
| generates | FND-*  | Findings that created this task |

Example with relationships:

```json
{"@context": "context.jsonld", "@id": "TSK-E3K8S6V2", "@type": "Task", "title": "Implement lexer", "module": {"@id": "MOD-A4F8R2X1"}, "processPhase": "implementation", "status": "ready", "dependsOn": [{"@id": "TSK-G5M2R8X4"}, {"@id": "TSK-D3S1GN01"}]}
```

## Templates

Tasks can be created from templates:

```bash
arci taskcreate --template quick-review --module MOD-A4F8R2X1
arci taskcreate --template feature-implementation --module MOD-A4F8R2X1 \
  --param priority=high
```

See CON-T3MPL8TZ for the templating system.

## Execution

Tasks execute in atomic Claude Code sessions:

```bash
arci contexttask TSK-E3K8S6V2    # Load task context for session
```

The context includes:

- Task details and instructions
- Module information
- Related requirements
- Dependency status
- Previous session notes (if resuming)

### Session state

During execution, session state tracks:

- Active task
- Phase constraints
- Progress checkpoints
- Pending deliverables

If a session ends early, the next session can resume from checkpoints.

## Implementation architecture

Task functionality follows the three-layer architecture (see CON-GR4PH4RC).

### Typed node

TaskNode is an independent dataclass with typed fields. Graph stores TaskNode directly, and IO serializes directly to/from TaskNode (see CON-GR4PH4RC for the full architecture).

```python
@dataclass(frozen=True, slots=True)
class TaskNode:
    """Task module—atomic work unit."""
    id: str
    title: str
    status: TaskStatus             # Typed enum
    process_phase: ProcessPhase    # Typed enum
    priority: TaskPriority         # Typed enum
    task_type: str = ""            # Type-specific field
    content: str = ""              # Path to prose file
    assignee: str = ""
    started: datetime | None = None
    completed: datetime | None = None
    deliverables: tuple[dict, ...] = ()  # Type-specific field
    description: str = ""
    created: datetime | None = None
    updated: datetime | None = None
```

All fields use proper types. The IO layer creates TaskNode directly from JSON-LD records, preserving all type-specific fields like `process_phase`, `task_type`, and `deliverables`.

### Core layer (arci.core.task)

Pure functions and typed data structures:

```python
# Types
class TaskStatus(StrEnum):
    PENDING = "pending"
    READY = "ready"
    IN_PROGRESS = "in_progress"
    BLOCKED = "blocked"
    COMPLETE = "complete"
    CANCELLED = "cancelled"

class ProcessPhase(StrEnum):
    ARCHITECTURE = "architecture"
    DESIGN = "design"
    IMPLEMENTATION = "implementation"
    INTEGRATION = "integration"
    VERIFICATION = "verification"
    VALIDATION = "validation"

class TaskPriority(StrEnum):
    HIGH = "high"
    MEDIUM = "medium"
    LOW = "low"

# Typed node
@dataclass(frozen=True, slots=True)
class TaskNode(Node):
    process_phase: ProcessPhase = ProcessPhase.ARCHITECTURE
    task_type: str = ""
    priority: TaskPriority = TaskPriority.MEDIUM
    content: str = ""
    assignee: str = ""
    started: datetime | None = None
    completed: datetime | None = None
    deliverables: tuple[dict, ...] = ()

# Operations (pure functions)
def from_node(node: Node) -> TaskNode: ...
def with_status(task: TaskNode, status: TaskStatus) -> TaskNode: ...
def can_transition(task: TaskNode, target: TaskStatus) -> bool: ...
def is_ready(graph: Graph, task_id: str) -> bool: ...

# Queries (pure, take Graph)
def get(graph: Graph, task_id: str) -> TaskNode | None: ...
def list_all(graph: Graph) -> tuple[TaskNode, ...]: ...
def list_by_module(graph: Graph, module_id: str) -> tuple[TaskNode, ...]: ...
def list_by_status(graph: Graph, status: TaskStatus) -> tuple[TaskNode, ...]: ...
def list_by_phase(graph: Graph, phase: ProcessPhase) -> tuple[TaskNode, ...]: ...
def list_ready(graph: Graph) -> tuple[TaskNode, ...]: ...
def depends_on(graph: Graph, task_id: str) -> frozenset[str]: ...
def dependents(graph: Graph, task_id: str) -> frozenset[str]: ...
def ancestors(graph: Graph, task_id: str) -> frozenset[str]: ...
def descendants(graph: Graph, task_id: str) -> frozenset[str]: ...
def blocking(graph: Graph, task_id: str) -> frozenset[str]: ...
def critical_path(graph: Graph, task_id: str) -> tuple[str, ...]: ...
```

### Service layer (arci.service.task)

Orchestrates core and IO:

```python
# Thin query wrappers
def get(store: GraphStore, task_id: str) -> TaskNode | None: ...
def list_all(store: GraphStore) -> tuple[TaskNode, ...]: ...
def list_ready(store: GraphStore) -> tuple[TaskNode, ...]: ...
def blocking(store: GraphStore, task_id: str) -> tuple[TaskNode, ...]: ...

# Mutations
def create(
    store: GraphStore,
    module_id: str,
    title: str,
    process_phase: ProcessPhase,
    task_type: str = "",
    depends_on: list[str] | None = None,
    ...
) -> TaskNode: ...
def update(store: GraphStore, task_id: str, **fields) -> TaskNode: ...
def delete(store: GraphStore, task_id: str) -> None: ...

# Lifecycle
def start(store: GraphStore, task_id: str) -> TaskNode: ...
def complete(store: GraphStore, task_id: str, deliverables: list[dict] | None = None) -> TaskNode: ...
def block(store: GraphStore, task_id: str, reason: str) -> TaskNode: ...
def cancel(store: GraphStore, task_id: str, reason: str) -> TaskNode: ...

# Dependencies
def add_dependency(store: GraphStore, task_id: str, depends_on: str) -> TaskNode: ...
def remove_dependency(store: GraphStore, task_id: str, depends_on: str) -> TaskNode: ...

# Deliverables
def add_deliverable(store: GraphStore, task_id: str, deliverable: dict) -> TaskNode: ...
```

## CLI commands

```bash
# CRUD
arci taskcreate --module MOD-A4F8R2X1 --title "Implement lexer" \
  --phase implementation --task-type feature-implementation
arci taskshow TSK-E3K8S6V2
arci tasklist
arci tasklist --module MOD-A4F8R2X1 --phase implementation --status ready
arci taskupdate TSK-E3K8S6V2 --status in_progress
arci taskdelete TSK-E3K8S6V2

# Dependencies
arci taskdepend TSK-E3K8S6V2 --on TSK-G5M2R8X4
arci taskundepend TSK-E3K8S6V2 --on TSK-G5M2R8X4

# DAG queries
arci taskancestors TSK-E3K8S6V2
arci taskdescendants TSK-E3K8S6V2
arci taskblocking TSK-E3K8S6V2
arci taskready
arci taskcritical-path TSK-R7V3W9Y1

# Execution
arci taskstart TSK-E3K8S6V2      # Mark in_progress
arci taskcomplete TSK-E3K8S6V2   # Mark complete
arci taskblock TSK-E3K8S6V2 --reason "Waiting on API decision"
arci taskcancel TSK-E3K8S6V2 --reason "No longer needed"

# Deliverables
arci taskdeliverable TSK-E3K8S6V2 --kind commit --sha a1b2c3d4
arci taskdeliverables TSK-E3K8S6V2  # List deliverables

# Context
arci contexttask TSK-E3K8S6V2
```

## Examples

### Architecture task

```json
{"@context": "context.jsonld", "@id": "TSK-4RCH0001", "@type": "Task", "title": "Parser module decomposition", "module": {"@id": "MOD-A4F8R2X1"}, "processPhase": "architecture", "taskType": "module-decomposition", "status": "complete", "deliverables": [{"kind": "document", "type": "architecture-doc", "path": "modules/MOD-A4F8R2X1/architecture.md"}]}
```

### Implementation task

```json
{"@context": "context.jsonld", "@id": "TSK-E3K8S6V2", "@type": "Task", "title": "Implement lexer tokenization", "module": {"@id": "MOD-A4F8R2X1"}, "processPhase": "implementation", "taskType": "feature-implementation", "status": "complete", "dependsOn": [{"@id": "TSK-D3S1GN01"}], "deliverables": [{"kind": "commit", "sha": "a1b2c3d4e5f6", "message": "Implement lexer tokenization"}]}
```

### Verification task

```json
{"@context": "context.jsonld", "@id": "TSK-V3R1FY01", "@type": "Task", "title": "Parser code review", "module": {"@id": "MOD-A4F8R2X1"}, "processPhase": "verification", "taskType": "code-review", "status": "complete", "deliverables": [{"kind": "findings", "ids": ["FND-F1L4T7W5", "FND-K8Q3V6X2"]}, {"kind": "document", "type": "review-report", "path": "modules/MOD-A4F8R2X1/reviews/2026-01-03.md"}]}
```

### Milestone task

```json
{"@context": "context.jsonld", "@id": "TSK-R7V3W9Y1", "@type": "Task", "title": "Release 1.0", "module": {"@id": "MOD-OAPSROOT"}, "processPhase": "validation", "taskType": "release", "status": "pending", "dependsOn": [{"@id": "TSK-V3R1FY01"}, {"@id": "TSK-D0CS0001"}]}
```

## Implementation status

| Layer | Status | Notes |
|-------|--------|-------|
| Core | Implemented | Typed node, operations, queries in `arci.core.task` |
| IO | Implemented | JSON-LD serialization via `arci.io.graph` |
| Service | Implemented | Full CRUD, transitions (start, complete, block, cancel), dependency management |
| CLI | Implemented | 23 commands: CRUD, transitions, dependencies, deliverables, assign |

## Summary

Tasks are atomic work units that:

- Form a DAG with dependsOn relationships
- Belong to a process phase (architecture → validation)
- Track status from pending through complete
- Record deliverables with kind-specific fields
- Execute in atomic Claude Code sessions
- Replace plan containers with graph queries
- Store metadata in graph.jsonlt with no frontmatter in prose files
- Implemented following three-layer architecture (core/io/service)

The task DAG is the work organization—milestones are downstream tasks, scopes are transitive closures, and "the plan" is a query over the graph.
