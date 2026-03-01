# Architecture

This document describes the ARCI architecture following the [arc42](https://arc42.org/) template. Each section links to detail docs for depth.

## 1. Introduction and goals

ARCI is a tool for Claude Code that structures software development around an INCOSE-inspired knowledge graph, enforces process integrity through declarative hook policies, and encodes development methods as agent skills and subagents. The knowledge graph is the source of truth for what the team must build and why. Hook policies control what the agent can do and ensure it follows the process. Skills and subagents are how the agent actually does the work.

### Knowledge graph

The knowledge graph stores typed nodes (CON, MOD, NEED, REQ, TC, TASK, DEF, BSL) connected by semantic predicates. A formal transformation chain runs concept → need → requirement → task → deliverable. Phase-gated execution constrains module advancement. Suspect link propagation surfaces downstream impacts of upstream changes.

See [Graph overview](graph/index.md) for the full graph documentation.

### Hook policies

The hook system intercepts Claude Code lifecycle events through declarative policies. Policies match events, check conditions, and produce admission decisions (allow, deny, warn) or mutations. A six-stage evaluation pipeline processes policies efficiently; fail-open semantics ensure hooks never block the assistant unexpectedly. Hooks also inject graph context into the agent's conversation and steer the agent toward correct next steps when blocking an action.

See [Hooks overview](hooks/index.md) for the full hook documentation.

### Skills and subagents

Skills encode ARCI's development method as structured instructions that Claude Code follows during execution. Each skill maps to a workflow (formalization, review, task execution) and combines preprocessing commands that inject graph context at load time with instructed commands the agent runs during execution. Subagents provide isolated execution contexts for workflows that benefit from a fresh context window, restricted tool access, or a different posture (a review subagent shouldn't carry the build context). The project stores skills and subagents as files (`.claude/skills/`, `.claude/agents/`) and bundles them with the ARCI plugin.

See [Workflows](workflows/index.md) for the agent interaction layer documentation.

### Quality goals

| Priority | Goal | Description |
|----------|------|-------------|
| 1 | Safety | Fail-open semantics; errors never block Claude Code |
| 2 | Traceability | Unbroken derivation chain from concept through need to requirement to task to deliverable |
| 3 | Queryability | Project questions reduce to graph queries, not document searches |
| 4 | Low latency | Hook evaluation completes within agent-invisible time budgets |
| 5 | Extensibility | Extension system for policies, custom functions, and effect handlers |

### Stakeholders

| Stakeholder | Role | Concern |
|-------------|------|---------|
| Developer using Claude Code | Primary user | Policies don't block workflow; graph provides useful context |
| Policy author | Writes hook policies | Policies are debuggable, testable, and composable |
| Team lead / security | Governs agent behavior | Deny-wins aggregation, audit trail, managed config |
| Extension author | Distributes reusable policies | Extension packaging, trust model, capability tiers |
| ARCI contributor | Develops ARCI itself | Clean architecture, testability, clear boundaries |

## 2. Architecture constraints

### Integration constraints

