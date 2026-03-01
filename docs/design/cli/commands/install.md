# Install / uninstall

The install and uninstall commands manage integration with Claude Code. These commands change Claude Code configuration files to wire up ARCI hooks.

## Synopsis

```text
arci install [options]
arci uninstall [options]
```

## Description

### Install

The `arci install` command configures Claude Code to invoke ARCI. Without arguments, it enters an interactive mode that guides the user through configuration. With explicit flags, it can run non-interactively for scripting.

### Uninstall

The `arci uninstall` command reverses the installation, removing ARCI hook entries from Claude Code configuration.

## Options

### Install options

- `--scope <scope>`: Choose global, project, or both levels.
- `--scaffold`: Create starter configuration files.
- `--non-interactive`: Skip prompts (for scripting).
- `--dry-run`: Preview changes without applying them.

### Uninstall options

- `--scope <scope>`: Choose which level to uninstall from.
- `--purge`: Also remove ARCI's own configuration directories.

## Examples

```bash
# Interactive installation
arci install

# Non-interactive project-level install with starter files
arci install --scope project --scaffold --non-interactive

# Preview what install would do
arci install --dry-run

# Remove arci hooks from project config
arci uninstall --scope project

# Full removal including config directories
arci uninstall --purge
```

## See also

- [Installation](../../installation.md)
