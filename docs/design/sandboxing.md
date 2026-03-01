# Sandboxing

ARCI can run inside a sandbox that constrains filesystem access, network connectivity, and resource consumption. The sandbox provides defense in depth: even if a malicious rule or prompt injection attack tricks ARCI into executing dangerous commands, the sandbox limits the blast radius.

Sandboxing is optional and off by default. Users opt in via the `--sandbox` flag or configuration when the security benefit outweighs operational constraints. Organizations can enforce sandboxing via managed configuration so that users cannot turn it off. The build uses platform-native technologies: bubblewrap on Linux and WSL2, Seatbelt via sandbox-exec on macOS.

## The re-exec pattern

Sandboxing uses a re-exec pattern where the ARCI binary detects a sandboxing request, constructs a platform-specific wrapper command, and executes itself inside the sandbox. The sandboxed process sees that it's already inside a sandbox and proceeds with normal execution.

```text
User invokes:
  arci hook apply --sandbox claude ...

CLI detects --sandbox, not already sandboxed:
  exec: bwrap [options] -- arci hook apply claude ...
        (Linux)

  exec: sandbox-exec -f profile.sb arci hook apply claude ...
        (macOS)

Sandboxed process detects ARCI_SANDBOXED=1:
  Proceeds with normal evaluation
```

The `ARCI_SANDBOXED=1` environment variable signals to the re-exec'd process that it's already inside a sandbox and shouldn't attempt to wrap itself again. This prevents infinite re-exec loops.

This same pattern applies to three scenarios: daemonless CLI execution, daemon process startup, and the combination where a sandboxed CLI spawns a sandboxed daemon.

## Configuration model

Sandbox configuration uses a cross-platform abstraction that maps to platform-specific implementations. The configuration separates filesystem and network concerns, following the model established by Claude Code's sandboxing system.

### Filesystem configuration

Filesystem settings control which paths the sandboxed process can read and write:

```yaml
sandbox:
  enabled: true
  filesystem:
    writablePaths:
      - "{{ project_dir }}"
      - "{{ state_dir }}"
      - "/tmp"
    readablePaths:
      - "{{ system_config_dir }}"
      - "{{ user_config_dir }}"
      - "/usr"
      - "/lib"
      - "/bin"
    deniedPaths:
      - "{{ project_dir }}/.env"
      - "{{ project_dir }}/secrets"
      - "~/.ssh"
      - "~/.aws"
      - "~/.gnupg"
```

The `writablePaths` list specifies directories where the sandbox can create, modify, and delete files. The `readablePaths` list specifies additional directories with read-only access. The `deniedPaths` list explicitly blocks access even if a parent directory would otherwise allow it.

The sandbox expands template variables at setup time using the resolved configuration:

| Variable | Description |
|----------|-------------|
| `{{ project_dir }}` | Project root directory |
| `{{ project_config_dir }}` | Project configuration directory (`.arci/` or `.config/arci/`) |
| `{{ user_config_dir }}` | User configuration directory (`~/.config/arci` on Linux) |
| `{{ system_config_dir }}` | System configuration directory (`/etc/xdg/arci` on Linux) |
| `{{ state_dir }}` | State store directory |
| `{{ socket_dir }}` | Daemon socket directory |

### Network configuration

Network settings control which hosts the sandboxed process can connect to:

```yaml
sandbox:
  enabled: true
  network:
    mode: deny  # deny, allow, or restricted
    allowedHosts:
      - "api.github.com"
      - "*.anthropic.com"
```

The `mode` field controls the default network policy. `deny` blocks all network access (the default). `allow` permits all network access. `restricted` blocks by default but permits connections to hosts in `allowedHosts`.

For most ARCI use cases, `mode: deny` is appropriate. The core evaluation engine doesn't need network access. You can configure shell actions that require network with explicit host allowlists or run them outside the sandbox.

### Built-in allowances

When you enable sandboxing, ARCI automatically includes paths required for its own operation. These built-in allowances ensure the sandbox doesn't break the configuration cascade or state management.

The following paths always have read access when you enable sandboxing:

```text
# System configuration (managed policies live here)
<system_config_dir>/                     # /etc/xdg/arci on Linux
<system_config_dir>/managed/recommended/
<system_config_dir>/managed/required/

# User configuration
<user_config_dir>/                       # ~/.config/arci on Linux

# Project configuration (one of these, based on detection)
<project_dir>/.arci/
<project_dir>/.config/arci/
<project_dir>/arci.yaml
<project_dir>/.arci.yaml

# System libraries and binaries (for shell actions)
/usr/
/lib/
/lib64/
/bin/
/etc/resolv.conf
```

The following paths always have write access when you enable sandboxing:

```text
# State store
<state_dir>/                             # ~/.local/share/arci on Linux

# Daemon socket
<socket_dir>/                            # /tmp/arci on Linux

# Project directory (where shell actions typically operate)
<project_dir>/

# Temporary files
/tmp/
```

