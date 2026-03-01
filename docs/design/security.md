# Security model

This document describes ARCI's security model, including trust relationships, threat scenarios, and the guardrails around rule execution and extension loading.

## Trust model overview

ARCI operates in a privileged position: it intercepts Claude Code operations and can execute code through shell and script actions. Understanding who trusts whom is essential for reasoning about security.

The trust relationships form a chain: the user trusts their Claude Code to execute code on their behalf. The Claude Code trusts ARCI (via configured hooks) to evaluate operations. ARCI trusts its configuration sources to define legitimate rules. ARCI trusts installed extensions to provide safe rule packs and Starlark scripts.

Breaking any link in this chain has security implications. A malicious rule can exfiltrate data via shell actions. A malicious extension can ship rules that execute arbitrary commands. A compromised daemon can intercept all hook traffic.

Starlark scripts provide a safer alternative to shell actions because the runtime sandboxes them by default. Scripts cannot access the filesystem or network directly; they can only interact with the outside world through APIs that ARCI explicitly exposes, such as the state store functions.

## Threat scenarios

Enumerate and analyze threat scenarios.

Topics to cover:

### Malicious rules

- Rules that exfiltrate sensitive data via shell actions
- Rules that modify input to inject malicious commands
- Rules that deny legitimate operations (denial of service)
- Rules that leak information through timing or output
- Sandbox escapes that allow shell actions to exceed their configured restrictions

### Malicious extensions

- Extensions with malicious Starlark scripts in bundled rule packs
- Extensions that ship rules with dangerous shell actions
- Supply chain attacks via git dependencies or compromised registries
- Extensions that exfiltrate state store contents through allowed APIs
- Typosquatting on extension package names

### Configuration tampering

- Unauthorized modification of user-level configuration
- Project configuration that overrides safety rules
- Symlink attacks on configuration directories
- Race conditions during configuration reload

### Daemon compromise

- Unauthorized access to the daemon API
- Exploitation of daemon vulnerabilities
- Man-in-the-middle between CLI and daemon
- State store manipulation

### Extension supply chain

- Compromised git repositories
- Lockfile manipulation
- Hash collision attacks (theoretical)
- Unsigned extension packages

## Security controls

Document existing and planned security controls.

Topics to cover:

### Configuration security

- File permission requirements for configuration directories
- Validation of configuration before loading
- Isolation between project configurations
- The role of `.local` files for sensitive personal rules

### Extension security

- Lockfile hashing and verification
- Commit pinning for git sources
- Audit surface for rules-only extensions (YAML inspection)
- Planned: signature verification for extensions

### Daemon security

- Localhost-only binding by default
- Unix socket permissions
- No authentication (current limitation)
- Planned: API key or token authentication

### Action execution security

ARCI provides multiple layers of security for action execution, with different guarantees for each action type.

**Script actions (Starlark)** run in a sandbox by default. The Starlark runtime provides built-in isolation with no filesystem access, no network access, and instruction counting to prevent infinite loops. Scripts can only interact with the outside world through APIs that ARCI explicitly exposes: the state store functions (`session_get`, `session_set`, `project_get`, `project_set`), git context functions, and path helper functions. Memory limits prevent unbounded allocation. This makes script actions inherently safer than shell actions for complex logic.

**Shell actions** execute commands in the user's environment. By default, shell actions run with user privileges and have full system access. For security-sensitive deployments, ARCI provides optional sandboxing that constrains filesystem access, network connectivity, and resource consumption.

Shell action sandboxing uses platform-native technologies:

- **Linux and WSL2**: bubblewrap (bwrap) for namespace-based isolation
- **macOS**: Seatbelt via sandbox-exec
- **Windows without WSL2**: Resource limits only via job objects (no filesystem/network isolation)

Predefined sandbox profiles simplify configuration: `restricted` (strong isolation, network blocked, filesystem read-only except cwd), `network-isolated` (allows filesystem writes, blocks network), `read-only` (blocks writes, allows network), and `unrestricted` (explicit opt-out).

For complete sandbox configuration options including per-rule overrides, fail-open versus fail-closed behavior, and platform-specific details, see [Shell action sandboxing](sandboxing.md).

Timeout enforcement applies to both shell and script actions. ARCI stops actions that exceed their timeout and fails open, contributing no directives to the evaluation.

Template substitution in action messages and commands uses Go's `text/template` with `missingkey=zero` for safe defaults. Unknown variables resolve to zero values rather than causing errors.

## Sensitive data handling

Document how to handle sensitive data.

Topics to cover:

- API keys and tokens in rule conditions
- Secrets in action outputs
- State store and sensitive values
- Log redaction
- Dashboard exposure of sensitive data

## Audit logging

Document audit capabilities.

Topics to cover:

- What events the system logs
- Log retention and rotation
- Tamper-evident logging (future)
- Integration with external audit systems

## Security recommendations

Provide security guidance for different deployment scenarios.

Topics to cover:

### Individual developers

- Extension vetting before installation
- Review of project-level configurations
- Preferring script actions over shell actions when possible
- Enabling shell action sandboxing for untrusted rule sources

### Teams

- Centralized rule management
- Extension allowlisting
- Code review for rule changes
- Sandboxing extension-provided rules by default

### Enterprises

- System-level configuration for mandatory rules
- MDM deployment of security policies
- Audit requirements
- Compliance considerations

## Known limitations

Document security limitations that users should understand.

Topics to cover:

- No authentication on daemon API
- Sandbox availability varies by platform (Windows without WSL2 has limited isolation)
- Fail-open semantics and security implications (configuration errors or unavailable sandboxing does not block the Claude Code by default)
- Trust in extension sources (git repositories, registries)

## Future security enhancements

Document planned security improvements.

Topics to cover:

- Extension signature verification
- Daemon authentication
- Audit logging improvements
- Security scanning for rules
- Enhanced sandbox escape detection

---

This document expands as the security model matures. Security is a process, not a feature, and this document should evolve as the team identifies new threats and deploys mitigations.
