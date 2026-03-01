# Need

The need command group manages stakeholder needs, expectations expressed from the stakeholder's perspective that drive requirement derivation.

## Synopsis

    arci need <subcommand> [options]

## Description

Needs capture what stakeholders expect a module to do or be. These commands handle the full need lifecycle from creation through validation and derivation into requirements. Needs sit between concepts (exploration) and requirements (obligation) in the formal transformation chain.

## Subcommands

### `create`

Creates a need node.

    arci need create --module <module-id> --stakeholder <stakeholder> \
      --statement <statement>

**Flags:**

- `--module <module-id>`: Module this need belongs to (required).
- `--stakeholder <stakeholder-id>`: One or more stakeholders who have this expectation. Accepts STK-* identifiers, comma-separated (required).
- `--statement <statement>`: The need statement in stakeholder terms (required).
- `--title <title>`: Human-readable title.
- `--rationale <text>`: Why this need exists.
- `--priority <priority>`: MoSCoW priority: `must`, `should`, `could`, or `wont`.
- `--tags <tags>`: Comma-separated tags.

### `show`

Displays detailed information about a need, including metadata, relationships, and derivation status.

    arci need show <need-id>

### `list`

Lists needs, optionally filtered by module or stakeholder.

    arci need list [options]

**Flags:**

- `--module <module-id>`: Filter to needs in this module.
- `--stakeholder <stakeholder-id>`: Filter to needs linked to this stakeholder.
- `--status <status>`: Filter by lifecycle state.
- `--priority <priority>`: Filter by MoSCoW priority.

### `update`

Modifies need fields.

    arci need update <need-id> [flags]

**Flags:**

- `--priority <priority>`: Change MoSCoW priority.
- `--statement <statement>`: Change need statement.
- `--title <title>`: Change title.
- `--rationale <text>`: Change rationale.
- `--tags <tags>`: Replace tags.

### `delete`

Removes a need node from the graph.

    arci need delete <need-id>

### `transition`

Advances or changes the need's lifecycle state with validation.

    arci need transition <need-id> --to <status>

**Flags:**

- `--to <status>`: Target lifecycle state (required).

### `validate`

Records stakeholder validation of a need. Transitions the need to `validated` status.

    arci need validate <need-id> --evidence <evidence>

**Flags:**

- `--evidence <evidence>`: Evidence of stakeholder validation (required).

### `derive`

Produces verifiable requirements from a validated need. Creates REQ-* records with derivesFrom relationships back to the need.

    arci need derive <need-id>

### `link`

Creates relationships between a need and other nodes.

    arci need link <need-id> [flags]

**Flags:**

- `--derives-from <concept-id>`: Add a derivesFrom relationship to a concept.

### `trace`

Displays the full traceability chain for a need: concept to need to derived requirements.

    arci need trace <need-id>

## See also

- [Needs](../../graph/nodes/needs.md): Graph node documentation for needs
- [Stakeholders](../../graph/nodes/stakeholders.md): Stakeholders referenced by needs
- [Concepts](../../graph/nodes/concepts.md): Upstream source for need formalization
- [Requirements](../../graph/nodes/requirements.md): Downstream derivation targets
