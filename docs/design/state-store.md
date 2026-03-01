# State store

ARCI includes a persistent state store for tracking data across hook invocations. Rules that depend on history can use the state store: warning on first occurrence and blocking on third, or tracking cumulative token usage across a session. This document describes the state store design and usage.

## Overview

The state store is a key-value store with metadata tracking. Each entry has a key, a value, timestamps for creation and last update, and author information. The store supports scoping by session or project, allowing both ephemeral session state and persistent project state.

## Storage backend

The primary backend uses SQLite via `modernc.org/sqlite` (a pure-Go driver) with `database/sql` for persistence. SQLite provides durability, atomic operations, and the ability to query state for the dashboard without complex serialization.

The database schema is simple:

```sql
CREATE TABLE state_store (
    session_id TEXT,      -- Empty string for project scope
    key TEXT NOT NULL,
    value BLOB,
    created_at TEXT NOT NULL,
    created_by TEXT,
    updated_at TEXT NOT NULL,
    updated_by TEXT,
    PRIMARY KEY (session_id, key)
);
```

The composite primary key of session_id and key allows the same key to exist in multiple sessions. Project-scoped entries use an empty string for session_id (rather than NULL) to ensure the ON CONFLICT clause works correctly.

An in-memory mock backend is also available for testing.

## Scoping

State entries scope to either a session or a project.

Session-scoped state ties to a specific Claude Code session. When Claude Code starts a new session, that session gets a unique ID. State stored with that session ID stays isolated from other sessions. Session state is appropriate for tracking things like prompt count within this session, tools used in this session, or temporary flags that should reset on new sessions.

Project-scoped state spans all sessions in a project. The session_id defaults to an empty string, and the project directory identifies each entry. Project state is appropriate for persistent data like cumulative statistics, configuration that persists across sessions, or flags that should survive session restarts.

## Interface

The `StateStore` interface defines methods for `Get` (returning a full entry with metadata), `GetValue` (returning just the value), `Set` (storing a value with optional author), `Delete` (removing an entry), `Contains` (checking existence), `Keys` (listing all keys for a session), and `AtomicIncrement` (atomically incrementing a counter). Each method takes a session ID and key as parameters.

A `StateEntry` contains the key, value, creation and update timestamps, and optional author fields for both creation and last update.

Basic operations include `get_value(session_id, key)` to get a value (returning `None` if not found), `get(session_id, key)` to get the full entry with metadata, `set(session_id, key, value, author)` to set or update a value, `delete(session_id, key)` to remove an entry, `contains(session_id, key)` to check existence, and `keys(session_id)` to list all keys for a session.

The `atomic_increment(session_id, key, amount, author)` method atomically increments a counter. If the key doesn't exist, the method initializes it to the amount. If it exists but isn't numeric, the method treats it as 0. A single SQL statement handles this to ensure atomicity.

## Usage in policies

Policies access the state store through the `$session_get(key)` and `$project_get(key)` built-in functions in CEL expressions. These functions return the value for the key, and you can combine them with the `??` operator to provide defaults when not found.

### Reading state

Policies typically read state in policy-level or rule-level variables, making the values available to conditions and validation expressions:

```yaml
version: 1
name: dangerous-command-tracking

match:
  tools: [Bash]

variables:
  - name: command
    expression: 'tool_input.command ?? ""'
  - name: rm_attempts
    expression: '$session_get("rm_rf_count") ?? 0'

conditions:
  - name: is-dangerous-rm
    expression: 'command.matches("rm\\s+-rf")'

rules:
  - name: block-after-warnings
    validate:
      expression: 'rm_attempts < 3'
      message: "Blocked after {{ rm_attempts }} dangerous rm attempts"
      action: deny
```

The `??` operator provides a default value when the key doesn't exist. This pattern keeps expressions clean and avoids null-checking complexity.

### Writing state

Effects with `type: setState` write state. Effects run after the engine determines the admission decision, so state updates reflect the current evaluation regardless of whether the tool call proceeds:

```yaml
rules:
  - name: track-rm-count
    effects:
      - type: setState
        scope: session
        key: rm_rf_count
        value: '{{ rm_attempts + 1 }}'
        when: always
```

The `scope` field is either `session` (isolated per conversation) or `project` (persists across sessions). The `key` and `value` fields support template expressions for computed values.

### Escalating enforcement pattern

A common pattern is warning on initial occurrences and blocking after repeated attempts. This combines state reading in variables with validation and state writing in effects:

