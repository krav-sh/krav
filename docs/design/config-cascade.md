# Configuration cascade

arci uses a layered configuration system that merges settings and policies from multiple sources. This document describes the cascade order, directory locations, and merging semantics.

## Cascade order

Configuration sources are loaded from lowest to highest precedence. Higher-precedence sources override lower-precedence ones, with managed/required configuration enforced at the top of the stack regardless of user or project settings.

```
Defaults (built-in)
    ↓
<system config dir>/managed/recommended/{arci.yaml, arci-policies.yaml, policies.d/*.yaml}
    ↓
<system config dir>/{arci.yaml, arci-policies.yaml, policies.d/*.yaml}
    ↓
<user config dir>/{arci.yaml, arci-policies.yaml, policies.d/*.yaml}
    ↓
<project dir>/{arci.yaml | .arci.yaml | .config/arci/arci.yaml | .arci/arci.yaml}
<project config dir>/{arci-policies.yaml, policies.d/*.yaml}
    ↓
<project dir>/{arci.local.yaml | .arci.local.yaml | .config/arci/arci.local.yaml | .arci/arci.local.yaml}
<project config dir>/{arci-policies.local.yaml, policies.local.d/*.yaml}
    ↓
ARCI_* environment variables
    ↓
CLI flags
    ↓
ARCI_CONFIG_FILE (replaces all arci*.yaml if set)
    ↓
ARCI_POLICIES_FILE (replaces all arci-policies*.yaml if set)
    ↓
ARCI_POLICIES_DIR (replaces all policies*.d/ if set)
    ↓
<system config dir>/managed/required/{arci.yaml, arci-policies.yaml, policies.d/*.yaml}
```

The managed/required tier at the top cannot be overridden by any user, project, or environment configuration. The managed/recommended tier near the bottom provides enterprise defaults that users and projects can customize.

## Directory locations

Directory paths follow platform conventions via the toolpaths module.

### System configuration directory

System-wide configuration managed by IT administrators. Regular users typically cannot modify these files.

| Platform | Path                                          |
| -------- | --------------------------------------------- |
| Linux    | `/etc/xdg/arci` (or `$XDG_CONFIG_DIRS`) |
| macOS    | `/Library/Application Support/arci`     |
| Windows  | `C:\ProgramData\arci`                   |

The system directory contains two managed subdirectories for enterprise policy enforcement:

```
<system config dir>/
├── managed/
│   ├── recommended/     # Enterprise defaults (overridable)
│   │   ├── arci.yaml
│   │   ├── arci-policies.yaml
│   │   └── policies.d/
│   │       └── security-baseline.yaml
│   └── required/        # Enterprise enforcement (not overridable)
│       ├── arci.yaml
│       ├── arci-policies.yaml
│       └── policies.d/
│           └── compliance.yaml
├── arci.yaml      # System defaults
├── arci-policies.yaml
└── policies.d/
```

### User configuration directory

Personal configuration that applies across all projects.

| Platform | Path                                                      |
| -------- | --------------------------------------------------------- |
| Linux    | `~/.config/arci` (or `$XDG_CONFIG_HOME/arci`) |
| macOS    | `~/Library/Application Support/arci`                |
| Windows  | `%APPDATA%\arci`                                    |

```
<user config dir>/
├── arci.yaml
├── arci-policies.yaml
└── policies.d/
    ├── personal-safety.yaml
    └── coding-style.yaml
```

### Project directory

The project directory (project root) is determined by a precedence chain of mutually exclusive sources. Only one source is used; they do not merge.

In order of precedence (highest wins):

1. `--project-dir` CLI flag
2. `ARCI_PROJECT_DIR` environment variable
3. Git worktree root (if in a worktree)
4. VCS root (Git, Mercurial, etc.)
5. Nearest ancestor directory containing a project marker

Project markers searched during ancestor traversal include VCS directories (`.git`, `.hg`, `.svn`, `.bzr`), arci config directories (`.arci`, `.config/arci`), and arci config files (`arci.yaml`, `.arci.yaml`, `arci.local.yaml`, `.arci.local.yaml`).

If none of these resolve, there is no project configuration.

### Project configuration directory

Within the project directory, the configuration directory can be located in one of two places. Only one is used, with the following precedence:

1. `.arci/` (highest)
2. `.config/arci/` (lowest)

The `.arci/` location is the default and recommended choice. The `.config/arci/` alternative is provided for projects that prefer to consolidate tooling configuration under a `.config/` directory to reduce clutter in the project root.

