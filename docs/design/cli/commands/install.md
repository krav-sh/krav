# Install / uninstall

The install and uninstall commands manage integration with Claude Code. These commands change Claude Code configuration files to wire up Krav hooks.

## Synopsis

```text
Krav install [options]
Krav uninstall [options]
```

## Description

### Install

The `krav install` command configures Claude Code to invoke Krav. Without arguments, it enters an interactive mode that guides the user through configuration. With explicit flags, it can run non-interactively for scripting.

### Uninstall

The `krav uninstall` command reverses the installation, removing Krav hook entries from Claude Code configuration.

## Options

### Install options

- `--scope <scope>`: Choose global, project, or both levels.
- `--scaffold`: Create starter configuration files.
- `--non-interactive`: Skip prompts (for scripting).
- `--dry-run`: Preview changes without applying them.

### Uninstall options

- `--scope <scope>`: Choose which level to uninstall from.
- `--purge`: Also remove Krav's own configuration directories.

## Examples

```bash
# Interactive installation
Krav install

# Non-interactive project-level install with starter files
Krav install --scope project --scaffold --non-interactive

# Preview what install would do
Krav install --dry-run

# Remove krav hooks from project config
Krav uninstall --scope project

# Full removal including config directories
Krav uninstall --purge
```

## See also

- [Installation](../../installation.md)
