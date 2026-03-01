# Krav development discipline

This file encodes Krav's development approach as instructions for Claude Code. It's the Stage 0 substitute for hooks, skills, and enforcement policies that don't exist yet. Follow these instructions when working on Krav.

## Before starting work

Check what's ready. The knowledge graph in `.krav/graph.jsonlt` tracks all work as TASK-* nodes in a dependency DAG. Before picking up work, find tasks whose dependencies are all complete:

```bash
jq -s '
  [.[] | select(."@type" == "Task" and .status == "complete") | ."@id"] as $done |
  .[] | select(."@type" == "Task" and .status != "complete" and .status != "cancelled") |
  select(
    (.dependsOn == null) or
    ([.dependsOn[] | ."@id"] | all(. as $id | $done | index($id)))
  ) | {id: ."@id", title: .title, phase: .processPhase, status: .status}
' .krav/graph.jsonlt
```

If there's a specific task assigned for this session, read its full context: the task node, its linked requirements (follow `implements` edges), and the module it belongs to. Understand what you need to build and why before writing code.

| Rule | Classification | Stage | Replacement |
|------|---------------|-------|-------------|
| Query eligible tasks via jq | Temporary | 1 | `krav task list --status ready` CLI command |
| Read full task context before coding | Permanent | n/a | n/a |

## Shell environment

**ALWAYS use `bash -lc "..."` when running tools installed via Homebrew or any tool manager.** The sandbox environment does not inherit PATH modifications from tool managers in non-login shells. If you run `which vale` or `gh auth status` and get "not found," you forgot the login shell. Do not debug PATH issues, do not reinstall tools, do not pass go. Just use `bash -lc`.

```bash
# WRONG — will fail to find brew-installed tools
vale docs/design/index.md
gh auth status
rumdl fmt file.md

# RIGHT — always use login shell for installed tools
bash -lc "vale docs/design/index.md"
bash -lc "gh auth status"
bash -lc "rumdl fmt file.md"
```

The `just` recipes handle this internally, so `just lint`, `just format`, etc. work without a login shell wrapper.

| Rule | Classification | Stage | Replacement |
|------|---------------|-------|-------------|
| Use login shell for brew-installed tools | Permanent | n/a | n/a |

## While working

### Graph awareness

The knowledge graph is the source of truth for what the project needs and why. When the work touches requirements, check whether they're verified. When creating new files, note which module owns them. When you discover a problem, it's a defect (DEF-*), not just a `TODO` comment.

| Rule | Classification | Stage | Replacement |
|------|---------------|-------|-------------|
| Graph is source of truth | Permanent | n/a | n/a |
| Check requirement verification status | Temporary | 2 | Hook-based verification reporting |
| Record module ownership for new files | Permanent | n/a | n/a |
| Create DEF-* nodes, not inline comments | Permanent | n/a | n/a |

### Code organization

Krav is a Go project. Source code lives in `cmd/` (entry points) and `internal/` (packages). The design docs at `docs/design/` specify the architecture in detail. Read the relevant design doc before building a subsystem. The graph engine follows a three-layer architecture: core (pure types and validation), I/O (file serialization), and service (query and mutation API).

| Rule | Classification | Stage | Replacement |
|------|---------------|-------|-------------|
| Source in `cmd/` and `internal/` | Permanent | n/a | n/a |
| Read design docs before building | Permanent | n/a | n/a |
| Graph engine three-layer architecture | Permanent | n/a | n/a |

### Documentation conventions

All documentation uses sentence case for headlines. Write in clear technical prose. Avoid bullet lists in favor of flowing paragraphs. Don't use AI-slop language. PlantUML for diagrams, sources in `docs/design/diagrams/`.

| Rule | Classification | Stage | Replacement |
|------|---------------|-------|-------------|
| Sentence case for headlines | Temporary | 1 | Vale rule enforcement |
| Clear technical prose | Permanent | n/a | n/a |
| Avoid bullet lists, use paragraphs | Temporary | 1 | Vale style rule |
| No AI-slop language | Permanent | n/a | n/a |
| PlantUML for diagrams | Permanent | n/a | n/a |

