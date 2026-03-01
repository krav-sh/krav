# Findings module type

## Overview

Findings (FND-*) are observations that emerge from review and analysis activities. They capture issues, recommendations, decisions, and questions discovered during verification, validation, or any review task.

Unlike test results (pass/fail on specific requirements), findings are qualitative observations that may require follow-up work, discussion, or decisions.

## Purpose

Findings serve multiple roles:

**Issue tracking**: problems discovered during review need tracking through resolution. Findings provide this without requiring external issue trackers.

**Decision capture**: review discussions produce decisions. Capturing these as findings preserves rationale.

**Recommendation pipeline**: not every observation is an issue. Recommendations capture improvement opportunities without blocking work.

**Question tracking**: ambiguities discovered during review need answers. Questions track these until resolved.

**Audit trail**: for compliance or quality purposes, findings document what the team reviewed and what it discovered.

## Finding types

The `findingType` field categorizes the nature of the observation:

| Type           | Description                       | Blocking? |
|----------------|-----------------------------------|-----------|
| issue          | Something wrong that needs fixing | Often     |
| recommendation | Suggested improvement             | No        |
| question       | Clarification needed              | Sometimes |
| decision       | Choice made during review         | No        |
| observation    | Neutral note                      | No        |

### Issues

Problems that typically need resolution before proceeding.

### Recommendations

Improvement suggestions that don't block progress.

### Questions

Ambiguities requiring clarification.

### Decisions

Choices made during review, captured for posterity.

### Observations

Neutral notes, neither good nor bad.

## Lifecycle

Findings progress through states:

```text
open → acknowledged → addressed → closed
                  ↘ wont_fix
```

| State        | Description                                  |
|--------------|----------------------------------------------|
| open         | Finding identified, not yet reviewed         |
| acknowledged | Finding reviewed and accepted as valid       |
| addressed    | Work completed to resolve the finding        |
| closed       | Finding verified as resolved                 |
| wont_fix     | Intentionally not addressed (with rationale) |

State transitions:

- `open → acknowledged`: Reviewer confirms finding is valid
- `acknowledged → addressed`: Resolution work complete
- `addressed → closed`: Resolution verified
- `acknowledged → wont_fix`: Decided not to address (with reason)

## Severity

For issues, severity indicates impact:

| Severity | Description                                    | Typical response |
|----------|------------------------------------------------|------------------|
| critical | Blocks release, major feature broken            | Immediate fix    |
| major    | Significant problem, should fix before release | Schedule fix     |
| minor    | Small issue, can defer if needed               | Backlog          |
| trivial  | Cosmetic, minimal impact                       | Optional fix     |

## Storage model

`graph.jsonlt` stores finding metadata as JSON-LD compact form. Prose files have no frontmatter; `graph.jsonlt` is the single source of truth for all structured data.

```json
{"@context": "context.jsonld", "@id": "FND-F1L4T7W5", "@type": "Finding", "title": "Missing line numbers in errors", "module": {"@id": "MOD-A4F8R2X1"}, "findingType": "issue", "severity": "major", "status": "open", "statement": "Error messages don't include line numbers", "rationale": "Makes debugging difficult for users", "regarding": {"@id": "MOD-A4F8R2X1"}}
```

Fields:

- `@id`: Unique identifier (FND-XXXXXXXX format)
- `@type`: Always "Finding"
- `title`: Human-readable title
- `module`: Module context for this finding (required)
- `findingType`: Type of finding (issue, recommendation, question, decision, observation)
- `severity`: For issues (critical, major, minor, trivial)
- `status`: Lifecycle state (open, acknowledged, addressed, closed, wont_fix)
- `statement`: The finding statement
- `rationale`: Why this is a finding / why it matters (optional)
- `resolutionNotes`: How the team resolved the finding (optional)
- `content`: Path to prose file for extended context (optional)
- `created`, `updated`, `closed`: ISO 8601 timestamps
- `tags`: Array of strings (optional)

## Relationships

