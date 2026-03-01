# Baselines

## Overview

Baselines (BSL-*) capture the state of the knowledge graph at a decision point. A baseline records which nodes existed, their states, and the relationships between them, anchoring everything to a specific git commit. Baselines enable milestone recording, change auditing, regression detection, and phase gate enforcement.

Traditional RE tools treat baselines as snapshots of a requirements database. Krav takes a lighter approach: since graph.jsonlt is version-controlled and append-only, the git history already contains every historical state. A baseline is a named reference into that history with metadata about why someone created it, what it covers, and who approved it.

## Purpose

Baselines serve multiple roles:

**Milestone recording**: "This is what the team agreed at the architecture review." A baseline freezes the graph state at a decision point, creating an unambiguous record of what existed when the team made a commitment.

**Change auditing**: semantic diff between baselines shows what changed in terms the project cares about (nodes that the team added, modified, or removed; relationships that changed; phases that advanced) rather than raw JSONLT line diffs.

**Phase gates**: phase advancement can require a baseline of the current phase before proceeding, ensuring the system records the pre-advancement state for later review. The architecture baseline captures what the team agreed before design begins; the design baseline captures what the team agreed before coding.

**Regression detection**: comparing the current graph against a baseline reveals unintended changes. A requirement that existed in the architecture baseline but is missing now warrants investigation.

**Suspect link review**: baselines interact with suspect propagation. When reviewing suspect links, the baseline provides a reference point: "this link was valid at baseline X, what changed since then?"

## Storage model

Baseline metadata lives in `graph.jsonlt` as JSON-LD compact form, like all other node types. The baseline record does not contain a full graph snapshot; it stores a git commit SHA that Krav uses to reconstruct the graph state at baseline time.

```json
{"@context": "context.jsonld", "@id": "BSL-R3L3AS31", "@type": "Baseline", "title": "Architecture baseline", "module": {"@id": "MOD-OAPSROOT"}, "scope": "subtree", "commitSha": "a1b2c3d4e5f6789...", "phase": "architecture", "status": "approved", "approvedBy": "tony", "approvedAt": "2026-02-28T16:00:00Z", "description": "Architecture phase complete for root module. All architecture tasks done, no blocking findings.", "statistics": {"modules": 5, "concepts": 12, "needs": 8, "requirements": 15, "verifications": 6, "tasks": 23, "findings": {"open": 0, "closed": 7}}}
```

Fields:

- `@id`: Unique identifier (BSL-XXXXXXXX format)
- `@type`: Always "Baseline"
- `title`: Human-readable title
- `module`: The root module for this baseline
- `scope`: What the baseline covers (see Scope below)
- `commitSha`: Git commit SHA anchoring the graph state
- `phase`: The lifecycle phase this baseline captures (optional, for phase-gate baselines)
- `status`: Lifecycle state (see Lifecycle below)
- `approvedBy`: Who approved this baseline (optional)
- `approvedAt`: When the approver approved it (optional)
- `description`: Why someone created this baseline
- `statistics`: Denormalized counts at baseline time (see Statistics below)
- `created`, `updated`: ISO 8601 timestamps
- `tags`: Array of strings (optional)

### Why git commit SHA, not a full snapshot?

The graph.jsonlt file is version-controlled. Checking out graph.jsonlt at the baseline's commit SHA reconstructs any historical state. This avoids duplicating the entire graph inside the baseline record (which would be expensive and redundant), while remaining fully reproducible.

The tradeoff: if someone rewrites git history (force push, rebase) and the baseline's commit SHA becomes unreachable, the baseline cannot resolve. This is a feature, as it surfaces history tampering. Projects that need tamper-evident baselines should protect the branch containing `.krav/` from force pushes.

### Statistics

