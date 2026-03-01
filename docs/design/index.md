# arci Design Documentation

arci — **Agentic Requirements Composition & Integration** — is a framework for Claude Code that combines policy-driven hooks with spec-driven development. It provides declarative rules to validate, transform, and control tool execution, alongside an INCOSE-inspired knowledge graph that structures requirements, verifications, and traceability.

## Two pillars

### Hooks

The hook system intercepts Claude Code tool execution through declarative policies. Policies match events, evaluate CEL conditions, and produce admission decisions (allow, deny, warn) or mutations. A daemon-based architecture keeps evaluation fast; fail-open semantics ensure hooks never block the assistant unexpectedly.

See [Hooks overview](hooks/index.md) for the full hook documentation.

### Specs

The spec system structures software development around INCOSE-inspired systems engineering practices. A knowledge graph of interconnected modules captures concepts, needs, requirements, verifications, tasks, and defects — providing full traceability from stakeholder expectations to verified implementations.

**Knowledge graph** — schema, relationships, constraints, and query patterns:
- [Graph overview](graph/index.md) — entry point for graph design documentation

**Intent** — capturing stakeholder expectations:
- [Concepts](intent/concepts.md) — exploration, design decisions, crystallized thinking
- [Needs](intent/needs.md) — stakeholder expectations, validated

**Requirements** — formalizing into verifiable obligations:
- [Modules](requirements/modules.md) — architectural containers with hierarchy and phase tracking
- [Requirements](requirements/requirements.md) — design obligations, verified
- [Baselines](requirements/baselines.md) — named snapshots of graph state at decision points

**Execution** — implementing the work:
- [Tasks](execution/tasks.md) — atomic work units in a DAG
- [Templating](execution/templating.md) — reusable task and decomposition templates

**Verification** — checking the work:
- [Verifications](verification/verifications.md) — evidence that requirements are satisfied
- [Defects](verification/defects.md) — identified problems requiring action

## Infrastructure

Shared components that support both pillars:

### Configuration

- [Configuration](configuration.md) — layered configuration system with precedence rules and hot reloading
- [Config cascade](config-cascade.md) — full precedence chain from built-in defaults to CLI flags
- [Managed config](config-managed.md) — managed/recommended and managed/required configuration

### Daemon

- [Daemon](daemon.md) — long-running process for fast evaluation, API endpoints, and lifecycle management
- [Daemon auto-spawn](daemon-auto-spawn.md) — automatic daemon startup and management

### CLI

- [CLI](cli/index.md) — command-line interface for hook evaluation and spec management

### State & storage

- [State store](state-store.md) — persistent key-value store for tracking data across hook invocations

### Security

- [Security model](security.md) — trust model, threat scenarios, and security controls
- [Sandboxing](sandboxing.md) — platform-native shell action isolation

### Observability

- [Logging](logging.md) — structured logging architecture
- [Error handling](errors.md) — error handling and diagnostics
- [Dashboard](dashboard/index.md) — web-based diagnostics dashboard

### Extensions

- [Extensions](extensions.md) — unified extension system for distributing policies and custom functions

### Testing & quality

- [Testing strategy](testing.md) — testing approaches for the engine, shells, and policies
- [Performance](performance.md) — latency expectations and optimization strategies

### Architecture & operations

- [Architecture](architecture.md) — high-level system architecture
- [Installation](installation.md) — installation and upgrade workflows
- [Versioning](versioning.md) — schema evolution and backward compatibility

### Reference

- [Glossary](glossary.md) — key terms used throughout arci documentation
