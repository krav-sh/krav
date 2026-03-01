# Stakeholder

The stakeholder command group manages project stakeholders: the named parties whose expectations drive the project's needs.

## Synopsis

    krav stakeholder <subcommand> [options]

## Description

Stakeholders represent people, roles, organizations, or communities with concerns about the system under construction. Each project defines its own stakeholders during initialization and can add or archive them as the project evolves. Needs reference stakeholders via the `stakeholder` object property, anchoring the top of the traceability chain.

No fixed taxonomy of stakeholder types exists. A solo developer's side project might have two stakeholders, while a regulated enterprise system might have a dozen.

## Subcommands

### `create`

Creates a stakeholder node.

    krav stakeholder create --title <title> [options]

**Flags:**

- `--title <title>`: Human-readable name for this stakeholder (required). Typically a role or group name like "CLI end user" or "compliance office."
- `--description <text>`: Who this stakeholder is and their relationship to the system.
- `--concerns <text>`: What this stakeholder cares about, in their terms.
- `--tags <tags>`: Comma-separated tags.

### `show`

Displays detailed information about a stakeholder, including metadata, concerns, and linked needs.

    krav stakeholder show <stakeholder-id>

### `list`

Lists stakeholders, optionally filtered by status.

    krav stakeholder list [options]

**Flags:**

- `--status <status>`: Filter by lifecycle state (`active` or `archived`).

### `update`

Modifies stakeholder fields.

    krav stakeholder update <stakeholder-id> [flags]

**Flags:**

- `--title <title>`: Change the stakeholder name.
- `--description <text>`: Change the description.
- `--concerns <text>`: Change the concerns.
- `--tags <tags>`: Replace tags.

### `delete`

Removes a stakeholder node from the graph. Fails if any needs still reference this stakeholder, unless the user passes `--force`.

    krav stakeholder delete <stakeholder-id>

**Flags:**

- `--force`: Delete even if needs reference this stakeholder. The command removes the `stakeholder` references on those needs. If a need would have zero stakeholders remaining, the command still rejects the deletion (constraint C-MULTI1).

### `archive`

Transitions a stakeholder to `archived` status. Archived stakeholders remain in the graph and their need references remain valid, but formalization workflows exclude them by default.

    krav stakeholder archive <stakeholder-id>

### `needs`

Lists all needs linked to a stakeholder.

    krav stakeholder needs <stakeholder-id> [options]

**Flags:**

- `--status <status>`: Filter needs by lifecycle state.
- `--module <module-id>`: Filter needs to a specific module.

## See also

- [Stakeholders](../../graph/nodes/stakeholders.md): Graph node documentation for stakeholders
- [Needs](../../graph/nodes/needs.md): Downstream consumers of stakeholder expectations
- [Formalizing concepts into needs](../../workflows/formalizing-concepts.md): The workflow that produces needs from concepts and stakeholders
