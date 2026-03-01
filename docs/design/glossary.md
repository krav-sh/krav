# Glossary

This document defines key terms used throughout Krav's documentation.

## Core concepts

**Hook.** A point in Claude Code's execution flow where external code can intercept and influence behavior. Krav receives hook invocations and evaluates policies against them. See [Hook schema](hooks/hook-schema.md) for the full list of event types.

**Hook event.** A specific type of hook invocation. Examples include pre_tool_call (before a tool executes), post_tool_call (after a tool executes), session_start (when a conversation begins), and pre_prompt (when the user sends a message). See [Hook schema](hooks/hook-schema.md) for event mapping.

**Policy.** A self-contained unit of hook logic in Krav. A policy declares its matching criteria, conditions, parameters, variables, macros, and rules all in one YAML document. See [Policy model](hooks/policy-model.md) for details.

**Rule.** A component within a policy that defines a specific check, mutation, or side effect. Rules contain one of validate (for admission decisions), mutate (for transforming input), or effects (for fire-and-forget actions), plus optional match constraints and conditions.

**Match constraints.** Structural filtering criteria that determine whether a policy or rule applies to a given hook event. Match constraints use OR-within-arrays, AND-across-fields logic and include events, tools, paths, and branches. Match constraints are fast to evaluate because they use structural comparison and indexing rather than CEL expressions.

**Match conditions.** CEL expressions in a `conditions` array that must all return true for a policy or rule to apply. Unlike validation expressions, failing a match condition causes the policy or rule to be silently skipped rather than producing a violation.

**Validate.** A rule component that evaluates a CEL expression and produces an admission decision when the expression returns false. Validate blocks include an expression, a message explaining the failure, and an action (deny, warn, or audit).

**Mutate.** A rule component that transforms the hook event before it proceeds to Claude Code. Mutate blocks contain a CEL expression that receives the current event state as `object` and returns the modified state.

**Effects.** fire-and-forget side actions that run after the evaluator reaches an admission decision. Effects include setState, log, and notify. Effects cannot influence whether the tool call proceeds.

**Condition.** A CEL expression used in a `conditions` array within a policy or rule. All conditions in the array must return true (AND logic with short-circuit evaluation).

## Spec-driven development

**Concept.** a high-level exploration of how something could work. Concepts capture design thinking, architectural options, and crystallized decisions. They are the raw material from which the agent derives needs through formal transformation.

**Module.** a named, addressable architectural container in the spec knowledge graph. Modules represent systems, subsystems, or components under construction. They form a hierarchy, own needs and requirements, and track lifecycle phase.

**Stakeholder.** a named party with expectations about the system under construction. Stakeholders are project-defined STK-* nodes representing actual people, roles, organizations, or communities. No fixed taxonomy of stakeholder types exists; each project defines its own stakeholders during initialization.

**Need.** a stakeholder expectation: what stakeholders need a module to do or be. Each need references one or more stakeholders via the `stakeholder` object property. The agent validates needs against stakeholder intent, and needs serve as the source for deriving requirements.

**Requirement.** a precise, testable design obligation derived from a need. Requirements use "shall" statements and must be verifiable. Test cases verify them.

**Test case.** a verification specification that defines what to check and how. Test cases use one of four methods: test, inspection, demonstration, or analysis. They follow their own lifecycle (draft, specified, built, executable, obsolete), while Krav tracks execution results separately via `currentResult` (pass/fail/skip/unknown). Test cases link to requirements via verifies/verifiedBy relationships.

**Task.** a unit of work to satisfy a requirement. Tasks form a DAG with dependency relationships and follow a process phase (architecture through validation).

**Defect.** An identified problem requiring action, discovered during reviews, testing, or analysis. Defects have a severity (critical, major, minor, trivial) and category (missing, incorrect, ambiguous, etc.). Their lifecycle runs open → confirmed → resolved → verified → closed. Defects may generate remediation tasks and link to the subject node (requirement, test case, etc.) where the reviewer discovered the problem.

**Spec.** A structured specification document containing modules. Specs use INCOSE-inspired systems engineering practices to organize requirements and traceability.

**Knowledge graph.** The interconnected graph of spec modules and their relationships. On disk, per-table NDJSON files under `.krav/graph/` store vertex and edge data in a git-friendly format. At runtime, DuckDB with the DuckPGQ extension hydrates these files into relational tables queryable via SQL and SQL/PGQ. The knowledge graph provides full traceability from concepts through verified implementations. Inline prose uses the `summary` field; extended prose lives in markdown files at convention-derived paths under `.krav/`.

## Configuration

**Project configuration.** Policies and settings that apply to a specific project, stored in the project's `.krav/` directory. Project policies have higher precedence than user-level policies.

**Drop-in directory.** a `policies.d/` directory containing individual YAML files that the loader merges into configuration. Local variants (`policies.local.d/`) are gitignored and take precedence over non-local policies at the same cascade layer.

**Precedence.** the order in which the loader merges configuration sources. Later sources override earlier ones. See [Config cascade](configuration/config-cascade.md) for details.

## Expression and scripting languages