The finding's JSON-LD record embeds relationships using `{"@id": "..."}` values.

### Outgoing relationships

| Property    | Target | Cardinality | Description                                            |
|-------------|--------|-------------|--------------------------------------------------------|
| module      | MOD-*  | Single      | Module context for this finding                        |
| regarding   | any    | Single      | What this finding is about (module, requirement, etc.) |
| generates   | TSK-*  | Single      | Task created to address this finding                   |

### Incoming relationships (queried via graph)

| Property    | Source | Description                     |
|-------------|--------|---------------------------------|
| addressedBy | TSK-*  | Task that resolved this finding |

### regarding targets

Findings can regard any module type:

```json
{"@context": "context.jsonld", "@id": "FND-F1L4T7W5", "@type": "Finding", "regarding": {"@id": "MOD-A4F8R2X1"}, ...}
{"@context": "context.jsonld", "@id": "FND-K8Q3V6X2", "@type": "Finding", "regarding": {"@id": "REQ-C2H6N4P8"}, ...}
{"@context": "context.jsonld", "@id": "FND-N33D1SS1", "@type": "Finding", "regarding": {"@id": "NED-B7G3M9K2"}, ...}
{"@context": "context.jsonld", "@id": "FND-C0NC3PT1", "@type": "Finding", "regarding": {"@id": "CON-K7M3NP2Q"}, ...}
```

Example with relationships:

```json
{"@context": "context.jsonld", "@id": "FND-F1L4T7W5", "@type": "Finding", "title": "Missing line numbers", "module": {"@id": "MOD-A4F8R2X1"}, "findingType": "issue", "severity": "major", "status": "addressed", "statement": "Error messages don't include line numbers", "regarding": {"@id": "MOD-A4F8R2X1"}, "generates": {"@id": "TSK-F1X00001"}}
```

## Finding sources

Verification and validation tasks produce findings as deliverables:

```json
{"@context": "context.jsonld", "@id": "TSK-V3R1FY01", "@type": "Task", "title": "Parser code review", "module": {"@id": "MOD-A4F8R2X1"}, "processPhase": "verification", "taskType": "code-review"}
```

The review task produces findings that link back via their module field.

## Finding resolution

When a finding needs work, generate a task:

```bash
arci findinggenerate-task FND-F1L4T7W5
```

This creates:

- A task to address the finding
- generates relationship from finding to task
- Sets the task's addressedBy relationship back to the finding

When the task completes:

```bash
arci findingclose FND-F1L4T7W5 --notes "Added line/column to errors"
```

This:

- Transitions finding to `closed`
- Records resolution notes

## Automatic findings

Some operations create findings automatically:

**Phase regression**: when a module regresses to an earlier phase, ARCI creates a finding:

```json
{"@context": "context.jsonld", "@id": "FND-R3GR3SS1", "@type": "Finding", "findingType": "issue", "severity": "major", "statement": "Boundary between lexer and tokenizer unclear", "rationale": "Module regression: MOD-A4F8R2X1 regressed to architecture"}
```

**Verification failures**: when tests fail, findings can capture the details:

```json
{"@context": "context.jsonld", "@id": "FND-T35TF41L", "@type": "Finding", "findingType": "issue", "severity": "critical", "statement": "Performance requirement REQ-C2H6N4P8 not met", "rationale": "p99 latency: 67ms (requirement: 50ms)"}
```

## Implementation architecture

Finding capability follows the three-layer architecture (see CON-GR4PH4RC).

### Typed node

FindingNode is an independent dataclass with typed fields. Graph stores FindingNode directly, and IO serializes directly to/from FindingNode (see CON-GR4PH4RC for the full architecture).

```python
@dataclass(frozen=True, slots=True)
class FindingNode:
    """Finding module—review observation."""
    id: str
    title: str
    status: FindingStatus          # Typed enum
    finding_type: FindingType      # Typed enum
    severity: Severity | None      # Typed enum or None
    statement: str                 # Type-specific field
    rationale: str = ""            # Type-specific field
    resolution_notes: str = ""     # Type-specific field
    content: str = ""              # Path to prose file
    description: str = ""
    created: datetime | None = None
    updated: datetime | None = None
```