If both directories exist, arci uses `.arci/` and logs a warning:

```
warning: multiple config directories found, using .arci/ (ignoring .config/arci/)
```

The `arci init` command creates `.arci/` by default. Use `arci init --config-dir .config/arci` to create the alternative structure.

### Project configuration files

Policy files always live in the project configuration directory:

```
<project config dir>/
├── arci-policies.yaml       # Policy enable/disable state (committed)
├── arci-policies.local.yaml # Personal policy state (gitignored)
├── policies.d/                    # Policy definitions (committed)
│   ├── api-standards.yaml
│   └── test-requirements.yaml
└── policies.local.d/              # Personal policies (gitignored)
    └── experiments.yaml
```

The main configuration file (`arci.yaml`) can be placed in one of four locations. Only one is used, with the following precedence (highest to lowest):

1. `.arci/arci.yaml`
2. `.config/arci/arci.yaml`
3. `.arci.yaml`
4. `arci.yaml`

The same precedence applies to local configuration overrides:

1. `.arci/arci.local.yaml`
2. `.config/arci/arci.local.yaml`
3. `.arci.local.yaml`
4. `arci.local.yaml`

If multiple config files exist, arci uses the highest-precedence file and logs a warning identifying which files are being ignored:

```
warning: multiple config files found, using .arci/arci.yaml (ignoring .arci.yaml, arci.yaml)
```

The directory structure (`.arci/arci.yaml` or `.config/arci/arci.yaml`) is recommended for projects that also use policies, as it keeps all arci configuration in one place. The dotfile (`.arci.yaml`) or plain file (`arci.yaml`) alternatives in the project root are convenient for simple projects that only need configuration without policies.

### Example project structures

Standard structure with dedicated directory:

```
my-project/
├── .arci/
│   ├── arci.yaml
│   ├── arci.local.yaml           # gitignored
│   ├── arci-policies.yaml
│   ├── arci-policies.local.yaml  # gitignored
│   ├── policies.d/
│   │   └── team-standards.yaml
│   └── policies.local.d/               # gitignored
│       └── experiments.yaml
└── src/
```

Alternative structure using `.config/`:

```
my-project/
├── .config/
│   └── arci/
│       ├── arci.yaml
│       ├── arci-policies.yaml
│       └── policies.d/
│           └── team-standards.yaml
└── src/
```

Minimal structure with just a config file:

```
my-project/
├── .arci.yaml
└── src/
```

### Gitignore patterns

Add these patterns to `.gitignore` to exclude local overrides:

```gitignore
# arci local overrides (standard location)
.arci/arci.local.yaml
.arci/arci-policies.local.yaml
.arci/policies.local.d/

# arci local overrides (.config location)
.config/arci/arci.local.yaml
.config/arci/arci-policies.local.yaml
.config/arci/policies.local.d/

# arci local overrides (project root)
.arci.local.yaml
arci.local.yaml
```

## File types and purposes

Each cascade layer can contain three types of configuration files.

### arci.yaml

Main configuration file containing settings like logging and daemon behavior. See [configuration.md](configuration.md) for the schema.

### arci-policies.yaml

Controls which policies are enabled or disabled without modifying policy definitions. This separation allows the `arci hook policy enable/disable` commands to manage policy state without touching policy definition files.

```yaml
$schema: https://arci.dev/schemas/arci-policies/v1.yaml

defaultBehavior: all-enabled

enabled:
  - security-baseline
  - coding-standards

disabled:
  - experimental-feature

audit:
  - new-security-policy
```

The `defaultBehavior` field controls what happens to unlisted policies. Valid values are `all-enabled` (default), `all-disabled`, and `all-audit`. The `audit` list places policies in dry-run mode where they are evaluated but their actions are downgraded: deny becomes warn, mutations are computed but not applied, and effects are logged but not executed.

### policies.d/\*.yaml

Directory containing policy definition files. Each file is a single or multi-document YAML file (using `---` delimiter) containing policy definitions as described in [policy-model.md](hooks/policy-model.md). Filenames are not semantically relevant; policies are identified by their `name` field.

Files are loaded in lexicographical order within each directory. Use numeric prefixes to control ordering when it matters:

```
policies.d/
├── 00-security.yaml
├── 10-git-workflow.yaml
└── 20-coding-style.yaml
```

## Merging semantics

### Configuration merging

Settings in `arci.yaml` files merge recursively. Scalar values are replaced by higher-precedence sources. Arrays are replaced entirely (not concatenated). Maps are merged key-by-key.

The `extends` field allows explicit inheritance from other configuration files before merging:

```yaml
extends:
  - ../shared/base-config.yaml
  - ./team-defaults.yaml
```

Extended files are merged in order, then the current file's settings are applied on top.

### Policy state merging

The `arci-policies.yaml` files merge across layers using per-policy precedence. The effective state for a policy is determined by the highest-precedence layer that mentions it. A policy appearing in multiple lists within the same manifest (for example, both `enabled` and `disabled`) is a validation error that rejects the manifest.

```
# System: defaultBehavior: all-enabled, disabled: [dangerous-policy]
# User: enabled: [dangerous-policy]  ← overrides system
# Project: disabled: [dangerous-policy]  ← overrides user
# Result: dangerous-policy is disabled
```

The three enforcement states (enabled, disabled, audit) are mutually exclusive. A policy in the `audit` list is evaluated in dry-run mode:

```
# User: enabled: [new-security-policy]
# Project: audit: [new-security-policy]  ← overrides user, runs in dry-run mode
# Result: new-security-policy runs in audit mode
```

Policies in `arci-policies.local.yaml` override policies in `arci-policies.yaml` at the same cascade level.

### Policy definition merging

Policy names must be unique within a cascade layer. If the same policy name appears in both `policies.d/` and `policies.local.d/` at the same layer, the `.local.d` version completely replaces the non-local version.

Across cascade layers, policies with the same name coexist. When querying or listing policies:

- Unqualified names (e.g., `security-baseline`) resolve to the highest-precedence layer
- Qualified names (e.g., `system/security-baseline`, `project/security-baseline`) select a specific layer

The layer names for qualification are: `managed-required`, `managed-recommended`, `system`, `user`, `project`, `project-local`.

## Environment variables

### Configuration overrides

Environment variables prefixed with `ARCI_` override corresponding settings in `arci.yaml`. Nested keys use double underscores:

```bash
ARCI_LOG_LEVEL=debug
ARCI_DAEMON__ENABLED=false
```

### Cascade override variables

Three special environment variables replace entire segments of the cascade:

| Variable                   | Effect                                                                                                                             |
| -------------------------- | ---------------------------------------------------------------------------------------------------------------------------------- |
| `ARCI_CONFIG_FILE`   | If set, only this file is loaded for `arci*.yaml` configuration. All other config files in the cascade are ignored.          |
| `ARCI_POLICIES_FILE` | If set, only this file is loaded for `arci-policies*.yaml` state. All other policy state files are ignored.                  |
| `ARCI_POLICIES_DIR`  | If set, only this directory is loaded for `policies*.d/` definitions. All other policy directories are ignored.                    |

These overrides are useful for testing, CI/CD, and debugging. They bypass the normal cascade entirely for their respective file types, but managed/required configuration still applies on top.

## Managed configuration

The managed/recommended and managed/required directories provide enterprise policy enforcement following the Chrome policy model.

### managed/recommended

Configuration in `managed/recommended/` provides enterprise defaults that users and projects can override. These settings are merged early in the cascade, giving higher layers the opportunity to customize them.

Use this for organizational preferences, suggested security policies, and default configurations that teams should be able to adjust for their needs.

### managed/required

Configuration in `managed/required/` is merged last and cannot be overridden. Settings here take absolute precedence over all user, project, and environment configuration.

Use this for security policies that must be enforced, compliance requirements, and configurations that users should not be able to disable.

IT administrators deploy managed configuration via MDM tools, configuration management systems, or manual installation. The files require elevated privileges to modify on most systems.

## Error handling

Parse errors in configuration files result in warnings, not failures. arci follows fail-open semantics: if a configuration file cannot be parsed, it is skipped and loading continues with the remaining sources.

This means a syntax error in a project's `arci.yaml` doesn't prevent the tool from running—it runs with the configuration that could be loaded successfully. Errors are logged and surfaced through the CLI's `config validate` command and the daemon's dashboard.

Critically, managed/required configuration errors are treated differently. If required managed configuration cannot be loaded, arci fails closed rather than proceeding without enterprise security policies.

## Diagnostics

The CLI provides commands for inspecting the effective configuration:

```bash
# Show resolved configuration with source annotations
arci config show

# Validate all configuration files
arci config validate

# Show where a specific setting comes from
arci config explain log_level

# List all policies with their source layers
arci hook policy list

# Show a specific policy definition
arci hook policy get security-baseline

# Show policy from a specific layer
arci hook policy get system/security-baseline
```

The daemon dashboard also displays configuration status, including any parse errors or validation warnings.