**CEL.** The Common Expression Language, used for policy conditions, validation expressions, and mutation expressions. CEL is a non-Turing-complete expression language designed by Google for security policy evaluation.

**Go template.** The Go standard library `text/template` engine with Sprig functions, used for validation messages, effect values, and dashboard rendering.

**Expression.** Code written in CEL, used in conditions, validate expressions, mutate expressions, and variable definitions.

**Starlark.** An embedded scripting language for complex action logic beyond what CEL can express. See [Starlark scripting](hooks/starlark-scripting.md) for details.

**GritQL.** A query language for structural code analysis using syntax tree patterns. See [GritQL](hooks/gritql.md) for details.

**Custom function.** an extension to CEL for policy evaluation. Policies invoke custom functions with a `$` prefix (such as `$file_exists(path)`, `$current_branch()`).

**Macro.** a reusable CEL expression fragment defined in a policy's `macros` array. Other CEL expressions can call macros with a `$` prefix.

**jsonPointer function.** A custom CEL function for accessing values in untyped JSON using RFC 6901 JSON Pointer syntax.

**Path expression.** CEL's dot notation for traversing nested data structures (such as `tool_input.command`).

## Execution model

**Priority level.** One of four levels (critical, high, medium, low) assigned to policies. Higher-priority policies evaluate first. See [Execution model](hooks/execution-model.md) for details.

**Validation aggregation.** The process of combining validation results from all matching policies. The most restrictive action wins: deny > warn > audit > allow.

**Mutation composition.** The process of combining mutations from multiple policies. Higher priority mutations apply first.

**Fail-open semantics.** The design principle that errors, timeouts, and failures should never block Claude Code operations. Only explicit deny decisions block operations.

**Deny-wins.** The permission resolution strategy where any deny results in a deny decision, regardless of other allow actions.

**Failure policy.** A per-policy setting that determines what happens when the policy errors during evaluation. Default is `allow` (fail open).

## State

**State store.** A persistent key-value store backed by DuckDB for tracking data across hook invocations. The state database at `.krav/state.duckdb` attaches to the same DuckDB instance as the knowledge graph at runtime, so queries can join across both domains. See [State store](state-store.md) for details.

**Session scope.** State tied to a specific Claude Code session.

**Project scope.** State shared across all sessions in a project.

## Architecture

**Functional core.** The pure evaluation engine with no I/O or side effects.

**Imperative shell.** Code that handles real-world concerns like reading files, managing connections, and executing side effects.

**Server.** Long-running process that caches configuration, pools connections, and serves an HTTP API. See [Server](server/index.md) for details.

**Direct execution.** Running `krav hook apply` without a server, where configuration loading happens on every invocation.

## Parameters and variables

**Parameter.** External data brought into policy evaluation. Parameters can come from static values, named providers, or inline providers (file, http, env).

**Variable.** A computed value derived from parameters, the hook event, built-in functions, or other variables.

**Provider.** a source of parameter values. The Krav configuration defines named providers. Inline providers include file, http, and env.

## Extensions

**Extension.** A package that provides policies, custom functions, or other capabilities. Distributed via Go modules, git repositories, or local paths. See [Extensions](extensions.md) for details.

**Policies-only extension.** An extension containing only YAML policy files with no custom code.

**Manifest.** The `extensions.toml` file declaring which extensions to install and their version constraints.

**Lockfile.** The `extensions.lock` file recording exactly what's installed.

## Claude Code terms

**Matcher.** a pattern that filters which tools trigger a hook. An empty matcher matches all tools. Krav's policy match constraints provide more powerful filtering.

**Tool.** An operation Claude Code can perform, like executing shell commands (Bash), writing files (Write), or reading files (Read). See [Hook schema](hooks/hook-schema.md) for canonical tool names.

**Tool input.** The parameters passed to a tool, available in expressions as `tool_input`. See [Hook schema](hooks/hook-schema.md) for tool input schemas.

## Data engine

**DuckDB.** Embedded analytical database engine used for both the knowledge graph runtime and the state store. Distributed as a statically linked library within the Go binary via `github.com/duckdb/duckdb-go/v2` (CGo). No separate database process.

**DuckPGQ.** DuckDB community extension that adds SQL/PGQ (SQL:2023 standard) for property graph queries over relational tables. Enables graph pattern matching, variable-length path traversals, and shortest path queries alongside standard SQL.

**Hydrate/dehydrate.** The process of loading per-table NDJSON files from `.krav/graph/` into DuckDB tables (hydrate) and serializing DuckDB tables back to sorted NDJSON files (dehydrate). The server hydrates on startup and dehydrates on checkpoint, baseline creation, or graceful shutdown.

**NDJSON.** Newline-delimited JSON format, one JSON object per line. Used for git-friendly serialization of knowledge graph tables under `.krav/graph/`. Each vertex table and edge table has its own NDJSON file, sorted deterministically for stable diffs.

**SQL/PGQ.** The property graph query extension to SQL, standardized in SQL:2023. Adds `MATCH` clause syntax for graph pattern matching over relational tables registered as a property graph. DuckPGQ provides this capability.

---

This glossary expands as Krav evolves. Add terms when they appear frequently in documentation or when vocabulary needs disambiguation.