Configuration cannot remove these built-in allowances. They represent the minimum access ARCI needs to function. The `deniedPaths` list can block subdirectories within allowed paths, such as denying `{{ project_dir }}/.env` while allowing `{{ project_dir }}`.

### Profiles

Profiles provide named presets that configure filesystem and network settings together:

```yaml
sandbox:
  enabled: true
  profile: standard
```

The `standard` profile includes the built-in allowances with network access denied. It's suitable for typical usage where shell actions operate on project files without network requirements.

The `network-allowed` profile adds `network.mode: allow` on top of the standard filesystem settings. Use this when shell actions need to fetch dependencies or call APIs.

The `minimal` profile restricts writable paths to only the project directory and state store. Network access defaults to deny. Configuration directories allow reads but not writes. This is appropriate for evaluating rules from untrusted sources.

Custom configuration overrides profile settings. If you specify a profile and also include explicit `filesystem` or `network` sections, the explicit settings merge with and override the profile defaults.

## Managed configuration enforcement

Organizations can enforce sandboxing via managed configuration so that users cannot turn it off. This uses the same managed configuration mechanism described in the configuration cascade document.

### Recommended sandboxing

Configuration in `<system_config_dir>/managed/recommended/` provides enterprise defaults that users can override:

```yaml
# /etc/xdg/arci/managed/recommended/arci.yaml
sandbox:
  enabled: true
  profile: standard
```

Users can turn off sandboxing or change the profile in their personal or project configuration. This suits environments that encourage sandboxing without requiring it.

### Required sandboxing

Configuration in `<system_config_dir>/managed/required/` cannot be overridden by any user, project, or environment configuration:

```yaml
# /etc/xdg/arci/managed/required/arci.yaml
sandbox:
  enabled: true
  filesystem:
    deniedPaths:
      - "~/.ssh"
      - "~/.aws"
      - "~/.gnupg"
      - "~/.config/gh"
  network:
    mode: deny
```

With required managed configuration, users cannot turn off sandboxing or add the denied paths to their allowed list. IT administrators deploy managed configuration via MDM tools, configuration management systems, or manual installation. The files require elevated privileges to modify on most systems.

If managed/required configuration enables sandboxing but the platform cannot provide it (bubblewrap not installed, etc.), ARCI fails closed rather than proceeding unsandboxed. This differs from the default fail-open behavior for user-requested sandboxing.

### Combining with policy enforcement

Managed configuration can also require specific policies that complement sandboxing. An organization might require both sandboxing and a policy that blocks dangerous shell patterns:

```yaml
# /etc/xdg/arci/managed/required/arci.yaml
sandbox:
  enabled: true
  profile: standard

# /etc/xdg/arci/managed/required/arci-policies.yaml
defaultBehavior: all-enabled
enabled:
  - security-baseline

# /etc/xdg/arci/managed/required/policies.d/security-baseline.yaml
name: security-baseline
description: Enterprise security baseline
bindings:
  - events: [PreToolUse]
    condition: tool_name == "Bash"
validate:
  - condition: not (tool_input.command matches "rm\\s+-rf\\s+/")
    deny: "Dangerous rm command blocked by enterprise policy"
```

This defense-in-depth approach uses policies to block known-dangerous patterns and sandboxing to limit damage from patterns that slip through.

## Platform implementations

### Linux and WSL2

On Linux systems including WSL2, ARCI uses bubblewrap (bwrap) for sandboxing. Bubblewrap creates lightweight containers using Linux namespaces without requiring root privileges. It's widely deployed as the sandbox for Flatpak applications and has a strong security track record.

Bubblewrap's overhead is minimal. Benchmarks show roughly 5-8 ms of setup time for namespace creation and bind mounts, compared to 250-300 ms for Docker container startup. For a CLI targeting sub-50 ms response times, this overhead is acceptable, and daemon mode pays the cost only once at startup.

The sandbox configuration maps to bwrap arguments:

| Configuration | bwrap argument |
|---------------|----------------|
| `writablePaths` entry | `--bind <path> <path>` |
| `readablePaths` entry | `--ro-bind <path> <path>` |
| `deniedPaths` entry | Path excluded from binds |
| `network.mode: deny` | `--unshare-net` |
| `network.mode: restricted` | Custom network namespace with proxy |

A typical bwrap invocation for the standard profile:

```bash
bwrap \
  --ro-bind /usr /usr \
  --ro-bind /lib /lib \
  --ro-bind /lib64 /lib64 \
  --ro-bind /bin /bin \
  --ro-bind /etc/resolv.conf /etc/resolv.conf \
  --ro-bind ~/.config/arci ~/.config/arci \
  --ro-bind /etc/xdg/arci /etc/xdg/arci \
  --bind "${PROJECT_DIR}" "${PROJECT_DIR}" \
  --bind ~/.local/share/arci ~/.local/share/arci \
  --bind /tmp/arci /tmp/arci \
  --dev /dev \
  --proc /proc \
  --tmpfs /tmp \
  --unshare-net \
  --die-with-parent \
  --setenv ARCI_SANDBOXED 1 \
  -- arci hook apply claude ...
```

Install bubblewrap separately. On Debian and Ubuntu, install with `apt install bubblewrap`. On Fedora and RHEL, use `dnf install bubblewrap`. If bubblewrap is unavailable and the user requests sandboxing (but managed config does not require it), ARCI follows fail-open semantics: it logs a warning and executes without sandboxing.

### macOS

On macOS, ARCI uses Seatbelt via the sandbox-exec command. Seatbelt is Apple's mandatory access control system, built into macOS since Leopard. Unlike bubblewrap, it requires no installation and is available on every Mac.

Seatbelt's overhead model differs from bubblewrap. Rather than creating isolated namespaces at startup, Seatbelt attaches a policy to the process that the kernel checks at syscall time via MACF (Mandatory Access Control Framework) hooks. Startup overhead is negligible, just parsing the profile and calling `sandbox_init()`. The per-syscall overhead is minimal for typical workloads.

The sandbox configuration maps to Seatbelt profile rules:

| Configuration | Seatbelt rule |
|---------------|---------------|
| `writablePaths` entry | `(allow file-write* (subpath "..."))` |
| `readablePaths` entry | `(allow file-read* (subpath "..."))` |
| `deniedPaths` entry | `(deny file-read* (subpath "..."))` before allows |
| `network.mode: deny` | `(deny network*)` |
| `network.mode: restricted` | `(deny network*)` with `(allow network* (remote tcp "..."))` |

ARCI generates a temporary Seatbelt profile from the configuration. The invocation:

```bash
sandbox-exec -f /tmp/arci-profile.sb \
  -D SYSTEM_CONFIG_DIR="/Library/Application Support/arci" \
  -D USER_CONFIG_DIR="$HOME/Library/Application Support/arci" \
  -D PROJECT_DIR="$PROJECT_DIR" \
  -D STATE_DIR="$HOME/Library/Application Support/arci/state" \
  -D SOCKET_DIR="/tmp/arci" \
  arci hook apply claude ...
```

Apple has marked sandbox-exec as deprecated but continues using Seatbelt internally for system services and as the foundation of the App Sandbox. The capability is unlikely to disappear, though the API may change between macOS versions.

### Windows without WSL2

Native Windows lacks the kernel features needed for full sandboxing. Without WSL2, ARCI provides only resource limits through Windows job objects. These constrain CPU time, memory usage, working set size, and process count, but do not provide filesystem or network isolation.

For users who need full sandboxing on Windows, the recommended approach is to run ARCI under WSL2 where the Linux build applies normally. When native Windows users request sandboxing without WSL2, ARCI logs a warning indicating that it enforces only resource limits.

## Sandboxing modes

The re-exec pattern enables three sandboxing modes that work independently or in combination.

### CLI sandboxing

When the user passes `--sandbox` to the CLI (or sets `sandbox.enabled: true` in configuration), the entire evaluation runs inside a sandbox:

```bash
arci hook apply --sandbox --event=PreToolUse claude < event.json
```

The sandbox constrains what ARCI itself can access. Configuration paths allow only reads. The project directory allows writes so that shell actions can modify project files. The sandbox blocks network access by default.

CLI sandboxing pays the sandbox setup cost on every invocation. On Linux with bubblewrap, this adds roughly 5-8 ms per call. On macOS with Seatbelt, the overhead is negligible. For most use cases this is acceptable, but high-frequency invocations may prefer daemon mode, which amortizes the cost.

### Daemon sandboxing

When the daemon starts with sandboxing enabled, the entire daemon process runs inside a sandbox. Every RPC the daemon handles executes within the already-established sandbox with zero additional setup overhead:

```bash
arci daemon --sandbox
```

The socket path must be in a location that's bind-mounted into the sandbox so that unsandboxed CLI clients can connect. The default `/tmp/arci/<project-hash>/daemon.sock` works because the sandbox's writable paths include `/tmp/arci`.

Daemon sandboxing is the most efficient mode when you require sandboxing. The namespace or policy setup happens once at daemon startup. Subsequent requests benefit from the cached configuration and established sandbox without paying setup costs repeatedly.

### Combined mode

When you enable both CLI and daemon sandboxing, ARCI auto-spawns a sandboxed daemon from a sandboxed CLI. The CLI runs sandboxed, detects no running daemon, and spawns a new daemon that also runs sandboxed:

```bash
arci hook apply --sandbox --event=PreToolUse claude < event.json
  # CLI is sandboxed
  # Detects no daemon, spawns one
  # Daemon inherits sandbox configuration
  # Subsequent calls connect to sandboxed daemon
```

The sandbox constrains the daemon even if an attacker compromises it.

## Fail-open behavior

Sandboxing integrates with ARCI's fail-open philosophy, but security-sensitive failure modes deserve explicit attention.

When the user requests sandboxing via `--sandbox` or user/project configuration but the platform cannot provide it (bubblewrap not installed, sandbox-exec unavailable, etc.), the default behavior is fail-open: log a warning and proceed without sandboxing. This matches the principle that configuration errors should not block the Claude Code:

```yaml
sandbox:
  enabled: true
  onUnavailable: warn  # Default: log warning, proceed unsandboxed
```

For deployments where proceeding without sandboxing is unacceptable, `onUnavailable: fail` enables fail-closed behavior:

```yaml
sandbox:
  enabled: true
  onUnavailable: fail  # Abort if sandbox unavailable
```

A third option, `onUnavailable: skip`, causes ARCI to exit successfully without evaluating anything:

```yaml
sandbox:
  enabled: true
  onUnavailable: skip  # Exit 0, no evaluation
```

When managed/required configuration enforces sandboxing, `onUnavailable` is implicitly `fail`. Enterprise-mandated sandboxing should never silently degrade to unsandboxed execution.

## Security considerations

### What sandboxing protects against

Sandboxing limits damage from malicious or buggy rules. The sandbox blocks a rule that attempts to read `~/.ssh/id_rsa` or write to `/etc/passwd` if those paths aren't in the allow list. Network isolation prevents data exfiltration; a compromised rule cannot send project contents to an external server.

The sandbox provides defense in depth against prompt injection. If an attacker tricks Claude Code into invoking ARCI with malicious input, the sandbox constrains what the resulting evaluation can access. Even a successful injection can only reach the project directory rather than the entire filesystem.

Resource limits (where supported) prevent denial-of-service. A rule that spawns an infinite loop or allocates unbounded memory stops when it exceeds its limits.

### What sandboxing does not protect against

Sandboxing does not make dangerous commands safe. A sandboxed `rm -rf {{ project_dir }}` still deletes the project. The sandbox constrains where damage can occur, not what operations run within allowed paths. Rule conditions and policies remain the first line of defense against dangerous commands.

Sandboxing cannot prevent reading files within allowed paths. If the sandbox permits reading the project directory and the project contains secrets in `.env`, any rule can access those secrets. Use `deniedPaths` to explicitly block sensitive files even within allowed directories.

Sandboxing provides limited protection against side channels. Timing attacks, existence checks, and error message analysis can leak information even when the sandbox blocks direct access.

Network filtering operates at the domain level and does not inspect traffic content. If the sandbox allows `github.com`, an attacker could potentially exfiltrate data through that domain. Be cautious about allowing broad domains.

Sandboxing doesn't protect against vulnerabilities in the sandbox code itself. Bubblewrap and Seatbelt have strong track records, but sandbox escapes are theoretically possible. Defense in depth (combining sandboxing with policies, input validation, and monitoring) provides the strongest security posture.

### The sandbox is not a complete security boundary

The sandbox is one layer in a defense-in-depth strategy. It's most effective when combined with policies that block known-dangerous patterns before they reach execution:

```yaml
# Managed required policy
name: security-baseline
bindings:
  - events: [PreToolUse]
    condition: tool_name == "Bash"
validate:
  - condition: not (tool_input.command matches "(~/.ssh|~/.aws|~/.gnupg)")
    deny: "Access to credential directories blocked"
  - condition: not (tool_input.command matches "curl.*\\|.*sh")
    deny: "Piping curl to shell blocked"
```

The policy catches known-dangerous patterns. The sandbox limits damage from patterns that slip through.

## Implementation notes

Sandbox setup lives in the CLI shell layer. The functional core knows nothing about sandboxing; it evaluates rules and returns actions. The shell layer decides whether and how to wrap execution.

On Linux, the CLI constructs a bwrap command line from the resolved sandbox configuration. The CLI substitutes path templates, sets the ARCI_SANDBOXED environment variable, and execs the wrapper.

On macOS, the CLI writes a temporary Seatbelt profile to `/tmp`, substituting path parameters, then execs sandbox-exec with the profile path and parameter definitions.

The re-exec detection is simple: if the process finds `ARCI_SANDBOXED=1`, it skips sandbox setup and proceeds with normal execution. The sandbox wrapper always sets this environment variable, ensuring the inner process knows it's already sandboxed.

For daemon mode, the same logic applies at daemon startup. The daemon checks for the sandboxed environment variable, and if the configuration requests sandboxing but the process has not yet applied it, the daemon re-execs itself inside the sandbox before binding the socket and accepting connections.