```yaml
version: 1
name: escalating-rm-protection

match:
  tools: [Bash]

variables:
  - name: command
    expression: 'tool_input.command ?? ""'
  - name: attempts
    expression: '$session_get("rm_attempts") ?? 0'

conditions:
  - name: is-dangerous-rm
    expression: 'command.matches("rm\\s+-rf")'

rules:
  - name: escalating-block
    validate:
      expression: 'attempts < 3'
      message: "Blocked after {{ attempts }} dangerous rm attempts"
      action: deny
    effects:
      - type: setState
        scope: session
        key: rm_attempts
        value: '{{ attempts + 1 }}'
        when: always
```

The `attempts` variable reads the current count before any rules evaluate. The validation blocks when the count reaches the threshold, and the effect always increments the counter for subsequent evaluations.

For more advanced escalation with different messages at each level, use multiple rules with conditions:

```yaml
rules:
  - name: block-after-three
    validate:
      expression: 'attempts < 3'
      message: "Blocked after {{ attempts }} dangerous rm attempts"
      action: deny
    effects:
      - type: setState
        scope: session
        key: rm_attempts
        value: '{{ attempts + 1 }}'
        when: always

  - name: second-warning
    conditions:
      - expression: 'attempts == 1'
    validate:
      expression: 'false'
      message: "Second dangerous rm command. One more and I'll block."
      action: warn

  - name: first-warning
    conditions:
      - expression: 'attempts == 0'
    validate:
      expression: 'false'
      message: "First dangerous rm command. Be careful."
      action: warn
```

The condition on each warning rule ensures only the appropriate message appears based on the current attempt count.

### Script effects for complex logic

For state logic that you cannot express declaratively, use script effects with Starlark. Script effects run after the admission decision and have access to state functions:

```yaml
version: 1
name: command-pattern-tracking

match:
  tools: [Bash]

variables:
  - name: command
    expression: 'tool_input.command ?? ""'

rules:
  - name: track-command-patterns
    effects:
      - type: script
        language: starlark
        source: |
          # Track command history per branch
          branch = current_branch()
          key = "commands_" + branch

          history = session_get(key)
          if history == None:
              history = []

          # Keep last 50 commands
          history = history[-49:] + [command]
          session_set(key, history)

          # Log warning if repetitive patterns detected
          if len(history) >= 10:
              recent = history[-10:]
              unique = len(set(recent))
              if unique <= 2:
                  log("warning", "Repetitive command pattern: %d unique in last 10" % unique)
        timeout_ms: 5000
        when: always
```

Starlark scripts have access to state functions:

- `session_get(key)` - read session-scoped state (returns None if not found)
- `session_set(key, value)` - write session-scoped state
- `project_get(key)` - read project-scoped state (returns None if not found)
- `project_set(key, value)` - write project-scoped state

Scripts run in a sandbox with no filesystem, network, or environment access unless explicitly granted. They cannot influence the admission decision since they execute after the engine makes that decision.

## Built-in state tracking

ARCI can optionally track common state automatically. This is configurable and includes things like `arci.prompts.count` for total prompts in the session, `arci.tools.total_count` for total tool invocations, `arci.tools.<name>.count` for per-tool invocation counts, `arci.session.started_at` for session start timestamp, and similar metrics.

These built-in state updates happen after rule evaluation, so rules see the state from before the current event. This avoids confusion about whether the count includes the current event.

## Database location

The state database location depends on the configuration. For project-scoped state, the database lives within the project directory structure, typically at `.arci/state.db`.

For user-level state that spans projects, the database lives in the user config directory at `~/.config/arci/state.db`.

The daemon manages database connections using `database/sql`'s built-in connection pooling for performance.

## Cleanup and lifecycle

Session state naturally becomes orphaned when sessions end. The state store does not automatically clean up old session data. A periodic cleanup task or explicit cleanup command can remove state for sessions that haven't been active for a configurable period.

Project state persists indefinitely until someone explicitly deletes it.

## Concurrency

SQLite handles concurrency through its built-in locking. The `database/sql` connection pool manages multiple connections, allowing concurrent read access while serializing writes. The daemon coordinates writes to avoid contention.

The `atomic_increment` operation uses a single `INSERT ON CONFLICT` statement to ensure atomicity even under concurrent access.

## Dashboard integration

The dashboard provides a state browser that shows all entries for the current project and session. Users can view keys, values, metadata, and timestamps. For debugging, users can also set or delete entries manually through the dashboard.
