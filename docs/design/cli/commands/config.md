# Config

The config command group provides configuration inspection, validation, and modification.

## Synopsis

    krav config <subcommand> [options]

## Description

These commands operate on Krav's layered configuration system. The system merges configuration from all sources in precedence order. These commands let you inspect the effective result, check structure, and change individual layers.

## Subcommands

### `validate`

Checks all configuration files for structural errors. Reports YAML syntax errors, unknown fields, and schema violations. This verifies the configuration file structure; use `krav hook policy validate` to check policy and rule expressions.

**Exit codes:**

- `0`: Configuration is valid.
- `1`: One or more errors found.

### `list`

Shows all discovered configuration sources in precedence order. For each source, it displays the path (or indicates built-in/environment), whether the file exists, and how many policies it contributes.

### `show`

Dumps the merged configuration as YAML. This shows the effective configuration after the system merges all sources.

### `where`

Prints the paths where Krav looks for configuration.

**Flags:**

- `--project`: Show project-level paths for a specific directory.

### `get`

Retrieves a configuration value by key.

    krav config get <key>

Keys use dot notation like `server.port` and `logging.level`.

**Flags:**

- `--scope <scope>`: Read from a specific configuration layer. Without this flag, shows the effective merged value.

### `set`

Sets a configuration value.

    krav config set <key> <value>

Writes to the YAML file on disk and triggers a configuration reload if the server is running.

**Flags:**

- `--scope <scope>`: Which configuration file to edit (defaults to local/project).

### `unset`

Removes a configuration value from a file.

    krav config unset <key>

**Flags:**

- `--scope <scope>`: Which configuration file to edit (defaults to local).

## See also

- [Configuration](../../configuration/configuration.md)
- [Config Cascade](../../configuration/config-cascade.md)
