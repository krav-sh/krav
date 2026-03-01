# Configuration

This document describes the `arci.yaml` configuration file format, including schema versioning and extension mechanisms.

## Loading architecture

Configuration loading is a shell-layer concern. The config loader handles all filesystem interaction and YAML parsing; the functional core receives only typed domain objects.

The loader performs four operations. Discovery determines which paths to check using platform-appropriate directory resolution, project root detection, and the cascade rules described in [config-cascade.md](config-cascade.md). Loading reads file contents from disk, handling missing files and permission errors gracefully. Parsing deserializes YAML and validates against the configuration schema. Materialization transforms validated structures into typed domain objects that the core can use.

This separation keeps the core pure and testable. The core has no knowledge of YAML, file paths, or precedence layers; it operates on domain types like `Rule`, `Policy`, and `Config`.

## File format

Configuration files use YAML format with a required `$schema` field for version identification:

```yaml
$schema: https://arci.dev/schemas/arci/v1.yaml

log_level: info

default:
  daemon:
    enabled: true
    port: 9100
  failure_policy: allow
```

## Schema versioning

Every `arci.yaml` file requires the `$schema` field. It identifies the configuration schema version and enables tooling support (editor autocompletion, validation).

```yaml
$schema: https://arci.dev/schemas/arci/v1.yaml
```

Schema versions follow semantic versioning. Breaking changes increment the major version. The loader validates files against their declared schema and reports errors for unknown or incompatible versions.

## Structure

The configuration file uses a `default` section containing all settings:

### Default section

The `default` section contains all configuration settings:

```yaml
default:
  log_level: info

  daemon:
    enabled: true
    port: 9100
    hot_reload: true

  failure_policy: allow

  state:
    backend: sqlite
    path: null # uses platform default

  providers:
    specs:
      enabled: false
      endpoint: null
```

## Extension mechanism

The `extends` field allows a configuration file to inherit from other files:

```yaml
$schema: https://arci.dev/schemas/arci/v1.yaml

extends:
  - ../shared/team-base.yaml
  - ./security-overlay.yaml

default:
  log_level: debug # Overrides anything from extended files
```

The loader processes extended files in order. Each file's settings merge on top of the previous result. Finally, the current file's settings merge on top. This allows building configuration from composable pieces.

Paths in `extends` are relative to the file containing the `extends` field. Absolute paths are also supported but discouraged for portability.

Extended files can themselves extend other files. The loader detects circular references and raises a validation error.

## Settings reference

### log_level

Controls logging verbosity. Values: `error`, `warn`, `info`, `debug`, `trace`. Default: `info`.

```yaml
default:
  log_level: info
```

### failure_policy

Determines behavior when policy evaluation fails (expression errors, provider timeouts, etc.). Values: `allow` (fail-open, continue without the failing policy) or `deny` (fail-closed, block the operation). Default: `allow`.

```yaml
default:
  failure_policy: allow
```

This is a critical safety property. With `allow`, configuration errors never block Claude Code from operating. Only explicit deny decisions from successfully evaluated policies block operations.

### Daemon

Controls the optional daemon process that caches configuration and serves an HTTP API.

```yaml
default:
  daemon:
    enabled: true
    host: 127.0.0.1
    port: 9100
    hot_reload: true
    metrics:
      enabled: true
      port: 9101
```

| Field             | Description                              | Default     |
| ----------------- | ---------------------------------------- | ----------- |
| `enabled`         | Whether to use the daemon for evaluation | `true`      |
| `host`            | Address to bind                          | `127.0.0.1` |
| `port`            | Port for the evaluation API              | `9100`      |
| `hot_reload`      | Watch config files and reload on change  | `true`      |
| `metrics.enabled` | Expose Prometheus metrics                | `true`      |
| `metrics.port`    | Metrics endpoint port                    | `9101`      |

