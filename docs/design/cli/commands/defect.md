# Defect

The defect command group tracks and manages defects, identified problems that deviate from requirements, standards, or expectations.

## Synopsis

    arci defect <subcommand> [options]

## Description

Defects record what went wrong, the severity, the disposition, and the resolution. These commands cover the full defect lifecycle from creation through disposition, resolution, verification, and closure. Critical and major defects block module phase advancement by default.

## Subcommands

### Core operations

#### `create`

Creates a defect node.

    arci defect create --module <module-id> --severity <severity> \
      --statement <statement> [flags]

**Flags:**

- `--module <module-id>`: Module containing the defect (required).
- `--severity <severity>`: Impact level: `critical`, `major`, `minor`, or `trivial` (required).
- `--statement <statement>`: Full description of the problem (required).
- `--title <title>`: Short description.
- `--subject <node-id>`: The defective item (requirement, module, etc.).
- `--category <category>`: Defect category: `missing`, `incorrect`, `ambiguous`, `inconsistent`, `non-verifiable`, `non-traceable`, `incomplete`, `superfluous`, `non-conformant`, or `regression`.
- `--detected-by <task-id>`: The examination task that found this defect.

#### `show`

Displays detailed information about a defect, including metadata, relationships, and resolution history.

    arci defect show <defect-id>

#### `list`

Lists defects, optionally filtered by module, severity, or status.

    arci defect list [options]

**Flags:**

- `--module <module-id>`: Filter to defects in this module.
- `--severity <severity>`: Filter by severity level.
- `--status <status>`: Filter by lifecycle state.

#### `update`

Modifies defect fields.

    arci defect update <defect-id> [flags]

**Flags:**

- `--severity <severity>`: Change severity level.
- `--title <title>`: Change title.
- `--category <category>`: Change defect category.
- `--statement <statement>`: Change problem description.

#### `delete`

Removes a defect node from the graph.

    arci defect delete <defect-id>

### Disposition

#### `confirm`

Confirms a defect as a real problem requiring remediation. Transitions from `open` to `confirmed`.

    arci defect confirm <defect-id>

#### `reject`

Rejects a defect as not a problem (duplicate, by-design, invalid, or out of scope). Transitions from `open` to `rejected`. The user must provide a rationale.

    arci defect reject <defect-id> --rationale <rationale>

**Flags:**

- `--rationale <rationale>`: Reason for rejecting this defect (required).

#### `defer`

Postpones a confirmed defect with justification. Transitions from `confirmed` to `deferred`.

    arci defect defer <defect-id> --rationale <rationale> --target <target>

**Flags:**

- `--rationale <rationale>`: Reason for deferring this defect (required).
- `--target <target>`: Milestone, phase, or baseline that triggers re-evaluation (required).

### Resolution

#### `generate-task`

Creates a remediation task linked to the defect via the `generates` relationship.

    arci defect generate-task <defect-id>

#### `resolve`

Marks a defect as remediated. Transitions from `confirmed` to `resolved`.

    arci defect resolve <defect-id> --notes <notes>

**Flags:**

- `--notes <notes>`: Description of the fix (required).

#### `verify`

Confirms a fix is adequate after re-examination. Transitions from `resolved` to `verified`.

    arci defect verify <defect-id>

#### `close`

Administrative closure of a verified defect. Transitions from `verified` to `closed`.

    arci defect close <defect-id>

#### `reopen`

Reopens a defect that needs further attention.

    arci defect reopen <defect-id>

### Queries

#### `open`

Lists all open or confirmed defects.

    arci defect open [options]

**Flags:**

- `--module <module-id>`: Filter to a specific module.

#### `blocking`

Lists defects that block module phase advancement (critical and major severity with open/confirmed status).

    arci defect blocking

#### `deferred`

Lists all deferred defects.

    arci defect deferred

#### `by-review`

Lists defects found by a specific review or examination task.

    arci defect by-review <task-id>

#### `by-subject`

Lists defects about a specific node.

    arci defect by-subject <node-id>

#### `by-category`

Lists defects matching a specific category.

    arci defect by-category <category>

#### `summary`

Displays aggregate defect counts grouped by status and severity.

    arci defect summary

## See also

- [Defects](../../graph/nodes/defects.md): Graph node documentation for defects
- [Lifecycle coordination](../../graph/lifecycle-coordination.md): Phase gate enforcement with defects
