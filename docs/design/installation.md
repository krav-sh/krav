# Installation

This document describes how users install arci and configure it to work with Claude Code. Go binaries are self-contained with no runtime dependencies, making installation straightforward across all platforms.

## Quick start

The simplest way to get started is with go install:

```bash
go install github.com/tbhb/arci@latest
arci install
```

This installs the `arci` binary to `$GOPATH/bin/` (or `~/go/bin/` by default), then runs the install subcommand to configure Claude Code.

For users who prefer pre-built binaries, download the appropriate release for your platform from the GitHub releases page and place it somewhere on your PATH.

## Installation methods

### go install

The standard Go installation method works for anyone with the Go toolchain installed:

```bash
go install github.com/tbhb/arci@latest
```

This downloads, compiles, and installs arci. Compilation takes a few seconds on most machines.

### goreleaser / pre-built binaries

For installation without the Go toolchain, download pre-built binaries from the GitHub releases page or use a package manager like Homebrew.

### Homebrew

On macOS and Linux, Homebrew provides the simplest installation:

```bash
brew install arci
```

### Pre-built binaries

Each release includes pre-built binaries for common platforms:

- `arci-linux-amd64.tar.gz` (Linux x86_64)
- `arci-linux-arm64.tar.gz` (Linux ARM64)
- `arci-darwin-amd64.tar.gz` (macOS Intel)
- `arci-darwin-arm64.tar.gz` (macOS Apple Silicon)
- `arci-windows-amd64.zip` (Windows x86_64)

Download the appropriate archive, extract it, and move the binary to a directory on your PATH:

```bash
# macOS/Linux example
curl -L https://github.com/tbhb/arci/releases/latest/download/arci-aarch64-apple-darwin.tar.gz | tar xz
mv arci ~/.local/bin/
```

## Development installation

Contributors to arci itself can build and install from a local checkout:

```bash
go install ./cmd/arci
```

For iterative development without reinstalling, use go build and run the binary directly:

```bash
go build -o arci ./cmd/arci
./arci --help
```

Or use go run during development:

```bash
go run ./cmd/arci -- install --dry-run
```

## The install command

Once arci is installed, the `arci install` command configures integration with Claude Code.

```bash
arci install [options]
```

The install command handles several tasks: detecting the Claude Code installation, modifying Claude Code configuration to invoke arci on hook events, creating initial arci configuration directories, and optionally scaffolding starter configuration files.

### Interactive mode

Running `arci install` without arguments enters an interactive flow:

1. arci detects the Claude Code installation
2. The user chooses the installation scope (global, project, or both)
3. arci writes the necessary configuration and reports what it created
4. Since Claude Code requires manual approval of hook changes, arci explains the next steps (the `/hooks` review)

The interactive mode provides a guided experience for users who are setting up arci for the first time or who want to understand what changes are being made.

### Non-interactive mode

For scripted installation or CI/CD pipelines, non-interactive mode skips prompts and uses explicit flags:

```bash
arci install --non-interactive --scope global
```

In non-interactive mode, all options must be specified via flags. If a required decision cannot be made from the provided flags, the command fails with an error explaining what's missing.

### Command options

The `--scope` flag determines where hook configuration is written. Valid values are `global` (user-level configuration), `project` (project-level configuration), and `all` (both levels). The default in non-interactive mode is `global`.

The `--project` flag specifies the project directory for project-scope installation. It defaults to the current working directory. This allows installing project hooks for a repo without `cd`ing into it first.

The `--scaffold` flag creates starter configuration files with commented examples. Without this flag, install only writes the minimum hook wiring; with it, the user gets a `config.yaml` template they can customize:

```bash
arci install --scaffold
```

The `--dry-run` flag shows what would be changed without making any modifications. This is useful for understanding the installation's effects before committing to them.

## Claude Code detection

Claude Code is detected by checking for the `claude` binary on the PATH and the existence of `~/.claude/` directory. If either is present, Claude Code is considered available.

The presence of `.claude/` in the current directory (or specified project directory) indicates project-level Claude Code configuration exists, which affects what scope options are offered.

## Hook configuration

The install command adds hook entries that invoke `arci hook apply` for relevant Claude Code events.

### Claude Code hook configuration

Installation modifies `settings.json` files. At the global scope, this is `~/.claude/settings.json`. At the project scope, this is `.claude/settings.json` (or `.claude/settings.local.json` if the user prefers uncommitted configuration).

