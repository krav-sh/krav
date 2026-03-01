# State

The state command group provides access to the state store.

## Synopsis

    arci state <subcommand> [options]

## Description

The state store holds key-value data scoped to sessions or projects. These commands let you inspect, modify, and clear state entries. By default, commands operate on project-scoped state unless the user passes a `--session` flag.

## Subcommands

### List

Shows all entries in the state store.

**Flags:**

- `--session <id>`: filter to a specific session.
- `--project`: show project-scoped entries.

### Get

Retrieves a specific entry, showing its value and metadata (creation time, last update, author).

    arci state get <key>

### Set

Sets a state entry.

    arci state set <key> <value>

**Flags:**

- `--session <id>`: store as session-scoped entry. Uses project scope by default.

### Unset

Removes a specific entry.

    arci state unset <key>

**Flags:**

- `--session <id>`: remove from session scope. Targets project scope by default.

### Clear

Removes multiple entries.

**Flags:**

- `--session <id>`: clear entries for a specific session.
- `--project`: clear project-scoped entries.
- `--all`: clear everything.

## See also

- [State Store](../state-store.md)
