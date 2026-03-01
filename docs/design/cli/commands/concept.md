# Concept

The concept command group manages concept nodes, which capture explorations of how something could work, including design thinking, architectural options, and decision rationale.

## Synopsis

    arci concept <subcommand> [options]

## Description

Concepts capture exploration and crystallized thinking that informs the project. These commands handle the full concept lifecycle from creation through formalization into needs. Concepts progress through draft, exploring, crystallized, formalized, and superseded states.

## Subcommands

### `create`

Creates a concept node.

    arci concept create --title <title> --type <concept-type>

**Flags:**

- `--title <title>`: Human-readable title (required).
- `--type <concept-type>`: Type of exploration: `architectural`, `operational`, `technical`, `interface`, `process`, or `integration` (required).
- `--module <module-id>`: Module this concept informs (sets the `informs` relationship).
- `--description <text>`: Brief description.
- `--tags <tags>`: Comma-separated tags.

### `show`

Displays detailed information about a concept, including metadata, relationships, and prose file location.

    arci concept show <concept-id>

### `list`

Lists concepts, optionally filtered by status or type.

    arci concept list [options]

**Flags:**

- `--status <status>`: Filter by lifecycle state (draft, exploring, crystallized, formalized, superseded).
- `--type <concept-type>`: Filter by concept type.

### `update`

Modifies concept fields.

    arci concept update <concept-id> [flags]

**Flags:**

- `--status <status>`: Change lifecycle state.
- `--title <title>`: Change title.
- `--type <concept-type>`: Change concept type.
- `--description <text>`: Change description.
- `--tags <tags>`: Replace tags.

### `delete`

Removes a concept node from the graph.

    arci concept delete <concept-id>

### `transition`

Advances or changes the concept's lifecycle state with validation.

    arci concept transition <concept-id> --to <status>

**Flags:**

- `--to <status>`: Target lifecycle state (required).

### `formalize`

Extracts stakeholder expectations from a crystallized concept and creates need nodes. This interactive process identifies stakeholders, extracts expectations, creates NEED-* records with derivesFrom relationships, and transitions the concept to `formalized` state.

    arci concept formalize <concept-id>

### `link`

Creates relationships between a concept and other nodes.

    arci concept link <concept-id> [flags]

**Flags:**

- `--supersedes <concept-id>`: Mark this concept as superseding another.
- `--informs <module-id>`: Set the module this concept informs.
- `--derives-from <concept-id>`: Add a derivesFrom relationship to another concept.

## See also

- [Concepts](../../graph/nodes/concepts.md): Graph node documentation for concepts
- [Needs](../../graph/nodes/needs.md): Needs derived from concept formalization
