# `tc`

The `tc` command group manages test cases, verification specifications that provide evidence of requirement satisfaction.

## Synopsis

    krav tc <subcommand> [options]

## Description

Test cases define what to check and how, using one of four INCOSE verification methods (inspection, demonstration, test, analysis). These commands handle CRUD operations, requirement linkage, execution result recording, and coverage analysis. Test cases are specifications; the system tracks execution results separately via `currentResult` and `lastRunAt` fields.

## Subcommands

### `create`

Creates a test case node.

    krav tc create --module <module-id> --title <title> --method <method>

**Flags:**

- `--module <module-id>`: Module this test case belongs to (required).
- `--title <title>`: Human-readable title (required).
- `--method <method>`: Verification method: `inspection`, `demonstration`, `test`, or `analysis` (required).
- `--level <level>`: Test case level: `unit`, `integration`, `system`, or `acceptance`.
- `--implementation <path>`: Path to test code or procedure.
- `--acceptance-criteria <criteria>`: Explicit pass/fail criteria.
- `--description <text>`: What this test case verifies.
- `--tags <tags>`: Comma-separated tags.

### `show`

Displays detailed information about a test case, including metadata, linked requirements, and execution results.

    krav tc show <tc-id>

### `list`

Lists test cases, optionally filtered by module or result.

    krav tc list [options]

**Flags:**

- `--module <module-id>`: Filter to test cases in this module.
- `--result <result>`: Filter by current result (pass, fail, skip, unknown).
- `--method <method>`: Filter by verification method.
- `--status <status>`: Filter by specification lifecycle state.

### `update`

Modifies test case fields.

    krav tc update <tc-id> [flags]

**Flags:**

- `--status <status>`: Change specification lifecycle state.
- `--title <title>`: Change title.
- `--method <method>`: Change verification method.
- `--level <level>`: Change test case level.
- `--implementation <path>`: Change test code path.
- `--acceptance-criteria <criteria>`: Change acceptance criteria.
- `--tags <tags>`: Replace tags.

### `delete`

Removes a test case node from the graph.

    krav tc delete <tc-id>

### `link`

Creates a verifies relationship between a test case and a requirement.

    krav tc link <tc-id> --verifies <req-id>

**Flags:**

- `--verifies <req-id>`: Requirement this test case verifies (required).

### `unlink`

Removes a verifies relationship between a test case and a requirement.

    krav tc unlink <tc-id> --verifies <req-id>

**Flags:**

- `--verifies <req-id>`: Requirement to unlink (required).

### `record`

Records an execution result for a test case. Updates `currentResult` and `lastRunAt`.

    krav tc record <tc-id> --result <result>

**Flags:**

- `--result <result>`: Execution result: `pass`, `fail`, or `skip` (required).
- `--duration <ms>`: Execution duration in milliseconds.
- `--details <text>`: Additional execution details.

### `coverage`

Reports verification coverage, the ratio of requirements with at least one passing test case.

    krav tc coverage [options]

**Flags:**

- `--module <module-id>`: Limit to a specific module.

### `untested`

Lists requirements that have no linked test cases.

    krav tc untested

## See also

- [Test cases](../../graph/nodes/test-cases.md): Graph node documentation for test cases
- [Requirements](../../graph/nodes/requirements.md): Requirements verified by test cases
- [Defects](../../graph/nodes/defects.md): Defects created from verification failures
