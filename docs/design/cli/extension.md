# Extension

The extension command group manages arci extensions.

## Synopsis

    arci extension <subcommand> [options]

## Description

Extensions add policies, functions, and action handlers to arci. They can be installed from Go module paths, git URLs, or local paths. Extension dependencies are tracked in a manifest and resolved to a lockfile for reproducible installations.

## Subcommands

### list

Shows installed extensions. For each extension, it displays the name, version, source (Go module, git, or path), and what the extension provides (policies, functions, action handlers).

### add

Installs an extension.

    arci extension add <spec>

The spec can be a Go module path for published packages, a git URL, or a local path. Updates the manifest, resolves dependencies, updates the lockfile, and installs the extension.

**Flags:**

- `--project` — Add to the project manifest instead of user manifest.
- `--path` — Treat spec as a local path.

### remove

Uninstalls an extension.

    arci extension remove <name>

Removes the extension from the manifest and updates the lockfile.

### lock

Resolves the manifest to a lockfile without installing. Useful for generating a lockfile in CI before a separate install step.

### sync

Installs exactly what's in the lockfile without re-resolving. Ensures reproducible installations across machines.

### upgrade

Upgrades extensions within their manifest constraints.

    arci extension upgrade [name]

Without a name argument, upgrades all extensions. With a name, upgrades only that extension.

### init

Scaffolds a new extension package.

    arci extension init <name>

**Flags:**

- `--policies-only` — Generate a minimal policies-only extension structure instead of a full extension with stub function and action handler implementations.

## See also

- [Extensions](../extensions.md)
