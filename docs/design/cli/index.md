# CLI

The arci CLI provides both the fast evaluation path called by Claude Code and management commands for configuration, diagnostics, and daemon control. This document describes the command structure and behavior.

## Overview

The CLI has two distinct usage patterns. The `arci hook apply` command is called by Claude Code on the hot path and must be as fast as possible. All other commands are for human use and can take normal startup time.

```bash
arci <command> [options]
```

The CLI is built with [Cobra](https://github.com/spf13/cobra) for command structure and [pflag](https://github.com/spf13/pflag) for flag parsing. Global options are defined once and inherited by all subcommands.

### Terminology

This document uses "policy" and "rule" with specific meanings that match the [policy model](../hooks/policy-model.md):

A **policy** is a self-contained YAML document that declares matching criteria, parameters, variables, and rules. Policies have names like `security-baseline` or `coding-standards-injection`. Most CLI commands operate at the policy level.

A **rule** is a component within a policy that defines a specific validation, mutation, or side effect. Rules have names like `block-rm-rf` or `inject-coding-standards`. Rules are always scoped to their containing policy.

## Installation

Install arci using go install:

```bash
go install github.com/tbhb/arci@latest
```

Or build from source:

```bash
go build -o arci ./cmd/arci
```

Pre-built binaries are available from the releases page for major platforms (Linux, macOS, Windows) and architectures (x86_64, aarch64).

## Contents

### Commands

- [hook](commands/hook.md) — Hook commands (`arci hook apply`, `arci hook policy`, `arci hook logs`, `arci hook stats`)
- [doctor](doctor.md) — Health checks for installation, configuration, and integrations
- [install](install.md) — Manage Claude Code hook integration
- [daemon](daemon.md) — Control the optional background daemon process
- [dashboard](dashboard.md) — Start the diagnostics web interface
- [config](config.md) — Inspect, validate, and modify configuration
- [state](state.md) — Access the state store
- [extension](extension.md) — Manage arci extensions

### Reference

- [selectors](selectors.md) — Selector syntax for matching policies and rules
- [global-options](global-options.md) — Flags inherited by all subcommands
- [environment-variables](environment-variables.md) — `ARCI_` environment variable defaults
- [exit-codes](exit-codes.md) — Exit code conventions

## Design principles

- **Fast apply path.** The `arci hook apply` command is on the Claude Code hot path and must minimize latency. Direct mode targets 50-200ms; daemon mode provides sub-millisecond cached evaluation.
- **Fail-open semantics.** Configuration errors and policy evaluation failures never block Claude Code. Only explicit deny decisions from successfully-evaluated validation rules block operations.
- **Cobra/pflag conventions.** The CLI follows standard Cobra patterns: persistent flags for global options, subcommand groups for related functionality, and consistent help/completion output.

## See also

- [Policy model](../hooks/policy-model.md) — definitions of policy and rule concepts
- [Hook schema](../hooks/hook-schema.md) — event types consumed by `arci hook apply`
- [Execution model](../hooks/execution-model.md) — the six-stage policy evaluation pipeline
