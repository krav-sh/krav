# Configuration cascade

Krav uses a layered configuration system that merges settings and policies from multiple sources. This document describes the cascade order, directory locations, and merging semantics.

## Cascade order

The loader reads configuration sources from lowest to highest precedence. Higher-precedence sources override lower-precedence ones, with managed/required configuration enforced at the top of the stack regardless of user or project settings.

```text
Defaults (built-in)
    ↓
<system config dir>/managed/recommended/{krav.yaml, krav-policies.yaml, policies.d/*.yaml}
    ↓
<system config dir>/{krav.yaml, krav-policies.yaml, policies.d/*.yaml}
    ↓
<user config dir>/{krav.yaml, krav-policies.yaml, policies.d/*.yaml}
    ↓
<project dir>/{krav.yaml | .krav.yaml | .config/krav/krav.yaml | .krav/krav.yaml}
<project config dir>/{krav-policies.yaml, policies.d/*.yaml}
    ↓
<project dir>/{krav.local.yaml | .krav.local.yaml | .config/krav/krav.local.yaml | .krav/krav.local.yaml}
<project config dir>/{krav-policies.local.yaml, policies.local.d/*.yaml}
    ↓
KRAV_* environment variables
    ↓
CLI flags
    ↓
KRAV_CONFIG_FILE (replaces all krav*.yaml if set)
    ↓
KRAV_POLICIES_FILE (replaces all krav-policies*.yaml if set)
    ↓
KRAV_POLICIES_DIR (replaces all policies*.d/ if set)
    ↓
<system config dir>/managed/required/{krav.yaml, krav-policies.yaml, policies.d/*.yaml}
```

No user, project, or environment configuration can override the managed/required tier at the top. The managed/recommended tier near the bottom provides enterprise defaults that users and projects can customize.

## Directory locations

Directory paths follow platform conventions via the toolpaths module.

### System configuration directory

IT administrators manage system-wide configuration. Regular users typically cannot modify these files.

| Platform | Path                                          |
| -------- | --------------------------------------------- |
| Linux    | `/etc/xdg/krav` (or `$XDG_CONFIG_DIRS`) |
| macOS    | `/Library/Application Support/krav`     |
| Windows  | `C:\ProgramData\krav`                   |

The system directory contains two managed subdirectories for enterprise policy enforcement:

```text
<system config dir>/
├── managed/
│   ├── recommended/     # Enterprise defaults (overridable)
│   │   ├── krav.yaml
│   │   ├── krav-policies.yaml
│   │   └── policies.d/
│   │       └── security-baseline.yaml
│   └── required/        # Enterprise enforcement (not overridable)
│       ├── krav.yaml
│       ├── krav-policies.yaml
│       └── policies.d/
│           └── compliance.yaml
├── krav.yaml      # System defaults
├── krav-policies.yaml
└── policies.d/
```

### User configuration directory

Personal configuration that applies across all projects.

| Platform | Path                                                      |
| -------- | --------------------------------------------------------- |
| Linux    | `~/.config/krav` (or `$XDG_CONFIG_HOME/krav`) |
| macOS    | `~/Library/Application Support/krav`                |
| Windows  | `%APPDATA%\krav`                                    |

```text
<user config dir>/
├── krav.yaml
├── krav-policies.yaml
└── policies.d/
    ├── personal-safety.yaml
    └── coding-style.yaml
```

### Project directory

The loader determines the project directory (project root) through a precedence chain of mutually exclusive sources. Only one source applies; they do not merge.

In order of precedence (highest wins):

1. `--project-dir` CLI flag
2. `KRAV_PROJECT_DIR` environment variable
3. Git worktree root (if in a worktree)
4. VCS root (Git, Mercurial, etc.)
5. Nearest ancestor directory containing a project marker

The ancestor traversal searches for project markers including VCS directories (`.git`, `.hg`, `.svn`, `.bzr`), Krav config directories (`.krav`, `.config/krav`), and Krav config files (`krav.yaml`, `.krav.yaml`, `krav.local.yaml`, `.krav.local.yaml`).

If none of these resolve, there is no project configuration.

### Project configuration directory

Within the project directory, the configuration directory can exist in one of two places. The loader uses only one, with the following precedence:

1. `.krav/` (highest)
2. `.config/krav/` (lowest)

The `.krav/` location is the default and recommended choice. The `.config/krav/` alternative supports projects that prefer to consolidate tooling configuration under a `.config/` directory to reduce clutter in the project root.

If both directories exist, Krav uses `.krav/` and logs a warning:

