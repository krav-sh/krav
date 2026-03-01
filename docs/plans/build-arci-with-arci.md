# Build ARCI with ARCI

This plan describes how to build ARCI using ARCI's own methodology, staged so that each phase produces tooling that the next phase uses. The classic compiler bootstrapping problem: you can't use ARCI to build ARCI until ARCI exists, but you can follow ARCI's methodology before the tooling enforces it, and progressively replace manual discipline with real enforcement.

## Current state

The project has extensive design documentation covering the full architecture: 12 node types in the knowledge graph ontology, 15 predicates, 21 workflows, a hook system with policy model and evaluation pipeline, skills and subagent integration, CLI command specs, daemon architecture, configuration cascade, and a web dashboard. The ontology has been through vocabulary alignment against OSLC, PROV-O, and Dublin Core.

On the implementation side there's Go scaffolding: a CLI entry point using cobra, a config cascade loader, structured logging, a build metadata package, and a command factory. The `.arci/` directory structure exists but is empty. Tooling exists (vale for prose linting, golangci-lint, pre-commit hooks, biome for JS/TS, justfile for tasks). A Claude plugin skeleton exists at `claude/`.

No graph operations exist. No working CLI commands beyond `arci --help`. No hook engine. No skills or subagents.

## The bootstrapping insight

ARCI is a development methodology backed by tooling. The methodology is available before the tooling exists. The transformation chain (concept → need → requirement → task → deliverable) is a way of thinking about work, not just a set of CLI commands. Following it manually, even clumsily, tests the ontology under real conditions and catches design problems before they're encoded in Go.

The risk of skipping manual practice is that you build ARCI without ever using ARCI's approach, discover the approach has problems, and have to retrofit. The risk of overdoing it is that you spend so long hand-crafting JSON-LD that you never ship code. The plan below tries to thread that needle.

## Stage 0: proto-ARCI

Build ARCI's own knowledge graph by hand, using the design docs as the source of design decisions and the JSON-LD schema as the file format. This stage has no ARCI code dependencies. A text editor, shell scripts, and Claude Code following instructions in CLAUDE.md handle everything.

### What gets built

A `graph.jsonlt` file (JSON Lines of JSON-LD compact form) containing ARCI's own development graph. The root module for ARCI itself. Child modules matching the architecture (graph, hooks, CLI, daemon, skills, dashboard). A developer node for Tony. Stakeholder nodes for the primary user personas from the architecture doc. Concepts capturing key design decisions already made during the design phase. Needs extracted from the most critical concepts. Requirements derived from those needs for the Stage 1 scope only. A task breakdown for Stage 1 work, with dependency edges forming the execution DAG.

The graph doesn't need to be exhaustive. Capture enough structure to plan Stage 1 work and test the ontology under real use, not to model every requirement ARCI has.

### What gets validated

The JSON-LD schema works for real data, not just examples in docs. The identifier scheme (PREFIX-nanoid) works in practice. The `derivesFrom` chain from concept through need to requirement feels natural, not forced. The task DAG structure adequately models the dependency relationships between implementation work. The module hierarchy maps cleanly to the codebase's package structure.

### What gets written

A CLAUDE.md that encodes ARCI development discipline as agent instructions: check the task DAG before starting work, update task status and record deliverables when done, don't modify files owned by baselined modules. This is the degraded-mode version of what hooks and skills eventually enforce.

Shell script approximations of core ARCI operations, installed as Claude Code skills in `.claude/skills/`. A "task context" skill that reads a task's JSON-LD and its linked requirements. A "ready tasks" skill that finds tasks with all dependencies complete. A "graph summary" skill that produces a status overview. These don't need to be production-quality; they need to prove the workflow patterns.

### Exit criteria

The Stage 1 task breakdown exists in `graph.jsonlt` and has guided actual implementation planning. At least one concept → need → requirement → task chain traces end to end using real ARCI content. The CLAUDE.md and shell skills are functional enough that Claude Code can check what's ready and record what's done.

## Stage 1: graph storage and CLI foundation

Build the minimum viable graph engine and enough CLI surface to replace the Stage 0 shell scripts. This is the bootstrap compiler. Once it works, ARCI can manage its own development graph.

