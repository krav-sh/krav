# Command-line interface

The ARCI command-line tool provides both the fast evaluation path that Claude Code invokes and management commands for configuration, diagnostics, and daemon control. This document describes the command structure and behavior.

## Overview

The command-line tool has two distinct usage patterns. Claude Code invokes the `arci hook apply` command on the hot path, so it must be as fast as possible. All other commands target human use and can take normal startup time.

```bash
arci <command> [options]
```

The command-line tool uses [Cobra](https://github.com/spf13/cobra) for command structure and [pflag](https://github.com/spf13/pflag) for flag parsing. Global options apply once and all subcommands inherit them.

### Terminology

This document uses "policy" and "rule" with specific meanings that match the [policy model](../hooks/policy-model.md):

A **policy** is a self-contained YAML document that declares matching criteria, parameters, variables, and rules. Policies have names like `security-baseline` or `coding-standards-injection`. Most commands operate at the policy level.

A **rule** is a component within a policy that defines a specific validation, mutation, or side effect. Rules have names like `block-rm-rf` or `inject-coding-standards`. Rules are always scoped to their containing policy.

## Installation

Install ARCI using go install:

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

- [hook](commands/hook.md): Hook commands (`arci hook apply`, `arci hook policy`, `arci hook logs`, `arci hook stats`)
- [doctor](commands/doctor.md): Health checks for installation, configuration, and integrations
- [install](commands/install.md): Manage Claude Code hook integration
- [server](commands/server.md): Start the ARCI server for the current project
- [config](commands/config.md): Inspect, check, and change configuration
- [state](commands/state.md): Access the state store
- [extension](commands/extension.md): Manage ARCI extensions
- [baseline](commands/baseline.md): Manage knowledge graph baselines
- [concept](commands/concept.md): Manage concept nodes
- [defect](commands/defect.md): Track and manage defects
- [module](commands/module.md): Manage architectural modules
- [need](commands/need.md): Manage stakeholder needs
- [`req`](commands/req.md): Manage requirements and traceability
- [stakeholder](commands/stakeholder.md): Manage project stakeholders
- [task](commands/task.md): Manage tasks and the work DAG
- [`tc`](commands/tc.md): Manage test cases and verification coverage

### Reference

- [global-options](global-options.md): Flags inherited by all subcommands
- [environment-variables](environment-variables.md): `ARCI_` environment variable defaults
- [exit-codes](exit-codes.md): Exit code conventions
- [errors](errors.md): Error presentation, validation output, health checks
- [logging](logging.md): Output verbosity flags, diagnostic tracing

## Design principles

- **Fast apply path.** The `arci hook apply` command is on the Claude Code hot path and must limit latency. Direct mode targets 50 to 200 ms; daemon mode provides sub-millisecond cached evaluation.
- **Fail-open semantics.** Configuration errors and policy evaluation failures never block Claude Code. Only explicit deny decisions from successfully evaluated validation rules block operations.
- **Cobra/pflag conventions.** The command-line tool follows standard Cobra patterns: persistent flags for global options, subcommand groups for related capability, and consistent help/completion output.

## See also

- [Policy model](../hooks/policy-model.md): definitions of policy and rule concepts
- [Hook schema](../hooks/hook-schema.md): event types consumed by `arci hook apply`
- [Execution model](../hooks/execution-model.md): the six-stage policy evaluation pipeline
