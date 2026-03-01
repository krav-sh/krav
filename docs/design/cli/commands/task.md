# Task

The task command group manages tasks, atomic units of work with verifiable deliverables that form a directed acyclic graph (DAG) via dependency relationships.

## Synopsis

    krav task <subcommand> [options]

## Description

Tasks break complex work into manageable units that Claude Code executes in focused sessions. These commands cover CRUD operations, dependency management, DAG queries, execution lifecycle, and deliverable tracking. Tasks belong to a process phase (architecture through validation) and form a flat DAG where the "plan" for any scope is the subgraph reachable from a target.

## Subcommands

### Core operations

#### `create`

Creates a task node.

    krav task create --module <module-id> --title <title> \
      --phase <phase> --task-type <type>

**Flags:**

- `--module <module-id>`: Module this task belongs to (required).
- `--title <title>`: Human-readable title (required).
- `--phase <phase>`: Process phase: `architecture`, `design`, `implementation`, `integration`, `verification`, or `validation` (required).
- `--task-type <type>`: Type within phase such as `feature-implementation` or `code-review` (required).
- `--template <template-name>`: Create from a task template.
- `--param <key=value>`: Template parameter (repeatable, used with `--template`).
- `--priority <priority>`: Priority level: `high`, `medium`, or `low`.
- `--description <text>`: Brief description.
- `--tags <tags>`: Comma-separated tags.

#### `show`

Displays detailed information about a task, including metadata, dependencies, deliverables, and execution state.

    krav task show <task-id>

#### `list`

Lists tasks, optionally filtered by module, phase, or status.

    krav task list [options]

**Flags:**

- `--module <module-id>`: Filter to tasks for this module.
- `--phase <phase>`: Filter by process phase.
- `--status <status>`: Filter by lifecycle state (pending, ready, in_progress, blocked, complete, cancelled).
- `--include-descendants`: Include tasks from child modules.

#### `update`

Modifies task fields.

    krav task update <task-id> [flags]

**Flags:**

- `--status <status>`: Change lifecycle state.
- `--title <title>`: Change title.
- `--priority <priority>`: Change priority level.
- `--description <text>`: Change description.
- `--tags <tags>`: Replace tags.

#### `delete`

Removes a task node from the graph.

    krav task delete <task-id>

### Dependencies

#### `depend`

Adds a dependency relationship between tasks.

    krav task depend <task-id> --on <dependency-id>

**Flags:**

- `--on <dependency-id>`: Task that must complete before this one (required).

#### `undepend`

Removes a dependency relationship between tasks.

    krav task undepend <task-id> --on <dependency-id>

**Flags:**

- `--on <dependency-id>`: Task to remove from dependencies (required).

### Graph queries

#### `ancestors`

Lists all tasks that a given task depends on (transitively).

    krav task ancestors <task-id>

#### `descendants`

Lists all tasks that depend on a given task (transitively).

    krav task descendants <task-id>

#### `blocking`

Lists incomplete ancestor tasks that block a given task.

    krav task blocking <task-id>

#### `ready`

Lists tasks with no incomplete dependencies, available to start.

    krav task ready

#### `critical-path`

Computes the longest dependency path to a target task.

    krav task critical-path <task-id>

### Execution

#### `start`

Marks a task as in progress.

    krav task start <task-id>

#### `complete`

Marks a task as complete.

    krav task complete <task-id>

#### `block`

Marks a task as blocked with a reason.

    krav task block <task-id> --reason <reason>

**Flags:**

- `--reason <reason>`: What is blocking this task (required).

#### `cancel`

Cancels a task with a reason.

    krav task cancel <task-id> --reason <reason>

**Flags:**

- `--reason <reason>`: Reason for cancelling this task (required).

### Deliverables

#### `deliverable`

Records a deliverable produced by a task.

    krav task deliverable <task-id> --kind <kind> [flags]

**Flags:**

- `--kind <kind>`: Deliverable type: `document`, `diagram`, `commit`, `file`, `test-results`, `findings`, `artifact`, or `external` (required).
- `--sha <sha>`: Git commit SHA (for `commit` kind).
- `--path <path>`: File path (for `document`, `diagram`, `file` kinds).
- `--url <url>`: External URL (for `external` kind).

#### `deliverables`

Lists all deliverables recorded for a task.

    krav task deliverables <task-id>

### Context

#### `context`

Loads task context for a Claude Code session. See `krav context` for the top-level context command.

    krav context <task-id>

## See also

- [Tasks](../../graph/nodes/tasks.md): Graph node documentation for tasks
- [Lifecycle coordination](../../graph/lifecycle-coordination.md): Phase constraints on task creation
- [Templating](../../execution/templating.md): Task template system