### Storage engine

The graph storage layer. Read and write `graph.jsonlt`. Parse and emit JSON-LD compact form. Validate nodes against the schema (type/prefix consistency, required fields, enum values). Generate nanoid identifiers. Resolve prose file paths from node identifiers.

This is the foundation everything else depends on. No graph operations means no CLI, no hooks, no skills.

The design docs specify a three-layer architecture (core, I/O, service) with rustworkx for graph operations. The implementation should follow this, with the core layer being pure Go types and validation, the I/O layer handling file serialization, and the service layer providing the query and mutation API that CLI commands call.

### CLI framework

Extend the existing cobra scaffolding into a working command dispatch system. Global flags (`--format`, `--verbose`, `--project-root`). Output formatting (human-readable, JSON, agent-optimized). Error handling following the design doc's error model. The command factory pattern from the existing scaffolding should scale to the full command set.

### Graph CRUD commands

The CLI commands that create, read, update, and delete graph nodes. These are the operations that replace hand-editing `graph.jsonlt`:

`arci module create/show/list/update`, `arci concept create/show/list/update`, `arci need create/show/list/update`, `arci req create/show/list/update`, `arci task create/show/list/update`, `arci defect create/show/list/update`, `arci tc create/show/list/update`, `arci baseline create/show/list`, `arci stakeholder create/show/list/update`.

Each command group follows the same pattern: create validates inputs and writes to graph.jsonlt, show reads and formats a single node, list queries with filters, update modifies fields. Delete comes later (soft-delete semantics need design work).

Priority ordering: module and task commands first (needed to manage the work), then requirement and need commands (needed for traceability), then everything else.

### Graph query commands

The commands that answer questions about the graph without changing it:

`arci task ready` (tasks with all dependencies complete), `arci task blocking TASK-*` (incomplete ancestors), `arci task ancestors/descendants TASK-*` (DAG traversal), `arci tc coverage` (verification coverage), `arci req trace REQ-*` (derivation chain).

These replace the Stage 0 shell scripts with real graph traversals. Skills and hooks invoke these commands to inject context.

### Exit criteria

`arci task list --status ready` returns the correct set of tasks from ARCI's own `graph.jsonlt`. `arci task show TASK-*` displays a task with its relationships resolved. `arci req trace REQ-*` walks the derivation chain. Skills that call real CLI commands replace the Stage 0 shell skills. ARCI is managing its own task DAG through its own CLI.

## Stage 2: hooks and enforcement

Build the hook engine. This is where ARCI's discipline becomes enforceable rather than advisory.

### Policy engine

Policy loading from the YAML cascade. CEL expression compilation and evaluation. The six-stage evaluation pipeline (match → filter → evaluate → aggregate → decide → effect). Admission decisions (allow, deny, warn) with deny-wins aggregation. Fail-open semantics on errors.

### Claude Code hook integration

The hook handler that receives events from Claude Code on stdin and returns decisions on stdout. Registration via `.claude/settings.json`. Support for the hook events that matter most for graph enforcement: PreToolUse (deny writes to baselined content), PostToolUse (inject graph context after mutations), Stop (enforce deliverable recording), and SessionStart (inject initial graph context).

### Built-in policies

The policies that ship with ARCI for enforcing development discipline. Baseline protection (deny writes to files owned by baselined modules). Task completion gates (block task completion until the agent records deliverables). Graph context injection (load task context at session start). These are the policies that ARCI uses on itself.

### Exit criteria

A PreToolUse hook denies a write to a file under a baselined module's scope. A Stop hook blocks session end when the current task has no recorded deliverables. A SessionStart hook injects the current task's context into Claude Code's conversation. ARCI is enforcing its own development discipline through its own hooks.

## Stage 3: skills and subagents

Package the workflow patterns as proper Claude Code skills and subagents, backed by real CLI commands.

### Core skills

The skills that encode ARCI's graph-aware workflows. Task execution skill (preprocesses `arci context TASK-*`, provides workflow instructions from rendered task type template). Formalization skill (reads concept prose, guides need extraction, calls `arci need create`). Review skill (loads module requirements, guides defect creation). Status skill (synthesizes graph state into a narrative summary).

