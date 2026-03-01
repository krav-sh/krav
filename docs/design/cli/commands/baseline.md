# Baseline

The baseline command group manages knowledge graph baselines: named references into git history that capture graph state at defined points.

## Synopsis

    arci baseline <subcommand> [options]

## Description

Baselines record which nodes existed, their states, and the relationships between them, anchored to a specific git commit. These commands create, inspect, approve, compare, and verify baselines. Since graph.jsonlt is version-controlled, a baseline stores a commit SHA rather than a full graph snapshot.

## Subcommands

### `create`

Creates a baseline for a module subtree at the current git commit.

    arci baseline create --module <module-id> --title <title>

**Flags:**

- `--module <module-id>`: Root module for the baseline (required).
- `--title <title>`: Human-readable title (required).
- `--phase <phase>`: Lifecycle phase this baseline captures (optional).
- `--scope <scope>`: Coverage scope, either `subtree` (default) or `module-only`.
- `--auto-approve`: Create the baseline in `approved` status immediately.
- `--approved-by <name>`: Who approved this baseline (used with `--auto-approve`).
- `--description <text>`: Reason for creating this baseline.

### `list`

Lists baselines, optionally filtered by module or phase.

    arci baseline list [options]

**Flags:**

- `--module <module-id>`: Filter to baselines rooted at this module.
- `--phase <phase>`: Filter to baselines for this lifecycle phase.

### `show`

Displays detailed information about a specific baseline, including metadata, statistics, and status.

    arci baseline show <baseline-id>

### `approve`

Approves a draft baseline, setting the approvedBy and approvedAt fields.

    arci baseline approve <baseline-id> --approved-by <name>

**Flags:**

- `--approved-by <name>`: Who approved this baseline (required).

### `diff`

Produces a semantic diff between two baselines, or between a baseline and the current graph state. The diff covers node changes, relationship changes, phase changes, coverage changes, and statistics deltas.

    arci baseline diff <baseline-a> [baseline-b]

When the user provides only one baseline, the command compares it against the current graph state.

### `verify`

Checks baseline integrity by confirming that the commit SHA is reachable and the stored statistics match the reconstructed graph state.

    arci baseline verify <baseline-id>

## See also

- [Baselines](../../graph/nodes/baselines.md): Graph node documentation for baselines
- [Lifecycle coordination](../../graph/lifecycle-coordination.md): Phase gate integration with baselines