The installed hooks look like:

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "",
        "hooks": [
          {
            "type": "command",
            "command": "arci hook apply"
          }
        ]
      }
    ],
    "PostToolUse": [
      {
        "matcher": "",
        "hooks": [
          {
            "type": "command",
            "command": "arci hook apply"
          }
        ]
      }
    ],
    "PermissionRequest": [
      {
        "matcher": "",
        "hooks": [
          {
            "type": "command",
            "command": "arci hook apply"
          }
        ]
      }
    ],
    "UserPromptSubmit": [
      {
        "matcher": "",
        "hooks": [
          {
            "type": "command",
            "command": "arci hook apply"
          }
        ]
      }
    ],
    "SessionStart": [
      {
        "matcher": "startup",
        "hooks": [
          {
            "type": "command",
            "command": "arci hook apply"
          }
        ]
      }
    ],
    "SessionEnd": [
      {
        "matcher": "",
        "hooks": [
          {
            "type": "command",
            "command": "arci hook apply"
          }
        ]
      }
    ],
    "Notification": [
      {
        "matcher": "",
        "hooks": [
          {
            "type": "command",
            "command": "arci hook apply"
          }
        ]
      }
    ],
    "Stop": [
      {
        "matcher": "",
        "hooks": [
          {
            "type": "command",
            "command": "arci hook apply"
          }
        ]
      }
    ]
  }
}
```

Empty matchers (`""`) match all tool names for tool-related events. The `SessionStart` matcher of `"startup"` is Claude Code's convention for the session initialization event.

When the settings file already exists, arci merges its hooks with any existing hooks rather than replacing them. If hooks for the same event type already exist, arci's hooks are added alongside them.

### Claude Code hook approval

Claude Code includes a security feature where direct edits to hook configuration require review before taking effect. After `arci install` modifies settings files, users need to approve the changes in Claude Code using the `/hooks` command.

The install command explains this clearly in its output:

```
✓ Claude Code hooks configured in ~/.claude/settings.json

Next step: Open Claude Code and run /hooks to review and approve the new hooks.
Claude Code requires manual approval of hook configuration changes for security.
```

This manual approval step is a feature, not a bug—it prevents malicious software from silently installing hooks that could exfiltrate data or execute arbitrary commands.

## Directory structure

Installation creates the arci configuration directory structure if it doesn't exist.

### Global configuration

After global installation:

```text
<user-config-dir>/arci/
├── config.yaml           # Universal rules (created if --scaffold)
└── config.claude.yaml    # Claude-specific rules (created if --scaffold)
```

The `<user-config-dir>` follows platform conventions: `~/.config/` on Linux, `~/Library/Application Support/` on macOS, and `%APPDATA%` on Windows.

### Project configuration

After project installation:

```text
.arci/
├── arci.yaml                # Project rules (created if --scaffold)
├── arci.claude.yaml         # Claude-specific project rules
├── arci.local.yaml          # Personal overrides (gitignored)
└── arci.local.claude.yaml   # Personal Claude overrides
```

The install command suggests adding `.arci/arci.local.*` to `.gitignore` if the file doesn't already ignore it.

> **Planned feature**: Drop-in directories (`hooks.d/`) for modular rule organization are planned but not yet implemented. See the [configuration documentation](configuration.md) for details.

## Uninstallation

The `arci uninstall` command reverses the installation process:

```bash
arci uninstall [--scope <scope>]
```

Uninstallation removes the arci hook entries from Claude Code configuration but leaves arci's own configuration directories intact. Users who want to completely remove arci configuration can delete those directories manually.

The `--purge` flag removes both hook configuration and arci's configuration directories:

```bash
arci uninstall --purge
```

To remove the arci binary itself:

```bash
# If installed via go install
rm $(go env GOPATH)/bin/arci

# If installed via Homebrew
brew uninstall arci

# If installed from pre-built binary
rm ~/.local/bin/arci  # or wherever you placed it
```

## Daemon considerations

The `arci install` command configures hooks but does not start the daemon. The daemon must be running for hook evaluation to work. There are several approaches to daemon lifecycle management.

### Manual daemon management

Users can start the daemon manually in a terminal:

```bash
arci daemon start
```

This is the recommended approach during development or initial setup, as it provides visibility into daemon logs and makes it easy to restart after configuration changes.

### Automatic daemon startup

When `arci hook apply` runs and cannot connect to the daemon, it fails open (allowing the operation to proceed) as described in the fail-open semantics. However, having the daemon not running means no rules are evaluated.

A future enhancement could have `arci hook apply` automatically start the daemon if it's not running. This would add latency to the first hook invocation but ensure rules are always evaluated. The tradeoff is complexity and potential issues with multiple processes trying to start the daemon simultaneously.

### System service installation

For users who want the daemon to start automatically at login or boot, arci could provide a command to install a system service:

```bash
arci daemon install-service
```

This would create a launchd plist on macOS, a systemd unit on Linux, or equivalent on other platforms. The service would start the daemon automatically and restart it if it crashes.

This is planned for a future release but not part of the initial implementation.

## Upgrade workflow

When a new version of arci is released, upgrade using the same method you used to install:

```bash
# If installed via go install
go install github.com/tbhb/arci@latest

# If installed via Homebrew
brew upgrade arci
```

For pre-built binaries, download the new version and replace the existing binary.

Hook configuration typically doesn't need to change between versions. However, major version upgrades might require re-running `arci install` if the hook invocation format changes.

The CLI will warn if the installed hook configuration appears incompatible with the current arci version.

## Troubleshooting

Common installation issues and their solutions:

If `arci` is not found after installation, ensure `$GOPATH/bin/` (default `~/go/bin/`) is on your PATH. The Go installer typically adds this, but you may need to restart your shell or source your profile.

If hooks aren't being evaluated, verify the daemon is running with `arci daemon status`. If it's not running, start it with `arci daemon start`.

If Claude Code shows hooks as pending review, use `/hooks` in Claude Code to approve the configuration. This is Claude Code's security feature, not an arci bug.

If detection fails to find Claude Code, check that the `claude` binary is on your PATH and the `~/.claude/` directory exists.