```text
warning: multiple config directories found, using .krav/ (ignoring .config/krav/)
```

The `krav init` command creates `.krav/` by default. Use `krav init --config-dir .config/krav` to create the alternative structure.

### Project configuration files

Policy files always live in the project configuration directory:

```text
<project config dir>/
├── krav-policies.yaml       # Policy enable/disable state (committed)
├── krav-policies.local.yaml # Personal policy state (gitignored)
├── policies.d/                    # Policy definitions (committed)
│   ├── api-standards.yaml
│   └── test-requirements.yaml
└── policies.local.d/              # Personal policies (gitignored)
    └── experiments.yaml
```

The main configuration file (`krav.yaml`) can exist in one of four locations. The loader uses only one, with the following precedence (highest to lowest):

1. `.krav/krav.yaml`
2. `.config/krav/krav.yaml`
3. `.krav.yaml`
4. `krav.yaml`

The same precedence applies to local configuration overrides:

1. `.krav/krav.local.yaml`
2. `.config/krav/krav.local.yaml`
3. `.krav.local.yaml`
4. `krav.local.yaml`

If multiple config files exist, Krav uses the highest-precedence file and logs a warning identifying the ignored files:

```text
warning: multiple config files found, using .krav/krav.yaml (ignoring .krav.yaml, krav.yaml)
```

The directory structure (`.krav/krav.yaml` or `.config/krav/krav.yaml`) works best for projects that also use policies, as it keeps all Krav configuration in one place. The dotfile (`.krav.yaml`) or plain file (`krav.yaml`) alternatives in the project root are convenient for simple projects that only need configuration without policies.

### Example project structures

Standard structure with dedicated directory:

```text
my-project/
├── .krav/
│   ├── krav.yaml
│   ├── krav.local.yaml           # gitignored
│   ├── krav-policies.yaml
│   ├── krav-policies.local.yaml  # gitignored
│   ├── policies.d/
│   │   └── team-standards.yaml
│   └── policies.local.d/               # gitignored
│       └── experiments.yaml
└── src/
```

Alternative structure using `.config/`:

```text
my-project/
├── .config/
│   └── krav/
│       ├── krav.yaml
│       ├── krav-policies.yaml
│       └── policies.d/
│           └── team-standards.yaml
└── src/
```

Minimal structure with just a config file:

```text
my-project/
├── .krav.yaml
└── src/
```

### Gitignore patterns

Add these patterns to `.gitignore` to exclude local overrides:

```gitignore
# krav local overrides (standard location)
.krav/krav.local.yaml
.krav/krav-policies.local.yaml
.krav/policies.local.d/

# krav local overrides (.config location)
.config/krav/krav.local.yaml
.config/krav/krav-policies.local.yaml
.config/krav/policies.local.d/

# krav local overrides (project root)
.krav.local.yaml
krav.local.yaml
```

## File types and purposes

Each cascade layer can contain three types of configuration files.

### `krav.yaml`

Main configuration file containing settings like logging and server behavior. See [configuration.md](configuration.md) for the schema.

### `krav-policies.yaml`

Controls which policies the system enables or turns off without modifying policy definitions. The `krav hook policy enable/disable` commands manage policy state through this file rather than touching policy definition files.

```yaml
$schema: https://krav.sh/schemas/krav-policies/v1.yaml

defaultBehavior: all-enabled

enabled:
  - security-baseline
  - coding-standards

disabled:
  - experimental-feature

audit:
  - new-security-policy
```

The `defaultBehavior` field controls what happens to unlisted policies. Valid values are `all-enabled` (default), `all-disabled`, and `all-audit`. The `audit` list places policies in dry-run mode where the system evaluates them but downgrades their actions: deny becomes warn, the system computes mutations but does not apply them, and it logs effects but does not execute them.

### policies.d/\*.yaml

Directory containing policy definition files. Each file is a single or multi-document YAML file (using `---` delimiter) containing policy definitions as described in [policy-model.md](../hooks/policy-model.md). Filenames are not semantically relevant; the `name` field identifies each policy.

The loader reads files in lexicographical order within each directory. Use numeric prefixes to control ordering when it matters:

```text
policies.d/
├── 00-security.yaml
├── 10-git-workflow.yaml
└── 20-coding-style.yaml
```

## Merging semantics

### Configuration merging

Settings in `krav.yaml` files merge recursively. Higher-precedence sources replace scalar values. Higher-precedence sources replace arrays entirely (not concatenated). Maps merge key-by-key.

