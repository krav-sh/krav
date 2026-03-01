# ARCI design documentation

ARCI (Agentic Requirements Composition & Integration) is a system for Claude Code that combines policy-driven hooks with spec-driven development. It provides declarative rules to validate, transform, and control tool execution, alongside an INCOSE-inspired knowledge graph that structures requirements, test cases, and traceability.

## Two pillars

### Hooks

The hook system intercepts Claude Code tool execution through declarative policies. Policies match events, evaluate CEL conditions, and produce admission decisions (allow, deny, warn) or mutations. A server-based architecture keeps evaluation fast; fail-open semantics ensure hooks never block the assistant unexpectedly.

See [Hooks overview](hooks/index.md) for the full hook documentation.

### Specs

The spec system structures software development around INCOSE-inspired systems engineering practices. A knowledge graph of typed nodes captures concepts, needs, requirements, test cases, tasks, defects, baselines, and test plans, providing full traceability from stakeholder expectations to verified implementations.

Twelve node types form the knowledge graph:

| Prefix | Type                  | Role                                                  |
|--------|-----------------------|-------------------------------------------------------|
| CON-*  | Concept               | Exploration, decisions, design thinking               |
| MOD-*  | Module                | Architectural container, owns requirements            |
| NEED-* | Need                  | Stakeholder expectation (validated)                   |
| REQ-*  | Requirement           | Design obligation (verified)                          |
| TC-*   | Test case             | Verification case specification                       |
| TASK-* | Task                  | Atomic work unit in DAG                               |
| DEF-*  | Defect                | Identified problem requiring action                   |
| BSL-*  | Baseline              | Named graph state reference at a decision point       |
| STK-*  | Stakeholder           | Named party with concerns about the system            |
| TP-*   | Test plan             | Coordinated verification evidence package             |
| DEV-*  | Developer             | Human actor who initiates sessions and makes decisions |
| AGT-*  | Agent                 | Claude Code session or subagent, ephemeral            |

Node identifiers use variable-length prefixes with 8-character Crockford Base32 nanoids: `REQ-C2H6N4P8`, `NEED-B7G3M9K2`, `TASK-E3K8S6V2`.

The formal transformation chain runs: concept → need → requirement → task → deliverable. Test cases verify that the system meets requirements. Defects track problems found during review and verification. Baselines freeze the graph state at milestones. Test plans capture coordinated verification evidence for releases and phase gates.

## Knowledge graph

Schema, relationships, constraints, and query patterns:

- [Graph overview](graph/index.md): entry point for graph design documentation, reading order for ontology docs

Reference builds and architecture notes (predating the graph/ directory, retained for detail and performance benchmarks):

- Data model: core data model with storage format, typed nodes, CLI surface, and examples
- Knowledge graph architecture: three-layer architecture, serialization, rustworkx integration, performance characteristics

## Node type specifications

### Intent: Capturing stakeholder expectations

- [Concepts](graph/nodes/concepts.md) (CON-*): exploration, design decisions, crystallized thinking
- [Needs](graph/nodes/needs.md) (NEED-*): stakeholder expectations, validated against intent

### Requirements: Formalizing into verifiable obligations

- [Modules](graph/nodes/modules.md) (MOD-*): architectural containers with hierarchy and phase tracking
- [Requirements](graph/nodes/requirements.md) (REQ-*): design obligations with verification criteria
- [Baselines](graph/nodes/baselines.md) (BSL-*): named snapshots of graph state, anchored to git commits, with semantic diff

### Execution: Building the work

- [Tasks](graph/nodes/tasks.md) (TASK-*): atomic work units in a DAG, organized by process phase
- [Templating](execution/templating.md): reusable task and decomposition templates

### Cross-cutting

- [Domain context](domain-context.md): project-wide and module-scoped domain knowledge for agent context injection
- [Workflows](workflows/index.md): human-initiated workflows covering the full development lifecycle

### Verification: Checking the work