All fields use proper types. The IO layer creates FindingNode directly from JSON-LD records, preserving all type-specific fields like `finding_type`, `severity`, `statement`, and `resolution_notes`.

### Core layer (`arci.core.finding`)

Pure functions and typed data structures:

```python
# Types
class FindingStatus(StrEnum):
    OPEN = "open"
    ACKNOWLEDGED = "acknowledged"
    ADDRESSED = "addressed"
    CLOSED = "closed"
    WONT_FIX = "wont_fix"

class FindingType(StrEnum):
    ISSUE = "issue"
    RECOMMENDATION = "recommendation"
    QUESTION = "question"
    DECISION = "decision"
    OBSERVATION = "observation"

class Severity(StrEnum):
    CRITICAL = "critical"
    MAJOR = "major"
    MINOR = "minor"
    TRIVIAL = "trivial"

# Typed node
@dataclass(frozen=True, slots=True)
class FindingNode(Node):
    finding_type: FindingType = FindingType.OBSERVATION
    severity: Severity | None = None
    statement: str = ""
    rationale: str = ""
    resolution_notes: str = ""
    content: str = ""

# Operations (pure functions)
def from_node(node: Node) -> FindingNode: ...
def with_status(finding: FindingNode, status: FindingStatus) -> FindingNode: ...
def can_transition(finding: FindingNode, target: FindingStatus) -> bool: ...
def is_blocking(finding: FindingNode) -> bool: ...

# Queries (pure, take Graph)
def get(graph: Graph, finding_id: str) -> FindingNode | None: ...
def list_all(graph: Graph) -> tuple[FindingNode, ...]: ...
def list_by_module(graph: Graph, module_id: str) -> tuple[FindingNode, ...]: ...
def list_by_status(graph: Graph, status: FindingStatus) -> tuple[FindingNode, ...]: ...
def list_open(graph: Graph) -> tuple[FindingNode, ...]: ...
def list_blocking(graph: Graph) -> tuple[FindingNode, ...]: ...
def regarding(graph: Graph, finding_id: str) -> str | None: ...
def generated_task(graph: Graph, finding_id: str) -> str | None: ...
```

### Service layer (`arci.service.finding`)

Orchestrates core and IO:

```python
# Thin query wrappers
def get(store: GraphStore, finding_id: str) -> FindingNode | None: ...
def list_all(store: GraphStore) -> tuple[FindingNode, ...]: ...
def list_open(store: GraphStore) -> tuple[FindingNode, ...]: ...
def list_blocking(store: GraphStore) -> tuple[FindingNode, ...]: ...

# Mutations
def create(
    store: GraphStore,
    module_id: str,
    finding_type: FindingType,
    statement: str,
    regarding: str | None = None,
    severity: Severity | None = None,
    ...
) -> FindingNode: ...
def update(store: GraphStore, finding_id: str, **fields) -> FindingNode: ...
def delete(store: GraphStore, finding_id: str) -> None: ...
def transition(store: GraphStore, finding_id: str, target: FindingStatus) -> FindingNode: ...

# Workflows
def acknowledge(store: GraphStore, finding_id: str) -> FindingNode: ...
def generate_task(store: GraphStore, finding_id: str) -> TaskNode: ...
def close(store: GraphStore, finding_id: str, notes: str = "") -> FindingNode: ...
def wont_fix(store: GraphStore, finding_id: str, reason: str) -> FindingNode: ...
```

## CLI commands