### Linting and formatting

Run `just lint` before committing. Run `just format` to auto-fix formatting issues. The project uses golangci-lint for Go, vale for prose, rumdl for markdown structure, yamllint for YAML, biome for JS/TS, and codespell for typos.

| Rule | Classification | Stage | Replacement |
|------|---------------|-------|-------------|
| Run `just lint` before committing | Temporary | 2 | Pre-commit hook via krav hooks |
| Run `just format` to auto-fix | Temporary | 2 | Pre-commit hook via krav hooks |
| Linter tool selection | Permanent | n/a | n/a |

## When finishing work

### Update the task

When a task is complete, update its status and record its deliverables. Edit the task's entry in `.krav/graph.jsonlt`:

Set `status` to `"complete"`. Set `completed` to the current ISO 8601 timestamp. Add deliverables to the `deliverables` array with the appropriate `kind` discriminator:

```jsonl
{"@context":"context.jsonld","@id":"TASK-E3K8S6V2","@type":"Task","title":"Implement lexer","status":"complete","completed":"2026-03-01T15:30:00Z","deliverables":[{"kind":"commit","sha":"a1b2c3d","message":"Implement lexer tokenization"},{"kind":"file","path":"internal/graph/lexer.go","action":"created"}]}
```

If the task isn't finished, update `status` to `"in_progress"` and add a note to the task's `summary` field describing where things stand, what's done, and what's left.

| Rule | Classification | Stage | Replacement |
|------|---------------|-------|-------------|
| Set status to complete and add timestamp | Temporary | 1 | `krav task complete` CLI command |
| Record deliverables with kind discriminator | Temporary | 1 | `krav task complete --deliverable` CLI command |
| Set in-progress status with summary | Temporary | 1 | `krav task start` CLI command |

### Record defects

If you find a problem during development or review, create a DEF-* node in `graph.jsonlt`. Include `subject` (the node with the problem), `category` (missing, incorrect, ambiguous, inconsistent, etc.), and `severity` (critical, major, minor, trivial). If the defect needs a remediation task, create the TASK-* node and link them with `generates`.

| Rule | Classification | Stage | Replacement |
|------|---------------|-------|-------------|
| Create DEF-* nodes for discovered problems | Permanent | n/a | n/a |
| Include subject, category, severity fields | Permanent | n/a | n/a |
| Link defect to remediation task via generates | Permanent | n/a | n/a |
| Manual graph editing to create defect nodes | Temporary | 1 | `krav defect create` CLI command |

### Commit discipline

Commit messages reference the task ID: `feat(graph): implement JSONLT parser (TASK-E3K8S6V2)`. Use conventional commits (feat, fix, refactor, docs, test, chore). Each commit should be a coherent unit of work, not a dump of everything changed in a session.

| Rule | Classification | Stage | Replacement |
|------|---------------|-------|-------------|
| Reference task ID in commit messages | Temporary | 2 | Git commit hook via krav hooks |
| Use conventional commits format | Temporary | 2 | Commitlint via krav hooks |
| One coherent unit of work per commit | Permanent | n/a | n/a |

## Graph file format

`.krav/graph.jsonlt` stores the knowledge graph as a JSON Lines file where each line is a JSON-LD compact form document representing one node. The file is the single source of truth for all structured data. Prose files have no frontmatter.

When editing `graph.jsonlt`:

Each line must be valid JSON and valid JSON-LD compact form. The `@id` prefix must match the `@type` (CON-* for Concept, TASK-* for Task, etc.). Identifiers use 8-character Crockford Base32 nanoids. Object property values use `{"@id": "TARGET-ID"}` form. Multi-valued properties use arrays. Keep lines sorted by `@id` for easier diffing.