The `statistics` field captures a denormalized snapshot of graph counts at baseline time. This is redundant with the commit SHA (you could recompute it) but serves two purposes: quick inspection without materializing the historical graph, and detection of history tampering (if the reconstructed counts don't match the stored statistics, something changed).

```json
{
  "statistics": {
    "modules": 5,
    "concepts": 12,
    "needs": 8,
    "requirements": 15,
    "verifications": 6,
    "tasks": 23,
    "findings": {
      "open": 0,
      "acknowledged": 2,
      "addressed": 3,
      "closed": 7,
      "wont_fix": 1
    },
    "suspectLinks": 0,
    "verificationCoverage": 0.87
  }
}
```

The `verificationCoverage` field records the ratio of requirements that have at least one passing verification. The `suspectLinks` count records how many links carried the suspect flag at baseline time; a healthy baseline should have zero.

## Scope

Baselines scope to an module subtree. The `scope` field indicates what's included:

**subtree** (default): the baseline covers the specified module and all its descendants, including all nodes owned by any module in the subtree (needs, requirements, verifications, tasks, findings) and all relationships between them. This is the common case: baseline a module to capture its complete state.

**module-only**: the baseline covers only the specified module's directly owned nodes, not descendants. Useful for component-level baselines where child modules receive independent baselines.

The module field determines the root of the scope. Baselining the root module with `subtree` scope captures the entire project.

```bash
# Baseline the whole project
Krav baseline create --module MOD-OAPSROOT --title "Architecture baseline"

# Baseline a subsystem
Krav baseline create --module MOD-A4F8R2X1 --title "Parser design baseline" --scope subtree

# Baseline a single component
Krav baseline create --module MOD-L3X3R001 --title "Lexer implementation baseline" --scope module-only
```

## Lifecycle

Baselines have a simple lifecycle:

```text
draft → approved → superseded
```

| State      | Description                                               |
|------------|-----------------------------------------------------------|
| draft      | Baseline exists but nobody has reviewed or approved it yet |
| approved   | Reviewer accepted the baseline as an official record      |
| superseded | A newer baseline for the same module and phase exists     |

State transitions:

- `draft → approved`: Reviewer approves the baseline, setting approvedBy and approvedAt.
- `approved → superseded`: A newer baseline gains approval for the same module and phase. The older baseline remains in the graph for historical reference but no longer serves as the active baseline for that scope.

Draft baselines are useful when a phase gate requires a baseline but the team defers approval (the developer creates the baseline, a reviewer approves it later). For solo projects or less formal workflows, baselines can enter `approved` status directly at creation.

## Phase gate integration

Baselines integrate with module phase advancement. A hook policy can require the team to baseline the current phase before advancing:

```yaml
policies:
  - name: require-baseline-before-advance
    description: Ensure the current phase is baselined before advancing
    match:
      tool: krav
      args:
        - match: "module"
          position: 0
        - match: "advance"
          position: 1
    conditions:
      - expr: >
          !baselines.exists(b,
            b.module == input.args.module &&
            b.phase == state.currentPhase &&
            b.status == 'approved')
    rules:
      - effect: deny
        message: "Create and approve a baseline for the current phase before advancing"
```

When phase advancement triggers a baseline:

1. The CLI commits any pending changes to graph.jsonlt
2. The CLI creates a BSL-* record with the current commit SHA
3. The CLI computes statistics from the current graph state
4. If the configuration enables auto-approve, the baseline enters `approved` status immediately
5. Phase advancement proceeds

The resulting baseline records exactly what existed when the module left that phase. Later, `krav baseline diff` can show what changed between phases.

## Semantic diff

The primary analytical operation on baselines is semantic diff: given two baselines (or a baseline and the current state), produce a structured comparison at the graph level.

### Reconstruction

To diff two baselines, Krav materializes the graph at each commit:

1. Read graph.jsonlt at baseline A's commit SHA (via `git show <sha>:.krav/graph.jsonlt`)
2. Read graph.jsonlt at baseline B's commit SHA (or current working tree)
3. Materialize both into in-memory Graph instances
4. Scope each graph to the baseline's module subtree
5. Compute structural diff

### Diff output

The diff produces a structured result covering five dimensions.

Node changes identify nodes that the team added, modified, or removed between baselines. A modification is any change to a node's fields (status, statement, phase, etc.). The diff reports which fields changed and their old/new values.

Relationship changes identify links that the team added, removed, or modified (suspect flag set, budget changed). This is where suspect propagation becomes visible: a link that was healthy at baseline A but suspect at baseline B shows up here.

Phase changes show module phase transitions between baselines, which modules advanced, regressed, or remained unchanged.

Coverage changes show how verification coverage shifted: new requirements without verifications, newly verified requirements, verifications that changed status.

Statistics delta compares the aggregate counts between baselines.

### CLI output

```text
$ krav baseline diff BSL-4RCH0001 BSL-D3S1GN01

Comparing "Architecture baseline" → "Design baseline" for MOD-OAPSROOT
  Time span: 2026-01-15 → 2026-02-28
  Commit range: a1b2c3d..f6e5d4c

Nodes added (7):
  REQ-N3WR3Q01  "API response time < 50ms" (requirement, approved)
  REQ-N3WR3Q02  "Request ID in all responses" (requirement, approved)
  REQ-N3WR3Q03  "Structured error responses" (requirement, draft)
  VRF-N3WV3R01  "Response time benchmark" (verification, ready)
  TSK-D3S1GN01  "Design parser API" (task, done)
  TSK-D3S1GN02  "Design error catalog" (task, done)
  FND-R3V13W01  "Error catalog incomplete" (finding, addressed)

Nodes modified (3):
  NED-B7G3M9K2  statement: "fast feedback" → "sub-second feedback"
  MOD-A4F8R2X1  phase: architecture → design
  MOD-OAPSROOT  phase: architecture → design

Nodes removed (1):
  CON-OLDCON01  "Alternative parser approach" (superseded)

Suspect links (1):
  REQ-C2H6N4P8 ←derivesFrom— NED-B7G3M9K2
    Reason: NED-B7G3M9K2 statement modified after link established

Coverage: 60% → 73% (+13%)
  Newly verified: REQ-C2H6N4P8
  Still unverified: REQ-N3WR3Q02, REQ-N3WR3Q03
```

## Relationships

### Outgoing relationships

| Property | Target | Cardinality | Description                                |
|----------|--------|-------------|--------------------------------------------|
| module   | MOD-*  | Single      | Root module this baseline covers           |

### Incoming relationships (queried via graph)

Baselines don't typically have incoming relationships from other node types. Baselines serve as reference points, not targets of traceability.

### Baseline-to-baseline ordering

Baselines for the same module and phase form a temporal sequence via their `created` timestamps and the superseded lifecycle state. The most recent approved, non-superseded baseline for a given module/phase pair is the "current" baseline for that scope.

## Implementation architecture

Baseline capability follows the three-layer architecture.

### Typed node

```python
@dataclass(frozen=True, slots=True)
class BaselineNode:
    """Baseline—named graph state reference."""
    id: str
    title: str
    status: BaselineStatus             # Typed enum
    module_id: str                     # Root module for scope
    scope: BaselineScope               # Typed enum
    commit_sha: str                    # Git commit anchor
    phase: ModulePhase | None = None   # Phase gate context
    approved_by: str = ""
    approved_at: datetime | None = None
    description: str = ""
    statistics: BaselineStatistics | None = None
    created: datetime | None = None
    updated: datetime | None = None
```

### Types

```python
class BaselineStatus(StrEnum):
    DRAFT = "draft"
    APPROVED = "approved"
    SUPERSEDED = "superseded"

class BaselineScope(StrEnum):
    SUBTREE = "subtree"
    MODULE_ONLY = "module-only"

@dataclass(frozen=True, slots=True)
class BaselineStatistics:
    modules: int = 0
    concepts: int = 0
    needs: int = 0
    requirements: int = 0
    verifications: int = 0
    tasks: int = 0
    findings_open: int = 0
    findings_closed: int = 0
    suspect_links: int = 0
    verification_coverage: float = 0.0
```

### Core layer (`krav.core.baseline`)

Pure functions for baseline operations:

```python
# Operations
def can_transition(baseline: BaselineNode, target: BaselineStatus) -> bool: ...
def with_status(baseline: BaselineNode, status: BaselineStatus) -> BaselineNode: ...
def with_approval(baseline: BaselineNode, approved_by: str, approved_at: datetime) -> BaselineNode: ...

# Queries (pure, take Graph)
def get(graph: Graph, baseline_id: str) -> BaselineNode | None: ...
def list_all(graph: Graph) -> tuple[BaselineNode, ...]: ...
def list_by_module(graph: Graph, module_id: str) -> tuple[BaselineNode, ...]: ...
def current_for_phase(graph: Graph, module_id: str, phase: ModulePhase) -> BaselineNode | None: ...
def has_approved_baseline(graph: Graph, module_id: str, phase: ModulePhase) -> bool: ...
```

### Core layer (`krav.core.baseline_diff`)

Pure functions for semantic diff. These operate on two Graph instances and produce a structured diff result:

```python
@dataclass(frozen=True, slots=True)
class NodeChange:
    id: str
    node_type: str
    title: str
    change_type: Literal["added", "modified", "removed"]
    field_changes: tuple[FieldChange, ...] = ()  # For modified nodes

@dataclass(frozen=True, slots=True)
class FieldChange:
    field: str
    old_value: str
    new_value: str

@dataclass(frozen=True, slots=True)
class LinkChange:
    source_id: str
    predicate: str
    target_id: str
    change_type: Literal["added", "removed", "suspect_set", "suspect_cleared"]

@dataclass(frozen=True, slots=True)
class BaselineDiff:
    node_changes: tuple[NodeChange, ...]
    link_changes: tuple[LinkChange, ...]
    phase_changes: tuple[tuple[str, str, str], ...]  # (module_id, old_phase, new_phase)
    coverage_before: float
    coverage_after: float
    statistics_before: BaselineStatistics
    statistics_after: BaselineStatistics

def diff_graphs(
    before: Graph,
    after: Graph,
    module_id: str,
    scope: BaselineScope,
) -> BaselineDiff:
    """Compute semantic diff between two graph states scoped to an module."""
    ...
```

### IO layer

Baseline persistence uses the same GraphStore append mechanism as all other nodes. The git interaction (reading graph.jsonlt at a historical commit) requires a new IO component:

```python
# krav/io/git.py

def read_file_at_commit(repo_root: Path, commit_sha: str, relative_path: str) -> str:
    """Read a file's contents at a specific git commit."""
    ...

def current_commit_sha(repo_root: Path) -> str:
    """Get the current HEAD commit SHA."""
    ...

def has_uncommitted_changes(repo_root: Path, path: str) -> bool:
    """Check if a file has uncommitted changes."""
    ...
```

### Service layer (`krav.service.baseline`)

Orchestrates baseline creation, approval, and diff:

```python
def create(
    store: GraphStore,
    module_id: str,
    title: str,
    scope: BaselineScope = BaselineScope.SUBTREE,
    phase: ModulePhase | None = None,
    auto_approve: bool = False,
    approved_by: str = "",
    description: str = "",
) -> BaselineNode:
    """Create a baseline anchored to the current commit."""
    # Verify no uncommitted changes to graph.jsonlt
    # Get current commit SHA
    # Compute statistics from current graph scoped to module
    # Create BaselineNode
    # If phase specified, supersede previous baseline for same module/phase
    # Persist
    ...

def approve(
    store: GraphStore,
    baseline_id: str,
    approved_by: str,
) -> BaselineNode:
    """Approve a draft baseline."""
    ...

def diff(
    store: GraphStore,
    baseline_a_id: str,
    baseline_b_id: str | None = None,  # None = compare against current
) -> BaselineDiff:
    """Compute semantic diff between two baselines (or baseline vs current)."""
    # Load baseline A record
    # Load graph.jsonlt at baseline A's commit
    # Materialize graph A
    # Load graph B (from baseline B's commit or current store.graph)
    # Call core diff_graphs
    ...
```

## CLI commands

```bash
# Create
Krav baseline create --module MOD-OAPSROOT --title "Architecture baseline"
Krav baseline create --module MOD-OAPSROOT --title "Architecture baseline" \
  --phase architecture --auto-approve --approved-by tony

# List and show
Krav baseline list
Krav baseline list --module MOD-OAPSROOT
Krav baseline list --module MOD-OAPSROOT --phase architecture
Krav baseline show BSL-R3L3AS31

# Approve
Krav baseline approve BSL-R3L3AS31 --approved-by tony

# Diff
Krav baseline diff BSL-4RCH0001 BSL-D3S1GN01
Krav baseline diff BSL-D3S1GN01              # Compare against current state

# Verify integrity
Krav baseline verify BSL-R3L3AS31            # Check commit is reachable, statistics match
```

## Interaction with other features

### Phase advancement

Phase advancement can optionally require an approved baseline. Hook policy configures this requirement (see the preceding Phase gate integration section) rather than hardcoding it into the advancement logic.

### Suspect links

When reviewing suspect links, baselines provide temporal context. The review finding can reference the baseline where the link was last known-good: "This link was valid at BSL-4RCH0001. Someone modified NED-B7G3M9K2 in commit f6e5d4c."

### Findings

Baseline creation can itself produce findings. If open blocking findings or suspect links exist at baseline time, the create command can warn or (via hook policy) refuse to create the baseline until someone resolves them.

### Templates

A baseline-creation task template could standardize the baselining process: review open findings, clear suspect links, run verification suite, create and approve baseline, advance phase.

## Examples

### Architecture phase gate baseline

```json
{"@context": "context.jsonld", "@id": "BSL-4RCH0001", "@type": "Baseline", "title": "Architecture baseline", "module": {"@id": "MOD-OAPSROOT"}, "scope": "subtree", "commitSha": "a1b2c3d4e5f6789abcdef0123456789abcdef01", "phase": "architecture", "status": "approved", "approvedBy": "tony", "approvedAt": "2026-01-15T14:30:00Z", "description": "Architecture phase complete. Module hierarchy established, key interfaces identified, architecture review findings all closed.", "statistics": {"modules": 5, "concepts": 12, "needs": 8, "requirements": 15, "verifications": 6, "tasks": 23, "findings": {"open": 0, "closed": 7}, "suspectLinks": 0, "verificationCoverage": 0.4}}
```

### Subsystem design baseline

```json
{"@context": "context.jsonld", "@id": "BSL-D3S1GN01", "@type": "Baseline", "title": "Parser design baseline", "module": {"@id": "MOD-A4F8R2X1"}, "scope": "subtree", "commitSha": "f6e5d4c3b2a19876543210fedcba9876543210fe", "phase": "design", "status": "approved", "approvedBy": "tony", "approvedAt": "2026-02-28T16:00:00Z", "description": "Parser API design finalized. All design tasks complete, API spec and data model documented.", "statistics": {"modules": 3, "concepts": 4, "needs": 3, "requirements": 8, "verifications": 5, "tasks": 11, "findings": {"open": 0, "closed": 4}, "suspectLinks": 0, "verificationCoverage": 0.625}}
```

### Draft baseline pending review

```json
{"@context": "context.jsonld", "@id": "BSL-1MPL0001", "@type": "Baseline", "title": "Implementation checkpoint", "module": {"@id": "MOD-A4F8R2X1"}, "scope": "subtree", "commitSha": "1234567890abcdef1234567890abcdef12345678", "status": "draft", "description": "Implementation checkpoint before refactoring parser internals."}
```

## Implementation status

| Layer | Status | Notes |
|-------|--------|-------|
| Core | Not yet | BaselineNode, operations, queries, diff |
| IO | Not yet | Git commit reading, baseline serialization |
| Service | Not yet | Create, approve, diff orchestration |
| CLI | Not yet | Commands for create, list, show, approve, diff, verify |

## Summary

Each baseline serves as a named reference into git history, capturing the knowledge graph state at a decision point:

- Anchor to git commit SHAs rather than full graph snapshots
- Scope to module subtrees for targeted baselining
- Integrate with phase gates via hook policies
- Semantic diff produces structured changelogs at the graph level (not JSONLT line diffs)
- Statistics snapshot enables quick inspection and integrity verification
- Temporal sequencing via lifecycle (draft → approved → superseded)
- Store metadata in graph.jsonlt like all other node types
