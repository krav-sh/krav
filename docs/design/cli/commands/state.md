# State

The state command group provides access to the state store.

## Synopsis

    arci state <subcommand> [options]

## Description

The state store holds key-value data scoped to sessions or projects. These commands let you inspect, change, and clear state entries. By default, commands operate on project-scoped state unless you pass a `--session` flag.

## Subcommands

### List

Shows all entries in the state store.

**Flags:**

- `--session <id>`: Filter to a specific session.
- `--project`: Show project-scoped entries.

### Get

Retrieves a specific entry, showing its value and metadata (creation time, last update, author).

    arci state get <key>

### Set

Sets a state entry.

    arci state set <key> <value>

**Flags:**

- `--session <id>`: Store as session-scoped entry. Uses project scope by default.

### Unset

Removes a specific entry.

    arci state unset <key>

**Flags:**

- `--session <id>`: Remove from session scope. Targets project scope by default.

### Clear

Removes entries in bulk.

**Flags:**

- `--session <id>`: Clear entries for a specific session.
- `--project`: Clear project-scoped entries.
- `--all`: Clear everything.

## See also

- [State Store](../../state-store.md)