When `daemon.enabled` is `false`, the CLI performs direct evaluation by loading configuration on each invocation.

### State

Configures the persistent state store used for session and project state.

```yaml
default:
  state:
    backend: sqlite
    path: null # Platform default
```

| Field     | Description                | Default                  |
| --------- | -------------------------- | ------------------------ |
| `backend` | Storage backend (`sqlite`) | `sqlite`                 |
| `path`    | Database path              | Platform state directory |

The state store enables patterns like rate limiting ("warn on first occurrence, block on third") with data that persists across sessions.

### Providers

Named parameter providers that policies can reference.

```yaml
default:
  providers:
    specs:
      type: http
      endpoint: http://localhost:8080/params
      timeout: 100ms
      cache_ttl: 5m

    team-config:
      type: file
      path: .arci/team.json
      watch: true
```

Policies reference providers by name in their `parameters` section:

```yaml
parameters:
  - name: blockedCommands
    from: specs
    defaults: []
```

See [policy-model.md](../hooks/policy-model.md) for details on parameter resolution.

## Environment variable mapping

Settings can be overridden via environment variables. The mapping follows these rules:

1. Prefix with `ARCI_`
2. Convert to uppercase
3. Replace dots and nested keys with double-underscore separators

Examples:

| Setting                  | Environment Variable          |
| ------------------------ | ----------------------------- |
| `default.log_level`      | `ARCI_LOG_LEVEL`        |
| `default.daemon.enabled` | `ARCI_DAEMON__ENABLED`  |
| `default.daemon.port`    | `ARCI_DAEMON__PORT`     |

## Validation

Configuration validation happens at two layers.

Schema validation in the config loader catches YAML syntax errors, unknown fields (with configurable strictness), type mismatches, missing required fields, and invalid enum values.

Semantic validation during materialization catches invalid provider configurations, circular extends references, and incompatible setting combinations.

The CLI provides validation commands:

```bash
# Validate all configuration files in the cascade
arci config validate

# Validate a specific file
arci config validate --file .arci/arci.yaml

# Show the effective merged configuration
arci config show
```

## Error handling

Parse errors in configuration files produce warnings, not hard failures. The loader skips the file and continues with remaining sources. This fail-open behavior ensures that a typo in a project config doesn't prevent Claude Code from operating.

The system logs and reports errors through:

- CLI output when running commands
- Daemon dashboard status panel
- The `arci config validate` command

Managed/required configuration is the exception. Errors in required managed config cause a hard failure, as proceeding without enterprise security policies would violate the security model.

## Hot reloading

When the daemon is running with `hot_reload: true`, it watches all configuration files for changes using filesystem notifications. On change:

1. The modified file is re-read and re-parsed
2. The full cascade is re-merged
3. Policies are re-compiled
4. The new configuration is atomically swapped in

Requests in flight complete with the old configuration. New requests use the new configuration. If reloading fails due to errors, the daemon continues with the previous valid configuration and logs the error.

## Example configurations

### Minimal project configuration

```yaml
$schema: https://arci.dev/schemas/arci/v1.yaml

default:
  log_level: info
```

### Development configuration with debugging

```yaml
$schema: https://arci.dev/schemas/arci/v1.yaml

default:
  log_level: debug
  daemon:
    enabled: true
    hot_reload: true
```

### Enterprise configuration with ARCI specs integration

```yaml
$schema: https://arci.dev/schemas/arci/v1.yaml

extends:
  - /etc/arci/managed/recommended/arci.yaml

default:
  failure_policy: deny

  providers:
    specs:
      type: http
      endpoint: https://specs.internal.company.com/params
      timeout: 200ms
      headers:
        Authorization: "Bearer ${SPECS_TOKEN}"

  daemon:
    enabled: true
    metrics:
      enabled: true
```

### Personal user configuration

```yaml
$schema: https://arci.dev/schemas/arci/v1.yaml

default:
  log_level: warn
```