- **Claude Code hook protocol.** stdin/stdout JSON with exit codes; ARCI cannot change this interface.
- **Claude Code skills and subagents.** Markdown files with YAML frontmatter, following the [Agent Skills](https://agentskills.io) open standard with Claude Code extensions. ARCI's skills and subagents must conform to Claude Code's loading model, frontmatter schema, and execution semantics.
- **MCP protocol.** stdio-based Model Context Protocol for diagnostics and graph queries.
- **Single-binary distribution.** No external runtime dependencies for core operation.
- **No external dependencies for state storage.** Embedded relational store, no separate database process.

### Standards alignment

- **INCOSE NRM.** Needs/requirements distinction (expectation vs. obligation).
- **ISO/IEC/IEEE 15288.** Lifecycle process phases for task organization and phase gating.
- **W3C RDF/JSON-LD.** Graph serialization format.
- **PROV-O.** Provenance tracking for graph modifications.

### Design constraints

- **Fail-open non-negotiable.** Errors in policy evaluation always result in allow.
- **Deny-wins aggregation.** When two or more policies match, the most restrictive action wins.
- **Pre-1.0 instability.** Schema and API may change; see [Versioning](versioning.md).

## 3. System scope and context

### Business context

| Actor | Interaction with ARCI | Direction |
|-------|----------------------|-----------|
| Developer | Authors policies, manages modules and graph nodes, reviews traceability, runs diagnostics | Bidirectional |
| Claude Code | Sends hook events for policy evaluation, issues graph queries, executes skill-driven workflows | Bidirectional |
| Git | Provides branch context; stores graph and policies in version control | ARCI reads and writes |
| GitHub | CI/CD integration, repository context, OSS contributor intake | ARCI reads |
| Parameter providers | Supply policy parameters at evaluation time | ARCI reads |

### Technical context

| Interface | Protocol / format | Purpose |
|-----------|------------------|---------|
| Hook evaluation | stdin/stdout JSON + exit codes | Claude Code → ARCI policy evaluation |
| Skills | Markdown + YAML frontmatter (Agent Skills standard) | ARCI workflow instructions for Claude Code |
| Subagents | Markdown + YAML frontmatter | Isolated agent contexts with preloaded skills |
| Command-line tool | Terminal invocation | Developer and agent → ARCI graph management, diagnostics |
| Server API | HTTP REST + WebSocket on localhost | Command-line delegation, live events, dashboard |
| MCP server | stdio (Model Context Protocol) | Claude Code → ARCI diagnostics and graph queries |
| State store | Embedded relational DB | Session-scoped and project-scoped persistent state |
| Knowledge graph | JSON-LD compact form (JSONLT) | Typed nodes with embedded relationships |
| Configuration | Cascading config files | Policies, settings, managed config layers |

![System context diagram](diagrams/arci-c4-context.puml)

## 4. Solution strategy

Key architectural decisions with rationale:

1. **Shared infrastructure.** The knowledge graph, hook policies, and agent interaction layer are separate concerns that share the command-line tool, server, state store, and config cascade. The CLI is the common interface: hooks call it to check graph state, skills instruct the agent to call it for graph mutations, and the developer calls it directly for management and diagnostics.

2. **Single binary, three modes.** Command-line direct execution, server delegation, and dashboard all share one binary. No version skew, simple distribution.

3. **Knowledge graph over document management.** The graph stores requirements, test cases, and tasks as typed nodes with semantic relationships. Views (graph queries) replace documents (SRS, test plans, traceability matrices).

4. **Method encoded as skills and subagents.** The INCOSE-inspired development process isn't just enforced by hooks; it's taught to the agent through skills that contain step-by-step workflow instructions and subagents that provide isolated execution contexts. Skills use `!`command`` preprocessing to inject graph context at load time and instructed commands for interactive graph queries during execution. Subagents preload relevant skills and operate with restricted tool access. See [Workflows](workflows/index.md).

5. **Formal transformation chain.** Concept → need → requirement → task with unbroken `derivesFrom` edges. Each transformation has preconditions and postconditions. See [Transformations](graph/transformations.md).

6. **Phase-gated execution.** Modules track lifecycle phases independently; phase gates enforce quality criteria per module. See [Lifecycle coordination](graph/lifecycle-coordination.md).

7. **Six-stage evaluation pipeline.** Progressive narrowing: structural match → conditions → parameters → variables → rules → effects. Unidirectional data flow with no cycles. See [Execution model](hooks/execution-model.md).

8. **Cascading configuration.** Seven precedence layers from built-in defaults through managed/required. Same model for settings, policies, and state. See [Configuration](configuration/configuration.md).

9. **Fail-open by default.** Errors never block Claude Code. Only an explicit deny from a policy that evaluated without error blocks operations.

## 5. Building block view

### Level 2: Containers

**ARCI command-line tool.** Unified entry point for hook evaluation, graph management, developer commands, and server/dashboard/MCP control. Operates in direct execution mode (loads config, evaluates locally) or server delegation mode (forwards to server API). See [command-line tool](cli/index.md).

**Server.** Long-running process with config cache, compiled policies, connection pooling, REST API, WebSocket events, and hot-reload. Amortizes expensive operations (config loading, policy compilation, parameter resolution) across many requests. See [Server](server/index.md).

**Dashboard.** Web diagnostics interface for live event streaming, policy testing, state browsing, coverage reports, and graph browsing. Reads from the server's cached state for consistency. See [Dashboard](dashboard/index.md).

**MCP Server.** Exposes policy diagnostics and graph queries to Claude Code via the Model Context Protocol. Delegates to the server API.

**State Store.** Embedded relational store with session-scoped and project-scoped persistent state for hook evaluations and graph operations. See [State store](state-store.md).

**Knowledge Graph.** JSON-LD graph file (`graph.jsonlt`) containing all typed nodes with embedded relationships. Single source of truth for structured metadata. Inline prose uses the `summary` field; extended prose lives in `.arci/` markdown files at convention-derived paths (no frontmatter). See [Graph overview](graph/index.md).

**Policies + Settings.** Configuration across cascade layers (built-in defaults, user, project, local, managed/recommended, managed/required, command-line flags). See [Configuration](configuration/configuration.md).

**Skills.** Markdown files (`.claude/skills/` in the project, `~/.claude/skills/` personal, or bundled with the ARCI plugin) that encode ARCI's development workflows as structured instructions for Claude Code. Each skill has YAML frontmatter (tool restrictions, invocation control, hooks, model overrides) and a markdown body with the workflow steps. Skills use `!`command`` preprocessing to inject live graph data at load time (task context, module requirements, domain context) and instructed commands that tell the agent to run `arci` CLI commands during execution. Claude Code loads skill descriptions into context at session start; full skill content loads on invocation. See [Workflows](workflows/index.md).

**Subagents.** Markdown files (`.claude/agents/` in the project, `~/.claude/agents/` personal, or bundled with the ARCI plugin) that define specialized agents with their own context window, system prompt, tool access, and optionally their own model. Subagents preload specified skills at startup (full content, not just descriptions) and can define lifecycle-scoped hooks in their frontmatter. ARCI uses subagents for workflows that benefit from isolation: code review, security audit, focused analysis. The subagent's restricted tool access and fresh context window prevent the reviewing agent from carrying build context. See [Workflows](workflows/index.md).

**CLAUDE.md and hooks configuration.** CLAUDE.md provides session-level instructions and project context that Claude Code reads at startup. The hooks configuration (`.claude/hooks.json` or managed settings) registers ARCI's hook handlers for Claude Code lifecycle events. Together with skills and subagents, these files form the agent interaction layer that sits between the developer's intent and ARCI's CLI/graph operations.

![Container diagram](diagrams/arci-c4-container.puml)

### Level 3: components (within the ARCI binary)

#### Domain logic

**Policy Engine.** Policy compilation and six-stage evaluation pipeline (structural match, conditions, parameters, variables, rules, effects). Action resolution with deny-wins aggregation. See [Execution model](hooks/execution-model.md).

**Graph Engine.** Eight typed nodes (CON, MOD, NEED, REQ, TC, TASK, DEF, BSL). Graph algorithms, traversal, validation, lifecycle state machines, phase constraints, and transformation chain enforcement. See [Schema](graph/schema.md).

**Template Engine.** Task and decomposition template resolution with context interpolation, inheritance, and parameter handling. DAG construction from patterns. See [Templating](execution/templating.md).

#### Persistence and integration

**Config Loader.** Discovery, loading, parsing, precedence merging, and materialization of domain objects. See [Configuration](configuration/configuration.md).

**Graph Store.** JSONLT persistence (append, compact, load) and JSON-LD serialization to/from typed nodes. Prose file handling.

**State Manager.** Embedded relational store with connection pooling, session/project scoping. Hook effects (setState, counters) use this component.

**Parameter Resolver.** Resolves policy parameters from named providers, inline definitions, env vars, and static values. Cacheable with TTL.

**Git Context.** Reads current branch, dirty state, and staged files. Provides context for policy conditions and graph operations.

#### Server API

**REST API.** Evaluation endpoint, state queries, configuration status, and graph queries.

**Event Stream.** WebSocket endpoint for live hook evaluations, policy matches, and state changes.

**Config Cache.** In-memory cache of compiled policies and materialized graph. File watcher triggers atomic hot-reload.

**Metrics.** Policy match counts, evaluation timing, and validation results. Exposed via API and dashboard.

![Component diagram](diagrams/arci-c4-component.puml)

### Domain model: Knowledge graph

Nine node types form the knowledge graph, organized by semantic category:

| Category | Node types | Purpose |
|----------|-----------|---------|
| Intent | CON (Concept), NEED (Need) | Capture exploration and stakeholder expectations |
| Structure | MOD (Module) | Architectural containers with hierarchy and phase tracking |
| Obligation | REQ (Requirement) | Design obligations with verification criteria |
| Evidence | TC (Test Case) | Verification case specifications |
| Execution | TASK (Task) | Atomic work units in a dependency DAG |
| Quality | DEF (Defect) | Identified problems with disposition and resolution |
| Configuration | BSL (Baseline) | Named graph state snapshots anchored to commits |

The formal transformation chain: concept → need → requirement → task → deliverable. Each step produces `derivesFrom` edges maintaining full traceability. Test cases link to requirements via `verifiedBy`. Defects link to any node via `subject` and generate remediation tasks.

![Knowledge graph domain model](diagrams/arci-knowledge-graph.puml)

See also [Schema](graph/schema.md) · [Predicates](graph/predicates.md) · [Constraints](graph/constraints.md)

## 6. Runtime view

### Scenario 1: intent capture and formalization

1. Developer (or agent) explores a problem area, creates a concept (CON-\*) in `draft` state
2. Concept progresses through `exploring` as the team evaluates options and decisions crystallize
3. Concept reaches `crystallized`, meaning thinking is complete and ready for formalization
4. Formalization extracts stakeholder expectations as needs (NEED-\*), each with `derivesFrom` edges back to the concept
5. Different stakeholders yield different needs from the same concept (1:N relationship)
6. Each need belongs to a module (MOD-\*) via the `module` property
7. Concept transitions to `formalized` and becomes reference material

See [Concepts](graph/nodes/concepts.md), [Needs](graph/nodes/needs.md), [Transformations](graph/transformations.md).

### Scenario 2: requirements derivation and flow-down

1. Stakeholders confirm needs as real expectations → `validated` status
2. Derivation transforms each need into one or more requirements (REQ-\*) with `derivesFrom` edges
3. Requirements are more specific and constrained than needs, stated as binding obligations that must be verifiable
4. Requirements link to modules; parent requirements flow down to child modules via `allocatesTo`
5. Test cases (TC-\*) link to requirements via `verifiedBy`, specifying what to verify, which method, and acceptance criteria
6. Requirements progress through `draft` → `approved`, at which point they become binding obligations

See [Requirements](graph/nodes/requirements.md), [Modules](graph/nodes/modules.md), [Transformations](graph/transformations.md).

### Scenario 3: Building via phase-gated task execution

1. Requirements decompose into tasks (TASK-\*) in a DAG with `dependsOn` edges
2. Each task has a `processPhase` (architecture, design, coding, integration, verification, validation) aligning with the module's lifecycle
3. Tasks become `ready` when all dependencies complete; agents or developers execute ready tasks
4. Task completion produces deliverables (commits, files, test results, documents)
5. When all tasks for a phase complete and no blocking defects (critical/major) remain open, the module can advance to the next phase
6. Each module's phase is independent. Cross-module coordination uses task dependencies and baseline policies.
7. "Plans" are not stored entities. They are graph queries over the task DAG for a given module scope.

See [Tasks](graph/nodes/tasks.md), [Modules](graph/nodes/modules.md), [Lifecycle coordination](graph/lifecycle-coordination.md).

### Scenario 4: verification, defects, and upward propagation

1. Verification tasks execute test cases; `currentResult` on TC-\* nodes updates to pass/fail/skip
2. Reviews (architecture-review, design-review, code-review) produce defects (DEF-\*) as deliverables
3. Each defect has `subject` (what's wrong) and `detectedBy` (which review found it); defects `generate` remediation tasks
4. Defect lifecycle: `open` → `confirmed` → `resolved` (fix complete) → `verified` (fix confirmed) → `closed`
5. Blocking defects (critical/major, open/confirmed) prevent module phase advancement
6. When all test cases for a requirement pass, the requirement can transition to `verified`
7. When all requirements derived from a need reach `verified`, the need can transition to `satisfied`
8. When upstream nodes change, ARCI marks downstream traceability links (`derivesFrom`, `verifiedBy`, `allocatesTo`) as suspect for reviewer triage
9. Baselines (BSL-\*) freeze the graph state at milestones, anchored to git commits; semantic diff between baselines produces structured changelogs

See [Verifications](graph/nodes/test-cases.md), [Defects](graph/nodes/defects.md), [Baselines](graph/nodes/baselines.md), [Lifecycle coordination](graph/lifecycle-coordination.md).

### Scenario 5: skill-driven task execution

1. Developer says "work on TASK-42" → Claude Code recognizes this as a task execution request and invokes the `arci:task` skill with `TASK-42` as the argument
2. Skill preprocessing runs `` !`arci taskcontext TASK-42` ``, which queries the graph and injects the task's requirements, deliverables, dependencies, module domain context, and current status into the skill content before Claude sees it
3. Claude receives the rendered skill with full task context and begins following the workflow instructions
4. Skill instructions tell Claude to build the task's requirements, running `arci` CLI commands during execution to record deliverables (`arci task update TASK-42 --add-deliverable src/parser.ts`) and check coverage (`arci reqcoverage --module MOD-3`)
5. PostToolUse hooks fire after each `arci` CLI command, injecting updated graph state as `additionalContext` ("deliverable recorded, 2 of 4 requirements now have deliverables")
6. If Claude tries to write to a baselined file, a PreToolUse hook denies the write and steers the agent: "this module is baselined; create a defect or run `arci moduleunlock` with justification"
7. When Claude finishes, a TaskCompleted hook checks whether all required deliverables exist and graph mutations are complete; if not, it blocks with a `reason` that tells Claude exactly what's missing
8. For review tasks, the skill delegates to a review subagent with `context: fork` and `agent: arci-review`. The subagent starts with the `arci:review` skill preloaded (full content, including preprocessed requirements data) and operates with read-only tool access

See [Workflows](workflows/index.md).

### Scenario 6: hook evaluation (direct execution)

1. Claude Code fires hook event → invokes ARCI with JSON on stdin
2. Config Loader discovers and loads configuration, materializes policies
3. Policy Engine compiles policies, resolves parameters
4. Six-stage evaluation pipeline runs (structural match → conditions → parameters → variables → rules → effects)
5. Action resolution aggregates results (deny > warn > audit > allow)
6. ARCI persists state mutations and writes JSON response to stdout

See [Execution model](hooks/execution-model.md).

### Scenario 7: hook evaluation (server delegation)

1. ARCI detects server enabled and delegates to the server's evaluate endpoint
2. Server uses cached compiled policies and pooled connections
3. Core evaluates, server returns JSON response
4. The command-line tool forwards response to stdout; latency drops because the server avoids per-invocation config loading and policy compilation

See [Server](server/index.md).

## 7. Deployment view

**Single-binary deployment.** One binary serves all roles: command-line tool, server, dashboard, and MCP server. No version skew between components.

**Installation methods.** See [Installation](installation.md) for supported installation workflows.

**Server options.** Manual foreground, system service, auto-start on unavailable, container deployment. See [Server](server/index.md).

**Platform considerations.** Configuration and state directories follow platform conventions.

**Enterprise deployment.** Managed configuration via MDM for organization-wide policy distribution. See [Managed config](configuration/config-managed.md).

## 8. Cross-cutting concepts

**Fail-open semantics.** Every error path has a permissive fallback. Evaluation errors, config parse failures, and parameter resolution failures all result in allow.

**Cascading configuration.** Seven-layer precedence chain from built-in defaults through managed/required. See [Config cascade](configuration/config-cascade.md).

**Deny-wins aggregation.** The most restrictive action wins across all matching policies. If any policy denies, the result is deny regardless of other policies.

**Knowledge graph as single source of truth.** All structured metadata lives in the graph. Prose lives in markdown files at derived paths or inline via `summary`. Views (queries) replace documents.

**Formal transformation chain.** Concept → need → requirement → task with unbroken `derivesFrom` edges. Each transformation has preconditions and postconditions. See [Transformations](graph/transformations.md).

**Phase-gated execution.** Modules track lifecycle phases independently; phase gates enforce quality criteria per module (no blocking defects, tasks complete). Cross-module synchronization uses task dependencies and baseline policies. See [Lifecycle coordination](graph/lifecycle-coordination.md).

**Suspect link propagation.** Node changes mark downstream traceability links as suspect for reviewer triage. Suspect links don't auto-generate defects; reviewers clear the flag, create a defect, or update the downstream node.

**Agent interaction through skills, subagents, and hooks.** The agent doesn't interact with the knowledge graph by accident or by reading documentation alone. Skills encode each development workflow as structured instructions with live graph context. Subagents provide isolated execution for workflows like review where context separation matters. Hooks enforce invariants and inject ambient graph state. All three use ARCI CLI commands as the interface to the graph: skills instruct the agent to run them, hooks call them to check preconditions, and hook denial output steers the agent toward the right command when blocking an action. See [Workflows](workflows/index.md).

**Change discipline through layered enforcement.** Hooks, skills, subagents, pre-commit hooks, CI, and the CLI all protect baselined content. Defense in depth rather than a single enforcement point. Hook PreToolUse policies deny writes to baselined files. Skills include instructions to check baseline status before modifying files. Review subagents verify changes against requirements. Each layer catches what the others miss.

**Unidirectional data flow in evaluation.** Parameters → variables → conditions → validate/mutate → effects → state. No cycles in the evaluation pipeline.

## 9. Architecture decisions

Key design decisions live inline across the design docs today. See [Design documentation index](index.md) for the current catalog.

## 10. Quality requirements

| Category | Attribute | Scenario |
|----------|-----------|----------|
| Safety | Fail-open | Any internal error results in allow; Claude Code is never blocked by ARCI failures |
| Performance | Evaluation latency | Hook evaluation within agent-invisible time budget (direct and server modes) |
| Testability | Domain logic | Domain logic testable with plain data, no mocking required |
| Testability | Policy testing | Policy authors can dry-run and test configurations before deployment |
| Reliability | Config resilience | Parse errors skip the file; system continues with remaining config |
| Reliability | Hot reload atomicity | Config reload is atomic; no partial state visible to evaluation |
| Security | Sandbox isolation | Platform-native sandboxing constrains shell actions |
| Security | Extension trust | Extension trust model with tiered capabilities |
| Traceability | Requirements chain | Unbroken `derivesFrom` chain from concept to requirement |
| Traceability | Suspect propagation | Changes to upstream nodes surface downstream for review |
| Extensibility | Extension system | New policies and custom functions without modifying core |

## 11. Risks and technical debt

### Risks

- **Server authentication.** The server API currently has no authentication; localhost binding provides minimal protection.
- **Platform sandbox variation.** Sandbox capabilities vary across platforms.
- **Pre-1.0 instability.** Schema and API may change; see [Versioning](versioning.md).
- **Expression debugging.** Complex expressions may be hard for policy authors to debug.

### Technical debt

- Security model has placeholder sections (threat scenarios, controls, audit logging)
- Versioning guarantees unspecified
- Test plans (TP-\*) not yet designed
- Performance characteristics not yet documented

## 12. Glossary

> See [Glossary](glossary.md) for terms used throughout ARCI documentation.
