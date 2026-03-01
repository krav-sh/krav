# Task types

This directory contains design specifications for the task types that Krav supports. Each task type defines what kind of work a TASK-\* node represents, what deliverables it produces, what completion criteria apply, and which skill typically creates it.

Task types populate the `taskType` field on TASK-\* nodes. They organize by process phase (from ISO/IEC/IEEE 15288), which populates the `processPhase` field. A task's type and phase together determine what the `krav:task` skill expects: what context to preprocess, what instructions to give the agent, and what the task-completion-gate policy checks at the end.

The decompose skill creates most tasks. The defect skill creates remediation tasks. Developers also create some task types directly via `krav task create --task-type <type>`.

Task types are an extensibility point. A template-processed markdown file with frontmatter defines each type, and the system loads them through a cascade. Krav ships with the built-in types described below, but projects can override them or define new ones. See [extensibility](#extensibility) and the [task type definitions](../../graph/nodes/tasks.md#task-type-definitions) section of the task node doc for the file format, cascade, and template context.

## Phase-independent task types

Some task types can occur at any point in the lifecycle. A module's current phase does not gate them.

| Task type | What it does | Expected deliverables | Created by |
|-----------|-------------|----------------------|------------|
| spike | Timeboxed investigation to reduce uncertainty before committing to an approach. Can happen during any phase: an architecture spike explores component boundaries, a design spike prototypes an API surface, a coding spike tests a library's real-world behavior | Investigation summary with findings and recommendation; may produce prototype code but that code is explicitly not a deliverable | `krav:decompose`, manual |
| remediate-defect | Fixes a specific defect identified during review or verification. The DEF-\* node links to this task via `generates`. The `processPhase` on the task reflects the phase of the work under remediation, not a fixed value: an architecture defect produces updated architecture docs, a design defect produces updated specs, a coding defect produces code fixes | Commits and modified deliverables appropriate to the defect's phase; must address the specific defect's subject | `krav:defect` |

## Architecture phase

Tasks in this phase identify components, boundaries, and interfaces. They typically produce documents and diagrams rather than code. The module must be in the `architecture` phase or later.

| Task type | What it does | Expected deliverables | Created by |
|-----------|-------------|----------------------|------------|
| decompose-module | Breaks a module into child modules with responsibilities, boundaries, and interfaces | Architecture doc, module hierarchy diagram, child MOD-\* nodes | `krav:decompose`, `krav:module-add` |
| define-interface | Defines the contract between two modules or between a module and an external system | Interface spec document, data flow diagram | `krav:decompose` |
| decide-architecture | Evaluates alternatives and records a binding architectural decision with rationale | Decision record (ADR-style prose in concept's content file) | `krav:explore` |
| select-technology | Evaluates technology options against requirements and selects one | Evaluation matrix, decision record, any proof-of-concept code | `krav:decompose`, manual |

## Design phase

Tasks in this phase define APIs, data models, algorithms, and protocols. They bridge architecture decisions and coding. The module must be in the `design` phase or later.

| Task type | What it does | Expected deliverables | Created by |
|-----------|-------------|----------------------|------------|
| design-api | Designs a public or internal API surface: endpoints, signatures, error handling, versioning | API spec document (OpenAPI, protobuf, Go interfaces, etc.) | `krav:decompose` |
| design-data-model | Designs data structures, storage schemas, or serialization formats | Schema definition, entity relationship diagram, migration plan if updating existing schema | `krav:decompose` |
| design-algorithm | Designs a non-trivial algorithm with complexity analysis and edge case handling | Algorithm spec with pseudocode or formal description, complexity bounds, test case outlines | `krav:decompose` |
| design-protocol | Designs a wire protocol, message format, or interaction pattern between components | Protocol spec with message formats, sequence diagrams, error handling | `krav:decompose` |

## Coding phase

Tasks in this phase build the thing. They produce code, configuration, and documentation. The module must be in the `implementation` phase or later.

| Task type                                     | What it does                                                                                                            | Expected deliverables                                                                     | Created by                      |
| --------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------- | ------------------------------- |
| `implement-feature`     | Writes a feature or capability to satisfy one or more requirements                                                  | Commits, source files (created/modified), inline documentation                            | `krav:decompose`                |
| `implement-tests`         | Writes test code for TC-\* specifications. May run tests as a development feedback loop to verify they compile and exercise the right paths, but does not record results on TC-\* nodes. That's verification's job | Commits, test source files, test fixtures                                                 | `krav:decompose`                |
| refactor                       | Restructures existing code without changing behavior to improve quality or enable future work                           | Commits, modified source files; must demonstrate behavior preservation (tests still pass) | `krav:decompose`, `krav:defect` |
| write-documentation | Writes or updates documentation: API docs, user guides, architecture docs, README updates                               | Document files (markdown, generated docs)                                                 | `krav:decompose`, manual        |
| configure                     | Sets up build configuration, CI pipelines, deployment config, environment setup                                         | Config files, CI definitions, scripts                                                     | `krav:decompose`, manual        |
| migrate                         | Migrates data, APIs, or dependencies from one version or format to another                                              | Migration scripts, rollback plan, verification that migration completes successfully      | `krav:decompose`, `krav:defect` |

## Integration phase

Tasks in this phase assemble components and resolve interfaces between them. The module must be in the `integration` phase or later.

| Task type | What it does | Expected deliverables | Created by |
|-----------|-------------|----------------------|------------|
| integrate-components | Wires child modules together, resolves interface contracts, handles cross-cutting concerns | Integration code, integration test results, updated interface documentation if contracts changed | `krav:decompose` |
| integrate-deployment | Sets up deployment pipelines, container definitions, infrastructure-as-code for the module | Deployment config, container definitions, deployment verification results | `krav:decompose` |

## Verification phase

Tasks in this phase test against requirements and review deliverables. The module must be in the `verification` phase or later. Most verification tasks run inside subagents for isolation.

| Task type | What it does | Expected deliverables | Created by |
|-----------|-------------|----------------------|------------|
| review-code | Reviews coding deliverables against requirements. Runs inside `krav-reviewer` subagent | Review report, DEF-\* nodes for problems found, structured assessment of requirement satisfaction | `krav:decompose`, `krav:advance` |
| review-design | Reviews design deliverables against requirements and architecture decisions | Review report, DEF-\* nodes for problems found | `krav:decompose`, `krav:advance` |
| review-architecture | Reviews architecture deliverables against needs and constraints | Review report, DEF-\* nodes for problems found | `krav:decompose`, `krav:advance` |
| audit-security | Reviews code and configuration for security vulnerabilities and compliance issues | Security findings report, DEF-\* nodes for vulnerabilities | `krav:decompose`, manual |
| execute-tests | Runs test cases and records results on TC-\* nodes. Runs inside `krav-verifier` subagent. May run the same test commands as `implement-tests`, but does not modify test code. Its job is formal verification, not authoring | Test results recorded on TC-\* nodes, test execution report | `krav:decompose`, `krav:advance` |
| benchmark-performance | Runs performance benchmarks against requirements with quantitative bounds | Benchmark results, comparison against requirement thresholds, DEF-\* nodes if thresholds exceeded | `krav:decompose` |

## Validation phase

Tasks in this phase confirm that the system satisfies stakeholder needs. The module must be in the `validation` phase or later.

| Task type                               | What it does                                                                                                                                                                   | Expected deliverables                                                    | Created by               |
| --------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------ | ------------------------ |
| accept-user           | Validates that the code satisfies the original stakeholder needs (not just requirements). Walks the derivation chain from needs through requirements to deliverables | Acceptance report with per-need assessment, DEF-\* nodes for unmet needs | `krav:decompose`         |
| demo-stakeholder | Demonstrates the code to stakeholders for feedback. Less formal than acceptance testing                                                                              | Demo notes, stakeholder feedback, any new CON-\* nodes from feedback     | `krav:decompose`, manual |
| release                   | Milestone gate that aggregates all verification and validation results for a module scope. Typically a downstream task that depends on all other tasks                         | Release checklist, BSL-\* baseline creation, changelog                   | `krav:decompose`, manual |

## Task type selection

The `krav:decompose` skill selects task types based on the requirements under decomposition and the module's current phase. A requirement about API behavior produces `design-api` then `implement-feature` then `implement-tests` then `execute-tests` then `review-code` as a typical chain. A requirement about performance adds `benchmark-performance`. A requirement about data storage adds `design-data-model`.

The `krav:defect` skill always creates `remediate-defect` tasks, though depending on the defect's nature the remediation might look like a `refactor` or `migrate` in practice. The task type is `remediate-defect` regardless, because the completion criteria differ: the developer must address and verify the specific defect, not just achieve general code improvement.

Manual task creation via `krav task create --task-type <type>` is always available for cases where the decompose skill didn't anticipate a need or the developer wants to add work outside the normal decomposition flow.

## Completion criteria

Each task type defines expected deliverables that the task-completion-gate policy checks. The gate verifies structural completion (are deliverables recorded on the TASK-\* node?) rather than quality (are the deliverables good?). Quality is the review and verification tasks' job.

Task types in the verification phase have additional completion criteria: review tasks must produce at least one DEF-\* node or an explicit "no defects found" assessment. Test execution tasks must record results on every TC-\* node in scope. The review-completion-gate policy on the subagent enforces these criteria.

## Extensibility

Markdown files with YAML frontmatter define task types, and the system loads them through a cascade. The preceding built-in types ship with Krav and cover the standard ISO/IEC/IEEE 15288 lifecycle. Projects extend this set by adding definition files to their `task-types.d/` directory.

A custom type works the moment its file lands in the cascade. The `krav:task` skill doesn't know or care whether a type ships as a default or comes from the project: it inlines `krav task render` via command preprocessing, which resolves the definition from the cascade and produces rendered instructions. The skill requires no changes and the plugin requires no updates.

The `krav:decompose` skill uses built-in types by default when generating task DAGs from requirements. Custom types participate in decomposition through two paths. The first is explicit: a developer creates a task manually with `krav task create --task-type train-model` and wires it into the DAG. The second is automatic: the decompose skill inlines a command that renders a table of eligible task types for the work under decomposition. The CLI does deterministic filtering first (phase eligibility based on the module's current phase, phase-independent types always included, potentially tag or module property matching), then the LLM selects from the filtered set when deciding what work a requirement implies. A well-named custom type with a clear description in its frontmatter is often enough for the LLM to pick it up without any additional configuration.

Overriding a built-in type replaces it entirely. If a project puts an `implement-feature.md` in its `task-types.d/`, that definition applies to all `implement-feature` tasks in the project. The original built-in version is still accessible via qualified name (`builtin/implement-feature`) but won't get selected during normal resolution. This is useful when a team's coding workflow diverges enough from the default that tweaking the template body isn't sufficient and the frontmatter (expected deliverables, completion criteria) also needs to change.

See the [task type definitions](../../graph/nodes/tasks.md#task-type-definitions) section of the task node doc for the full file format, frontmatter fields, template context contract, and cascade resolution rules.

## Open questions

What deterministic filters `krav task-type list --eligible-for` should support beyond phase gating. Phase eligibility is straightforward (only show types whose `processPhase` matches the module's current phase or later, plus phase-independent types). Filtering by tags, module properties, or requirement characteristics could further narrow the set, but the value depends on how large custom type catalogs get in practice. Starting with phase filtering alone and adding more filters as needed is probably the right approach.

Whether task type definitions should include a `description` field in frontmatter. The decompose skill needs a concise summary of each type's purpose to present in the eligibility table. The type's markdown body is too long to inline for every eligible type. A short `description` (one or two sentences) in frontmatter would serve this purpose and also be useful for `krav task-type list` output.