- [Test cases](graph/nodes/test-cases.md) (TC-*): verification case specifications decoupled from execution results
- [Defects](graph/nodes/defects.md) (DEF-*): identified problems with disposition and resolution lifecycles
- Test plans (TP-*): coordinated verification events for milestones and releases (not yet spec'd)

### Provenance: Tracking who did the work

- [Developers](graph/nodes/developers.md) (DEV-*): human actors with persistent identity
- [Agents](graph/nodes/agents.md) (AGT-*): Claude Code sessions and subagents, ephemeral per invocation

## Key design decisions

The design process produced these decisions, which appear across the specs. This section collects them for orientation.

**Module naming and PROV-O**: the architectural container type changed from `Entity` to `Module` to avoid collision with `prov:Entity` from the W3C PROV-O ontology. PROV-O's provenance vocabulary (`prov:wasDerivedFrom`, `prov:wasAttributedTo`, `prov:wasGeneratedBy`) maps well onto ARCI's needs for tracking which agent session created or modified graph nodes. The JSON-LD context can import PROV-O alongside the ARCI vocabulary.

**External vocabulary alignment**: the JSON-LD context uses Dublin Core Terms (`dcterms:title`, `dcterms:description`, `dcterms:created`, `dcterms:modified`) directly for common metadata properties where ARCI adds no additional semantics. For node types and predicates where ARCI adds lifecycle states, suspect propagation, DAG enforcement, or other constraints, the schema declares `rdfs:subClassOf` and `rdfs:subPropertyOf` relationships to OSLC and PROV-O vocabularies in the T-Box without changing instance data. Concept and Need have no external counterparts since OSLC conflates needs and requirements and has no pre-requirement exploration phase. See [Vocabulary alignment](graph/vocabulary-alignment.md) for the complete mapping.

**Defects replace findings**: the original Findings type (FND-*) conflated five concerns: issues, recommendations, questions, decisions, and observations. The redesigned Defects type (DEF-*) narrows to actual problems only. Decisions moved to concepts (CON-*). Questions are task-level context. Observations are review deliverable prose. Recommendations, if important enough to track, become needs (NEED-*).

**Test case specification decoupled from execution**: the original Verification type combined specification and execution state. The redesigned Test Case type (Tc-*) is a specification: what to verify, which method, acceptance criteria, requirements linkage. Everyday execution updates a `currentResult` field on the Tc node without creating graph entities. Formal milestone verification produces a test plan (TP-*) that captures the full evidence package. This decoupling matters in the agent execution model because agent-written test builds (benchmarks, e2e tests, analyses) are themselves verifiable artifacts that reviewers can examine and file defects against.

**Link metadata and suspect propagation**: relationships carry optional metadata beyond the target `@id`: timestamps, rationale, and a `suspect` flag. When a node changes, ARCI marks downstream traceability links (derivesFrom, verifiedBy, allocatesTo) as suspect. Suspect links surface in views for reviewer triage; reviewers clear the flag, create a defect, or update the downstream node. Suspect links don't auto-generate defects to avoid flooding the defect list with items that may not be real problems.

**Baselines use git commit anchoring, not full snapshots**: baselines store a git commit SHA rather than duplicating the entire graph. ARCI reconstructs historical graph state by reading graph.jsonlt at the baseline's commit. Semantic diff between baselines produces structured changelogs at the graph level. Baselines scope to module subtrees and integrate with phase gates.

**Views replace document management**: traditional RE document management (SRS, test plans, traceability matrices, V&V plans) maps to views on the graph, not to separate document entities. Requirements specifications are queries over requirements by module scope. Test plans are the set of Tc-* nodes grouped by level. Traceability matrices are bipartite projections of REQ and Tc. The CLI and dashboard expose these as commands and pages.

**Change discipline through layered enforcement**: defense in depth protects baselined content rather than a change-proposal workflow entity. Hooks deny writes to baselined nodes, skills instruct agents not to modify baselined content, pre-commit hooks catch violations, GitHub Actions validate on push, and CLI commands refuse mutations. If baselined content needs to change, the system regresses the module's phase (creating a defect automatically), applies changes, and re-baselines and re-advances the module.

**Developer and Agent provenance**: two node types track who performed actions. Developer (DEV-*) models persistent human actors. Agent (AGT-*) models ephemeral Claude Code sessions and subagents, distinguished by a required `sessionId` and nullable `subagentId`. Both declare `rdfs:subClassOf prov:Agent`. The `operator` predicate links agents to the developer who initiated the session. Provenance properties `generatedBy` (`prov:wasGeneratedBy`) and `attributedTo` (`prov:wasAttributedTo`) are available on all node types to record which task created a node and which actor was responsible. See [Vocabulary alignment](graph/vocabulary-alignment.md) for the external vocabulary mapping.

## Architecture diagrams

C4 and domain model diagrams in PlantUML, with rendered PNG/SVG:

- [diagrams/](diagrams/): PlantUML sources
  - `arci-c4-context.puml`: system context diagram (Level 1)
  - `arci-c4-container.puml`: container diagram (Level 2), runtime processes and data stores
  - `arci-c4-component.puml`: component diagram (Level 3), internals of the ARCI binary
  - `arci-knowledge-graph.puml`: domain model of all node types and relationships

Render locally: `plantuml -tpng -o png/ diagrams/*.puml`

## Infrastructure

Shared components that support both pillars:

### Configuration

- [Configuration](configuration/configuration.md): layered configuration system with precedence rules and hot reloading
- [Config cascade](configuration/config-cascade.md): full precedence chain from built-in defaults to CLI flags
- [Managed config](configuration/config-managed.md): managed/recommended and managed/required configuration

### Server

- [Server](server/index.md): long-running process that owns the knowledge graph, caches configuration, and serves the dashboard and API
- [Server discovery](server/discovery.md): how the CLI locates a running server via `.arci/server.json`

### MCP

- [MCP server](mcp/index.md): MCP integration for Claude Code, exposing diagnostic and introspection tools

### CLI

- [CLI](cli/index.md): command-line interface for hook evaluation and spec management

### State and storage

- [State store](state-store.md): persistent key-value store for tracking data across hook invocations

### Security

- [Security model](security.md): trust model, threat scenarios, and security controls
- [Sandboxing](sandboxing.md): platform-native shell action isolation

### Observability

- [Dashboard](dashboard/index.md): web-based diagnostics dashboard

### Extensions

- [Extensions](extensions.md): unified extension system for distributing policies and custom functions

### Testing and quality

- [Testing strategy](testing.md): testing approaches for the engine, shells, and policies
- Performance: latency expectations and optimization strategies

### Architecture and operations

- [Architecture](architecture.md): arc42 system architecture (hooks, knowledge graph, runtime views)
- [Installation](installation.md): installation and upgrade workflows
- [Versioning](versioning.md): schema evolution and backward compatibility

### Reference

- [Glossary](glossary.md): key terms used throughout ARCI documentation
