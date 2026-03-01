# Knowledge graph data

This directory contains the knowledge graph serialized as per-table NDJSON files. Each file represents one DuckDB table (vertex or edge), sorted deterministically for stable git diffs.

## Vertex tables

One file per node type. Each line is a JSON object representing one node.

| File | Node type | Sort key |
|------|-----------|----------|
| `concepts.ndjson` | CON-\* | `id` |
| `modules.ndjson` | MOD-\* | `id` |
| `needs.ndjson` | NEED-\* | `id` |
| `requirements.ndjson` | REQ-\* | `id` |
| `test_cases.ndjson` | TC-\* | `id` |
| `tasks.ndjson` | TASK-\* | `id` |
| `defects.ndjson` | DEF-\* | `id` |
| `baselines.ndjson` | BSL-\* | `id` |
| `stakeholders.ndjson` | STK-\* | `id` |
| `test_plans.ndjson` | TP-\* | `id` |
| `developers.ndjson` | DEV-\* | `id` |
| `agents.ndjson` | AGT-\* | `id` |

## Edge tables

One file per predicate. Each line has `src` and `dst` columns (plus optional metadata).

| File | Predicate | Sort key |
|------|-----------|----------|
| `child_of.ndjson` | childOf | `src, dst` |
| `derives_from.ndjson` | derivesFrom | `src, dst` |
| `module.ndjson` | module | `src, dst` |
| `stakeholder.ndjson` | stakeholder | `src, dst` |
| `allocates_to.ndjson` | allocatesTo | `src, dst` |
| `depends_on.ndjson` | dependsOn | `src, dst` |
| `verified_by.ndjson` | verifiedBy | `src, dst` |
| `subject.ndjson` | subject | `src, dst` |
| `detected_by.ndjson` | detectedBy | `src, dst` |
| `generates.ndjson` | generates | `src, dst` |
| `informs.ndjson` | informs | `src, dst` |
| `implements.ndjson` | implements | `src, dst` |
| `operator.ndjson` | operator | `src, dst` |
| `parent_agent.ndjson` | parentAgent | `src, dst` |

## Lifecycle

The ARCI server hydrates these files into an in-memory DuckDB instance on startup and dehydrates modified state back to these files on checkpoint, baseline creation, or graceful shutdown. The files are the git-tracked source of truth; DuckDB is the runtime query engine.

## Migration status

The `graph.jsonlt` file in the parent directory contains the current graph data in the legacy monolithic JSONLT format. Migration to per-table NDJSON files is pending. The migration involves splitting the monolithic file into per-table files, converting JSON-LD keys (`@id`, `@type`, `@context`) to plain JSON keys (`id`, `type`), and extracting relationships from inline properties to edge table rows.