```bash
# CRUD
arci findingcreate --module MOD-A4F8R2X1 --type issue \
  --statement "Error messages don't include line numbers"
arci findingshow FND-F1L4T7W5
arci findinglist
arci findinglist --module MOD-A4F8R2X1 --type issue --status open
arci findingupdate FND-F1L4T7W5 --severity critical
arci findingdelete FND-F1L4T7W5

# Lifecycle
arci findingacknowledge FND-F1L4T7W5
arci findinggenerate-task FND-F1L4T7W5
arci findingaddress FND-F1L4T7W5 --task TSK-F1X00001
arci findingclose FND-F1L4T7W5 --notes "Fixed in commit abc123"
arci findingwontfix FND-F1L4T7W5 --reason "Out of scope for v1"

# Queries
arci findingopen                     # All open findings
arci findingopen --module MOD-A4F8R2X1  # Open for module
arci findingblocking                 # Open issues blocking advancement
arci findingby-type issue            # All issues
arci findingby-severity critical     # Critical issues
```

## Examples

### Issue finding

```json
{"@context": "context.jsonld", "@id": "FND-F1L4T7W5", "@type": "Finding", "title": "Missing line numbers in errors", "module": {"@id": "MOD-A4F8R2X1"}, "findingType": "issue", "severity": "major", "status": "closed", "statement": "Error messages don't include line numbers", "regarding": {"@id": "MOD-A4F8R2X1"}, "generates": {"@id": "TSK-F1X00001"}, "resolutionNotes": "Added line/column to all error types"}
```

### Recommendation finding

```json
{"@context": "context.jsonld", "@id": "FND-R3C0MM01", "@type": "Finding", "title": "JSON output flag suggestion", "module": {"@id": "MOD-B9G3M7K2"}, "findingType": "recommendation", "status": "acknowledged", "statement": "Consider adding --json flag for machine-readable output", "regarding": {"@id": "MOD-B9G3M7K2"}}
```

### Decision finding

```json
{"@context": "context.jsonld", "@id": "FND-D3C1S10N", "@type": "Finding", "title": "Parser implementation approach", "module": {"@id": "MOD-A4F8R2X1"}, "findingType": "decision", "status": "closed", "statement": "Will use recursive descent parser", "rationale": "Simpler implementation, team familiarity, adequate for grammar complexity", "regarding": {"@id": "MOD-A4F8R2X1"}}
```

### Question finding

```json
{"@context": "context.jsonld", "@id": "FND-QU35T10N", "@type": "Finding", "title": "Windows line ending support", "module": {"@id": "MOD-OAPSROOT"}, "findingType": "question", "status": "open", "statement": "Should we support Windows line endings?", "regarding": {"@id": "MOD-OAPSROOT"}}
```

## Blocking findings

Findings can block module phase advancement:

```bash
arci moduleadvance MOD-A4F8R2X1 --to implementation
# Error: Cannot advance. 2 blocking findings:
#   FND-F1L4T7W5: Error messages don't include line numbers (major issue)
#   FND-QU35T10N: Should we support Windows line endings? (open question)
```

By default, open issues and questions block advancement. Recommendations and observations do not.

## Extended content

Complex findings can have prose files:

```json
{"@context": "context.jsonld", "@id": "FND-F1L4T7W5", "@type": "Finding", "content": "findings/F1L4T7W5-error-line-numbers.md", ...}
```

The prose file contains extended context, evidence, and discussion.

## Implementation status

| Layer | Status | Notes |
|-------|--------|-------|
| Core | Implemented | Typed node, operations, queries in `arci.core.finding` |
| IO | Implemented | JSON-LD serialization via `arci.io.graph` |
| Service | Implemented | All CRUD, transitions, and workflows in `arci.service.finding` |
| CLI | Implemented | Full command set in `arci.cli._commands._finding` |

## Summary

Findings capture review observations:

- Typed as issue, recommendation, question, decision, or observation
- Progress from open through acknowledged to closed or wont_fix
- Link to what they're about via regarding relationships
- Generate tasks via generates relationships
- Get resolved by tasks via addressedBy relationships
- Can block module phase advancement
- Produced by verification/validation tasks as deliverables
- Store metadata in graph.jsonlt with no frontmatter in prose files
- Implemented following three-layer architecture (core/io/service)

Findings provide structured tracking of review outcomes without requiring external issue trackers.
