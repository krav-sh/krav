# State

The state command group provides access to the state store.

## Synopsis

    arci state <subcommand> [options]

## Description

The state store holds key-value data scoped to sessions or projects. These commands let you inspect, modify, and clear state entries. By default, commands operate on project-scoped state unless a `--session` flag is provided.

## Subcommands

### list

Shows all entries in the state store.

**Flags:**

- `--session <id>` — Filter to a specific session.
- `--project` — Show project-scoped entries.

### get

Retrieves a specific entry, showing its value and metadata (creation time, last update, author).

    arci state get <key>

### set

Sets a state entry.

    arci state set <key> <value>

**Flags:**

- `--session <id>` — Store as session-scoped entry. Uses project scope by default.

### unset

Removes a specific entry.

    arci state unset <key>

**Flags:**

- `--session <id>` — Remove from session scope. Targets project scope by default.

### clear

Removes multiple entries.

**Flags:**

- `--session <id>` — Clear entries for a specific session.
- `--project` — Clear project-scoped entries.
- `--all` — Clear everything.

## See also

- [State Store](../state-store.md)