Don't remove or reorder fields in existing nodes without reason. Add new fields at the end of the object. This makes `git diff` output cleaner.

| Rule | Classification | Stage | Replacement |
|------|---------------|-------|-------------|
| Valid JSON per line | Temporary | 1 | Graph engine parser validation |
| Valid JSON-LD compact form | Temporary | 1 | Graph engine JSON-LD validation |
| @id prefix matches @type | Temporary | 1 | Graph engine schema validation |
| 8-character Crockford Base32 nanoids | Temporary | 1 | `krav` CLI nanoid generation |
| Object references use @id form | Temporary | 1 | Graph engine schema validation |
| Multi-valued properties use arrays | Temporary | 1 | Graph engine schema validation |
| Lines sorted by @id | Temporary | 1 | Graph engine formatter |
| Don't reorder fields without reason | Permanent | n/a | n/a |
| Add new fields at end of object | Permanent | n/a | n/a |

## Prose files

Extended prose for nodes lives in flat directories under `.krav/`:

| Node type   | Directory              |
|-------------|------------------------|
| Concept     | `.krav/concepts/`      |
| Module      | `.krav/modules/`       |
| Need        | `.krav/needs/`         |
| Requirement | `.krav/requirements/`  |
| Test case   | `.krav/test-cases/`    |
| Task        | `.krav/tasks/`         |
| Defect      | `.krav/defects/`       |
| Baseline    | `.krav/baselines/`     |
| Stakeholder | `.krav/stakeholders/`  |
| Developer   | `.krav/developers/`    |
| Agent       | `.krav/agents/`        |

Filenames: `{YYYYMMDDHHMMSS}-{NANOID}-{slug}.md`. The nanoid matches the node's `@id`. Not every node needs a prose file; the `summary` field handles a paragraph or two of inline context.

| Rule | Classification | Stage | Replacement |
|------|---------------|-------|-------------|
| Prose in flat directories under `.krav/` | Permanent | n/a | n/a |
| Filename convention with timestamp and nanoid | Temporary | 1 | `krav` CLI file generation |
| Nanoid in filename matches @id | Temporary | 1 | `krav` CLI file generation |
| Not every node needs a prose file | Permanent | n/a | n/a |

## What not to do

Don't modify files owned by a module that has a baseline. If you need to change baselined content, note the need as a defect first.

Don't create circular dependencies in the task DAG. `dependsOn` must form a DAG (no cycles).

Don't skip the transformation chain for important work. If something deserves a requirement, it deserves a need that justifies it and a concept that explains the thinking. Quick fixes and trivial changes can skip ceremony, but design decisions should be traceable.

Don't hand-edit graph.jsonlt without understanding the schema. Read `docs/design/graph/schema.md` and the relevant entity doc before creating or modifying nodes. The ontology has 12 node types, 15 predicates, and specific constraints on each.

| Rule | Classification | Stage | Replacement |
|------|---------------|-------|-------------|
| Don't modify baselined module content | Temporary | 2 | Baseline protection hook policy |
| No circular dependencies in task DAG | Temporary | 1 | Graph engine cycle detection |
| Don't skip transformation chain | Permanent | n/a | n/a |
| Quick fixes can skip ceremony | Permanent | n/a | n/a |
| Don't hand-edit graph without schema knowledge | Temporary | 1 | `krav` CLI with schema-aware validation |

## Reference

The design docs at `docs/design/` are the authoritative specification. Key entry points:

`docs/design/index.md` is the primary reference and table of contents.
`docs/design/graph/index.md` is the knowledge graph entry point.
`docs/design/graph/schema.md` defines the JSON-LD vocabulary and node types.
`docs/design/graph/predicates.md` defines all relationships.
`docs/design/graph/constraints.md` defines structural rules.
`docs/design/workflows/index.md` documents all 21 workflows.
`docs/plans/build-krav-with-krav.md` is the bootstrapping plan.