The `extends` field allows explicit inheritance from other configuration files before merging:

```yaml
extends:
  - ../shared/base-config.yaml
  - ./team-defaults.yaml
```

The loader merges extended files in order, then applies the current file's settings on top.

### Policy state merging

The `krav-policies.yaml` files merge across layers using per-policy precedence. The highest-precedence layer that mentions a policy determines its effective state. A policy appearing in multiple lists within the same manifest (both `enabled` and `disabled`) is a validation error that rejects the manifest.

```text
# System: defaultBehavior: all-enabled, disabled: [dangerous-policy]
# User: enabled: [dangerous-policy]  ← overrides system
# Project: disabled: [dangerous-policy]  ← overrides user
# Result: dangerous-policy is disabled
```

The three enforcement states (enabled, turned off, audit) are mutually exclusive. The system evaluates a policy in the `audit` list in dry-run mode:

```text
# User: enabled: [new-security-policy]
# Project: audit: [new-security-policy]  ← overrides user, runs in dry-run mode
# Result: new-security-policy runs in audit mode
```

Policies in `krav-policies.local.yaml` override policies in `krav-policies.yaml` at the same cascade level.

### Policy definition merging

Policy names must be unique within a cascade layer. If the same policy name appears in both `policies.d/` and `policies.local.d/` at the same layer, the `.local.d` version replaces the non-local version.

Across cascade layers, policies with the same name coexist. When querying or listing policies:

- Unqualified names (`security-baseline`) resolve to the highest-precedence layer
- Qualified names (`system/security-baseline`, `project/security-baseline`) select a specific layer

The layer names for qualification are: `managed-required`, `managed-recommended`, `system`, `user`, `project`, `project-local`.

## Environment variables

### Configuration overrides

Environment variables prefixed with `KRAV_` override corresponding settings in `krav.yaml`. Nested keys use double `_` separators to show hierarchy:

```bash
KRAV_LOG_LEVEL=debug
KRAV_SERVER__ENABLED=false
```

### Cascade override variables

Three special environment variables replace entire segments of the cascade:

| Variable                   | Effect                                                                                                                             |
| -------------------------- | ---------------------------------------------------------------------------------------------------------------------------------- |
| `KRAV_CONFIG_FILE`   | If set, the loader reads only this file for `krav*.yaml` configuration and ignores all other config files in the cascade.          |
| `KRAV_POLICIES_FILE` | If set, the loader reads only this file for `krav-policies*.yaml` state and ignores all other policy state files.                  |
| `KRAV_POLICIES_DIR`  | If set, the loader reads only this directory for `policies*.d/` definitions and ignores all other policy directories.               |

These overrides are useful for testing, CI/CD, and debugging. They bypass the normal cascade entirely for their respective file types, but managed/required configuration still applies on top.

## Managed configuration

The managed/recommended and managed/required directories provide enterprise policy enforcement following the Chrome policy model.

### Managed/recommended

Configuration in `managed/recommended/` provides enterprise defaults that users and projects can override. The loader merges these settings early in the cascade, giving higher layers the opportunity to customize them.

Use this for organizational preferences, suggested security policies, and default configurations that teams should be able to adjust for their needs.

### Managed/required

The loader merges configuration in `managed/required/` last, and nothing can override it. Settings here take absolute precedence over all user, project, and environment configuration.

Use this for security policies that must hold, compliance requirements, and configurations that users should not be able to turn off.

IT administrators deploy managed configuration via MDM tools, configuration management systems, or manual installation. Modifying these files requires elevated privileges on most systems.

## Error handling

Parse errors in configuration files produce warnings, not failures. Krav follows fail-open semantics: if the loader cannot parse a configuration file, it skips the file and continues with the remaining sources.

A syntax error in a project's `krav.yaml` does not prevent the tool from running. It runs with the configuration that loaded successfully, and the system logs errors and surfaces them through the CLI's `config validate` command and the server's dashboard.

Managed/required configuration errors follow different rules. If the loader cannot read required managed configuration, Krav fails closed rather than proceeding without enterprise security policies.

## Diagnostics

The CLI provides commands for inspecting the effective configuration:

```bash
# Show resolved configuration with source annotations
Krav config show

# Validate all configuration files
Krav config validate

# Show where a specific setting comes from
Krav config explain log_level

# List all policies with their source layers
Krav hook policy list

# Show a specific policy definition
Krav hook policy get security-baseline

# Show policy from a specific layer
Krav hook policy get system/security-baseline
```

The server dashboard also displays configuration status, including any parse errors or validation warnings.
