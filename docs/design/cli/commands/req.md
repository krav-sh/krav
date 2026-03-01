# `req`

The `req` command group manages requirements, verifiable design obligations that constrain the system and fulfill stakeholder needs.

## Synopsis

    krav req <subcommand> [options]

## Description

Requirements are formal "shall" statements that the system must satisfy. These commands handle the full requirement lifecycle from creation through verification, plus flow-down allocation to child modules and traceability queries. Every requirement links upward to needs (why it exists) and downward to test cases (how it gets verified).

## Subcommands

### `create`

Creates a requirement node.

    krav req create --module <module-id> \
      --statement <statement> --verification-method <method>

**Flags:**

- `--module <module-id>`: Module this requirement belongs to (required).
- `--statement <statement>`: The requirement statement in "shall" language (required).
- `--verification-method <method>`: How to verify: `inspection`, `demonstration`, `test`, or `analysis` (required).
- `--title <title>`: Human-readable title.
- `--priority <priority>`: MoSCoW priority: `must`, `should`, `could`, or `wont`.
- `--verification-criteria <criteria>`: Explicit pass/fail criteria.
- `--tags <tags>`: Comma-separated tags.

### `show`

Displays detailed information about a requirement, including metadata, relationships, verification status, and traceability.

    krav req show <req-id>

### `list`

Lists requirements, optionally filtered by module or status.

    krav req list [options]

**Flags:**

- `--module <module-id>`: Filter to requirements in this module.
- `--status <status>`: Filter by lifecycle state (draft, proposed, approved, implemented, verified, obsolete).

### `update`

Modifies requirement fields.

    krav req update <req-id> [flags]

**Flags:**

- `--status <status>`: Change lifecycle state.
- `--statement <statement>`: Change requirement statement.
- `--title <title>`: Change title.
- `--priority <priority>`: Change MoSCoW priority.
- `--verification-method <method>`: Change verification method.
- `--verification-criteria <criteria>`: Change verification criteria.
- `--tags <tags>`: Replace tags.

### `delete`

Removes a requirement node from the graph.

    krav req delete <req-id>

### `link`

Creates relationships between a requirement and other nodes.

    krav req link <req-id> [flags]

**Flags:**

- `--derives-from <need-id>`: Add a derivesFrom relationship to a need.
- `--verified-by <tc-id>`: Add a verifiedBy relationship to a test case.

### `derive`

Creates a derived requirement on a child module (flow-down).

    krav req derive <req-id> --to <module-id>

**Flags:**

- `--to <module-id>`: Target child module for the derived requirement (required).

### `allocate`

Allocates a parent module requirement to child modules with budget partitions.

    krav req allocate <req-id> --to <module-id> --budget <budget>

**Flags:**

- `--to <module-id>`: Target child module (required).
- `--budget <budget>`: Budget partition for this allocation (required).

### `trace`

Displays the full traceability chain: concept to need to requirement to test cases.

    krav req trace <req-id>

### `coverage`

Reports verification coverage, the ratio of requirements that have at least one passing test case.

    krav req coverage

### `unverified`

Lists requirements that have no linked test cases.

    krav req unverified

## See also

- [Requirements](../../graph/nodes/requirements.md): Graph node documentation for requirements
- [Needs](../../graph/nodes/needs.md): Upstream derivation source
- [Test cases](../../graph/nodes/test-cases.md): Downstream verification linkage
