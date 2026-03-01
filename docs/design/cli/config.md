# Config

The config command group provides configuration inspection, validation, and modification.

## Synopsis

    arci config <subcommand> [options]

## Description

These commands operate on ARCI's layered configuration system. ARCI merges configuration from multiple sources in precedence order; these commands let you inspect the effective result, validate structure, and modify individual layers.

## Subcommands

### Validate

Checks all configuration files for structural errors. Reports YAML syntax errors, unknown fields, and schema violations. This validates the configuration file structure; use `arci hook policy validate` to validate policy and rule expressions.

**Exit codes:**

- `0`: configuration is valid.
- `1`: one or more errors found.

### List

Shows all discovered configuration sources in precedence order. For each source, it displays the path (or indicates built-in/environment), whether the file exists, and how many policies it contributes.

### Show

Dumps the effective configuration as YAML after all sources merge.

### Where

Prints the paths where ARCI looks for configuration.

**Flags:**

- `--project`: show project-level paths for a specific directory.

### Get

Retrieves a configuration value by key.

    arci config get <key>

Keys use dot notation (like `daemon.port`, `logging.level`).

**Flags:**

- `--scope <scope>`: read from a specific configuration layer. Without this flag, shows the effective merged value.

### Set

Sets a configuration value.

    arci config set <key> <value>

Modifies the YAML file on disk and triggers a configuration reload if the daemon is running.

**Flags:**

- `--scope <scope>`: which configuration file to modify (defaults to local/project).

### Unset

Removes a configuration value from a file.

    arci config unset <key>

**Flags:**

- `--scope <scope>`: which configuration file to modify (defaults to local).

## See also

- [Configuration](../configuration.md)
- [Config Cascade](../config-cascade.md)