### Review subagent

An isolated agent with restricted tools (read-only access plus `arci defect create`) that reviews deliverables against requirements. Clean context window, no implementation bias. This is one of ARCI's highest-value features since it separates the "build" and "check" concerns into different agent contexts.

### Task type definitions

The built-in task type definitions (markdown files with YAML frontmatter) that `arci task render` processes into agent prompts. Each task type defines expected deliverables, completion criteria, and a template body with graph context placeholders. The design docs specify the full format and cascade resolution.

### Exit criteria

The task execution skill loads a task's context via preprocessing and guides Claude Code through implementation. The review subagent produces defects from a code review without carrying implementation context. Task type templates render correctly with graph data. ARCI skills drive ARCI's own development workflows.

## Stage 4: Full self-hosting

Everything that makes ARCI a complete system.

### Baselines and semantic diff

Baseline creation anchored to git commits. Graph state reconstruction at historical commits. Semantic diff between baselines producing structured changelogs. Phase gate integration (baselines capture module phase advancement).

### Test plans

The TestPlan (TP-*) entity and its associated workflows. Coordinated test execution scoped to milestones. Execution records with environment binding. Evidence packages for phase gate advancement.

### State store

SQLite-backed key-value store for hook state. Session-scoped and project-scoped state. CEL integration (`$session_get`, `$project_get`). Starlark script effects. Hooks use stored state to escalate enforcement across repeated violations and inject richer context from prior sessions.

### Daemon

Long-running process for fast hook evaluation. Unix socket API. Health monitoring. The daemon eliminates per-invocation startup cost for hooks, which matters when hooks fire on every tool call.

### Dashboard

Web-based diagnostics UI. Graph visualization. Policy testing. Hook activity timeline. Coverage reports. This is the lowest-priority module since the CLI covers all the same information, but it's high-value for understanding the system state at a glance.

### Exit criteria

ARCI manages its own development end to end. Baselines freeze graph state at milestones. The review subagent runs test plans. The daemon keeps hook evaluation fast. The dashboard shows the state of ARCI's own knowledge graph. The bootstrapping is complete.

## Ordering principles

Stage 0 should be light-ceremony. Create enough graph structure to plan Stage 1, not a complete requirements model for the entire system. The design docs are already an incredibly detailed specification; treat them as the concept and need layer and derive implementation requirements from them as needed.

Stage 1 is where most of the time goes. The graph engine and CLI are the foundation. Get them right and everything built on top works. Get them wrong and every later stage fights the foundation.

Stages 2 and 3 can overlap. Hook enforcement and skill authoring are somewhat independent, though skills benefit from hooks for context injection.

Stage 4 modules are independent of each other. Build them in whatever order the project needs demand. The dashboard is last because it's purely a visibility tool; everything it shows is accessible via CLI.

Within each stage, the task DAG in ARCI's own knowledge graph determines execution order. The preceding plan describes what gets built; the graph describes the dependency structure of how it gets built.

## Open questions

How much of Stage 0 is worth the investment before jumping into Go? The answer probably depends on how quickly the ontology questions get resolved. If the graph schema is stable, Stage 0 can be brief. If it's still shifting, Stage 0 provides a cheaper way to test changes than rewriting Go code.

Should the graph engine use rustworkx (via CGo or a subprocess) or a pure Go graph library? The design docs specify rustworkx for its performance characteristics, but the CGo bridge adds build complexity. A pure Go implementation might be simpler for bootstrapping, with rustworkx as a later optimization if needed.

How should the MCP server fit into the staging? The design docs mention it as an integration surface alongside the CLI. It could land in Stage 2 (alongside hook integration) or Stage 4 (as a refinement). If skills work well enough with CLI-via-Bash preprocessing, MCP might not be urgent.

What's the right ceremony level for ARCI's own development? The "implementing without ceremony" workflow exists for exactly this situation. During Stage 1, most work should be low-ceremony (task → implement → verify). The full transformation chain makes more sense for Stage 2+ when ARCI is managing its own requirements.
