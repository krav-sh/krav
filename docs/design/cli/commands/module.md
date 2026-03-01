# Module

The module command group manages architectural modules, the containers representing parts of the system hierarchy under construction.

## Synopsis

    arci module <subcommand> [options]

## Description

Modules form a hierarchy representing system decomposition. Each module owns needs, requirements, and tasks, and tracks its current lifecycle phase. These commands handle CRUD operations, hierarchy management, phase advancement, and work scoping.

## Subcommands

### Core operations

#### `create`

Creates a module node.

    arci module create --title <title> --parent <module-id>

**Flags:**

- `--title <title>`: Human-readable title (required).
- `--parent <module-id>`: Parent module in the hierarchy (required for non-root modules).
- `--description <text>`: Brief description.
- `--tags <tags>`: Comma-separated tags.

#### `show`

Displays detailed information about a module, including metadata, phase, hierarchy position, and owned nodes.

    arci module show <module-id>

#### `list`

Lists modules, optionally filtered by parent or phase.

    arci module list [options]

**Flags:**

- `--parent <module-id>`: Filter to children of this module.
- `--phase <phase>`: Filter by current lifecycle phase.

#### `update`

Modifies module fields.

    arci module update <module-id> [flags]

**Flags:**

- `--title <title>`: Change title.
- `--description <text>`: Change description.
- `--tags <tags>`: Replace tags.

#### `delete`

Removes a module node from the graph. The module must have no children.

    arci module delete <module-id>

### Hierarchy

#### `children`

Lists the direct children of a module.

    arci module children <module-id>

#### `tree`

Displays the full module subtree as a tree view.

    arci module tree <module-id>

#### `reparent`

Moves a module to a different parent in the hierarchy.

    arci module reparent <module-id> --to <new-parent-id>

**Flags:**

- `--to <new-parent-id>`: New parent module (required).

### Phase management

#### `phase`

Displays the current lifecycle phase of a module.

    arci module phase <module-id>

#### `advance`

Advances a module to the next lifecycle phase. All tasks for the current phase must be complete and verification tasks must have no blocking defects.

    arci module advance <module-id> --to <phase>

**Flags:**

- `--to <phase>`: Target phase (required).

#### `regress`

Moves a module back to an earlier lifecycle phase. Creates a defect (DEF-*) automatically with the provided reason.

    arci module regress <module-id> --to <phase> --reason <reason>

**Flags:**

- `--to <phase>`: Target phase (required).
- `--reason <reason>`: Why the module is regressing (required).

### Work scoping

#### `decompose`

Creates tasks for a module from a template.

    arci module decompose <module-id> --template <template-name>

**Flags:**

- `--template <template-name>`: Task template to apply (required).

#### `tasks`

Lists tasks for a module.

    arci module tasks <module-id> [options]

**Flags:**

- `--include-descendants`: Include tasks from child modules.

## See also

- [Modules](../../graph/nodes/modules.md): Graph node documentation for modules
- [Lifecycle coordination](../../graph/lifecycle-coordination.md): Phase advancement and constraints
