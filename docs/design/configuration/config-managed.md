# Managed configuration

This document describes the managed configuration system for enterprise deployments, including the bypass mechanism for policy developers.

## Overview

Managed configuration allows enterprises to enforce policies across their fleet. It follows Chrome's two-tier model with recommended (overridable defaults) and required (enforced, non-overridable) tiers.

```text
<system config dir>/
├── managed/
│   ├── recommended/    # Enterprise defaults, user/project can override
│   │   ├── arci.yaml
│   │   ├── arci-policies.yaml
│   │   └── policies.d/
│   └── required/       # Enforced policies, cannot be overridden
│       ├── arci.yaml
│       ├── arci-policies.yaml
│       └── policies.d/
```

IT administrators deploy managed configuration via MDM (Mobile Device Management), configuration management tools like Ansible/Chef/Puppet, or manual installation. Modifying the files requires elevated privileges on most systems.

## Cascade behavior

Managed recommended loads early in the cascade, after built-in defaults. Users, projects, environment variables, and CLI flags can all override these settings.

Managed required loads last in the cascade, after all other sources including CLI flags. Nothing can override these settings, providing absolute enforcement for security-critical policies.

If managed required configuration exists but fails to parse, ARCI fails closed: it refuses to run rather than proceeding without enterprise security policies. This is the only source with fail-closed semantics.

## Bypass mechanism for policy developers

Developers writing managed policies need to test their changes without production managed policies interfering. The custom config override (`ARCI_CONFIG_FILE`, etc.) doesn't help because managed required still applies on top.

A bypass mechanism allows authorized developers to turn off managed policy loading entirely. This requires cryptographic proof of authorization to prevent abuse.

### Threat model

The bypass mechanism must protect against:

- End users turning off security policies to circumvent controls
- Leaked credentials enabling bypass on production machines
- Malware using bypass to turn off protective policies

It does not need to protect against:

- Developers with administrator access on their own machines (they could modify the managed files directly)
- Nation-state attackers with persistent access (out of scope for this control)

### Key distribution

The bypass mechanism uses asymmetric cryptography. Policy developers hold private keys and generate signed bypass tokens. Machines that should accept bypass have the corresponding public key installed.

MDM does NOT distribute the public key. Production machines have managed policies but no public key, so the system cannot verify bypass tokens and bypass simply does not work.

Developer machine setup is a manual process:

1. The machine enrolls in MDM and receives managed policies like all other machines
2. Developer manually installs `bypass.pub` to system config directory (requires administrator access)
3. Developer can now use signed bypass tokens

The presence of the public key is itself the opt-in to bypass capability. A leaked token is useless on machines without the key.

```text
<system config dir>/
├── managed/
│   ├── recommended/
│   └── required/
└── bypass.pub          # Only on developer machines
```

### Token format

Bypass tokens are JSON payloads with an Ed25519 signature, base64-encoded for use in environment variables:

```text
ARCI_BYPASS_MANAGED=<base64(payload + signature)>
```

Payload structure:

```json
{
  "v": 1,
  "exp": "2025-02-01T00:00:00Z",
  "sub": "tony@example.com",
  "scope": ["config", "policies"],
  "machine": "tony-mbp.local"
}
```

Fields:

| Field | Required | Description |
|-------|----------|-------------|
| `v` | Yes | Token format version, currently `1` |
| `exp` | Yes | Expiration timestamp (RFC 3339). Tokens should be short-lived. |
| `sub` | Yes | Subject identifier, typically email. For audit logging. |
| `scope` | No | Which bypass types the token allows: `config`, `policies`, `policies-dir`. Default is all. |
| `machine` | No | If present, token only valid on this machine hostname. |

### Token validation

When the user sets `ARCI_BYPASS_MANAGED`:

1. Check if `<system config dir>/bypass.pub` exists. If not, ignore the token entirely (no warning, as this is the expected state on production machines).

2. Decode the base64 token and split into payload and signature.

3. Verify the Ed25519 signature against the public key. If invalid, log warning and ignore bypass.

4. Parse the JSON payload and validate:
   - `v` must be `1`
   - `exp` must be in the future
   - `machine`, if present, must match current hostname

5. If valid, skip loading managed recommended and managed required based on `scope`.

6. Log that bypass is active, including `sub` for audit trail.

### Key management

Organizations should maintain their own bypass keypair. The private key should be:

- Stored securely (hardware token, secrets manager, or encrypted at rest)
- Accessible only to authorized policy developers
- Used only on developer machines, never on CI/CD or shared systems

The public key should be:

- Manually distributed to developer machines (not via MDM)
- Rotatable if compromised (install new public key, revoke old tokens)

A simple key generation flow:

```bash
# Generate keypair (policy admin does this once)
arci managed keygen --out bypass.key --pub bypass.pub

# Generate a bypass token (developer does this as needed)
arci managed token --key bypass.key \
  --expires 24h \
  --subject tony@example.com \
  --machine tony-mbp.local

# Output: ARCI_BYPASS_MANAGED=eyJ2IjoxLC...
```

### CLI commands

```bash
# Generate a new keypair
arci managed keygen [--out FILE] [--pub FILE]

# Generate a bypass token
arci managed token --key FILE [--expires DURATION] [--subject EMAIL] [--machine HOSTNAME] [--scope SCOPE...]

# Verify a token (for debugging)
arci managed verify --pub FILE --token TOKEN

# Show managed config status
arci managed status
```

The `status` command shows:

- Whether managed recommended exists and loaded successfully
- Whether managed required exists and loaded successfully
- Whether bypass is active and for whom
- Whether bypass.pub is present (on developer machines)

### Security considerations

**Token expiration**: tokens should be short-lived. Recommend 24 hours for interactive development, 1 hour for CI/CD (if ever needed). The `token` command should warn if expiration exceeds 7 days.

**Audit logging**: when bypass is active, every ARCI invocation should log the `sub` from the token. This creates an audit trail of who bypassed policies and when.

**No bypass.pub in version control**: do not commit the public key to the ARCI policies repository. It is a machine-local file that developers install manually.

**Separate keys per environment**: large organizations might want separate keypairs for different teams or environments. The public key filename could be configurable or support multiple keys.

**Revocation**: no revocation mechanism exists beyond expiration. If an attacker compromises a private key, rotate to a new keypair and distribute the new public key to developer machines. Short-lived tokens limit exposure.

## Future considerations

### ADMX/mobileconfig integration

For organizations using Windows Group Policy or macOS MDM profiles, the system could read a small subset of scalar settings from OS policy stores:

- `ArciEnabled` (bool) - emergency stop
<<<<<<< Updated upstream
- `FailurePolicy` (string) - allow/deny
=======
- `FailurePolicy` (string) - allow/deny
>>>>>>> Stashed changes
- `ServerEnabled` (bool)
- `LogLevel` (string)

These would slot preceding managed required in precedence, providing an ultimate override capability for emergency situations (turning off ARCI fleet-wide if a bug causes outages).

File-based managed config remains the primary mechanism for complex policy definitions.

### Managed policy signing

Beyond bypass tokens, organizations might want to sign the managed policy files themselves, ensuring no one has tampered with them. This is a separate concern from bypass and would require a different key distribution model (public key via MDM, since all machines need to verify).

Not planned for the initial release.
